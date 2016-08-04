// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package storetesting

import (
	"testing"

	"github.com/stratumn/go/store"
)

// TestGetInfo tests what happens when you get information about the adapter.
func TestGetInfo(t *testing.T, a store.Adapter) {
	info, err := a.GetInfo()

	if info == nil {
		t.Fatal("info is nil")
	}

	if err != nil {
		t.Fatal(err)
	}
}
