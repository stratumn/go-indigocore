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

package jsonhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	s := New(&Config{})
	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPost(t *testing.T) {
	s := New(&Config{})
	s.Post("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPut(t *testing.T) {
	s := New(&Config{})
	s.Put("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PUT", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestDelete(t *testing.T) {
	s := New(&Config{})
	s.Delete("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPatch(t *testing.T) {
	s := New(&Config{})
	s.Patch("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PATCH", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestOptions(t *testing.T) {
	s := New(&Config{})
	s.Options("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "OPTIONS", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestCORS(t *testing.T) {
	s := New(&Config{EnableCORS: true})
	// Once CORS is enabled, options calls are automatically handled on all
	// registered routes.
	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "OPTIONS", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNotFound(t *testing.T) {
	s := New(&Config{})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Not Found", body["error"])
	assert.Equal(t, http.StatusNotFound, int(body["status"].(float64)))
}

func TestErrHTTP(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		validate func(*testing.T, interface{})
	}{{
		"standard error",
		errors.New("nothing for you today sir"),
		func(t *testing.T, err interface{}) {
			assert.Equal(t, "nothing for you today sir", err)
		},
	}, {
		"structured error",
		types.WrapError(
			types.WrapError(
				errors.New("inner error"),
				errorcode.FailedPrecondition,
				"A",
				"inner message",
			),
			errorcode.Internal,
			"B",
			"outer message",
		),
		func(t *testing.T, err interface{}) {
			expected := map[string]interface{}{
				"category": "B",
				"code":     float64(13),
				"inner": map[string]interface{}{
					"category": "A",
					"code":     float64(9),
					"inner":    "inner error",
					"message":  "inner message",
				},
				"message": "outer message",
			}
			assert.Equal(t, expected, err)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := New(&Config{})
			s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
				return nil, ErrHTTP{err: tt.err, status: http.StatusBadRequest}
			})

			var body map[string]interface{}
			w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)

			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, http.StatusBadRequest, int(body["status"].(float64)))
			tt.validate(t, body["error"])
		})
	}
}

func TestError(t *testing.T) {
	s := New(&Config{})

	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return nil, errors.New("no")
	})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error", body["error"])
	assert.Equal(t, http.StatusInternalServerError, int(body["status"].(float64)))
}
