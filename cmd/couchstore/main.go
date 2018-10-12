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

// The command filestore starts a storehttp server with a couchstore.

package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-core/couchstore"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storehttp"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/utils"
	"github.com/stratumn/go-core/validation"
)

var (
	endpoint = flag.String("endpoint", "http://localhost:5984", "CouchDB endpoint")
	version  = "x.x.x"
	commit   = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	monitoring.RegisterFlags()
	validation.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", couchstore.Description, version, commit[:7])

	var a store.Adapter
	var storeErr error

	err := utils.Retry(func(attempt int) (retry bool, err error) {
		a, storeErr = couchstore.New(&couchstore.Config{
			Address: *endpoint,
			Version: version,
			Commit:  commit,
		})

		if storeErr == nil {
			return false, nil
		}

		structErr, ok := storeErr.(*types.Error)
		if ok && structErr.Code == errorcode.Unavailable {
			log.Infof("Unable to connect to couchdb (%v). Retrying in 5s.", storeErr.Error())
			time.Sleep(5 * time.Second)
			return true, storeErr
		}

		return false, storeErr
	}, 10)

	if err != nil {
		log.Fatal(storeErr)
	}

	a, err = validation.WrapStoreWithConfigFile(a, validation.ConfigurationFromFlags())
	if err != nil {
		log.Fatal(err)
	}

	storehttp.RunWithFlags(monitoring.WrapStore(a, "couchstore"))
}
