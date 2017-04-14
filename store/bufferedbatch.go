// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package store

import (
	"bytes"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

// BufferedBatch can be used as a base class for types
// that want to implement github.com/stratumn/sdk/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type BufferedBatch struct {
	originalStore Adapter
	ValueOps      []ValueOperation
	SegmentOps    []SegmentOperation
}

// OpType represents a operation type on the Batch.
type OpType int

const (
	// OpTypeSet set represents a save operation.
	OpTypeSet = iota

	// OpTypeDelete set represents a delete operation.
	OpTypeDelete
)

// ValueOperation represents a operation on a value.
type ValueOperation struct {
	OpType
	Key   []byte
	Value []byte
}

// SegmentOperation represents a operation on a segment.
type SegmentOperation struct {
	OpType
	LinkHash *types.Bytes32
	Segment  *cs.Segment
}

// NewBufferedBatch creates a new Batch.
func NewBufferedBatch(a Adapter) *BufferedBatch {
	return &BufferedBatch{originalStore: a}
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (b *BufferedBatch) SaveValue(key, value []byte) error {
	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeSet, key, value})
	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (b *BufferedBatch) DeleteValue(key []byte) (value []byte, err error) {
	ops := make([]ValueOperation, len(b.ValueOps))
	copy(ops, b.ValueOps)

	// remove all existing save operations and get the last saved value.
	for i, sOp := range ops {
		if bytes.Compare(sOp.Key, key) == 0 {
			value = sOp.Value
			b.ValueOps = append(b.ValueOps[:i], b.ValueOps[i+1:]...)
		}
	}

	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeDelete, key, nil})

	if value != nil {
		return value, nil
	}
	return b.originalStore.GetValue(key)
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (b *BufferedBatch) SaveSegment(segment *cs.Segment) error {
	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeSet, nil, segment})
	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (b *BufferedBatch) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	ops := make([]SegmentOperation, len(b.SegmentOps))
	copy(ops, b.SegmentOps)
	// remove all existing save operations and get the last saved value.
	for i, sOp := range ops {
		if sOp.LinkHash == linkHash {
			segment = sOp.Segment
			b.SegmentOps = append(b.SegmentOps[:i], b.SegmentOps[i+1:]...)
		}
	}

	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeDelete, linkHash, nil})

	if segment != nil {
		return segment, nil
	}
	return b.originalStore.GetSegment(linkHash)
}

// TODO: read from buffer in addition to store

// GetSegment delegates the call to a underlying store
func (b *BufferedBatch) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	return b.originalStore.GetSegment(linkHash)
}

// FindSegments delegates the call to a underlying store
func (b *BufferedBatch) FindSegments(filter *Filter) (cs.SegmentSlice, error) {
	return b.originalStore.FindSegments(filter)
}

// GetMapIDs delegates the call to a underlying store
func (b *BufferedBatch) GetMapIDs(pagination *Pagination) ([]string, error) {
	return b.originalStore.GetMapIDs(pagination)
}

// GetValue delegates the call to a underlying store
func (b *BufferedBatch) GetValue(key []byte) ([]byte, error) {
	return b.originalStore.GetValue(key)
}
