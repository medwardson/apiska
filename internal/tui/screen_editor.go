// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/launchers"
	"github.com/ixti/apiska/internal/rds"
	"github.com/ixti/apiska/internal/tui/styles"
)

type editorScreen struct {
	client   *rds.Client
	textarea textarea.Model
	err      error

	width, height int
}

func newEditorScreen(client *rds.Client, initialSQL string) *editorScreen {
	ta := textarea.New()
	ta.Placeholder = "SELECT * FROM ..."
	ta.SetValue(initialSQL)
	ta.Focus()
	ta.CharLimit = rds.MaxQuerySize

	// Style the textarea - no border since TitledBox provides it
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	return &editorScreen{
		client:   client,
		textarea: ta,
	}
}

func (s *editorScreen) Title() string {
	return "NEW QUERY"
}

func (s *editorScreen) KeyHints() string {
	return formatHint("F5/ctrl+r", "execute") +
		styles.HintSep.String() + formatHint("ctrl+e", "external editor")
}

func (s *editorScreen) Init() tea.Cmd {
	return textarea.Blink
}

func (s *editorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 1 // account for footer
		s.updateTextareaSize()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyF5, tea.KeyCtrlR:
			return s, s.executeQuery()
		case tea.KeyCtrlE:
			return s, s.openExternalEditor()
		}

	case editorFinishedMsg:
		if msg.err != nil {
			s.err = msg.err
			return s, nil
		}
		defer msg.editor.Cleanup()

		content, err := msg.editor.ReadContent()
		if err != nil {
			s.err = err
			return s, nil
		}

		s.textarea.SetValue(content)
		return s, nil

	case queryExecutedMsg:
		// This shouldn't happen anymore since query screen handles execution
		// but keep for safety
		if msg.err != nil {
			s.err = msg.err
			return s, nil
		}
		return s, nil
	}

	var cmd tea.Cmd
	s.textarea, cmd = s.textarea.Update(msg)
	return s, cmd
}

func (s *editorScreen) View() string {
	var content string

	if s.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(styles.Red)
		content = s.textarea.View() + "\n" + errStyle.Render("Error: "+s.err.Error())
	} else {
		content = s.textarea.View()
	}

	return styles.TitledBoxTopLeft(s.Title(), content, s.width, s.height)
}

func (s *editorScreen) updateTextareaSize() {
	// Account for TitledBox borders (2 left/right)
	s.textarea.SetWidth(s.width - 2)
	// Account for TitledBox borders (2 top/bottom) + extra line for potential error
	s.textarea.SetHeight(s.height - 4)
}

func (s *editorScreen) executeQuery() tea.Cmd {
	sql := strings.TrimSpace(s.textarea.Value())
	if sql == "" {
		return nil
	}

	query, err := s.client.NewQuery(sql)
	if err != nil {
		s.err = err
		return nil
	}

	// Create query screen that will execute in background
	queryScreen := newQueryScreen(query)
	queryScreen.client = s.client
	queryScreen.executing = true

	return func() tea.Msg {
		return submitQueryMsg{query: query, screen: queryScreen}
	}
}

func (s *editorScreen) openExternalEditor() tea.Cmd {
	cmd, editor, err := launchers.OpenExternalEditor(".sql", s.textarea.Value())
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorFinishedMsg{editor: editor, err: err}
	})
}


