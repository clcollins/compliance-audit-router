package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CARPrometheusLabels = map[string]string{"name": "compliance-audit-router"}
)

var (
	// SPLUNK WEBHOOK AND ALERT PROCESSING

	// MetricSplunkWebhookReceived is the number of Splunk alert webhooks received
	MetricSplunkWebhookReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_splunk_webhook_received",
		Help:        "Number of Splunk alert webhooks received",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricSplunkWebhookProcessFailures tracks the number of Splunk alert webhooks that failed to be processed
	MetricSplunkWebhookProcessFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_splunk_webhook_process_failures",
		Help:        "Number of Splunk alert webhooks that failed to be processed",
		ConstLabels: CARPrometheusLabels},
		[]string{"error_type", "uuid", "process"},
	)
	// MetricSplunkAlertSIDReceived is the number of Splunk alert SIDs received
	MetricSplunkAlertSIDReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_splunk_alert_sid_received",
		Help:        "Number of Splunk alert SIDs received",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricSplunkSearchResultQueryFailures is the number of Splunk search result queries that failed
	MetricSplunkSearchResultQueryFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_splunk_search_result_query_failures",
		Help:        "Number of Splunk search result queries that failed",
		ConstLabels: CARPrometheusLabels},
		[]string{"error_type", "uuid", "process"},
	)

	// COMPLIANCE EVENT PROCESSING

	// MetricComplianceEventsFound is the number of compliance events found in Splunk
	// There may be more than one compliance event in a given webhook search result.
	MetricComplianceEventsFound = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_compliance_events_found",
		Help:        "Number of compliance events found in Splunk webhook search results",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricComplianceEventsProcessed is the number of compliance events passed on to the next stage of processing
	MetricComplianceEventsProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_compliance_events_processed",
		Help:        "Number of compliance events passed on to the next stage of processing",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)

	// JIRA ISSUE CREATION FOR EVENTS

	// MetricJiraClientCreateFailures is the number of failures to create a Jira client
	MetricJiraClientCreateFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_client_create_failures",
		Help:        "Number of failures to create a Jira client",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricJiraIssuesCreated is the number of Jira issues created
	MetricJiraIssueCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_issues_created",
		Help:        "Number of Jira issues created",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricJiraErrorIssuesCreated is the number of Jira issues created tracking errors in processing
	MetricJiraErrorIssuesCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_error_issues_created",
		Help:        "Number of Jira issues created tracking errors in processing",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricJiraIssueCreateFailures is the number of compliance events that failed to be created as Jira issues
	MetricJiraIssueCreateFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_issue_create_failures",
		Help:        "Number of Jira issues that failed to be created",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)

	// JIRA WEBHOOK PROCESSING

	// MetricJiraWebhookReceived is the number of Jira notification webhooks received
	// Jira notification webhooks are received when a Jira issue is updated
	MetricJiraWebhookReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_webhook_received",
		Help:        "Number of Jira notification webhooks received",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)
	// MetricJiraWebhookProcessFailures is the number of Jira notification webhooks that failed to be processed
	MetricJiraWebhookProcessFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_webhook_process_failures",
		Help:        "Number of Jira notification webhooks notifications that failed to be processed",
		ConstLabels: CARPrometheusLabels},
		[]string{"error_type", "uuid", "process"},
	)
	// MetricJiraIssueUpdateFailures is the number of failures updating issues based on received webhook events
	MetricJiraIssueUpdateFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_jira_issue_update_failures",
		Help:        "Number of Jira issues that failed to be updated based on received webhook",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)

	// LDAP LOOKUP PROCESSING

	// MetricLDAPLookupFailures is the number of LDAP lookups that failed
	// Other LDAP metrics are not currently tracked, as they are not integral to the process
	MetricLDAPLookupFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_ldap_lookup_failures",
		Help:        "Number of LDAP lookups that failed",
		ConstLabels: CARPrometheusLabels},
		[]string{"uuid", "process"},
	)

	// HTTP RESPONSES TO CLIENTS

	// MetricHTTPResponses is the number of HTTP successes returned by the application
	MetricHTTPResponses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "compliance_audit_router_http_successes",
		Help:        "HTTP responses returned by the application with the status code as a label",
		ConstLabels: CARPrometheusLabels},
		[]string{"code", "uuid", "process"},
	)

	MetricsList = []prometheus.Collector{
		MetricSplunkWebhookReceived,
		MetricSplunkWebhookProcessFailures,
		MetricSplunkAlertSIDReceived,
		MetricSplunkSearchResultQueryFailures,
		MetricComplianceEventsFound,
		MetricComplianceEventsProcessed,
		MetricJiraClientCreateFailures,
		MetricJiraIssueCreated,
		MetricJiraErrorIssuesCreated,
		MetricJiraIssueCreateFailures,
		MetricJiraWebhookReceived,
		MetricJiraWebhookProcessFailures,
		MetricJiraIssueUpdateFailures,
		MetricLDAPLookupFailures,
		MetricHTTPResponses,
	}
)

func RegisterMetrics() {
	for _, metric := range MetricsList {
		prometheus.MustRegister(metric)
	}
}
