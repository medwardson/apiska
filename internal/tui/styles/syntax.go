// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package styles

import (
	"bytes"
	"embed"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/colorprofile"
)

//go:embed syntax/*.xml
var syntaxStyleSpecs embed.FS

var syntaxFormatter chroma.Formatter
var syntaxStyle *chroma.Style
var syntaxLexer chroma.Lexer

// init loads all embedded syntax styles and sizes up your terminal's color
// chops. TrueColor? Full lasers. ANSI256? Disco ball. Plain ANSI? Colored
// spots. Nothing? House lights only.
func init() {
	flexokiDarkData, err := syntaxStyleSpecs.Open("syntax/flexoki-dark.xml")
	if err != nil {
		panic(err)
	}

	defer flexokiDarkData.Close()

	flexokiDarkStyle, err := chroma.NewXMLStyle(flexokiDarkData)
	if err != nil {
		panic(err)
	}

	syntaxStyle = flexokiDarkStyle
	syntaxLexer = lexers.Get("sql")

	switch colorprofile.Detect(os.Stdout, os.Environ()) {
	case colorprofile.TrueColor:
		syntaxFormatter = formatters.Get("terminal16m")
	case colorprofile.ANSI256:
		syntaxFormatter = formatters.Get("terminal256")
	case colorprofile.ANSI:
		syntaxFormatter = formatters.Get("terminal")
	default:
		syntaxFormatter = formatters.Get("noop")
	}
}

// HighlightSQL drags your SQL to the makeup chair -- gets those keywords
// stage-ready. If the makeup artist bails, no sweat -- fans will just see
// your ugly SQL au naturel.
func HighlightSQL(sql string) string {
	iterator, err := syntaxLexer.Tokenise(nil, sql)
	if err != nil {
		return sql
	}

	var buf bytes.Buffer
	if err := syntaxFormatter.Format(&buf, syntaxStyle, iterator); err != nil {
		return sql
	}

	return buf.String()
}
