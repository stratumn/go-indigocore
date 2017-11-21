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

package filestore

import (
	"io/ioutil"
	"os"

	"github.com/stratumn/sdk/store"
)

func createAdapter() (store.Adapter, error) {
	path, err := ioutil.TempDir("", "filestore")
	if err != nil {
		return nil, err
	}
	fs, err := New(&Config{Path: path})
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func freeAdapter(s store.Adapter) {
	a := s.(*FileStore)
	defer os.RemoveAll(a.config.Path)
}
