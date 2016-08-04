// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// A fossilizer HTTP server with a dummy adapter.
package main

import (
	"flag"
	"log"

	"github.com/stratumn/go/dummyfossilizer"
	"github.com/stratumn/go/fossilizer/fossilizerhttp"
	"github.com/stratumn/go/jsonhttp"
)

var (
	port             = flag.String("port", fossilizerhttp.DefaultPort, "server port")
	certFile         = flag.String("tlscert", "", "TLS certificate file")
	keyFile          = flag.String("tlskey", "", "TLS private key file")
	numResultWorkers = flag.Int("workers", fossilizerhttp.DefaultNumResultWorkers, "number of result workers")
	minDataLen       = flag.Int("mindata", fossilizerhttp.DefaultMinDataLen, "minimum data length")
	maxDataLen       = flag.Int("maxdata", fossilizerhttp.DefaultMaxDataLen, "maximum data length")
	verbose          = flag.Bool("verbose", fossilizerhttp.DefaultVerbose, "verbose output")
	version          = ""
)

func init() {
	log.SetPrefix("dummyfossilizer ")
}

func main() {
	flag.Parse()

	a := dummyfossilizer.New(version)
	c := &fossilizerhttp.Config{
		Config: jsonhttp.Config{
			Port:     *port,
			CertFile: *certFile,
			KeyFile:  *keyFile,
			Verbose:  *verbose,
		},
		NumResultWorkers: *numResultWorkers,
		MinDataLen:       *minDataLen,
		MaxDataLen:       *maxDataLen,
	}
	h := fossilizerhttp.New(a, c)

	log.Printf("Listening on %s", *port)
	log.Fatal(h.ListenAndServe())
}
