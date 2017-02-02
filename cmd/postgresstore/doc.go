// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// The command postgresstore starts an HTTP server with a postgresstore.
//
// Usage
//
// The following flags are available:
//
//	$ postgresstore -h
//        -create
//          	create tables and indexes then exit
//        -drop
//          	drop tables and indexes then exit
//        -http string
//          	HTTP address (default ":5000")
//	  -maxmsgsize int
//	    	Maximum size of a received web socket message (default 32768)
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -url string
//          	URL of the PostgreSQL database (default "postgres://postgres@postgres/postgres?sslmode=disable")
//	  -wspinginterval duration
//	    	Interval between web socket pings (default 54s)
//	  -wspongtimeout duration
//	    	Timeout for a web socket expected pong (default 1m0s)
//	  -wsreadbufsize int
//	    	Web socket read buffer size (default 1024)
//	  -wswritebufsize int
//	    	Web socket write buffer size (default 1024)
//	  -wswritechansize int
//	    	Size of a web socket connection write channel (default 256)
//	  -wswritetimeout duration
//	    	Timeout for a web socket write (default 10s
//
// Env
//
//      POSTGRESSTORE_URL="xxx" // URL of the PostgreSQL database
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 5000:5000 stratumn/progresstore postgresstore \
//              -url 'postgres://postgres@localhost/postgres?sslmode=disable'
package main
