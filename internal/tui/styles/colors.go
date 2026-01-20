// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package styles

import "github.com/charmbracelet/lipgloss"

// Flexoki Dark palette - https://stephango.com/flexoki
// The ink that makes the press pop. Every color hand-picked
// for readability when the lights go down.
var (
	// Base tones - the black wax and paper white
	Black  = lipgloss.Color("#100F0F")
	Base   = lipgloss.Color("#1C1B1A")
	Raised = lipgloss.Color("#282726")

	// Text hierarchy - from headline to fine print
	Text      = lipgloss.Color("#CECDC3")
	TextMuted = lipgloss.Color("#878580")
	TextFaint = lipgloss.Color("#6F6E69")

	// Accent colors - the brass section
	Red     = lipgloss.Color("#D14D41")
	Orange  = lipgloss.Color("#DA702C")
	Yellow  = lipgloss.Color("#D0A215")
	Green   = lipgloss.Color("#879A39")
	Cyan    = lipgloss.Color("#3AA99F")
	Blue    = lipgloss.Color("#4385BE")
	Purple  = lipgloss.Color("#8B7EC8")
	Magenta = lipgloss.Color("#CE5D97")
)
