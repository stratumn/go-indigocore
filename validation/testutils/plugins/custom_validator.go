// Copyright 2017 Stratumn SAS. All rights reserved.
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

package main

import (
	"context"
	"errors"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/store"
)

// Init validates the transition towards the "init" state
func Init(storeReader store.SegmentReader, l *chainscript.Link) error {
	return nil
}

// FetchLink fetches a link and returns a nil error
func FetchLink(storeReader store.SegmentReader, l *chainscript.Link) error {
	_, err := storeReader.FindSegments(context.Background(), &store.SegmentFilter{
		MapIDs: []string{l.Meta.MapId},
	})
	return err
}

// Invalid validates the transition towards the "invalid" state
func Invalid(storeReader store.SegmentReader, l *chainscript.Link) error {
	return errors.New("error")
}

// BadSignature is an example of validator which is not of type ScriptValidatorFunc
func BadSignature() error {
	return nil
}

func main() {}
