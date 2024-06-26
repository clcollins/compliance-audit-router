/*
Copyright 2021-2024 Red Hat, Inc

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

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/listeners"

	"github.com/openshift/compliance-audit-router/pkg/metrics"
)

func init() {
	config.LoadConfig()
}

func main() {
	if config.AppConfig.DryRun {
		log.Printf("running in dry-run mode; no actions will be taken")
	}

	var portString = ":" + fmt.Sprint(config.AppConfig.ListenPort)

	r := chi.NewRouter()
	r.Use(middleware.DefaultLogger)

	metrics.RegisterMetrics()

	log.Printf("initializing routes")
	listeners.InitRoutes(r)

	log.Printf("listening on %s", portString)
	log.Fatal(http.ListenAndServe(portString, r))
}
