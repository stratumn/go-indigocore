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

// Package cs defines types to work with Chainscripts.
package cs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"reflect"

	"github.com/pkg/errors"

	"github.com/stratumn/sdk/types"

	cj "github.com/gibson042/canonicaljson-go"
)

// Segment contains a link and meta data about the link.
type Segment struct {
	Link Link        `json:"link"`
	Meta SegmentMeta `json:"meta"`
}

// GetLinkHash returns the link ID as bytes.
// It assumes the segment is valid.
func (s *Segment) GetLinkHash() *types.Bytes32 {
	return s.Meta.GetLinkHash()
}

// GetLinkHashString returns the link ID as a string.
// It assumes the segment is valid.
func (s *Segment) GetLinkHashString() string {
	return s.Meta.GetLinkHashString()
}

// HashLink hashes the segment link and stores it into the Meta
func (s *Segment) HashLink() (string, error) {
	return s.Link.HashString()
}

// SetLinkHash overwrites the segment LinkHash using HashLink()
func (s *Segment) SetLinkHash() error {
	linkHash, err := s.HashLink()
	if err != nil {
		return err
	}

	s.Meta.LinkHash = linkHash
	return nil
}

// GetSegmentFunc is the function signature to retrieve a Segment
type GetSegmentFunc func(linkHash *types.Bytes32) (*Segment, error)

// Validate checks for errors in a segment.
func (s *Segment) Validate(getSegment GetSegmentFunc) error {
	if s.Meta.LinkHash == "" {
		return errors.New("meta.linkHash should be a non empty string")
	}
	if process, ok := s.Link.Meta["process"].(string); !ok || process == "" {
		return errors.New("link.meta.process should be a non empty string")
	}
	if mapID, ok := s.Link.Meta["mapId"].(string); !ok || mapID == "" {
		return errors.New("link.meta.mapId should be a non empty string")
	}
	if v, ok := s.Link.Meta["prevLinkHash"]; ok {
		if prevLinkHash, ok := v.(string); !ok || prevLinkHash == "" {
			return errors.New("link.meta.prevLinkHash should be a non empty string")
		}
	}

	if v, ok := s.Link.Meta["tags"]; ok {
		tags, ok := v.([]interface{})
		if !ok {
			return errors.New("link.meta.tags should be an array of non empty string")
		}
		for _, t := range tags {
			if tag, ok := t.(string); !ok || tag == "" {
				return errors.New("link.meta.tags should be an array of non empty string")
			}
		}
	}

	if v, ok := s.Link.Meta["priority"]; ok {
		if _, ok := v.(float64); !ok {
			return errors.New("link.meta.priority should be a float64")
		}
	}

	want, err := s.HashLink()
	if err != nil {
		return err
	}
	if got := s.GetLinkHashString(); want != got {
		return errors.New("meta.linkHash is not in sync with link")
	}

	return s.validateReferences(getSegment)
}

func (s *Segment) validateReferences(getSegment GetSegmentFunc) error {
	if refs, ok := s.Link.Meta["refs"].([]interface{}); ok {
		for refIdx, refChild := range refs {
			ref, ok := refChild.(map[string]interface{})
			if !ok {
				return errors.Errorf("link.meta.refs[%d] should be a map", refIdx)
			}
			if jsonSeg, ok := ref["segment"].(string); ok {
				var seg Segment
				if err := json.Unmarshal([]byte(jsonSeg), &seg); err != nil {
					return errors.Errorf("link.meta.refs[%d].segment should be a valid json segment", refIdx)
				}
				if err := seg.Validate(getSegment); err != nil {
					return errors.WithMessage(err, fmt.Sprintf("invalid link.meta.refs[%d].segment", refIdx))
				}
			} else {
				process, ok := ref["process"].(string)
				if !ok || process == "" {
					return errors.Errorf("link.meta.refs[%d].process should be a non empty string", refIdx)
				}
				linkHashStr, ok := ref["linkHash"].(string)
				if !ok || linkHashStr == "" {
					return errors.Errorf("link.meta.refs[%d].linkHash should be a non empty string", refIdx)
				}
				linkHash, err := types.NewBytes32FromString(linkHashStr)
				if err != nil {
					return errors.Errorf("link.meta.refs[%d].linkHash should be a bytes32 field", refIdx)
				}
				if s.Link.Meta["process"].(string) == process && getSegment != nil {
					if seg, err := getSegment(linkHash); err != nil {
						return errors.Wrapf(err, "link.meta.refs[%d] segment should be retrieved", refIdx)
					} else if seg == nil {
						return errors.Errorf("link.meta.refs[%d] segment is nil", refIdx)
					}
				}
				// Segment from another process is not retrieved because it could be in another store
			}
		}
	}
	return nil
}

// MergeMeta updates the current segment meta from an updated one
// It returns the newly created segment
// If the updated segment contains new evidences, they will be added to the current ones
// The current segment.Meta.Data will be updated with the new segment's own Meta.Data
func (s *Segment) MergeMeta(updated *Segment) (*Segment, error) {
	if strings.Compare(s.Meta.LinkHash, updated.Meta.LinkHash) != 0 {
		return nil, errors.New("trying to merge segment meta with different linkHash")
	}
	s.Meta = s.Meta.Merge(updated.Meta)
	return s, nil
}

// IsEmpty checks if a segment is empty (nil)
func (s *Segment) IsEmpty() bool {
	return reflect.DeepEqual(*s, Segment{})
}

