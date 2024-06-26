// Copyright 2021-2024 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ldap

import (
	"io"
	"log"
	"os"
	"testing"
)

// Silence logto
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestGetUID(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		expectedError  string
	}{
		{
			name:           "valid input",
			input:          "uid=avulaj,ou=users,dc=redhat,dc=com",
			expectedResult: "avulaj",
		},
		{
			name:          "no uid present",
			input:         "ou=users,dc=redhat,dc=com",
			expectedError: "ldap.getUID(): no uid field found for given ldap string",
		},
		{
			name:          "malformed dn",
			input:         "uid:avulaj",
			expectedError: "ldap.getUID(): error parsing dn: DN ended with incomplete type, value pair",
		},
		{
			name:          "empty input",
			expectedError: "ldap.getUID(): no uid field found for given ldap string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getUID(tc.input)
			if tc.expectedError != "" && err.Error() != tc.expectedError {
				t.Fatalf("Did not receive the expected error.\nExpected: %v\nActual: %v", tc.expectedError, err.Error())
			}
			if result != tc.expectedResult {
				t.Fatalf("Expected %v, but got %v", tc.expectedResult, result)
			}
		})
	}
}
