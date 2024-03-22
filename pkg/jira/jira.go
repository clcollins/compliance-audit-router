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

package jira

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/openshift/compliance-audit-router/pkg/config"
)

const (
	managedLabel    = "compliance-audit-router/managed"
	sreLabelKey     = "compliance-audit-router/sre"
	managerLabelKey = "compliance-audit-router/manager"
	sreLabel        = sreLabelKey + ":%v"
	managerLabel    = managerLabelKey + ":%v"
	unknownUser     = "unknown"

	initialTransitionKey = "initial"
	sreTransitionKey     = "sre"
	managerTransitionKey = "manager"

	ticketSummary = "Compliance Alert: SRE Cluster Admin Elevation"
)

type Webhook struct {
	Issue   jira.Issue
	Comment jira.Comment
}

func DefaultClient() (*jira.Client, error) {
	var transportClient *http.Client
	if config.AppConfig.JiraConfig.Username != "" {
		log.Printf("jira.DefaultClient(): WARNING: Using basic auth for Jira client development\n")
		transportClient = basicAuthClient(config.AppConfig.JiraConfig.Username, config.AppConfig.JiraConfig.Token)
	} else {
		transportClient = patAuthClient(config.AppConfig.JiraConfig.Token)
	}

	return jira.NewClient(transportClient, config.AppConfig.JiraConfig.Host)
}

func CreateTicket(userService *jira.UserService, issueService *jira.IssueService, user string, manager string, description string) error {

	if config.AppConfig.DryRun {
		log.Printf("jira.CreateTicket(): dry-run mode: would have created Jira ticket with user, manager, description: %+v, %+v, %+v", user, manager, description)
		if config.AppConfig.Verbose {
			log.Printf("jira.CreateTicket(): dry-run mode: *jira.UserService: %+v", userService)
			log.Printf("jira.CreateTicket(): dry-run mode: *jira.issueService: %+v", issueService)
		}
	}

	reporterUser, _, err := userService.GetSelf()
	if err != nil {
		return fmt.Errorf("failed to get Jira user for reporter: %w", err)
	}

	sreUser, err := getUserByName(userService, user)
	if err != nil {
		log.Printf("jira.CreateTicket(): failed to fetch SRE's Jira account. The ticket will be created with no assignee and need to be managed manually: %v\n", err)
		sreUser = &jira.User{AccountID: unknownUser}
	}

	managerUser, err := getUserByName(userService, manager)
	if err != nil {
		log.Printf("jira.CreateTicket(): failed to fetch manager's Jira account: %v\n", err)
		managerUser = &jira.User{AccountID: unknownUser}
	}

	jiraIssue := &jira.Issue{
		Fields: &jira.IssueFields{
			Reporter:    reporterUser,
			Description: description,
			Type:        jira.IssueType{Name: config.AppConfig.JiraConfig.IssueType},
			Project:     jira.Project{Key: config.AppConfig.JiraConfig.Key},
			Summary:     ticketSummary,
		},
	}

	if sreUser.AccountID != unknownUser {
		jiraIssue.Fields.Assignee = sreUser
		jiraIssue.Fields.Labels = []string{managedLabel, fmt.Sprintf(sreLabel, sreUser.AccountID), fmt.Sprintf(managerLabel, managerUser.AccountID)}
	}

	var createdIssue *jira.Issue
	if config.AppConfig.DryRun {
		log.Printf("jira.CreateTicket(): dry-run mode: would have created Jira ticket with the following fields: %+v", jiraIssue)
		createdIssue = &jira.Issue{}
		createdIssue.Key = "DRY-RUN-0000"
		err = nil
	} else {
		createdIssue, _, err = issueService.Create(jiraIssue)
	}

	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	log.Printf("jira.CreateTicket(): created new issue with key %v", createdIssue.Key)

	messageTemplate, err := template.New("messageTemplate").Parse(config.AppConfig.MessageTemplate)
	if err != nil {
		if config.AppConfig.Verbose {
			log.Printf("jira.CreateTicket(): failed to parse message template from AppConfig; template: %v\n", config.AppConfig.MessageTemplate)
		}
		return fmt.Errorf("failed to parse message template from AppConfig: %w", err)
	}

	var message bytes.Buffer
	err = messageTemplate.Execute(&message, map[string]string{"Username": fmt.Sprintf("[~accountid:%v]", sreUser.AccountID)})
	if err != nil {
		return fmt.Errorf("failed to apply parsed template to the specified data object: %w", err)
	}

	comment := &jira.Comment{Body: message.String()}

	if config.AppConfig.DryRun {
		log.Printf("jira.CreateTicket(): dry-run mode: would have added comment to Jira ticket with the following body: %+v", comment)
		err = nil
	} else {
		_, _, err = issueService.AddComment(createdIssue.ID, comment)
	}

	if err != nil {
		return fmt.Errorf("issue %v was successfully created but failed to apply initial comment: %w", createdIssue.Key, err)
	}

	log.Printf("jira.CreateTicket(): initial comment successfully left on issue %v\n", createdIssue.Key)

	initialStatusName := config.AppConfig.JiraConfig.Transitions[initialTransitionKey]

	initialStatusId, err := getTransitionId(issueService, createdIssue.ID, initialStatusName)
	if err != nil {
		return fmt.Errorf("failed to fetch ID for status %v: %w", initialStatusName, err)
	}

	if config.AppConfig.DryRun {
		log.Printf("jira.CreateTicket(): dry-run mode: would have transitioned Jira ticket to status %v", initialStatusName)
	} else {
		_, err = issueService.DoTransition(createdIssue.ID, initialStatusId)
		if err != nil {
			return fmt.Errorf("failed to transition issue %v to status %v: %w", createdIssue.Key, initialStatusName, err)
		}
	}

	log.Printf("jira.CreateTicket(): issue %v has been transitioned to state %v", createdIssue.Key, initialStatusName)

	return nil
}

