// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package launchers

import (
	"fmt"
	"os"
	"os/exec"
)

// ExternalEditor is your backstage pass to the temp file we toss at your
// editor of choice. Holds onto the path like a rude boy holds onto the mic.
type ExternalEditor struct {
	tempPath string
}

// detectEditor checks who's headlining tonight. APISKA_EDITOR gets top billing,
// EDITOR is the reliable opener if the headliner's a no-show. No bands? No gig.
func detectEditor() (string, error) {
	editor := os.Getenv("APISKA_EDITOR")

	if editor != "" {
		return editor, nil
	}

	editor = os.Getenv("EDITOR")

	if editor != "" {
		return editor, nil 
	}

	return editor, ErrEditorNotConfigured
}

// OpenExternalEditor drops the needle on your favorite editor -- pick it up!
// Pass a file extension (like ".sql") and initial content to get the horn
// section warmed up. Checks APISKA_EDITOR first, then EDITOR env vars --
// returns ErrEditorNotConfigured if neither is set (no show without a venue).
// Don't forget to call Cleanup() when done -- leave the venue how you found it!
func OpenExternalEditor(fileExtension, initialContent string) (*exec.Cmd, *ExternalEditor, error) {
	editor, err := detectEditor()

	if err != nil {
		return nil, nil, err
	}

	tmpFile, err := os.CreateTemp("", "apiska-*"+fileExtension)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(initialContent); err != nil {
		tmpFile.Close()
		os.Remove(tempPath)
		return nil, nil, fmt.Errorf("failed to write initial content: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command(editor, tempPath)
	return cmd, &ExternalEditor{tempPath: tempPath}, nil
}

// ReadContent skanks back with whatever you wrote after your editing solo.
// Like checking the setlist after the show -- returns the goods or an error.
func (e *ExternalEditor) ReadContent() (string, error) {
	content, err := os.ReadFile(e.tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to read editor content: %w", err)
	}
	return string(content), nil
}

// Cleanup picks up the temp file and tosses it in the pit. Always call this
// when you're done -- don't be that guy who leaves Mahou cans at the venue!
func (e *ExternalEditor) Cleanup() {
	os.Remove(e.tempPath)
}
