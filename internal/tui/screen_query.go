// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/exporters"
	"github.com/ixti/apiska/internal/formatters"
	"github.com/ixti/apiska/internal/launchers"
	"github.com/ixti/apiska/internal/rds"
	"github.com/ixti/apiska/internal/tui/styles"
)

// Height constants for query screen layout
const (
	sqlBoxHeight = 8 // Fixed height for SQL display box
)

// csvExportedMsg is sent when CSV export completes
type csvExportedMsg struct {
	path string
	err  error
}

type queryScreen struct {
	client    *rds.Client
	query     *rds.Query
	table     table.Model
	executing bool
	err       error
	status    string // Status message (e.g., "Exported to ...")

	// Horizontal scroll: which column index to start displaying from
	colOffset int

	width, height int
}

func newQueryScreen(query *rds.Query) *queryScreen {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
	)

	return &queryScreen{
		query: query,
		table: t,
	}
}

func (s *queryScreen) Title() string {
	return "QUERIES / " + s.query.Label
}

func (s *queryScreen) KeyHints() string {
	hints := formatHint("ctrl+n", "clone") +
		styles.HintSep.String() + formatHint("ctrl+e", "export")

	if s.query.Result != nil && len(s.query.Result.Rows) > 0 {
		rowHints := formatHint("↑↓/jk", "rows")
		if len(s.query.Result.Columns) > maxVisibleCols {
			rowHints += styles.HintSep.String() + formatHint("←→/hl", "columns")
		}
		hints = rowHints + styles.HintSep.String() + hints
	}

	return hints
}

func (s *queryScreen) Init() tea.Cmd {
	if s.executing && s.client != nil {
		return s.executeQuery()
	}
	return nil
}

func (s *queryScreen) executeQuery() tea.Cmd {
	return func() tea.Msg {
		err := s.client.ExecuteQuery(context.Background(), s.query)
		return queryExecutedMsg{query: s.query, err: err}
	}
}

func (s *queryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 1 // account for footer
		s.updateLayout()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlN:
			// Clone: pop current screen and open editor with current SQL
			initialSQL := s.query.SQL
			return s, tea.Sequence(
				func() tea.Msg { return PopScreenMsg{} },
				func() tea.Msg { return PushScreenMsg{Screen: newEditorScreen(s.client, initialSQL)} },
			)
		case tea.KeyCtrlE:
			// Export as CSV
			s.status = ""
			return s, s.exportCSV()
		}

		// Handle horizontal scrolling
		switch msg.String() {
		case "left", "h":
			if s.colOffset > 0 {
				s.colOffset--
				s.rebuildTableContent()
			}
			return s, nil
		case "right", "l":
			maxOffset := s.maxColOffset()
			if s.colOffset < maxOffset {
				s.colOffset++
				s.rebuildTableContent()
			}
			return s, nil
		}

	case csvExportedMsg:
		if msg.err != nil {
			s.err = msg.err
		} else {
			s.status = "Exported to " + msg.path
		}
		return s, nil

	case queryExecutedMsg:
		s.executing = false
		if msg.err != nil {
			s.err = msg.err
		}
		s.colOffset = 0 // Reset horizontal scroll on new results
		s.buildTableContent()
		return s, nil
	}

	var cmd tea.Cmd
	s.table, cmd = s.table.Update(msg)
	return s, cmd
}

func (s *queryScreen) View() string {
	// Need dimensions to render properly
	if s.width == 0 || s.height == 0 {
		return "Loading..."
	}

	// SQL box at the top (content-sized with max height)
	sqlBox := s.renderSQLBox()
	sqlBoxActualHeight := lipgloss.Height(sqlBox)

	// Results section fills remaining space
	resultsHeight := max(s.height - sqlBoxActualHeight, 3)

	resultsContent := s.renderResultsContent()
	resultsBox := styles.TitledBoxTopLeft("RESULTS", resultsContent, s.width, resultsHeight)

	return lipgloss.JoinVertical(lipgloss.Top, sqlBox, resultsBox)
}

func (s *queryScreen) renderResultsContent() string {
	if s.executing {
		return styles.HintDesc.Render("Executing query...")
	}
	if s.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(styles.Red)
		return errStyle.Render("Error: " + s.err.Error())
	}
	if s.query.Result == nil {
		return styles.HintDesc.Render("No results")
	}
	if len(s.query.Result.Rows) == 0 {
		return s.query.Result.Message()
	}

	content := s.table.View()
	if s.status != "" {
		statusStyle := lipgloss.NewStyle().Foreground(styles.Green)
		content += "\n" + statusStyle.Render(s.status)
	}
	return content
}

