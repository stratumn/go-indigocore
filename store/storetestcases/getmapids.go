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

package storetestcases

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMapIDs tests what happens when you get map IDs.
func (f Factory) TestGetMapIDs(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	processNames := [2]string{"Foo", "Bar"}
	testPageSize := 3
	maps := make(map[string]*chainscript.Link, 4)
	for i := 0; i < testPageSize; i++ {
		mapID := fmt.Sprintf("map%d", i)
		for j := 0; j < testPageSize; j++ {
			l := chainscripttest.NewLinkBuilder(t).
				WithRandomData().
				WithProcess(processNames[i%2]).
				WithMapID(mapID).
				Build()
			_, err := a.CreateLink(context.Background(), l)
			require.NoError(t, err)
			maps[mapID] = l
		}
	}

	mapID := "other-map2"
	l := chainscripttest.NewLinkBuilder(t).
		WithMapID(mapID).
		Build()
	_, err := a.CreateLink(context.Background(), l)
	require.NoError(t, err)
	maps[mapID] = l

	linksCnt := testPageSize*testPageSize + 1

	t.Run("Getting all map IDs should work", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
		})
		assert.NoError(t, err)
		assert.Equal(t, testPageSize+1, len(slice), "Invalid number of map IDs")

		for i := 0; i < testPageSize; i++ {
			mapID := fmt.Sprintf("map%d", i)
			assert.True(t, testutil.ContainsString(slice, mapID),
				"slice does not contain %s", mapID)
		}
		assert.True(t, testutil.ContainsString(slice, "other-map2"),
			"slice does not contain %s", "other-map2")
	})

	t.Run("Map ID pagination should work", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Offset: 1, Limit: 2},
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(slice), "Invalid number of map IDs found")
	})

	t.Run("Map ID outside pagination limits should return an empty slice", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Offset: 100000, Limit: 5},
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(slice), "Invalid number of map IDs found")
	})

	t.Run("Filtering by process should work", func(t *testing.T) {
		ctx := context.Background()
		processName := processNames[0]
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
			Process:    processName,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(slice), "Invalid number of maps for %s", processName)

		for i := 0; i < testPageSize; i += 2 {
			expectedMapID := fmt.Sprintf("map%d", i)
			assert.True(t, testutil.ContainsString(slice, expectedMapID),
				"slice does not contain %q", expectedMapID)
		}
	})

	t.Run("Filtering by prefix should work", func(t *testing.T) {
		ctx := context.Background()
		prefix := "map"
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
			Prefix:     prefix,
		})
		assert.NoError(t, err)
		assert.Equal(t, testPageSize, len(slice))

		for _, mapID := range slice {
			assert.True(t, strings.HasPrefix(mapID, prefix))
		}
	})

	t.Run("Filtering by prefix and process should work", func(t *testing.T) {
		ctx := context.Background()
		prefix := "map"
		process := "Foo"
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
			Prefix:     prefix,
			Process:    process,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(slice))

		for _, mapID := range slice {
			assert.True(t, strings.HasPrefix(mapID, prefix))
			assert.Equal(t, process, maps[mapID].Meta.Process.Name)
		}
	})

	t.Run("Filtering by suffix should work", func(t *testing.T) {
		ctx := context.Background()
		suffix := "ap2"
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
			Suffix:     suffix,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(slice))

		for _, mapID := range slice {
			assert.True(t, strings.HasSuffix(mapID, suffix))
		}
	})

	t.Run("Filtering by suffix and process should work", func(t *testing.T) {
		ctx := context.Background()
		suffix := "ap2"
		process := "Foo"
		slice, err := a.GetMapIDs(ctx, &store.MapFilter{
			Pagination: store.Pagination{Limit: linksCnt},
			Suffix:     suffix,
			Process:    process,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(slice))

		for _, mapID := range slice {
			assert.True(t, strings.HasSuffix(mapID, suffix))
			assert.Equal(t, process, maps[mapID].Meta.Process.Name)
		}
	})

}

// BenchmarkGetMapIDs benchmarks getting map IDs.
func (f Factory) BenchmarkGetMapIDs(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc MapFilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		_, err := a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
		if err != nil {
			b.Fatal(err)
		}
	}

	filters := make([]*store.MapFilter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numLinks, i)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.GetMapIDs(context.Background(), filters[i]); err != nil {
			b.Fatal(err)
		} else if s == nil {
			b.Error("s = nil want []string")
		}
	}
}

// BenchmarkGetMapIDs100 benchmarks getting map IDs within 100 segments.
func (f Factory) BenchmarkGetMapIDs100(b *testing.B) {
	f.BenchmarkGetMapIDs(b, 100, RandomLink, RandomPaginationOffset)
}

// BenchmarkGetMapIDs1000 benchmarks getting map IDs within 1000 segments.
func (f Factory) BenchmarkGetMapIDs1000(b *testing.B) {
	f.BenchmarkGetMapIDs(b, 1000, RandomLink, RandomPaginationOffset)
}

// BenchmarkGetMapIDs10000 benchmarks getting map IDs within 10000 segments.
func (f Factory) BenchmarkGetMapIDs10000(b *testing.B) {
	f.BenchmarkGetMapIDs(b, 10000, RandomLink, RandomPaginationOffset)
}

// BenchmarkGetMapIDsParallel benchmarks getting map IDs in parallel.
func (f Factory) BenchmarkGetMapIDsParallel(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc MapFilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		_, err := a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
		if err != nil {
			b.Fatal(err)
		}
	}

	filters := make([]*store.MapFilter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numLinks, i)
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddUint64(&counter, 1) - 1)
			if s, err := a.GetMapIDs(context.Background(), filters[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want []string")
			}
		}
	})
}

// BenchmarkGetMapIDs100Parallel benchmarks gettiBenchmarkFindSegmentsPrevLinkHashTags100Parallelng map IDs within 100 segments
// in parallel.
func (f Factory) BenchmarkGetMapIDs100Parallel(b *testing.B) {
	f.BenchmarkGetMapIDsParallel(b, 100, RandomLink, RandomPaginationOffset)
}

// BenchmarkGetMapIDs1000Parallel benchmarks getting map IDs within 1000
// segments in parallel.
func (f Factory) BenchmarkGetMapIDs1000Parallel(b *testing.B) {
	f.BenchmarkGetMapIDsParallel(b, 1000, RandomLink, RandomPaginationOffset)
}

// BenchmarkGetMapIDs10000Parallel benchmarks getting map IDs within 10000
// segments in parallel.
func (f Factory) BenchmarkGetMapIDs10000Parallel(b *testing.B) {
	f.BenchmarkGetMapIDsParallel(b, 10000, RandomLink, RandomPaginationOffset)
}
