// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package launchers

import (
	"errors"
)

var ErrEditorNotConfigured = errors.New("editor not configured: set APISKA_EDITOR or EDITOR environment variable")
var ErrUnsupportedPlatform = errors.New("unsupported platform")
