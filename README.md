<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [compliance-audit-router](#compliance-audit-router)
  - [Configuration](#configuration)
    - [Configuration Values](#configuration-values)
      - [General Configuration](#general-configuration)
      - [LDAP Configuration](#ldap-configuration)
      - [Splunk Configuration](#splunk-configuration)
      - [Jira Configuration](#jira-configuration)
    - [Example compliance-audit-router.yaml file](#example-compliance-audit-routeryaml-file)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# compliance-audit-router

A tool to receive compliance alert webhooks from an external source (eg. Splunk), look up the responsible engineer's information (eg. from LDAP), and create a compliance report ticket (eg. Jira) assigned to the engineer for follow-up.

## Configuration

Configuration is managed in the `~/.config/compliance-audit-router/compliance-audit-router.yaml` file.

Alternatively, configuration options may be set using environment variables according to the [Viper environmental variable setup](https://github.com/spf13/viper#working-with-environment-variables), with the prefix `CAR_` (eg. `CAR_LISTENPORT=8080`).

### Configuration Values

#### General Configuration

verbose
: Turns on more verbose logging output. Default: false

listenport
: The port on which Compliance Audit Router will listen for SIEM (ie. Splunk) alert webhooks. Default: 8080

#### LDAP Configuration

ldapconfig.host
: The LDAP server to query for user information. May or may not include `ldap://` or `ldaps://` schema, as appropriate. (eg: `ldaps://ldap.example.org`)

ldapconfig.username
: The username with which to authenticate to the LDAP server. Requires `ldapconfig.password`. If no username is provided, Compliance Audit Router will attempt an unauthenticated bind.

ldapconfig.password
: The password with which to authenticate to the LDAP server. Requires `ldapconfig.username`.

ldapconfig.searchbase
: The LDAP Search Base directory from which to begin object searches.

ldapconfig.scope
: The LDAP scope depth for queries.

ldapconfig.attributes
: The LDAP attributes to look up for the provided query.

#### Splunk Configuration

splunkconfig.host
: The Splunk server to query for alert search results. Must include the scheme and port. (eg: `https://splunk.example.org:8089`)

splunkconfig.token
: An API token to authenticate to the Splunk API.

splunkconfig.allowinsecure
: Boolean. When `true`, allows insecure TLS connections. Don't do this.

#### Jira Configuration

jiraconfig.host
: The Jira instance in which to create and manage compliance alert issues. Must include the scheme. May optionally include the port. (eg: `https://jira.example.org:8443`)

jiraconfig.username
: The (optional) username with which to authenticate to the Jira API. Requires `jiraconfig.token`. Setting this causes Compliance Audit Router to use Jira's Basic Authentication method. This should only be done for development. (eg: `jira-user@example.org`)

jiraconfig.token
: The API token to authenticate to the Jira API. Setting this without setting `jiraconfig.username` causes Compliance Audit Router to use Jira's Personal Access Token (PAT) authentication method.

jiraconfig.allowinsecure
: Boolean. When `true`, allows insecure TLS connections. Don't do this.

jiraconfig.key
: The Jira Project key of the project in which Compliance Audit Router will create and manage compliance alert issues.

jiraconfig.issuetype
: The Jira Issue type that new compliance alerts will be created as. (eg. "Task")

jiraconfig.transitions
: TODO - document the transitions


### Example compliance-audit-router.yaml file

```yaml
---
verbose: false
listenport: 8080

ldapconfig:
  host: ldaps://ldap.example.org
  username: <username>
  password: <password>
  searchbase: dc=example,dc=org
  scope: sub
  attributes:
    - manager
    - alternateID

splunkconfig:
  host: https://splunk.example.org:8089
  token: <token>
  allowinsecure: false

jiraconfig:
  host: https://jira.example.org:443
  allowinsecure: false
  username: <username>
  token: <token>
  key: <Jira project key>
  issuetype: <type of issue to create>
  dev: false
  transitions:

messagetemplate: |
  {{.Username}},

  This action required business justification from the engineer who used this access, and management approval.

  If this action is unexpected or unexplained, please contact the Security team immediately for further investigation.
```
