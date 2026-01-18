// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !darwin && !freebsd && !linux && !windows

package launchers

// OpenDefaultApp on unknown platforms is like showing up to a venue that
// doesn't exist -- sorry mate, no show tonight. Returns ErrUnsupportedPlatform
// because we don't know how to open files on whatever exotic OS you're running.
// Plan 9? Haiku? Respect the hustle, but we ain't got the hooks.
func OpenDefaultApp(filename string) error {
	return ErrUnsupportedPlatform
}
