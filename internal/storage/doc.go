// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package storage handles persistence of user data -- your personal crate of
// vinyl that survives the gig. Saved queries get pressed to disk and stay there
// until you're ready to spin them again. Everything lives in ~/.apiska/ so your
// setlists follow you wherever you go.
package storage
