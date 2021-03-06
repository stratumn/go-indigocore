// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package monitoring

import "flag"

var (
	monitor     bool
	metricsPort int
	exporter    string

	// These should be set by the executable at start-up.
	version string
	commit  string
)

// RegisterFlags registers the command-line monitoring flags.
func RegisterFlags() {
	flag.BoolVar(&monitor, "monitoring.active", true, "Set to true to activate monitoring")
	flag.IntVar(&metricsPort, "monitoring.metrics.port", 0, "Port to use to expose metrics, for example 5001")
	flag.StringVar(&exporter, "monitoring.exporter", PrometheusExporter, "Exporter for metrics and traces (either prometheus or elastic)")
}

// SetVersion sets the current code's version and commit.
func SetVersion(v, c string) {
	version = v
	commit = c
}

// ConfigurationFromFlags builds configuration from user-provided
// command-line flags.
func ConfigurationFromFlags() *Config {
	config := &Config{
		Monitor:     monitor,
		Exporter:    exporter,
		MetricsPort: metricsPort,
	}

	return config
}
