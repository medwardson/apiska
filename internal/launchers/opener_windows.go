// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package launchers

import (
	"os/exec"
	"syscall"
)

// OpenDefaultApp tells Windows to open your file with whatever app it feels
// like today -- because consistency was never Redmond's strong suit. Uses
// `cmd /c start` because apparently that's the only way to ask Windows nicely.
// Spawns in its own process group so the child can outlive us, like a ballad
// that overstays its welcome at a punk show. Fire and forget -- we're out.
func OpenDefaultApp(filename string) error {
	cmd := exec.Command("cmd", "/c", "start", "", filename)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	return cmd.Start()
}
