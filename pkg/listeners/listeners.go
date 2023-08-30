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

	"github.com/go-chi/chi/v5"
	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/helpers"
	"github.com/openshift/compliance-audit-router/pkg/jira"
	"github.com/openshift/compliance-audit-router/pkg/ldap"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

// genericErrorMsg is a text string returned to the client
// when an error occurs that we don't want to accidentally expose
// data from. All error messages should be logged to the application log.
const genericErrorMsg = "The request could not be completed. Please contact the system administrator."

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

	jiraClient, err := jira.DefaultClient()
	if err != nil {
		log.Printf("failed creating Jira client: %s\n", err.Error())
		set500Response(w)
		// Should this panic?  Return?
		// How do we notify ourselves when Jira client failures are occurring?
		// Alert on some metric?
		return
	}

	err = helpers.DecodeJSONRequestBody(w, r, &alert)
	if err != nil {
		var mr *helpers.MalformedRequest
		if errors.As(err, &mr) {
			log.Printf("received malformed request: %s\n", mr.Msg)
			// This is a client error, so we return the status code and message
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Printf("failed decoding JSON request body: %s\n", err.Error())
			set500Response(w)
		}
		return
	}

	log.Println("received alert from Splunk:", alert.Sid)

	searchResults, err := splunk.Server(config.AppConfig.SplunkConfig).RetrieveSearchFromAlert(alert.Sid)

	if err != nil {
		log.Printf("error retrieving search results from Splunk: %s", err.Error())

		alertJson, jsonErr := json.MarshalIndent(alert, "", "  ")
		if jsonErr != nil {
			log.Printf("error marshalling webhook data to JSON: %s", jsonErr.Error())
		}

		ticketDetails := fmt.Sprintf(
			"A Compliance Alert was received from Splunk, but the alert details could not be retrieved. "+
				"Please review:\n"+
				"Splunk Webhook Data: %s\n"+
				"\nError: %s\n", string(alertJson), err.Error(),
		)

		// Add a note to the ticket details if the webhook data might be incomplete
		if jsonErr != nil {
			ticketDetails += fmt.Sprintf("\nNOTE: The Splunk webhook data could not be marshalled to JSON. This may indicate that the webhook data is incomplete. The error was: %s\n", err.Error())
		}

		createErr := jira.CreateTicket(jiraClient.User, jiraClient.Issue, "", "", ticketDetails)
		if createErr != nil {
			log.Printf("failed creating Jira ticket: %s", createErr.Error())
			return
			// Should we panic here? How do we notify ourselves when Jira ticket creation is failing?
		}

		// Return a 500 for any error case
		set500Response(w)

		return
	}

	for _, result := range searchResults.Details() {
		log.Println(result)
		//user, manager, err := ldap.LookupUser(searchResults.UserName)
		user, manager, err := ldap.LookupUser(result.User)
		if err != nil {
			log.Printf("failed ldap lookup: %s\n", err.Error())
			set500Response(w)
			return
		}
		err = jira.CreateTicket(jiraClient.User, jiraClient.Issue, user, manager, result.AlertName)
		if err != nil {
			log.Printf("failed creating Jira ticket: %s", err.Error())
			set500Response(w)
			return
		}

	}

	set200Response(w)
}

func ProcessJiraWebhook(w http.ResponseWriter, r *http.Request) {
	webhook := jira.Webhook{}
	err := helpers.DecodeJSONRequestBody(w, r, &webhook)
	if err != nil {
		var mr *helpers.MalformedRequest
		if errors.As(err, &mr) {
			log.Printf("received malformed request: %s\n", mr.Msg)
			// This is a client error, so we return the status code and message
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Printf("failed decoding JSON request body: %s\n", err.Error())
			set500Response(w)
		}
		return
	}

	client, err := jira.DefaultClient()
	if err != nil {
		log.Println(err)
		set500Response(w)
	}

	err = jira.HandleUpdate(client.Issue, webhook)
	if err != nil {
		log.Println(err)
		set500Response(w)
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

func set500Response(w http.ResponseWriter) {
	setResponse(w, http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain"}, genericErrorMsg)
}
func set200Response(w http.ResponseWriter) {
	setResponse(w, http.StatusOK, map[string]string{"Content-Type": "text/plain"}, "ok")
}
