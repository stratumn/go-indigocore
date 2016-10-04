// Copyright 2016 Stratumn SAS. All rights reserved.
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

// The command start contains various subcommands related to Stratumn.
//
// Usage
//
//	$ strat
//	Subcommands for CLI:
//		info             print program info
//		update           update the CLI or generators
//		version          print version info
//
//	Subcommands for generator:
//		generate         generate a project
//		generators       list generators
//
//	Subcommands for help:
//		commands         list all command names
//		flags            describe all known top-level flags
//		help             describe subcommands and their syntax
//
//	Subcommands for project:
//		build            build project
//		deploy           deploy project to an environment
//		down             stop services
//		pull             pull updates
//		push             push updates
//		run              run script by name
//		test             run tests
//		up               start services
//
// Env
//
//      GITHUB_TOKEN="xxxx" # Github token to access private repos
package main
