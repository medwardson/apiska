// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SaveAndList(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	store := &Store{path: path, queries: []SavedQuery{}}

	// Save a query
	err := store.Save("Test Query", "SELECT * FROM users")
	require.NoError(t, err)

	// List should return the saved query
	queries := store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "Test Query", queries[0].Name)
	assert.Equal(t, "SELECT * FROM users", queries[0].SQL)
	assert.NotEmpty(t, queries[0].ID)
	assert.False(t, queries[0].CreatedAt.IsZero())
}

func TestStore_SaveMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	store := &Store{path: path, queries: []SavedQuery{}}

	err := store.Save("Query 1", "SELECT 1")
	require.NoError(t, err)

	err = store.Save("Query 2", "SELECT 2")
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 2)

	// Newest should be first
	assert.Equal(t, "Query 2", queries[0].Name)
	assert.Equal(t, "Query 1", queries[1].Name)
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	store := &Store{path: path, queries: []SavedQuery{}}

	err := store.Save("Query 1", "SELECT 1")
	require.NoError(t, err)

	err = store.Save("Query 2", "SELECT 2")
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 2)

	// Delete first query
	err = store.Delete(queries[0].ID)
	require.NoError(t, err)

	queries = store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "Query 1", queries[0].Name)
}

func TestStore_DeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	store := &Store{path: path, queries: []SavedQuery{}}

	// Deleting non-existent ID should not error
	err := store.Delete("nonexistent-id")
	require.NoError(t, err)
}

func TestStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	// Create first store and save
	store1 := &Store{path: path, queries: []SavedQuery{}}
	err := store1.Save("Persisted Query", "SELECT * FROM orders")
	require.NoError(t, err)

	// Create second store and load from same file
	store2 := &Store{path: path, queries: []SavedQuery{}}
	err = store2.load()
	require.NoError(t, err)

	queries := store2.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "Persisted Query", queries[0].Name)
	assert.Equal(t, "SELECT * FROM orders", queries[0].SQL)
}

func TestStore_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")

	store := &Store{path: path, queries: []SavedQuery{}}
	err := store.load()

	assert.True(t, os.IsNotExist(err))
}

func TestStore_ListReturnsDefensiveCopy(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "queries.json")

	store := &Store{path: path, queries: []SavedQuery{}}
	err := store.Save("Original", "SELECT 1")
	require.NoError(t, err)

	// Modify the returned slice
	queries := store.List()
	queries[0].Name = "Modified"

	// Original should be unchanged
	queries2 := store.List()
	assert.Equal(t, "Original", queries2[0].Name)
}
