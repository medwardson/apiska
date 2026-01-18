// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/stretchr/testify/assert"
)

func rdsDataStringValue(value string) types.Field {
	return &types.FieldMemberStringValue{Value: value}
}

func TestQueryResult_Message(t *testing.T) {
	t.Run("reports rows returned for SELECT results", func(t *testing.T) {
		qr := &QueryResult{
			Columns: []string{"id", "name"},
			Rows: [][]types.Field{
				{rdsDataStringValue("1"), rdsDataStringValue("rancid")},
				{rdsDataStringValue("2"), rdsDataStringValue("madness")},
			},
		}

		assert.Equal(t, "2 row(s) returned", qr.Message())
	})

	t.Run("reports single row returned", func(t *testing.T) {
		qr := &QueryResult{
			Columns: []string{"id"},
			Rows: [][]types.Field{
				{rdsDataStringValue("1")},
			},
		}

		assert.Equal(t, "1 row(s) returned", qr.Message())
	})

	t.Run("reports rows affected for INSERT/UPDATE/DELETE", func(t *testing.T) {
		qr := &QueryResult{
			NumberOfRowsUpdated: 42,
		}

		assert.Equal(t, "42 row(s) affected", qr.Message())
	})

	t.Run("reports single row affected", func(t *testing.T) {
		qr := &QueryResult{
			NumberOfRowsUpdated: 1,
		}

		assert.Equal(t, "1 row(s) affected", qr.Message())
	})

	t.Run("reports success when no rows returned or affected", func(t *testing.T) {
		qr := &QueryResult{}

		assert.Equal(t, "Query executed successfully", qr.Message())
	})

	t.Run("rows take precedence over rows updated", func(t *testing.T) {
		qr := &QueryResult{
			Rows: [][]types.Field{
				{rdsDataStringValue("1")},
			},
			NumberOfRowsUpdated: 99,
		}

		assert.Equal(t, "1 row(s) returned", qr.Message())
	})
}

func TestBuildQueryResult(t *testing.T) {
	t.Run("copies NumberOfRecordsUpdated", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			NumberOfRecordsUpdated: 42,
		}

		result := BuildQueryResult(output)

		assert.Equal(t, int64(42), result.NumberOfRowsUpdated)
	})

	t.Run("extracts column names from metadata Name field", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			ColumnMetadata: []types.ColumnMetadata{
				{Name: aws.String("id")},
				{Name: aws.String("band_name")},
			},
		}

		result := BuildQueryResult(output)

		assert.Equal(t, []string{"id", "band_name"}, result.Columns)
	})

	t.Run("falls back to Label when Name is empty", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			ColumnMetadata: []types.ColumnMetadata{
				{Name: aws.String(""), Label: aws.String("ID")},
				{Name: nil, Label: aws.String("Band")},
			},
		}

		result := BuildQueryResult(output)

		assert.Equal(t, []string{"ID", "Band"}, result.Columns)
	})

	t.Run("falls back to index when Name and Label are empty", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			ColumnMetadata: []types.ColumnMetadata{
				{Name: aws.String(""), Label: aws.String("")},
				{Name: nil, Label: nil},
			},
		}

		result := BuildQueryResult(output)

		assert.Equal(t, []string{"(1)", "(2)"}, result.Columns)
	})

	t.Run("uses mixed column name sources", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			ColumnMetadata: []types.ColumnMetadata{
				{Name: aws.String("id")},
				{Name: aws.String(""), Label: aws.String("Label")},
				{Name: nil, Label: nil},
			},
		}

		result := BuildQueryResult(output)

		assert.Equal(t, []string{"id", "Label", "(3)"}, result.Columns)
	})

	t.Run("copies rows from Records", func(t *testing.T) {
		rows := [][]types.Field{
			{rdsDataStringValue("1"), rdsDataStringValue("rancid")},
			{rdsDataStringValue("2"), rdsDataStringValue("madness")},
		}
		output := &rdsdata.ExecuteStatementOutput{
			ColumnMetadata: []types.ColumnMetadata{
				{Name: aws.String("id")},
				{Name: aws.String("name")},
			},
			Records: rows,
		}

		result := BuildQueryResult(output)

		assert.Equal(t, rows, result.Rows)
	})

	t.Run("generates column names from row width when metadata is missing", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{
			Records: [][]types.Field{
				{rdsDataStringValue("a"), rdsDataStringValue("b"), rdsDataStringValue("c")},
			},
		}

		result := BuildQueryResult(output)

		assert.Equal(t, []string{"(1)", "(2)", "(3)"}, result.Columns)
	})

	t.Run("returns empty result for empty output", func(t *testing.T) {
		output := &rdsdata.ExecuteStatementOutput{}

		result := BuildQueryResult(output)

		assert.Equal(t, int64(0), result.NumberOfRowsUpdated)
		assert.Nil(t, result.Columns)
		assert.Nil(t, result.Rows)
	})
}
