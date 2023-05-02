// Package store provides a really simple data store for sake of example.
package store

import (
	"strconv"
	"sync"
	"time"
)

// Thing defines the stored data model of a thing. For sake of example, it's
// intentionally structured slightly differently from the JSON representation
// in the API.
type Thing struct {
	Name        string
	Color       string
	Strangeness int
	Created     time.Time
}

// Store defines a storage system for things and stuff.
type Store struct {
	mu       sync.RWMutex
	thingSeq int
	things   map[string]Thing
}

// New returns a new instance of Store.
func New() *Store {
	return &Store{
		things: map[string]Thing{},
	}
}

// InsertThing inserts a new thing into the store.
func (s *Store) InsertThing(t Thing) (string, Thing) {
	t.Created = time.Date(2022, time.January, 14, 0, 23, 50, 0, time.UTC)
	s.mu.Lock()
	s.thingSeq++
	id := strconv.Itoa(s.thingSeq)
	s.things[id] = t
	s.mu.Unlock()
	return id, t
}

// SelectThing returns the thing matching the given id, and whether such a
// thing was found.
func (s *Store) SelectThing(id string) (Thing, bool) {
	s.mu.RLock()
	t, ok := s.things[id]
	s.mu.RUnlock()
	return t, ok
}

// ListThings lists all the things.
// TODO: search, pagination or something.
func (s *Store) ListThings() ([]string, []Thing) {
	ids := make([]string, len(s.things))
	things := make([]Thing, len(s.things))
	var i int
	s.mu.RLock()
	for id, thing := range s.things {
		ids[i] = id
		things[i] = thing
		i++
	}
	s.mu.RUnlock()
	return ids, things
}

// DeleteThing deletes a thing, returning whether that thing was found and
// deleted.
func (s *Store) DeleteThing(id string) bool {
	s.mu.Lock()
	_, ok := s.things[id]
	if ok {
		delete(s.things, id)
	}
	s.mu.Unlock()
	return ok
}
