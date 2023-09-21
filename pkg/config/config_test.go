package config

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestFieldsAreNotNil(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   []error
	}{
		{
			"Empty configs should fail",
			&Config{},
			[]error{
				configError{Err: "missing required configuration value: ldapconfig.host"},
				configError{Err: "missing required configuration value: splunkconfig.host"},
				configError{Err: "missing required configuration value: jiraconfig.host"},
				configError{Err: "missing required configuration value: jiraconfig.token"},
				configError{Err: "missing required configuration value: jiraconfig.key"},
				configError{Err: "missing required configuration value: jiraconfig.issuetype"},
				configError{Err: "missing required configuration value: ldapconfig.searchbase"},
				configError{Err: "missing required configuration value: ldapconfig.scope"},
			},
		},
		{
			"Provided configs should not fail",
			&Config{
				LDAPConfig: LDAPConfig{
					Host: "ldaps://ldap.example.org:636",
				},
				SplunkConfig: SplunkConfig{
					Host: "https://splunk.example.org:8089",
				},
				JiraConfig: JiraConfig{
					Host: "https://jira.example.org",
				},
			},
			[]error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldsAreNotNil(tt.config)
			var failed bool = false
			for _, err := range tt.want {
				if !slices.Contains(got, err) {
					t.Errorf("fieldsAreNotNil() missing expected error: %+v", err)
					failed = true
				}
			}
			// Placing this outside the loop so we don't print the whole list for each individual failure
			if failed {
				t.Errorf("fieldsAreNotNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostFieldsAreParsable(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   []error
	}{
		{
			// Valid host names should pass, including those with ports
			"Valid hosts should not fail",
			&Config{
				LDAPConfig: LDAPConfig{
					Host: "ldaps://ldap.example.org",
				},
				SplunkConfig: SplunkConfig{
					Host: "https://splunk.example.org:8089",
				},
				JiraConfig: JiraConfig{
					Host: "https://jira.example.org",
				},
			},
			[]error{},
		},
		{
			"URLs missing Scheme should fail",
			&Config{
				LDAPConfig: LDAPConfig{
					Host: "ldap.example.org",
				},
				SplunkConfig: SplunkConfig{
					Host: "splunk.example.org:8089",
				},
				JiraConfig: JiraConfig{
					Host: "jira.example.org",
				},
			},
			[]error{
				configError{Err: "ldapconfig.host missing scheme: ldap.example.org"},
				configError{Err: "splunkconfig.host missing scheme: splunk.example.org:8089"},
				configError{Err: "jiraconfig.host missing scheme: jira.example.org"},
			},
		},
		{
			"Unparsable URLs should fail",
			&Config{
				LDAPConfig: LDAPConfig{
					Host: "www.url",
				},
				SplunkConfig: SplunkConfig{
					Host: "not_URL",
				},
				JiraConfig: JiraConfig{
					Host: "example.org:abc",
				},
			},
			[]error{
				configError{Err: "ldapconfig.host invalid URL: www.url"},
				configError{Err: "splunkconfig.host invalid URL: not_URL"},
				configError{Err: "jiraconfig.host invalid URL: example.org:abc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostFieldsAreParsable(tt.config)
			var failed bool = false
			for _, err := range tt.want {
				if !slices.Contains(got, err) {
					t.Errorf("hostFieldsAreParsable() missing expected error: %+v", err)
					failed = true
				}
			}
			// Placing this outside the loop so we don't print the whole list for each individual failure
			if failed {
				t.Errorf("hostFieldsAreParsable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsurePasswordIfUsernameProvided(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   []error
	}{
		{
			"Username with password/token should pass",
			&Config{
				LDAPConfig: LDAPConfig{
					Username: "testUsername",
					Password: "testPassword",
				},
				JiraConfig: JiraConfig{
					Username: "testUsername",
					Token:    "testToken",
				},
			},
			[]error{},
		},
		{
			"Username without password/token should fail",
			&Config{
				LDAPConfig: LDAPConfig{
					Username: "testUsername",
				},
				JiraConfig: JiraConfig{
					Username: "testUsername",
				},
			},
			[]error{
				configError{Err: "ldapconfig.username provided without ldapconfig.password"},
				configError{Err: "jiraconfig.username provided without jiraconfig.token"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passwordOrTokenExistIfUsernameProvided(tt.config)
			var failed bool = false
			for _, err := range tt.want {
				if !slices.Contains(got, err) {
					t.Errorf("ensurePasswordOrTokenIfUsernameProvided() missing expected error: %+v", err)
					failed = true
				}
			}
			// Placing this outside the loop so we don't print the whole list for each individual failure
			if failed {
				t.Errorf("ensurePasswordOrTokenIfUsernameProvided() = %v, want %v", got, tt.want)
			}
		})
	}
}
