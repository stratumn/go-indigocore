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
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/postgresstore"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initBatch(t *testing.T, a store.Adapter) store.Batch {
	b, err := a.NewBatch(context.Background())
	require.NoError(t, err, "a.NewBatch()")
	require.NotNil(t, b, "Batch should not be nil")
	return b
}

// TestBatch runs all tests for the store.Batch interface
func (f Factory) TestBatch(t *testing.T) {
	ctx := context.Background()
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	// Initialize the adapter with a few links with specific map ids
	for i := 0; i < 6; i++ {
		link := chainscripttest.NewLinkBuilder(t).WithRandomData().WithMapID(fmt.Sprintf("map%d", i%3)).Build()
		_, err := a.CreateLink(ctx, link)
		require.NoError(t, err, "a.CreateLink()")
	}

	t.Run("CreateLink should not write to underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		link := chainscripttest.RandomLink(t)
		linkHash, err := b.CreateLink(ctx, link)
		assert.NoError(t, err, "b.CreateLink()")

		found, err := a.GetSegment(ctx, linkHash)
		assert.NoError(t, err, "a.GetSegment()")
		assert.Nil(t, found, "Link should not be found in adapter until Write is called")
	})

	t.Run("CreateLink should handle previous link in batch", func(t *testing.T) {
		ctx := context.Background()
		b := initBatch(t, a)

		l1 := chainscripttest.RandomLink(t)
		lh1, err := b.CreateLink(ctx, l1)
		require.NoError(t, err, "b.CreateLink()")

		l2 := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithParent(t, l1).
			WithProcess(l1.Meta.Process.Name).
			WithMapID(l1.Meta.MapId).
			Build()
		lh2, err := b.CreateLink(ctx, l2)
		require.NoError(t, err, "b.CreateLink()")

		err = b.Write(ctx)
		require.NoError(t, err, "b.Write()")

		for _, lh := range []chainscript.LinkHash{lh1, lh2} {
			found, err := a.GetSegment(ctx, lh)
			assert.NoError(t, err, "a.GetSegment()")
			require.NotNil(t, found, "a.GetSegment()")
		}
	})

	t.Run("CreateLink should reject links after failure", func(t *testing.T) {
		ctx := context.Background()
		b := initBatch(t, a)

		// Only the postgres batch actually enforces that at the moment.
		// Bufferedbatch fails when Write() is called.
		_, ok := b.(*postgresstore.Batch)
		if !ok {
			t.Skip("Test not applicable to current batch implementation")
		}

		parentNotInStore := chainscripttest.RandomLink(t)
		invalidLink := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithParent(t, parentNotInStore).
			WithProcess(parentNotInStore.Meta.Process.Name).
			WithMapID(parentNotInStore.Meta.MapId).
			Build()
		_, err := b.CreateLink(ctx, invalidLink)
		assert.Error(t, err)

		validLink := chainscripttest.RandomLink(t)
		_, err = b.CreateLink(ctx, validLink)
		assert.EqualError(t, err, store.ErrBatchFailed.Error())
	})

	t.Run("Write should write to the underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		link := chainscripttest.RandomLink(t)
		linkHash, err := b.CreateLink(ctx, link)
		assert.NoError(t, err, "b.CreateLink()")

		err = b.Write(ctx)
		assert.NoError(t, err, "b.Write()")

		found, err := a.GetSegment(ctx, linkHash)
		assert.NoError(t, err, "a.GetSegment()")
		require.NotNil(t, found, "a.GetSegment()")
		chainscripttest.LinksEqual(t, link, found.Link)
	})

	t.Run("Finding segments should find in both batch and underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		var segs *types.PaginatedSegments
		var err error
		segs, err = b.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.FindSegments()")
		require.NotNil(t, segs)
		assert.Len(t, segs.Segments, segs.TotalCount)
		adapterLinksCount := len(segs.Segments)

		_, err = b.CreateLink(ctx, chainscripttest.RandomLink(t))
		require.NoError(t, err, "b.CreateLink()")
		_, err = b.CreateLink(ctx, chainscripttest.RandomLink(t))
		require.NoError(t, err, "b.CreateLink()")

		segs, err = b.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.FindSegments()")
		require.NotNil(t, segs)
		assert.Len(t, segs.Segments, adapterLinksCount+2, "Invalid number of segments found")
	})

	t.Run("Finding maps should find in both batch and underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		mapIDs, err := b.GetMapIDs(ctx, &store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.GetMapIDs()")
		adapterMapIdsCount := len(mapIDs)

		for _, mapID := range []string{"map42", "map43"} {
			link := chainscripttest.NewLinkBuilder(t).WithMapID(mapID).Build()
			_, err = b.CreateLink(ctx, link)
			require.NoError(t, err, "b.CreateLink()")
		}

		mapIDs, err = b.GetMapIDs(ctx, &store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.GetMapIDs()")
		assert.Equal(t, adapterMapIdsCount+2, len(mapIDs), "Invalid number of maps")

		want := map[string]interface{}{"map0": nil, "map1": nil, "map2": nil, "map42": nil, "map43": nil}
		got := make(map[string]interface{}, len(mapIDs))
		for _, mapID := range mapIDs {
			got[mapID] = nil
		}

		for mapID := range want {
			_, exist := got[mapID]
			assert.True(t, exist, "Missing map: %s", mapID)
		}
	})
}
