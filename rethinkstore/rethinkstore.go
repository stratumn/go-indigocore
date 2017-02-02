// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// Package rethinkstore implements a store that saves all the segments in a
// RethinkDB database.
package rethinkstore

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

func init() {
	rethink.SetTags("json", "gorethink")
}

const (
	// Name is the name set in the store's information.
	Name = "rethink"

	// Description is the description set in the store's information.
	Description = "Stratumn RethinkDB Store"

	// DefaultURL is the default URL of the database.
	DefaultURL = "rethinkdb:28015"

	// DefaultDB is the default database.
	DefaultDB = "test"

	// DefaultHard is whether to use hard durability by default.
	DefaultHard = true
)

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// The URL of the PostgreSQL database, such as "localhost:28015" order
	// "localhost:28015,localhost:28016,localhost:28017".
	URL string

	// The database name
	DB string

	// Whether to use hard durability.
	Hard bool
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Store is the type that implements github.com/stratumn/go/store.Adapter.
type Store struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	session      *rethink.Session
	db           rethink.Term
	segments     rethink.Term
}

type wrapper struct {
	ID           []byte      `json:"id"`
	Content      *cs.Segment `json:"content"`
	Priority     float64     `json:"priority"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	MapID        string      `json:"mapId"`
	PrevLinkHash []byte      `json:"prevLinkHash"`
	Tags         []string    `json:"tags"`
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	opts := rethink.ConnectOpts{Addresses: strings.Split(config.URL, ",")}
	session, err := rethink.Connect(opts)
	if err != nil {
		return nil, err
	}
	db := rethink.DB(config.DB)
	return &Store{
		config:   config,
		session:  session,
		db:       db,
		segments: db.Table("segments"),
	}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/go/fossilizer.Store.AddDidSaveChannel.
func (a *Store) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (a *Store) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (a *Store) SaveSegment(segment *cs.Segment) error {
	var (
		linkHash     = segment.GetLinkHash()
		prevLinkHash = segment.Link.GetPrevLinkHash()
	)

	w := wrapper{
		ID:        segment.GetLinkHash()[:],
		Content:   segment,
		Priority:  segment.Link.GetPriority(),
		UpdatedAt: time.Now().UTC(),
		MapID:     segment.Link.GetMapID(),
		Tags:      segment.Link.GetTags(),
	}

	if prevLinkHash != nil {
		w.PrevLinkHash = prevLinkHash[:]
	}

	if err := a.segments.Get(linkHash).Replace(&w).Exec(a.session); err != nil {
		return err
	}

	// Send saved segment to all the save channels without blocking.
	go func(chans []chan *cs.Segment) {
		for _, c := range chans {
			c <- segment
		}
	}(a.didSaveChans)

	return nil
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (a *Store) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	cur, err := a.segments.Get(linkHash[:]).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var w wrapper
	if err := cur.One(&w); err != nil {
		if err == rethink.ErrEmptyResult {
			return nil, nil
		}
		return nil, err
	}

	return w.Content, nil
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (a *Store) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	res, err := a.segments.
		Get(linkHash[:]).
		Delete(rethink.DeleteOpts{ReturnChanges: true}).
		RunWrite(a.session)
	if err != nil {
		return nil, err
	}
	if res.Deleted < 1 {
		return nil, nil
	}

	b, err := json.Marshal(res.Changes[0].OldValue)
	if err != nil {
		return nil, err
	}

	var w wrapper
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, err
	}

	return w.Content, nil
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
func (a *Store) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	q := a.segments

	if prevLinkHash := filter.PrevLinkHash; prevLinkHash != nil {
		q = q.Between([]interface{}{
			prevLinkHash[:],
			rethink.MinVal,
		}, []interface{}{
			prevLinkHash[:],
			rethink.MaxVal,
		}, rethink.BetweenOpts{
			Index:      "prevLinkHashOrder",
			LeftBound:  "closed",
			RightBound: "closed",
		}).OrderBy(rethink.OrderByOpts{Index: "prevLinkHashOrder"})
	} else if mapID := filter.MapID; mapID != "" {
		q = q.Between([]interface{}{
			mapID,
			rethink.MinVal,
		}, []interface{}{
			mapID,
			rethink.MaxVal,
		}, rethink.BetweenOpts{
			Index:      "mapIdOrder",
			RightBound: "closed",
		}).OrderBy(rethink.OrderByOpts{Index: "mapIdOrder"})
	} else {
		q = q.OrderBy(rethink.OrderByOpts{Index: rethink.Desc("order")})
	}

	if tags := filter.Tags; len(tags) > 0 {
		t := make([]interface{}, len(tags))
		for i, v := range tags {
			t[i] = v
		}
		q = q.Filter(rethink.Row.Field("tags").Contains(t...))
	}

	q = q.Field("content")

	if skip := filter.Offset; skip > 0 {
		q = q.Skip(filter.Offset)
	}

	cur, err := q.Limit(filter.Limit).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	segments := make(cs.SegmentSlice, 0, filter.Limit)
	if err := cur.All(&segments); err != nil {
		return nil, err
	}

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
func (a *Store) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	cur, err := a.segments.
		Between(rethink.MinVal, rethink.MaxVal, rethink.BetweenOpts{
			Index: "mapId",
		}).
		OrderBy(rethink.OrderByOpts{Index: "mapId"}).
		Distinct(rethink.DistinctOpts{Index: "mapId"}).
		Skip(pagination.Offset).
		Limit(pagination.Limit).
		Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	mapIDs := []string{}
	if err = cur.All(&mapIDs); err != nil {
		return nil, err
	}

	return mapIDs, nil
}

// Create creates the database tables and indexes.
func (a *Store) Create() (err error) {
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}

	tblOpts := rethink.TableCreateOpts{}
	if !a.config.Hard {
		tblOpts.Durability = "soft"
	}

	exec(a.db.TableCreate("segments", tblOpts))
	exec(a.segments.Wait())
	exec(a.segments.IndexCreate("mapId"))
	exec(a.segments.IndexWait("mapId"))
	exec(a.segments.IndexCreateFunc("order", []interface{}{
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.segments.IndexWait("order"))
	exec(a.segments.IndexCreateFunc("mapIdOrder", []interface{}{
		rethink.Row.Field("mapId"),
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.segments.IndexWait("mapIdOrder"))
	exec(a.segments.IndexCreateFunc("prevLinkHashOrder", []interface{}{
		rethink.Row.Field("prevLinkHash"),
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.segments.IndexWait("prevLinkHashOrder"))

	return err
}

// Drop drops the database tables and indexes.
func (a *Store) Drop() error {
	return a.db.TableDrop("segments").Exec(a.session)
}

// Exists returns whether the database tables exists.
func (a *Store) Exists() (bool, error) {
	cur, err := a.db.TableList().Run(a.session)
	if err != nil {
		return false, err
	}
	defer cur.Close()

	var name string
	for cur.Next(&name) {
		if name == "segments" {
			return true, nil
		}
	}
	return false, nil
}
