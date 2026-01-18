// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package exporters

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	t.Run("creates new file and writes content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		absPath, err := WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("hello"))
			return err
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(content))
	})

	t.Run("creates new file with 0644 permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows")
		}

		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		_, err := WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("hello"))
			return err
		})

		require.NoError(t, err)

		info, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	})

	t.Run("truncates existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		err := os.WriteFile(filename, []byte("old content that is longer"), 0o644)
		require.NoError(t, err)

		_, err = WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("new"))
			return err
		})

		require.NoError(t, err)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "new", string(content))
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "dir")
		filename := filepath.Join(nestedDir, "test.txt")

		absPath, err := WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("content"))
			return err
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
	})

	t.Run("creates parent directories with 0755 permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows")
		}

		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "dir")
		filename := filepath.Join(nestedDir, "test.txt")

		_, err := WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("content"))
			return err
		})

		require.NoError(t, err)

		info, err := os.Stat(nestedDir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	})

	t.Run("propagates callback error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")
		expectedErr := errors.New("callback error")

		_, err := WriteFile(filename, func(w io.Writer) error {
			return expectedErr
		})

		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns absolute path", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		absPath, err := WriteFile(filename, func(w io.Writer) error {
			return nil
		})

		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(absPath))
	})

	t.Run("handles relative path with parent references", func(t *testing.T) {
		tmpDir := t.TempDir()
		nested := filepath.Join(tmpDir, "a", "b")
		err := os.MkdirAll(nested, 0o755)
		require.NoError(t, err)

		// Write from nested using ../
		filename := filepath.Join(nested, "..", "c", "test.txt")

		absPath, err := WriteFile(filename, func(w io.Writer) error {
			_, err := w.Write([]byte("content"))
			return err
		})

		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(absPath))

		// Verify the file ended up in the right place
		expectedDir := filepath.Join(tmpDir, "a", "c")
		content, err := os.ReadFile(filepath.Join(expectedDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
	})

	t.Run("fails gracefully with empty filename", func(t *testing.T) {
		_, err := WriteFile("", func(w io.Writer) error {
			return nil
		})

		// Empty filename resolves to cwd, which is a directory - should fail
		require.Error(t, err)
	})
}

func TestWriteCsv(t *testing.T) {
	t.Run("writes csv content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		absPath, err := WriteCsv(filename, func(w *csv.Writer) error {
			if err := w.Write([]string{"name", "age"}); err != nil {
				return err
			}
			return w.Write([]string{"alice", "30"})
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "name,age\nalice,30\n", string(content))
	})

	t.Run("writes csv content with WriteAll", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		absPath, err := WriteCsv(filename, func(w *csv.Writer) error {
			return w.WriteAll([][]string{
				{"name", "age"},
				{"alice", "30"},
				{"bob", "25"},
			})
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "name,age\nalice,30\nbob,25\n", string(content))
	})

	t.Run("creates file with 0644 permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows")
		}

		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		_, err := WriteCsv(filename, func(w *csv.Writer) error {
			return nil
		})

		require.NoError(t, err)

		info, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "nested", "dir", "test.csv")

		_, err := WriteCsv(filename, func(w *csv.Writer) error {
			return nil
		})

		require.NoError(t, err)

		_, err = os.Stat(filename)
		require.NoError(t, err)
	})

	t.Run("propagates callback error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")
		expectedErr := errors.New("callback error")

		_, err := WriteCsv(filename, func(w *csv.Writer) error {
			return expectedErr
		})

		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestWriteJson(t *testing.T) {
	t.Run("writes json content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.json")

		data := map[string]any{"name": "alice", "age": 30}

		absPath, err := WriteJson(filename, func(enc *json.Encoder) error {
			return enc.Encode(data)
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"alice","age":30}`, string(content))
	})

	t.Run("writes multiple json objects", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.json")

		absPath, err := WriteJson(filename, func(enc *json.Encoder) error {
			if err := enc.Encode(map[string]string{"name": "alice"}); err != nil {
				return err
			}
			return enc.Encode(map[string]string{"name": "bob"})
		})

		require.NoError(t, err)
		assert.Equal(t, filename, absPath)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "{\"name\":\"alice\"}\n{\"name\":\"bob\"}\n", string(content))
	})

	t.Run("streams json objects from a loop", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.jsonl")

		type Record struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		records := []Record{
			{ID: 1, Name: "alice"},
			{ID: 2, Name: "bob"},
			{ID: 3, Name: "charlie"},
		}

		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			for _, r := range records {
				if err := enc.Encode(r); err != nil {
					return err
				}
			}
			return nil
		})

		require.NoError(t, err)

		// Verify we can read back as NDJSON (newline-delimited JSON)
		f, err := os.Open(filename)
		require.NoError(t, err)
		defer f.Close()

		dec := json.NewDecoder(f)
		var decoded []Record
		for dec.More() {
			var r Record
			err := dec.Decode(&r)
			require.NoError(t, err)
			decoded = append(decoded, r)
		}

		assert.Equal(t, records, decoded)
	})

	t.Run("streams json objects from a channel", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.jsonl")

		type Event struct {
			Type    string `json:"type"`
			Payload int    `json:"payload"`
		}

		// Simulate streaming from a channel
		events := make(chan Event, 3)
		events <- Event{Type: "start", Payload: 0}
		events <- Event{Type: "data", Payload: 42}
		events <- Event{Type: "end", Payload: 100}
		close(events)

		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			for event := range events {
				if err := enc.Encode(event); err != nil {
					return err
				}
			}
			return nil
		})

		require.NoError(t, err)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)

		expected := "{\"type\":\"start\",\"payload\":0}\n" +
			"{\"type\":\"data\",\"payload\":42}\n" +
			"{\"type\":\"end\",\"payload\":100}\n"
		assert.Equal(t, expected, string(content))
	})

	t.Run("creates file with 0644 permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows")
		}

		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.json")

		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			return enc.Encode(map[string]string{})
		})

		require.NoError(t, err)

		info, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "nested", "dir", "test.json")

		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			return enc.Encode(map[string]string{})
		})

		require.NoError(t, err)

		_, err = os.Stat(filename)
		require.NoError(t, err)
	})

	t.Run("propagates callback error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.json")
		expectedErr := errors.New("callback error")

		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			return expectedErr
		})

		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("propagates encoding error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.json")

		// Channels cannot be encoded to JSON
		_, err := WriteJson(filename, func(enc *json.Encoder) error {
			return enc.Encode(make(chan int))
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})
}
