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

func TestStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	// Create a .sql file manually
	err := os.WriteFile(filepath.Join(tmpDir, "test_query.sql"), []byte("SELECT * FROM users"), 0644)
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "test_query", queries[0].Name)
	assert.Equal(t, "SELECT * FROM users", queries[0].SQL)
	assert.Equal(t, "test_query.sql", queries[0].ID)
	assert.False(t, queries[0].UpdatedAt.IsZero())
}

func TestStore_ListMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	err := os.WriteFile(filepath.Join(tmpDir, "query_1.sql"), []byte("SELECT 1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "query_2.sql"), []byte("SELECT 2"), 0644)
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 2)
}

func TestStore_ListIgnoresNonSqlFiles(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	err := os.WriteFile(filepath.Join(tmpDir, "query.sql"), []byte("SELECT 1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a query"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "notes.md"), []byte("some notes"), 0644)
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "query", queries[0].Name)
}

func TestStore_ListIgnoresDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{path: tmpDir}

	err := os.WriteFile(filepath.Join(tmpDir, "query.sql"), []byte("SELECT 1"), 0644)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "subdir.sql"), 0755)
	require.NoError(t, err)

	queries := store.List()
	require.Len(t, queries, 1)
	assert.Equal(t, "query", queries[0].Name)
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
