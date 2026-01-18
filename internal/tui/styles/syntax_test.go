// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package styles

import (
	"testing"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestHighlightSQL(t *testing.T) {
	syntaxFormatter = formatters.Get("terminal16m")

	t.Cleanup(func() {
		syntaxFormatter = formatters.Get("noop")
	})

	t.Run("returns valid output preserving SQL content", func(t *testing.T) {
		sql := "SELECT * FROM users"
		highlighted := HighlightSQL(sql)

		assert.NotEqual(t, sql, highlighted)
		assert.Equal(t, sql, ansi.Strip(highlighted))
	})

	t.Run("handles empty string", func(t *testing.T) {
		assert.Empty(t, HighlightSQL(""))
	})

	t.Run("handles multiline SQL", func(t *testing.T) {
		sql := "SELECT\n  id,\n  name\nFROM users\nWHERE active = true"
		highlighted := HighlightSQL(sql)

		assert.NotEqual(t, sql, highlighted)
		assert.Equal(t, sql, ansi.Strip(highlighted))
	})
}
