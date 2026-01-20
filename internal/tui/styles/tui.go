// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package styles

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Keyboard hints - the setlist taped to the floor
var (
	HintKey = lipgloss.NewStyle().
		Foreground(TextFaint)

	HintDesc = lipgloss.NewStyle().
		Foreground(TextMuted)

	HintSep = lipgloss.NewStyle().
		Foreground(Raised).
		SetString(" · ")
)

// Welcome screen - the marquee outside the venue
var (
	WelcomeTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple)

	WelcomeSubtitle = lipgloss.NewStyle().
			Foreground(TextMuted)

	boxBorder = lipgloss.NewStyle().
			Foreground(Purple)

	boxTitle = lipgloss.NewStyle().
			Foreground(Purple)
)

// TitledBox renders a full-size box with a title embedded in the top border.
// Content is centered both horizontally and vertically.
// Like the venue name on the marquee -- you know where you are before you walk in.
func TitledBox(title, content string, width, height int) string {
	if width < 4 || height < 3 {
		return content
	}

	innerWidth := width - 2 // account for left/right borders

	// Render content centered in the available space
	innerHeight := height - 2 // account for top/bottom borders
	centeredContent := lipgloss.Place(
		innerWidth,
		innerHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	return renderBox(title, centeredContent, innerWidth)
}

// TitledBoxTopLeft renders a full-size box with content aligned to top-left.
// Like stepping into the venue -- content fills the space from the stage out.
func TitledBoxTopLeft(title, content string, width, height int) string {
	if width < 4 || height < 3 {
		return content
	}

	innerWidth := width - 2  // account for left/right borders
	innerHeight := height - 2 // account for top/bottom borders

	// Place content at top-left, filling the box
	alignedContent := lipgloss.Place(
		innerWidth,
		innerHeight,
		lipgloss.Left,
		lipgloss.Top,
		content,
	)

	return renderBox(title, alignedContent, innerWidth)
}

// ContentTitledBox renders a box that fits its content with a title in the top border.
// Like a 7" single sleeve -- just enough room for the track.
func ContentTitledBox(title, content string, width int) string {
	if width < 4 {
		return content
	}

	innerWidth := width - 2 // account for left/right borders

	// Pad each line to fill the width
	lines := strings.Split(content, "\n")
	padded := make([]string, len(lines))
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < innerWidth {
			padded[i] = line + strings.Repeat(" ", innerWidth-lineWidth)
		} else {
			padded[i] = line
		}
	}

	return renderBox(title, strings.Join(padded, "\n"), innerWidth)
}

// ContentBox renders a box that fits its content without a title.
// Plain sleeve, no label -- the music speaks for itself.
func ContentBox(content string, width int) string {
	return ContentTitledBox("", content, width)
}

// renderBox is the pressing machine -- stamps out the final product.
func renderBox(title, content string, innerWidth int) string {
	// Build top border: ╭─TITLE───...───╮ or ╭───...───╮
	var topBorder string
	if title != "" {
		titleStr := boxTitle.Render(title)
		titleLen := lipgloss.Width(titleStr)
		dashesNeeded := innerWidth - titleLen - 1 // -1 for dash before title
		if dashesNeeded < 0 {
			dashesNeeded = 0
		}
		topBorder = boxBorder.Render("╭─") +
			titleStr +
			boxBorder.Render(strings.Repeat("─", dashesNeeded)+"╮")
	} else {
		topBorder = boxBorder.Render("╭" + strings.Repeat("─", innerWidth) + "╮")
	}

	// Build bottom border: ╰───...───╯
	bottomBorder := boxBorder.Render("╰" + strings.Repeat("─", innerWidth) + "╯")

	// Add side borders to each line
	lines := strings.Split(content, "\n")
	bordered := make([]string, len(lines))
	for i, line := range lines {
		// Ensure line is exactly innerWidth
		lineWidth := lipgloss.Width(line)
		if lineWidth < innerWidth {
			line += strings.Repeat(" ", innerWidth-lineWidth)
		}
		bordered[i] = boxBorder.Render("│") + line + boxBorder.Render("│")
	}

	return topBorder + "\n" + strings.Join(bordered, "\n") + "\n" + bottomBorder
}

// TableStyles returns table styles. Width parameter ensures full row highlighting.
func TableStyles(width int) table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(TextFaint).
		BorderBottom(true).
		Bold(false).
		Foreground(TextMuted)

	s.Selected = s.Selected.
		Foreground(Text).
		Background(Purple).
		Bold(false).
		Width(width).
		MaxWidth(width)

	s.Cell = s.Cell.
		Foreground(Text)

	return s
}




