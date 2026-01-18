// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build darwin || freebsd || linux

package launchers

import (
	"os/exec"
	"runtime"
	"syscall"
)

// OpenDefaultApp kicks down the door and lets your OS's default app handle the
// file -- like crowd-surfing your data to whatever program wants to catch it.
// On macOS we holler at `open`, on Linux/FreeBSD we page `xdg-open` (the unsung
// hero of desktop integration). We spawn the process in its own group -- that
// way the kid keeps moshing while we fade to the back. Oi! Oi! Oi!
func OpenDefaultApp(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filename)
	case "freebsd", "linux":
		cmd = exec.Command("xdg-open", filename)
	default:
		return ErrUnsupportedPlatform
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}
