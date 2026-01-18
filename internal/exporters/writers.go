// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package exporters

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFile is your roadie for getting data onto disk -- handles all the heavy
// lifting so you can focus on the show. Pass a filename and a callback that
// writes your content, and this bad boy will:
//
//   - Figure out the absolute path (no getting lost on the way to the venue)
//   - Create any missing directories (building the stage if needed)
//   - Open the file fresh (truncate it like a tight snare hit)
//   - Let your callback skank all over that io.Writer
//   - Sync to disk (making sure your tracks are pressed to vinyl)
//
// Returns the absolute path on success -- your golden ticket to the afterparty.
// Don't forget: errors get wrapped with context, we ain't leaving you in the mosh pit blind!
func WriteFile(filename string, fn func(io.Writer) error) (string, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", filename, err)
	}

	dirname := filepath.Dir(absolutePath)
	if err := os.MkdirAll(dirname, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directories for %s: %w", dirname, err)
	}

	f, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", absolutePath, err)
	}
	defer f.Close()

	if err := fn(f); err != nil {
		return "", err
	}

	if err := f.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync file %s: %w", absolutePath, err)
	}

	return absolutePath, nil
}

// WriteCsv is WriteFile's horn-blowing cousin -- just like Lagwagon to RKL.
// Similar energy but with csv.Writer fronting the band. Your callback gets to
// riff on the CSV writer while we handle the opening act (file setup) and the
// encore (flush & error check).
func WriteCsv(filename string, fn func(*csv.Writer) error) (string, error) {
	return WriteFile(filename, func(f io.Writer) error {
		w := csv.NewWriter(f)

		if err := fn(w); err != nil {
			return err
		}

		w.Flush()
		return w.Error()
	})
}

// WriteJson is WriteFile's well-dressed sibling -- think Less Than Jake at a
// wedding reception -- structured, formatted, but still knows how to party.
// Your callback gets a json.Encoder to serialize your data while we handle
// the venue booking (file setup) and cleanup crew (sync to disk).
// Perfect for when your data needs to look sharp in curly braces.
func WriteJson(filename string, fn func(*json.Encoder) error) (string, error) {
	return WriteFile(filename, func(f io.Writer) error {
		enc := json.NewEncoder(f)
		return fn(enc)
	})
}
