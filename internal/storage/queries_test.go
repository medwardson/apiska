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
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	actualName, err := store.Save("Test Query", "SELECT * FROM users")
	require.NoError(t, err)
	assert.Equal(t, "test_query.sql", actualName)

	queries := store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "test_query", queries[0].Name)
	assert.Equal(t, "SELECT * FROM users", queries[0].SQL)
	assert.Equal(t, "test_query.sql", queries[0].ID)
	assert.False(t, queries[0].UpdatedAt.IsZero())

	// Verify file exists
	_, err = os.Stat(filepath.Join(tmpDir, "test_query.sql"))
	require.NoError(t, err)
}

func TestStore_SaveMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	_, err := store.Save("Query 1", "SELECT 1")
	require.NoError(t, err)

	_, err = store.Save("Query 2", "SELECT 2")
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 2)

	// Newest should be first (sorted by mod time)
	assert.Equal(t, "query_2", queries[0].Name)
	assert.Equal(t, "query_1", queries[1].Name)
}

func TestStore_SaveDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	name1, err := store.Save("Query", "SELECT 1")
	require.NoError(t, err)
	assert.Equal(t, "query.sql", name1)

	name2, err := store.Save("Query", "SELECT 2")
	require.NoError(t, err)
	assert.Equal(t, "query_1.sql", name2)

	queries := store.List()
	require.Len(t, queries, 2)

	// Should have both query.sql and query_1.sql
	_, err = os.Stat(filepath.Join(tmpDir, "query.sql"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "query_1.sql"))
	require.NoError(t, err)
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	_, err := store.Save("Query 1", "SELECT 1")
	require.NoError(t, err)

	_, err = store.Save("Query 2", "SELECT 2")
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 2)

	// Delete newest query
	err = store.Delete(queries[0].ID)
	require.NoError(t, err)

	queries = store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "query_1", queries[0].Name)
}

func TestStore_DeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	// Deleting non-existent ID should not error
	err := store.Delete("nonexistent.sql")
	require.NoError(t, err)
}

func TestStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first store and save
	store1 := &Store{path: tmpDir}
	_, err := store1.Save("Persisted Query", "SELECT * FROM orders")
	require.NoError(t, err)

	// Create second store pointing to same directory
	store2 := &Store{path: tmpDir}

	queries := store2.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "persisted_query", queries[0].Name)
	assert.Equal(t, "SELECT * FROM orders", queries[0].SQL)
}

func TestStore_ListEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	queries := store.List()
	assert.Empty(t, queries)
}

func TestStore_ListNonExistentDirectory(t *testing.T) {
	store := &Store{path: "/nonexistent/path"}

	queries := store.List()
	assert.Empty(t, queries)
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Query", "simple_query"},
		{"Query/With/Slashes", "querywithslashes"},
		{"Query:With:Colons", "querywithcolons"},
		{"Query<>With<>Brackets", "querywithbrackets"},
		{"...", "query"},
		{"", "query"},
		{"   ", "query"},
		{"Normal_Name", "normal_name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
