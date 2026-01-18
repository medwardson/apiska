// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
)

// QueryResult is what comes back from the pit.
//
//   - SELECT fills the `Columns` and `Rows`
//   - INSERT/UPDATE/DELETE counts the damage in `NumberOfRowsUpdated`
type QueryResult struct {
	Columns             []string
	Rows                [][]types.Field
	NumberOfRowsUpdated int64
}

// Message calls out what went down -- rows returned, rows wrecked, or just
// a nod that the set played clean.
func (qr *QueryResult) Message() string {
	if len(qr.Rows) > 0 {
		return fmt.Sprintf("%d row(s) returned", len(qr.Rows))
	} else if qr.NumberOfRowsUpdated > 0 {
		return fmt.Sprintf("%d row(s) affected", qr.NumberOfRowsUpdated)
	} else {
		return "Query executed successfully"
	}
}

// BuildQueryResult masters the raw take from the soundboard -- pulls the
// setlist from metadata and presses the tracks into something you can spin.
func BuildQueryResult(output *rdsdata.ExecuteStatementOutput) *QueryResult {
	result := &QueryResult{NumberOfRowsUpdated: output.NumberOfRecordsUpdated}

	if len(output.ColumnMetadata) > 0 {
		result.Columns = make([]string, len(output.ColumnMetadata))

		for i, col := range output.ColumnMetadata {
			if col.Name != nil && *col.Name != "" {
				result.Columns[i] = *col.Name
			} else if col.Label != nil && *col.Label != "" {
				result.Columns[i] = *col.Label
			} else {
				result.Columns[i] = genericColumnNameByIndex(i)
			}
		}
	}

	if len(output.Records) > 0 {
		result.Rows = output.Records

		// Fallback: generate column names if metadata was not returned but rows exist
		if len(result.Columns) == 0 && len(result.Rows[0]) > 0 {
			result.Columns = make([]string, len(result.Rows[0]))

			for i := range result.Columns {
				result.Columns[i] = genericColumnNameByIndex(i)
			}
		}
	}

	return result
}

// genericColumnNameByIndex slaps a number on nameless tracks -- when the
// setlist's missing, at least you know which song you're on.
func genericColumnNameByIndex(i int) string {
	return fmt.Sprintf("(%d)", i+1)
}