func HandleUpdate(issueService *jira.IssueService, webhook Webhook) error {
	if config.AppConfig.DryRun {
		log.Printf("jira.HandleUpdate(): dry-run mode: would have handled Jira webhook with issue, comment: %+v, %+v", webhook.Issue, webhook.Comment)
		if config.AppConfig.Verbose {
			log.Printf("jiraHandleUpdate(): dry-run mode: *jira.issueService: %+v", issueService)
		}

		return nil
	}

	webhookIssue, _, err := issueService.Get(webhook.Issue.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get issue %v from jira webhook: %w", webhook.Issue.Key, err)
	}

	var sreId string
	var managerId string
	for _, label := range webhookIssue.Fields.Labels {
		if strings.Contains(label, sreLabelKey) {
			_, sreId, _ = strings.Cut(label, ":")
		}
		if strings.Contains(label, managerLabelKey) {
			_, managerId, _ = strings.Cut(label, ":")
		}
	}

	// If the comment isn't from the current assignee then we don't need to do anything.
	if webhook.Comment.Author.AccountID != webhookIssue.Fields.Assignee.AccountID {
		return nil
	}

	var transitionName string
	if sreId == webhook.Comment.Author.AccountID {
		transitionName = config.AppConfig.JiraConfig.Transitions[sreTransitionKey]
	} else if managerId == webhook.Comment.Author.AccountID {
		transitionName = config.AppConfig.JiraConfig.Transitions[managerTransitionKey]
	}

	transitionId, err := getTransitionId(issueService, webhookIssue.ID, transitionName)
	if err != nil {
		return fmt.Errorf("failed to get transition ID for status %v on issue %v: %w", transitionName, webhookIssue.Key, err)
	}

	_, err = issueService.DoTransition(webhookIssue.ID, transitionId)
	if err != nil {
		return fmt.Errorf("failed to transition issue %v to status %v: %w", webhookIssue.Key, transitionName, err)
	}
	log.Printf("jira.HandleUpdate(): successfully updated ticket %v to status %v after comment from %v", webhookIssue.Key, transitionName, webhook.Comment.Author.Name)

	return nil
}

func basicAuthClient(user, token string) *http.Client {
	transport := jira.BasicAuthTransport{
		Username: user,
		Password: token,
	}
	return transport.Client()
}

func patAuthClient(token string) *http.Client {
	transport := jira.PATAuthTransport{
		Token: token,
	}
	return transport.Client()
}

func getTransitionId(issueService *jira.IssueService, issueId string, status string) (string, error) {
	if config.AppConfig.DryRun {
		log.Printf("jira.GetTransitionId(): dry-run mode: would have fetched transitions for Jira issue %v", issueId)
		return "dry-run-transition-id", nil
	}

	transitions, _, err := issueService.GetTransitions(issueId)
	if err != nil {
		return "", err
	}

	for _, t := range transitions {
		if t.Name == status {
			return t.ID, nil
		}
	}
	return "", fmt.Errorf("did not find status %v", status)
}

func getUserByName(userService *jira.UserService, username string) (*jira.User, error) {
	if (username == "") && config.AppConfig.Verbose {
		log.Printf("jira.getUserByName() called with empty username")
	}
	users, _, err := userService.Find(username)
	if err != nil {
		return nil, err
	}

	if jiraUserLen := len(users); jiraUserLen != 1 {
		return nil, fmt.Errorf("error finding user '%v': expected 1 user but found %v", username, jiraUserLen)
	}
	return &users[0], nil
}

type UserService interface {
	Find(property string, tweaks ...func([]userSearchParam) []userSearchParam) ([]jira.User, *jira.Response, error)
}

type userSearchParam struct {
	name  string
	value string
}
