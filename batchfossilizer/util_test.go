// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package batchfossilizer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/merkle"
)

type fossilizeTest struct {
	data       []byte
	meta       []byte
	path       merkle.Path
	sleep      time.Duration
	fossilized bool
}

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest, start bool, fossilize bool) {
	rc := make(chan *fossilizer.Result, 1)
	a.AddResultChan(rc)

	if start {
		go func() {
			if err := a.Start(); err != nil {
				t.Errorf("a.Start(): err: %s", err)
			}
		}()
	}

	<-a.Started()

	if fossilize {
		for _, test := range tests {
			if err := a.Fossilize(test.data, test.meta); err != nil {
				t.Errorf("a.Fossilize(): err: %s", err)
			}
			if test.sleep > 0 {
				time.Sleep(test.sleep)
			}
		}
	}

RESULT_LOOP:
	for _ = range tests {
		r := <-rc
		for i := range tests {
			test := &tests[i]
			if fmt.Sprint(test.meta) == fmt.Sprint(r.Meta) {
				test.fossilized = true
				if !reflect.DeepEqual(r.Data, test.data) {
					got := fmt.Sprintf("%x", r.Data)
					want := fmt.Sprintf("%x", test.data)
					t.Errorf("test#%d: Data = %q want %q", i, got, want)
				}
				evidence := r.Evidence.(*EvidenceWrapper).Evidence
				if !reflect.DeepEqual(evidence.Path, test.path) {
					got, _ := json.MarshalIndent(evidence.Path, "", "  ")
					want, _ := json.MarshalIndent(test.path, "", "  ")
					t.Errorf("test#%d: Path = %s\nwant %s", i, got, want)
				}
				continue RESULT_LOOP
			}
		}
		a := fmt.Sprintf("%x", r.Meta)
		t.Errorf("unexpected Meta %q", a)
	}

	for i, test := range tests {
		if !test.fossilized {
			t.Errorf("test#%d: not fossilized", i)
		}
	}

	if start {
		a.Stop()
	}
}

func benchmarkFossilize(b *testing.B, config *Config) {
	n := b.N

	a, err := New(config)
	if err != nil {
		b.Fatalf("New(): err: %s", err)
	}

	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

	go func() {
		if err := a.Start(); err != nil {
			b.Errorf("a.Start(): err: %s", err)
		}
	}()

	data := make([][]byte, n)
	for i := 0; i < n; i++ {
		data[i] = atos(*testutil.RandomHash())
	}

	<-a.Started()

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	go func() {
		for i := 0; i < n; i++ {
			if err := a.Fossilize(data[i], data[i]); err != nil {
				b.Errorf("a.Fossilize(): err: %s", err)
			}
		}
		a.Stop()
	}()

	for i := 0; i < n; i++ {
		<-rc
	}

	b.StopTimer()
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
