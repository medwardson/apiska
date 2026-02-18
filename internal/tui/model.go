// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ixti/apiska/internal/launchers"
	"github.com/ixti/apiska/internal/rds"
	"github.com/ixti/apiska/internal/storage"
	"github.com/ixti/apiska/internal/tui/styles"
)

// Screen represents a single screen in the TUI stack.
// Each screen is a self-contained tea.Model that handles its own
// updates and rendering within the content area.
type Screen interface {
	tea.Model

	// Title returns the screen's title for the header.
	Title() string

	// KeyHints returns context-specific keyboard hints for the footer.
	KeyHints() string
}

// PushScreenMsg pushes a new screen onto the stack.
type PushScreenMsg struct {
	Screen Screen
}

// PopScreenMsg pops the current screen from the stack.
type PopScreenMsg struct{}

// queryAddedMsg is sent when a query should be added to history.
type queryAddedMsg struct {
	query *rds.Query
}

// queryExecutedMsg is sent when a query finishes executing.
type queryExecutedMsg struct {
	query *rds.Query
	err   error
}

// editorFinishedMsg is sent when the external editor closes.
type editorFinishedMsg struct {
	editor *launchers.ExternalEditor
	err    error
}

// submitQueryMsg is sent when a query is submitted for execution.
// It handles: adding to history, popping editor, pushing query screen, starting execution.
type submitQueryMsg struct {
	query  *rds.Query
	screen Screen
}

// loadSavedQueryMsg is sent when a saved query should be loaded into a new editor.
type loadSavedQueryMsg struct {
	sql string
}

// querySavedMsg is sent when a query has been saved successfully.
type querySavedMsg struct {
	name string
}

// model is the root model that manages the screen stack and chrome.
type model struct {
	client *rds.Client
	store  *storage.Store

	screens []Screen

	width, height int
}

func (m model) Init() tea.Cmd {
	if len(m.screens) == 0 {
		return nil
	}

	return m.screens[len(m.screens)-1].Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ, tea.KeyF10:
			// Graceful quit - will prompt for transaction handling in the future
			return m, tea.Quit
		case tea.KeyCtrlC:
			// Force quit - will silently rollback transactions in the future
			return m, tea.Quit
		case tea.KeyEsc:
			if len(m.screens) > 1 {
				m.screens = m.screens[:len(m.screens)-1]
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case PushScreenMsg:
		m.screens = append(m.screens, msg.Screen)
		// Forward current window size to the new screen
		if m.width > 0 && m.height > 0 {
			updated, _ := msg.Screen.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.screens[len(m.screens)-1] = updated.(Screen)
		}
		return m, msg.Screen.Init()

	case PopScreenMsg:
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
		}
		return m, nil

	case queryAddedMsg:
		// Forward to home screen (first in stack) to update query history
		if len(m.screens) > 0 {
			updated, cmd := m.screens[0].Update(msg)
			m.screens[0] = updated.(Screen)
			return m, cmd
		}
		return m, nil

	case queryExecutedMsg:
		// Forward to current screen, then update home screen to refresh result
		var cmd tea.Cmd
		if len(m.screens) > 0 {
			current := m.screens[len(m.screens)-1]
			updated, c := current.Update(msg)
			m.screens[len(m.screens)-1] = updated.(Screen)
			cmd = c
		}
		// Also forward to home screen to update the result column
		if len(m.screens) > 0 {
			updated, _ := m.screens[0].Update(msg)
			m.screens[0] = updated.(Screen)
		}
		return m, cmd

	case submitQueryMsg:
		// Add query to home screen history
		if len(m.screens) > 0 {
			updated, _ := m.screens[0].Update(queryAddedMsg{query: msg.query})
			m.screens[0] = updated.(Screen)
		}
		// Pop current screen (editor)
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
		}
		// Push query screen and start execution
		m.screens = append(m.screens, msg.screen)
		// Forward current window size to the new screen
		if m.width > 0 && m.height > 0 {
			updated, _ := msg.screen.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.screens[len(m.screens)-1] = updated.(Screen)
		}
		return m, msg.screen.Init()

	case loadSavedQueryMsg:
		// Push editor with SQL (keep saved queries screen in stack for back navigation)
		screen := newEditorScreen(m.client, m.store, msg.sql)
		m.screens = append(m.screens, screen)
		if m.width > 0 && m.height > 0 {
			updated, _ := screen.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.screens[len(m.screens)-1] = updated.(Screen)
		}
		return m, screen.Init()
	}

	// Delegate to the current screen
	if len(m.screens) > 0 {
		current := m.screens[len(m.screens)-1]
		updated, cmd := current.Update(msg)
		m.screens[len(m.screens)-1] = updated.(Screen)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	contentHeight := m.height - 1 // footer only
	content := ""
	if len(m.screens) > 0 {
		content = m.screens[len(m.screens)-1].View()
	}
	content = lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Render(content)

	hints := []string{}
	if len(m.screens) > 1 {
		hints = append(hints, formatHint("esc", "back"))
	}
	if len(m.screens) > 0 {
		if screenHints := m.screens[len(m.screens)-1].KeyHints(); screenHints != "" {
			hints = append(hints, screenHints)
		}
	}
	hints = append(hints, formatHint("ctrl+q", "quit"))

	footer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(strings.Join(hints, styles.HintSep.String()))

	return lipgloss.JoinVertical(lipgloss.Top, content, footer)
}

func formatHint(key, desc string) string {
	return styles.HintKey.Render(key) + " " + styles.HintDesc.Render(desc)
}

// ContentSize returns the available size for screen content.
func (m model) ContentSize() (width, height int) {
	return m.width, m.height - 1
}

func NewProgram(client *rds.Client, store *storage.Store) *tea.Program {
	return tea.NewProgram(model{
		client:  client,
		store:   store,
		screens: []Screen{newHomeScreen(client, store)},
	})
}


