func (s *queryScreen) renderSQLBox() string {
	highlightedSQL := styles.HighlightSQL(s.query.SQL)
	lines := strings.Split(highlightedSQL, "\n")
	maxLines := sqlBoxHeight - 2 // account for box borders

	if len(lines) > maxLines {
		// Truncate and add indicator
		lines = lines[:maxLines-1]
		lines = append(lines, styles.HintDesc.Render("..."))
	}

	content := strings.Join(lines, "\n")
	return styles.ContentTitledBox(s.Title(), content, s.width)
}

func (s *queryScreen) updateLayout() {
	// Calculate table dimensions
	// SQL box takes sqlBoxHeight, rest goes to results
	// Results box has 2 lines for borders, table header takes ~2 lines
	resultsInnerHeight := max(s.height - sqlBoxHeight - 2 - 2, 1)

	// Width: screen width - 2 for box borders
	innerWidth := max(s.width - 2, 10)

	s.table.SetWidth(innerWidth)
	s.table.SetHeight(resultsInnerHeight)
	s.buildTableContent()
}

// maxColOffset returns the maximum valid column offset (so last 3 columns are shown)
func (s *queryScreen) maxColOffset() int {
	if s.query.Result == nil {
		return 0
	}
	totalCols := len(s.query.Result.Columns)
	if totalCols <= maxVisibleCols {
		return 0
	}
	return totalCols - maxVisibleCols
}

// rebuildTableContent rebuilds table while preserving row selection
func (s *queryScreen) rebuildTableContent() {
	cursor := s.table.Cursor()
	s.buildTableContent()
	s.table.SetCursor(cursor)
}

// Maximum number of columns to show at once
const maxVisibleCols = 3

func (s *queryScreen) buildTableContent() {
	if s.query.Result == nil || len(s.query.Result.Columns) == 0 {
		s.table.SetColumns([]table.Column{})
		s.table.SetRows([]table.Row{})
		return
	}

	totalCols := len(s.query.Result.Columns)
	tableWidth := s.table.Width()
	if tableWidth < 20 {
		tableWidth = 80 // fallback
	}

	// Ensure colOffset is within bounds
	maxOffset := s.maxColOffset()
	if s.colOffset > maxOffset {
		s.colOffset = maxOffset
	}
	if s.colOffset < 0 {
		s.colOffset = 0
	}

	// Calculate how many columns to show (at most maxVisibleCols)
	visibleCols := min(totalCols - s.colOffset, maxVisibleCols)

	// Calculate column width based on maxVisibleCols for consistent sizing
	// This prevents columns from stretching too wide when showing fewer columns
	// Account for separators between columns (3 chars each)
	const separatorWidth = 3
	availableWidth := tableWidth - (maxVisibleCols-1)*separatorWidth
	colWidth := max(availableWidth/maxVisibleCols, 10)
	remainder := availableWidth % maxVisibleCols

	// Build columns, distributing remainder to first columns
	columns := make([]table.Column, visibleCols)
	for i := range visibleCols {
		srcIdx := s.colOffset + i
		w := colWidth
		if i < remainder {
			w++
		}
		columns[i] = table.Column{Title: s.query.Result.Columns[srcIdx], Width: w}
	}

	// Build rows
	rows := make([]table.Row, len(s.query.Result.Rows))
	for i, row := range s.query.Result.Rows {
		rowData := make(table.Row, visibleCols)
		for j := range visibleCols {
			srcIdx := s.colOffset + j
			if srcIdx < len(row) {
				rowData[j] = formatters.FieldToString(row[srcIdx])
			}
		}
		rows[i] = rowData
	}

	// Update styles with current width for full row highlighting
	s.table.SetStyles(styles.TableStyles(tableWidth))

	// Clear rows first, then set columns, then set rows
	// This avoids panic from mismatched column counts during render
	s.table.SetRows([]table.Row{})
	s.table.SetColumns(columns)
	s.table.SetRows(rows)
}

func (s *queryScreen) exportCSV() tea.Cmd {
	return func() tea.Msg {
		if s.query.Result == nil || len(s.query.Result.Rows) == 0 {
			return csvExportedMsg{err: fmt.Errorf("no results to export")}
		}

		filename := fmt.Sprintf("%s-%s.csv", s.query.Label, time.Now().Format("20060102-150405"))
		path, err := exporters.WriteCsv(filename, func(w *csv.Writer) error {
			// Write header
			if err := w.Write(s.query.Result.Columns); err != nil {
				return err
			}

			// Write rows
			for _, row := range s.query.Result.Rows {
				record := make([]string, len(row))
				for i, field := range row {
					record[i] = formatters.FieldToString(field)
				}
				if err := w.Write(record); err != nil {
					return err
				}
			}

			return nil
		})

		if err == nil {
			// ignore errors if external app can't be executed
			launchers.OpenDefaultApp(path)
		}

		return csvExportedMsg{path: path, err: err}
	}
}



