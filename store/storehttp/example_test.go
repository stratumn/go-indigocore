// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storehttp_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"time"

	"github.com/stratumn/go/dummystore"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/jsonws"
	"github.com/stratumn/go/store/storehttp"
)

// This example shows how to create a server from a dummystore.
// It also tests the root route of the server using net/http/httptest.
func Example() {
	// Create a dummy adapter.
	a := dummystore.New(&dummystore.Config{Version: "0.1.0", Commit: "abc"})
	httpConfig := &jsonhttp.Config{
		Address: "5555",
	}
	basicConfig := &jsonws.BasicConfig{}
	bufConnConfig := &jsonws.BufferedConnConfig{
		Size:         256,
		WriteTimeout: 10 * time.Second,
		PongTimeout:  70 * time.Second,
		PingInterval: time.Minute,
		MaxMsgSize:   1024,
	}

	// Create a server.
	s := storehttp.New(a, httpConfig, basicConfig, bufConnConfig)
	go s.Start()
	defer s.Shutdown()

	// Create a test server.
	ts := httptest.NewServer(s)
	defer ts.Close()

	// Test the root route.
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	info, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", info)
	// Output: {"adapter":{"name":"dummy","description":"Stratumn Dummy Store","version":"0.1.0","commit":"abc"}}
}
