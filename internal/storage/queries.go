// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var unsafeCharsRe = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

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

// Save stores a query with the given name -- pressing a new record.
// Returns the actual filename that was used.
func (s *Store) Save(name, sql string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.path, 0755); err != nil {
		return "", fmt.Errorf("failed to create .apiska directory: %w", err)
	}

	filename := sanitizeFilename(name) + ".sql"
	filePath := uniqueFilePath(filepath.Join(s.path, filename))

	if err := os.WriteFile(filePath, []byte(sql), 0644); err != nil {
		return "", err
	}

	return filepath.Base(filePath), nil
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

// Delete removes a saved query by ID -- pulling a record from the crate.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.path, id)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// sanitizeFilename converts a query name to a safe filename.
func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = unsafeCharsRe.ReplaceAllString(name, "")
	name = strings.Trim(name, "._ ")

	if name == "" {
		name = "query"
	}

	return name
}

// uniqueFilePath returns a unique file path by appending a numeric suffix if needed.
func uniqueFilePath(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
