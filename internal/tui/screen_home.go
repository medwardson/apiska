// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/rds"
	"github.com/ixti/apiska/internal/tui/styles"
)

type homeScreen struct {
	queries []*rds.Query
	table   table.Model

	width, height int
}

func newHomeScreen() *homeScreen {
	columns := []table.Column{
		{Title: "Label", Width: 20},
		{Title: "Time", Width: 20},
		{Title: "Result", Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	return &homeScreen{
		queries: []*rds.Query{},
		table:   t,
	}
}

func (s *homeScreen) Title() string {
	return "QUERIES"
}

func (s *homeScreen) KeyHints() string {
	hints := formatHint("ctrl+n", "new query") + styles.HintSep.String() + formatHint("ctrl+l", "saved")

	if len(s.queries) > 0 {
		hints = formatHint("↑↓", "navigate") +
			styles.HintSep.String() + formatHint("enter", "view") +
			styles.HintSep.String() + hints
	}

	return hints
}

func (s *homeScreen) Init() tea.Cmd {
	return nil
}

func (s *homeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 1 // account for footer
		s.updateTableSize()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if q := s.selectedQuery(); q != nil {
				return s, func() tea.Msg {
					return openQueryMsg{query: q, execute: false}
				}
			}
			return s, nil

		case tea.KeyCtrlN:
			return s, func() tea.Msg { return openEditorMsg{} }

		case tea.KeyCtrlL:
			return s, func() tea.Msg { return openSavedQueriesMsg{} }
		}

	case queryAddedMsg:
		s.AddQuery(msg.query)
		return s, nil

	case queryExecutedMsg:
		// Query result updated, refresh table to show it
		s.refreshTable()
		return s, nil
	}

	// Only update table if we have queries
	if len(s.queries) > 0 {
		var cmd tea.Cmd
		s.table, cmd = s.table.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s *homeScreen) View() string {
	if s.width == 0 || s.height == 0 {
		return "Loading..."
	}

	if len(s.queries) == 0 {
		title := styles.WelcomeTitle.Render("APISKA")
		subtitle := styles.WelcomeSubtitle.Render("A skankin' TUI for the AWS RDS Data API.")
		content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)
		return styles.TitledBox(s.Title(), content, s.width, s.height)
	}

	return styles.TitledBoxTopLeft(s.Title(), s.table.View(), s.width, s.height)
}

func (s *homeScreen) updateTableSize() {
	// Width: screen width - 2 for box borders
	tableWidth := s.width - 2
	if tableWidth < 10 {
		tableWidth = 10
	}

	// Height: screen height - 2 for box borders - 2 for table header
	tableHeight := s.height - 2 - 2
	if tableHeight < 1 {
		tableHeight = 1
	}

	s.table.SetWidth(tableWidth)
	s.table.SetHeight(tableHeight)
	s.table.SetStyles(styles.TableStyles(tableWidth))

	// Distribute column widths: fixed widths for label/time, rest for result
	const separatorWidth = 3
	labelWidth := 20
	timeWidth := 20
	resultWidth := tableWidth - labelWidth - timeWidth - separatorWidth*2

	if resultWidth < 20 {
		resultWidth = 20
	}

	s.table.SetColumns([]table.Column{
		{Title: "Label", Width: labelWidth},
		{Title: "Time", Width: timeWidth},
		{Title: "Result", Width: resultWidth},
	})
}

func (s *homeScreen) selectedQuery() *rds.Query {
	idx := s.table.Cursor()
	if idx >= 0 && idx < len(s.queries) {
		return s.queries[idx]
	}
	return nil
}

func (s *homeScreen) refreshTable() {
	rows := make([]table.Row, len(s.queries))
	for i, q := range s.queries {
		result := ""
		if q.Result != nil {
			result = q.Result.Message()
		}
		rows[i] = table.Row{
			q.Label,
			q.CreatedAt.Format("15:04:05"),
			result,
		}
	}
	s.table.SetRows(rows)
}

// AddQuery adds a query to the history and refreshes the table.
func (s *homeScreen) AddQuery(q *rds.Query) {
	// Prepend so newest is at top
	s.queries = append([]*rds.Query{q}, s.queries...)
	s.refreshTable()
}
