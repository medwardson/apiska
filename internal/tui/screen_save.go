// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/storage"
	"github.com/ixti/apiska/internal/tui/styles"
)

type saveScreen struct {
	store *storage.Store
	sql   string
	input textinput.Model

	width, height int
}

func newSaveScreen(store *storage.Store, sql string, width int) *saveScreen {
	input := textinput.New()
	input.Placeholder = "My query"
	input.Width = width - 20
	input.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.TextFaint)
	input.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Purple)
	input.Focus()

	return &saveScreen{
		store: store,
		sql:   sql,
		input: input,
		width: width,
	}
}

func (s *saveScreen) Title() string {
	return "SAVE QUERY"
}

func (s *saveScreen) KeyHints() string {
	return formatHint("enter", "save") +
		styles.HintSep.String() + formatHint("esc", "cancel")
}

func (s *saveScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (s *saveScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 1
		s.input.Width = s.width - 20

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return s, func() tea.Msg { return PopScreenMsg{} }

		case tea.KeyEnter:
			name := strings.TrimSpace(s.input.Value())
			if name == "" {
				return s, func() tea.Msg { return PopScreenMsg{} }
			}

			actualName, err := s.store.Save(name, s.sql)
			if err != nil {
				return s, func() tea.Msg { return PopScreenMsg{} }
			}

			return s, tea.Sequence(
				func() tea.Msg { return PopScreenMsg{} },
				func() tea.Msg { return querySavedMsg{name: actualName} },
			)
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *saveScreen) View() string {
	label := styles.HintDesc.Render("Query name: ")
	content := label + s.input.View()
	return styles.TitledBoxTopLeft(s.Title(), content, s.width, s.height)
}
