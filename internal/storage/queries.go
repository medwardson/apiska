// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package storage

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// SavedQuery represents a query saved for later use -- a track in your crate.
type SavedQuery struct {
	ID        string
	Name      string
	SQL       string
	UpdatedAt time.Time
}

// Store manages saved queries persistence -- the record store that never closes.
// Queries are stored as .sql files in ./.apiska/
type Store struct {
	mu   sync.RWMutex
	path string
}

// NewStore creates a new Store using ./.apiska/ for query files.
func NewStore() (*Store, error) {
	return &Store{path: ".apiska"}, nil
}

// List returns all saved queries -- browsing the crate.
func (s *Store) List() []SavedQuery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.path)
	if err != nil {
		return []SavedQuery{}
	}

	var queries []SavedQuery
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(s.path, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".sql")
		queries = append(queries, SavedQuery{
			ID:        entry.Name(),
			Name:      name,
			SQL:       string(content),
			UpdatedAt: info.ModTime(),
		})
	}

	sort.Slice(queries, func(i, j int) bool {
		return queries[i].UpdatedAt.After(queries[j].UpdatedAt)
	})

	return queries
}

