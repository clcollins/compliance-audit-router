/*
Copyright © 2021 Red Hat, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package listeners

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/openshift/compliance-audit-router/pkg/helpers"
	"github.com/openshift/compliance-audit-router/pkg/jira"
	"github.com/openshift/compliance-audit-router/pkg/ldap"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

type Listener struct {
	Path        string
	Methods     []string
	HandlerFunc http.HandlerFunc
}

var Listeners = []Listener{
	{
		Path:        "/readyz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: RespondOKHandler,
	},
	{
		Path:        "/healthz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: RespondOKHandler,
	},
	{
		Path:        "/api/v1/alert",
		Methods:     []string{http.MethodPost},
		HandlerFunc: ProcessAlertHandler,
	},
	{
		Path:        "/api/v1/jira_webhook",
		Methods:     []string{http.MethodPost},
		HandlerFunc: ProcessJiraWebhook,
	},
}

// InitRoutes initializes routes from the defined Listeners
func InitRoutes(router *chi.Mux) {
	for _, listener := range Listeners {
		for _, method := range listener.Methods {
			router.Method(method, listener.Path, listener.HandlerFunc)
		}
	}
}

// RespondOKHandler replies with a 200 OK and "OK" text to any request, for health checks
func RespondOKHandler(w http.ResponseWriter, _ *http.Request) {
	setResponse(w, http.StatusOK, map[string]string{"Content-Type": "text/plain"}, "OK")
}

// ProcessAlertHandler is the main logic processing alerts received from Splunk
func ProcessAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the alert search results
	var alert splunk.Webhook

	client, err := jira.DefaultClient()
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, "failed to create Jira client")
		// Should this panic?  Return?
		return
	}

	err = helpers.DecodeJSONRequestBody(w, r, &alert)
	if err != nil {
		var mr *helpers.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	searchResults, err := splunk.RetrieveSearchFromAlert(alert.Sid)
	if err != nil {
		log.Printf("error retrieving search results from Splunk: %v", err)

		alertJson, jsonErr := json.MarshalIndent(alert, "", "    ")
		if jsonErr != nil {
			log.Printf("error marshalling webhook data to JSON : %v", jsonErr)
		}

		ticketDetails := fmt.Sprintf(
			"A Compliance Alert was received from Splunk, but the alert details could not be retrieved. "+
				"Please review:\n"+
				"Splunk Webhook Data: %s\n"+
				"\nError: %s\n", string(alertJson), err,
		)

		// Add a note to the ticket details if the webhook data might be incomplete
		if jsonErr != nil {
			ticketDetails += fmt.Sprintf("\n\nIn addition, an error occurred attempting to unmarshal the incoming webhook details: %v", err)
		}

		createErr := jira.CreateTicket(client.User, client.Issue, "", "", ticketDetails)

		if createErr != nil {
			log.Println(createErr)
			setResponse(
				w, http.StatusInternalServerError,
				map[string]string{"Content-Type": "text/plain"},
				fmt.Sprintf("failed Splunk alert lookup and unable to create a follow up Jira ticket - Splunk Error: %v, Jira Error: %v", err, createErr),
			)
			return
		}

		setResponse(
			w, http.StatusInternalServerError,
			map[string]string{"Content-Type": "text/plain"},
			fmt.Sprintf("failed Splunk alert lookup: %v", err),
		)

		return
	}

	fmt.Printf("%+v\n", searchResults)

	os.Exit(1)

	//user, manager, err := ldap.LookupUser(searchResults.UserName)
	user, manager, err := ldap.LookupUser("TODO USERNAME GOES HERE")
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, "failed ldap lookup")
		return
	}

	err = jira.CreateTicket(client.User, client.Issue, user, manager, "test description")
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, "failed ticket creation")
		return
	}

	setResponse(w, http.StatusOK, map[string]string{"Content-Type": "text/plain"}, "ok")
	return
}

func ProcessJiraWebhook(w http.ResponseWriter, r *http.Request) {
	webhook := jira.Webhook{}
	err := helpers.DecodeJSONRequestBody(w, r, &webhook)
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusBadRequest, map[string]string{"Content-Type": "text/plain"}, "failed to parse webhook")
		return
	}

	client, err := jira.DefaultClient()
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, "failed to create Jira client")
	}

	err = jira.HandleUpdate(client.Issue, webhook)
	if err != nil {
		log.Println(err)
		setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, "failed to update JIRA issue")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setResponse(w http.ResponseWriter, statusCode int, headers map[string]string, body string) {
	w.WriteHeader(statusCode)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	_, _ = w.Write([]byte(body))
}
