// Copyright 2016 Stratumn SAS. All rights reserved.
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

package storehttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
)

const zeros = "0000000000000000000000000000000000000000000000000000000000000000"

func TestRoot(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.StatusCode = %d want %d", got, want)
	}
	if got, want := body["adapter"].(string), "test"; got != want {
		t.Errorf(`body["adapter"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestRoot_err(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestSaveSegment(t *testing.T) {
	s, a := createServer()
	a.MockSaveSegment.Fn = func(*cs.Segment) error { return nil }

	s1 := cstesting.RandomSegment()
	var s2 cs.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/segments", s1, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if !reflect.DeepEqual(a.MockSaveSegment.LastCalledWith, s1) {
		got, _ := json.MarshalIndent(a.MockSaveSegment.LastCalledWith, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("a.MockSaveSegment.LastCalledWith = %s\nwant %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(&s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockSaveSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockSaveSegment.CalledCount = %d want %d", got, want)
	}
}

func TestSaveSegment_err(t *testing.T) {
	s, a := createServer()
	a.MockSaveSegment.Fn = func(*cs.Segment) error { return errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/segments", cstesting.RandomSegment(), &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockSaveSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockSaveSegment.CalledCount = %d want %d", got, want)
	}
}

func TestSaveSegment_invalidSegment(t *testing.T) {
	s, a := createServer()

	s1 := cstesting.RandomSegment()
	s1.Meta["linkHash"] = true
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/segments", s1, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), "meta.linkHash should be a non empty string"; got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockSaveSegment.CalledCount, 0; got != want {
		t.Errorf("a.MockSaveSegment.CalledCount = %d want %d", got, want)
	}
}

func TestSaveSegment_invalidJSON(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/segments", "azertyuio", &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrBadRequest("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockSaveSegment.CalledCount, 0; got != want {
		t.Errorf("a.MockSaveSegment.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment(t *testing.T) {
	s, a := createServer()
	s1 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return s1, nil }

	var s2 cs.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(&s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment_notFound(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment_err(t *testing.T) {
	s, a := createServer()
	a.MockGetSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return nil, errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestDeleteSegment(t *testing.T) {
	s, a := createServer()
	s1 := cstesting.RandomSegment()
	a.MockDeleteSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return s1, nil }

	var s2 cs.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/segments/"+zeros, nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockDeleteSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockDeleteSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(&s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockDeleteSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockDeleteSegment.CalledCount = %d want %d", got, want)
	}
}

func TestDeleteSegment_notFound(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockDeleteSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockDeleteSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockDeleteSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockDeleteSegment.CalledCount = %d want %d", got, want)
	}
}

func TestDeleteSegment_err(t *testing.T) {
	s, a := createServer()
	a.MockDeleteSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return nil, errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockDeleteSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockDeleteSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockDeleteSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockDeleteSegment.CalledCount = %d want %d", got, want)
	}
}

func TestFindSegments(t *testing.T) {
	s, a := createServer()
	var s1 cs.SegmentSlice
	for i := 0; i < 10; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&mapId=123&prevLinkHash="+zeros+"&tags=one+two", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}

	f := a.MockFindSegments.LastCalledWith
	if got, want := f.Offset, 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := f.Limit, 2; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Limit = %d want %d", got, want)
	}
	if got, want := f.MapID, "123"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapID = %q want %q", got, want)
	}
	if got, want := f.PrevLinkHash.String(), zeros; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q want %q", got, want)
	}
	if got, want := f.Tags, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.LastCalledWith.Tags = %v want %v", got, want)
	}
}

func TestFindSegments_defaultLimit(t *testing.T) {
	s, a := createServer()
	var s1 cs.SegmentSlice
	for i := 0; i < 2; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&&mapId=123&prevLinkHash="+zeros+"&tags=one+two", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}

	f := a.MockFindSegments.LastCalledWith
	if got, want := f.Offset, 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := f.Limit, store.DefaultLimit; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Limit = %d want %d", got, want)
	}
	if got, want := f.MapID, "123"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapID = %q want %q", got, want)
	}
	if got, want := f.PrevLinkHash.String(), zeros; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q want %q", got, want)
	}
	if got, want := f.Tags, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.LastCalledWith.Tags = %v want %v", got, want)
	}
}

func TestFindSegments_err(t *testing.T) {
	s, a := createServer()
	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}
}

func TestFindSegments_invalidOffset(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=a", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrOffset("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 0; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs(t *testing.T) {
	s, a := createServer()
	s1 := []string{"one", "two", "three"}
	a.MockGetMapIDs.Fn = func(*store.Pagination) ([]string, error) { return s1, nil }

	var s2 []string
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?offset=20&limit=10", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := s2, s1; !reflect.DeepEqual(got, want) {
		t.Errorf("s2 = %v want %v", got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 1; got != want {
		t.Errorf("a.MockGetMapIDs(pagination).CalledCount = %d want %d", got, want)
	}

	p := a.MockGetMapIDs.LastCalledWith
	if got, want := p.Offset, 20; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := p.Limit, 10; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith.Limit = %d want %d", got, want)
	}
}

func TestGetMapIDs_err(t *testing.T) {
	s, a := createServer()
	a.MockGetMapIDs.Fn = func(*store.Pagination) ([]string, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 1; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs_invalidLimit(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?limit=-1", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrLimit("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 0; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs_limitTooLarge(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	limit := store.MaxLimit + 1
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", fmt.Sprintf("/maps?limit=%d", limit), nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrLimit("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 0; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
}