// SegmentMeta contains additional information about the segment and a proof of existence
type SegmentMeta struct {
	Evidences Evidences              `json:"evidences"`
	LinkHash  string                 `json:"linkHash"`
	Data      map[string]interface{} `json:"data"`
}

// GetLinkHash returns the link ID as bytes.
// It assumes the segment is valid.
func (s *SegmentMeta) GetLinkHash() *types.Bytes32 {
	b, _ := types.NewBytes32FromString(s.LinkHash)
	return b
}

// GetLinkHashString returns the link ID as a string.
// It assumes the segment is valid.
func (s *SegmentMeta) GetLinkHashString() string {
	return s.LinkHash
}

// AddEvidence sets the segment evidence
func (s *SegmentMeta) AddEvidence(evidence Evidence) error {
	return s.Evidences.AddEvidence(evidence)
}

// GetEvidence gets an evidence from a provider
func (s *SegmentMeta) GetEvidence(provider string) *Evidence {
	return s.Evidences.GetEvidence(provider)
}

// FindEvidences find all evidences generated by a specified backend ("TMPop", "bcbatchfossilizer"...)
func (s *SegmentMeta) FindEvidences(backend string) (res Evidences) {
	return s.Evidences.FindEvidences(backend)
}

// Merge updates the current meta from an updated one
// It returns the newly updated SegmentMeta
// If the updated Meta contains new evidences, they will be added to the current ones
// The current SegmentMeta.Data will be updated with the new SegmentMeta.Data
func (s *SegmentMeta) Merge(updated SegmentMeta) SegmentMeta {
	prevEvidence := &Evidence{}
	for _, e := range updated.Evidences {
		prevEvidence = s.GetEvidence(e.Provider)
		if prevEvidence != nil && prevEvidence.State == PendingEvidence {
			prevEvidence.State = e.State
			prevEvidence.Proof = e.Proof
		}
		if prevEvidence == nil {
			s.AddEvidence(*e)
		}
	}

	for key, value := range updated.Data {
		s.Data[key] = value
	}
	return *s
}

// Link contains a state and meta data about the state.
type Link struct {
	State map[string]interface{} `json:"state"`
	Meta  map[string]interface{} `json:"meta"`
}

// Hash hashes the link
func (l *Link) Hash() (*types.Bytes32, error) {
	jsonLink, err := cj.Marshal(l)
	if err != nil {
		return nil, err
	}
	byteLinkHash := sha256.Sum256(jsonLink)
	linkHash := types.Bytes32(byteLinkHash)
	return &linkHash, nil
}

// HashString hashes the link and returns a string
func (l *Link) HashString() (string, error) {
	jsonLink, err := cj.Marshal(l)
	if err != nil {
		return "", err
	}
	byteLinkHash := sha256.Sum256(jsonLink)
	return hex.EncodeToString(byteLinkHash[:sha256.Size]), nil
}

// GetPriority returns the priority as a float64
// It assumes the link is valid.
// If priority is nil, it will return -Infinity.
func (l *Link) GetPriority() float64 {
	if f64, ok := l.Meta["priority"].(float64); ok {
		return f64
	}
	return math.Inf(-1)
}

// GetMapID returns the map ID as a string.
// It assumes the link is valid.
func (l *Link) GetMapID() string {
	return l.Meta["mapId"].(string)
}

// GetPrevLinkHash returns the previous link hash as a bytes.
// It assumes the link is valid.
// It will return nil if the previous link hash is null.
func (l *Link) GetPrevLinkHash() *types.Bytes32 {
	if str, ok := l.Meta["prevLinkHash"].(string); ok {
		b, _ := types.NewBytes32FromString(str)
		return b
	}
	return nil
}

// GetPrevLinkHashString returns the previous link hash as a string.
// It assumes the link is valid.
// It will return an empty string if the previous link hash is null.
func (l *Link) GetPrevLinkHashString() string {
	if str, ok := l.Meta["prevLinkHash"].(string); ok {
		return str
	}
	return ""
}

// GetTags returns the tags as an array of string.
// It assumes the link is valid.
// It will return nil if there are no tags.
func (l *Link) GetTags() []string {
	if t, ok := l.Meta["tags"].([]interface{}); ok {
		tags := make([]string, len(t))
		for i, v := range t {
			tags[i] = v.(string)
		}
		return tags
	}
	return nil
}

// GetTagMap returns the tags as a map of string to empty structs (a set).
// It assumes the link is valid.
func (l *Link) GetTagMap() map[string]struct{} {
	tags := map[string]struct{}{}
	if t, ok := l.Meta["tags"].([]interface{}); ok {
		for _, v := range t {
			tags[v.(string)] = struct{}{}
		}
	}
	return tags
}

// GetProcess returns the process name as a string.
// It assumes the link is valid.
func (l *Link) GetProcess() string {
	return l.Meta["process"].(string)
}

// SegmentSlice is a slice of segment pointers.
type SegmentSlice []*Segment

// Len implements sort.Interface.Len.
func (s SegmentSlice) Len() int {
	return len(s)
}

// Swap implements sort.Interface.Swap.
func (s SegmentSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface.Less.
func (s SegmentSlice) Less(i, j int) bool {
	var (
		s1 = s[i]
		s2 = s[j]
		p1 = s1.Link.GetPriority()
		p2 = s2.Link.GetPriority()
	)

	if p1 > p2 {
		return true
	}

	if p1 < p2 {
		return false
	}

	return s1.GetLinkHashString() < s2.GetLinkHashString()
}
