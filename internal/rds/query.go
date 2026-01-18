// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import "time"

// Query is a tune waiting to be played -- SQL on deck, result on standby.
type Query struct {
	CreatedAt time.Time
	Label     string
	SQL       string
	Result    *QueryResult
}
