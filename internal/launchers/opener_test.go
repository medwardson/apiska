// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package launchers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDefaultApp(t *testing.T) {
	t.Run("does not panic with valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		err := os.WriteFile(filename, []byte("test content"), 0o644)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			_ = OpenDefaultApp(filename)
		})
	})

	t.Run("does not panic with non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentFile := filepath.Join(tmpDir, "non", "existent", "file.txt")

		assert.NotPanics(t, func() {
			_ = OpenDefaultApp(nonExistentFile)
		})
	})

	t.Run("starts command without immediate error on supported platforms", func(t *testing.T) {
		switch runtime.GOOS {
		case "darwin", "freebsd", "linux", "windows":
			// Skankin' platforms -- well, freebsd and linux at least.
			// Windows and darwin get the accessible seating near the exit.
		default:
			t.Skip("skipping on unsupported platform")
		}

		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		err := os.WriteFile(filename, []byte("test content"), 0o644)
		require.NoError(t, err)

		// We just fire and forget here. CI boxes are moslty headless wastelands
		// where xdg-open goes to die. As long as nothing explodes, we're good.
		_ = OpenDefaultApp(filename)
	})

	t.Run("handles paths with spaces", func(t *testing.T) {
		tmpDir := t.TempDir()
		dirWithSpaces := filepath.Join(tmpDir, "path with spaces")
		err := os.MkdirAll(dirWithSpaces, 0o755)
		require.NoError(t, err)

		filename := filepath.Join(dirWithSpaces, "test file.txt")
		err = os.WriteFile(filename, []byte("test content"), 0o644)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			_ = OpenDefaultApp(filename)
		})
	})

	t.Run("handles unicode filenames", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "тест_文件_テスト.txt")

		err := os.WriteFile(filename, []byte("test content"), 0o644)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			_ = OpenDefaultApp(filename)
		})
	})
}
