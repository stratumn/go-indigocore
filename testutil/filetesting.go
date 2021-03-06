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

package testutil

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateTempFile creates a temporary file for tests with data as content.
func CreateTempFile(t *testing.T, data string) string {
	tmpfile, err := ioutil.TempFile("", "core-tmpfile")
	require.NoError(t, err, "ioutil.TempFile()")

	_, err = tmpfile.WriteString(data)
	require.NoError(t, err, "tmpfile.WriteString()")
	return tmpfile.Name()
}
