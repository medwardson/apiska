// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package launchers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectEditor(t *testing.T) {
	t.Run("returns APISKA_EDITOR when set", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "custom-editor")
		t.Setenv("EDITOR", "fallback-editor")

		editor, err := detectEditor()
		require.NoError(t, err)
		assert.Equal(t, "custom-editor", editor)
	})

	t.Run("falls back to EDITOR when APISKA_EDITOR not set", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "")
		t.Setenv("EDITOR", "fallback-editor")

		editor, err := detectEditor()
		require.NoError(t, err)
		assert.Equal(t, "fallback-editor", editor)
	})

	t.Run("returns error when neither is set", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "")
		t.Setenv("EDITOR", "")

		editor, err := detectEditor()
		assert.Empty(t, editor)
		assert.ErrorIs(t, err, ErrEditorNotConfigured)
	})
}

func TestOpenExternalEditor(t *testing.T) {
	t.Run("returns error when no editor configured", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "")
		t.Setenv("EDITOR", "")

		cmd, editor, err := OpenExternalEditor(".sql", "test content")
		assert.Nil(t, cmd)
		assert.Nil(t, editor)
		assert.ErrorIs(t, err, ErrEditorNotConfigured)
	})

	t.Run("uses APISKA_EDITOR first", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "custom-editor")
		t.Setenv("EDITOR", "fallback-editor")

		cmd, editor, err := OpenExternalEditor(".sql", "test content")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		assert.Equal(t, "custom-editor", cmd.Args[0])
	})

	t.Run("falls back to EDITOR", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "")
		t.Setenv("EDITOR", "fallback-editor")

		cmd, editor, err := OpenExternalEditor(".sql", "test content")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		assert.Equal(t, "fallback-editor", cmd.Args[0])
	})

	t.Run("creates file with initial content", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		initialContent := "SELECT * FROM users;"
		cmd, editor, err := OpenExternalEditor(".sql", initialContent)
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		content, err := os.ReadFile(editor.tempPath)
		require.NoError(t, err)
		assert.Equal(t, initialContent, string(content))
	})

	t.Run("creates file with empty initial content", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		cmd, editor, err := OpenExternalEditor(".sql", "")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		content, err := os.ReadFile(editor.tempPath)
		require.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("handles multiline content with special chars", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		initialContent := "SELECT *\nFROM users\nWHERE name = 'O''Brien';\n-- horn section goes here"
		cmd, editor, err := OpenExternalEditor(".sql", initialContent)
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		content, err := os.ReadFile(editor.tempPath)
		require.NoError(t, err)
		assert.Equal(t, initialContent, string(content))
	})

	t.Run("creates file with correct extension", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		cmd, editor, err := OpenExternalEditor(".sql", "content")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)
		defer editor.Cleanup()

		tempPath := cmd.Args[1]
		assert.Contains(t, tempPath, ".sql")
	})
}

func TestExternalEditor_Cleanup(t *testing.T) {
	t.Run("removes temp file", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		cmd, editor, err := OpenExternalEditor(".sql", "content")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		require.NotNil(t, editor)

		tempPath := cmd.Args[1]

		_, err = os.Stat(tempPath)
		require.NoError(t, err)

		editor.Cleanup()

		_, err = os.Stat(tempPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		_, editor, err := OpenExternalEditor(".sql", "content")
		require.NoError(t, err)

		// Call cleanup multiple times - should not panic
		assert.NotPanics(t, func() {
			editor.Cleanup()
			editor.Cleanup()
			editor.Cleanup()
		})
	})
}

func TestExternalEditor_ReadContent(t *testing.T) {
	t.Run("returns externally modified content", func(t *testing.T) {
		t.Setenv("APISKA_EDITOR", "cat")

		_, editor, err := OpenExternalEditor(".sql", "initial")
		require.NoError(t, err)
		defer editor.Cleanup()

		err = os.WriteFile(editor.tempPath, []byte("modified content"), 0o644)
		require.NoError(t, err)

		content, err := editor.ReadContent()
		require.NoError(t, err)
		assert.Equal(t, "modified content", content)
		assert.NotEqual(t, "initial", content)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		editor := &ExternalEditor{tempPath: "/nonexistent/path/file.sql"}

		_, err := editor.ReadContent()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read editor content")
	})
}
