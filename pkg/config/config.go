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

package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Package config loads configuration details so they can be accessed
// by other packages

var Appname = "compliance-audit-router"
var defaultMessageTemplate = "{{.Username}}\n\n" +
	"This action requires justification." +
	"Please provide the justification in the comments section below."

var AppConfig Config

type Config struct {
	Verbose         bool
	ListenPort      int
	MessageTemplate string

	LDAPConfig   LDAPConfig
	SplunkConfig SplunkConfig
	JiraConfig   JiraConfig
}

type LDAPConfig struct {
	Host               string
	InsecureSkipVerify bool
	Username           string
	Password           string
	SearchBase         string
	Scope              string
	Attributes         []string
}

type SplunkConfig struct {
	Host          string
	Token         string
	AllowInsecure bool
}

type JiraConfig struct {
	Host          string
	AllowInsecure bool
	Token         string
	Username      string
	Key           string
	IssueType     string
	Transitions   map[string]string
}

// configError defines a custom error so we can compare the errors returned
type configError struct {
	Err string
}

func (ce configError) Error() string {
	return ce.Err
}

func LoadConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(".")                          // Look for config in the cwd
	viper.AddConfigPath(home + "/.config/" + Appname) // Look for config in $HOME/.config/compliance-audit-router
	viper.SetConfigType("yaml")
	viper.SetConfigName(Appname)

	viper.SetEnvPrefix("CAR")
	viper.AutomaticEnv() // read in environment variables that match

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(err)
	}

	viper.SetDefault("MessageTemplate", defaultMessageTemplate)
	viper.SetDefault("Verbose", false)
	viper.SetDefault("ListenPort", 8080)

	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		panic(err)
	}

	if !AppConfig.Valid() {
		// If the config is invalid, log the errors and exit right away
		log.Fatal("FATAL: Configuration invalid - exiting")
	}
}

// Valid() wraps the validation functions for the config struct
// and collects all the errors returned so they can be logged
// together rather than erroring out on the first one found.
// This allows the user to fix all the errors at once rather than
// having to fix one, run the program, fix the next one, etc.
func (a *Config) Valid() bool {
	var configErrors []error

	validationFunctions := []func(a *Config) []error{
		fieldsAreNotNil,
		hostFieldsAreParsable,
		passwordOrTokenExistIfUsernameProvided,
	}

	for _, f := range validationFunctions {
		configErrors = append(configErrors, f(a)...)
	}

	if configErrors != nil {
		log.Print(errors.Join(configErrors...))
		return false
	}

	return true
}

// fieldsAreNotNil tests that he required configs values have been set
func fieldsAreNotNil(a *Config) []error {
	var nilFieldErrors []error

	nilStringTests := []struct {
		name  string
		value string
	}{
		{
			name:  "LDAPConfig.Host",
			value: a.LDAPConfig.Host,
		},
		{
			name:  "SplunkConfig.Host",
			value: a.SplunkConfig.Host,
		},
		{
			name:  "JiraConfig.Host",
			value: a.JiraConfig.Host,
		},
		{
			name:  "LDAPConfig.SearchBase",
			value: a.LDAPConfig.SearchBase,
		},
		{
			name:  "LDAPConfig.Scope",
			value: a.LDAPConfig.Scope,
		},
		{
			name:  "SplunkConfig.Token",
			value: a.SplunkConfig.Token,
		},
		{
			name:  "JiraConfig.Token",
			value: a.JiraConfig.Token,
		},
		{
			name:  "JiraConfig.Key",
			value: a.JiraConfig.Key,
		},
		{
			name:  "JiraConfig.IssueType",
			value: a.JiraConfig.IssueType,
		},
	}

	for _, i := range nilStringTests {
		if i.value == "" {
			// Use strings.ToLower() to match the YAML in the config file to avoid confusion
			nilFieldErrors = append(nilFieldErrors, configError{Err: fmt.Sprintf("missing required configuration value: %s", strings.ToLower(i.name))})
		}
	}

	return nilFieldErrors
}

// hostFieldsAreParsable tests that the host fields are valid URLs
func hostFieldsAreParsable(a *Config) []error {
	var hostParseErrors []error

	hostParseTests := []struct {
		name  string
		value string
	}{
		{
			name:  "LDAPConfig.Host",
			value: a.LDAPConfig.Host,
		},
		{
			name:  "SplunkConfig.Host",
			value: a.SplunkConfig.Host,
		},
		{
			name:  "JiraConfig.Host",
			value: a.JiraConfig.Host,
		},
	}
	for _, i := range hostParseTests {
		if i.value != "" {
			u, err := url.Parse(i.value)
			if err != nil {
				// Use strings.ToLower() to match the YAML in the config file to avoid confusion
				hostParseErrors = append(hostParseErrors, configError{Err: fmt.Sprintf("%s failed to parse URL: %s", strings.ToLower(i.name), i.value)})
			}

			if u.Host == "" {
				// Use strings.ToLower() to match the YAML in the config file to avoid confusion
				hostParseErrors = append(hostParseErrors, configError{Err: fmt.Sprintf("%s invalid URL: %s", strings.ToLower(i.name), i.value)})

			}

			if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ldaps" {
				// Use strings.ToLower() to match the YAML in the config file to avoid confusion
				hostParseErrors = append(hostParseErrors, configError{Err: fmt.Sprintf("%s missing scheme: %s", strings.ToLower(i.name), i.value)})
			}

		}
	}

	return hostParseErrors
}

// If an LDAP or Jira username is provided, ensure the password is also provided
func passwordOrTokenExistIfUsernameProvided(a *Config) []error {
	var passwordErrors []error

	if a.LDAPConfig.Username != "" && a.LDAPConfig.Password == "" {
		passwordErrors = append(passwordErrors, configError{Err: "ldapconfig.username provided without ldapconfig.password"})
	}

	if a.JiraConfig.Username != "" && a.JiraConfig.Token == "" {
		passwordErrors = append(passwordErrors, configError{Err: "jiraconfig.username provided without jiraconfig.token"})
	}

	return passwordErrors
}
