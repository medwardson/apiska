// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/storage"
	"github.com/ixti/apiska/internal/tui/styles"
)

type savedScreen struct {
	store   *storage.Store
	queries []storage.SavedQuery
	table   table.Model

	width, height int
}

func newSavedScreen(store *storage.Store) *savedScreen {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Updated", Width: 20},
		{Title: "SQL", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	s := &savedScreen{
		store: store,
		table: t,
	}
	s.refreshTable()

	return s
}

func (s *savedScreen) Title() string {
	return "SAVED QUERIES"
}

func (s *savedScreen) KeyHints() string {
	if len(s.queries) == 0 {
		return ""
	}
	return formatHint("↑↓", "navigate") +
		styles.HintSep.String() + formatHint("enter", "load") +
		styles.HintSep.String() + formatHint("d", "delete")
}

func (s *savedScreen) Init() tea.Cmd {
	return nil
}

func (s *savedScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 1 // account for footer
		s.updateTableSize()

	case querySavedMsg:
		s.refreshTable()
		return s, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if q := s.selectedQuery(); q != nil {
				return s, func() tea.Msg {
					return openEditorMsg{sql: q.SQL}
				}
			}
			return s, nil
		}

		switch msg.String() {
		case "d":
			if q := s.selectedQuery(); q != nil {
				s.store.Delete(q.ID)
				s.refreshTable()
			}
			return s, nil
		}
	}

	// Only update table if we have queries
	if len(s.queries) > 0 {
		var cmd tea.Cmd
		s.table, cmd = s.table.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s *savedScreen) View() string {
	if s.width == 0 || s.height == 0 {
		return "Loading..."
	}

	if len(s.queries) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(styles.TextMuted).Render("No saved queries yet.")
		return styles.TitledBox(s.Title(), emptyMsg, s.width, s.height)
	}

	return styles.TitledBoxTopLeft(s.Title(), s.table.View(), s.width, s.height)
}

func (s *savedScreen) updateTableSize() {
	// Width: screen width - 2 for box borders
	tableWidth := max(s.width-2, 10)

	// Height: screen height - 2 for box borders - 2 for table header
	tableHeight := max(s.height-2-2, 1)

	s.table.SetWidth(tableWidth)
	s.table.SetHeight(tableHeight)
	s.table.SetStyles(styles.TableStyles(tableWidth))

	// Distribute column widths
	const separatorWidth = 3
	nameWidth := 25
	createdWidth := 20
	sqlWidth := max(tableWidth-nameWidth-createdWidth-separatorWidth*2, 20)

	s.table.SetColumns([]table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Updated", Width: createdWidth},
		{Title: "SQL", Width: sqlWidth},
	})
}

func (s *savedScreen) selectedQuery() *storage.SavedQuery {
	idx := s.table.Cursor()
	if idx >= 0 && idx < len(s.queries) {
		return &s.queries[idx]
	}
	return nil
}

func (s *savedScreen) refreshTable() {
	s.queries = s.store.List()

	rows := make([]table.Row, len(s.queries))
	for i, q := range s.queries {
		rows[i] = table.Row{
			q.Name,
			q.UpdatedAt.Format("2006-01-02 15:04"),
			truncateSQL(q.SQL, 50),
		}
	}
	s.table.SetRows(rows)
}

func truncateSQL(sql string, maxLen int) string {
	// Replace newlines with spaces
	result := ""
	for _, c := range sql {
		if c == '\n' || c == '\r' || c == '\t' {
			result += " "
		} else {
			result += string(c)
		}
	}
	if len(result) > maxLen {
		result = result[:maxLen-3] + "..."
	}
	return result
}
