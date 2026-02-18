// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ixti/apiska/internal/exporters"
)

// SavedQuery represents a query saved for later use -- a track in your crate.
type SavedQuery struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SQL       string    `json:"sql"`
	CreatedAt time.Time `json:"created_at"`
}

// queriesFile is the JSON structure for persisting queries to disk.
type queriesFile struct {
	Queries []SavedQuery `json:"queries"`
}

// Store manages saved queries persistence -- the record store that never closes.
type Store struct {
	mu      sync.RWMutex
	path    string
	queries []SavedQuery
}

// NewStore creates a new Store loading from ~/.apiska/queries.json.
// If the file doesn't exist, starts with an empty collection.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	path := filepath.Join(home, ".apiska", "queries.json")
	s := &Store{
		path:    path,
		queries: []SavedQuery{},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

// Save stores a query with the given name -- pressing a new record.
func (s *Store) Save(name, sql string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := SavedQuery{
		ID:        fmt.Sprintf("q-%d", time.Now().UnixNano()),
		Name:      name,
		SQL:       sql,
		CreatedAt: time.Now(),
	}

	s.queries = append([]SavedQuery{query}, s.queries...)
	return s.persist()
}

// List returns all saved queries -- browsing the crate.
func (s *Store) List() []SavedQuery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SavedQuery, len(s.queries))
	copy(result, s.queries)
	return result
}

// Delete removes a saved query by ID -- pulling a record from the crate.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, q := range s.queries {
		if q.ID == id {
			s.queries = append(s.queries[:i], s.queries[i+1:]...)
			return s.persist()
		}
	}
	return nil
}

// load reads queries from disk -- opening the crate.
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var file queriesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse queries file: %w", err)
	}

	s.queries = file.Queries
	return nil
}

// persist writes queries to disk -- closing the crate for the night.
func (s *Store) persist() error {
	file := queriesFile{Queries: s.queries}
	_, err := exporters.WriteJson(s.path, func(enc *json.Encoder) error {
		enc.SetIndent("", "  ")
		return enc.Encode(file)
	})
	return err
}
