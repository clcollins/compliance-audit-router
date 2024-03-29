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

package splunk

import (
	"fmt"
	"log"
	"time"
)

const SplunkTimeFormat = "2006-01-02T15:04:05.GMT"

func (a SearchResult) string(field string) string {
	if i, ok := a[field]; !ok {
		log.Printf("No such field: %s", field)
		return ""
	} else {
		return fmt.Sprint(i)
	}
}

func (a SearchResult) slice(field string) []string {
	if i, ok := a[field]; !ok {
		return []string{}
	} else {
		var values []string
		switch v := i.(type) {
		case string:
			values = append(values, v)
		case []string:
			values = append(values, v...)
		case []interface{}:
			for _, e := range v {
				values = append(values, e.(string))
			}
		default:
			log.Printf("Unknown type for field %s: %T", field, v)
		}
		return values
	}
}

func (a SearchResult) time(field string) time.Time {
	if s := a.string(field); s != "" {
		if t, err := time.Parse(SplunkTimeFormat, s); err == nil {
			return t
		} else {
			log.Printf("Error parsing timestamp: %v", err)
		}
	}
	return time.Time{}
}
