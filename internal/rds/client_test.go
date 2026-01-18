// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRDSDataAPI is a struct-based mock for the RDS Data API.
// Add new function fields here as the interface grows.
type mockRDSDataAPI struct {
	ExecuteStatementFunc func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error)
	// Future methods:
	// BatchExecuteStatementFunc func(...) (...)
	// BeginTransactionFunc func(...) (...)
	// CommitTransactionFunc func(...) (...)
	// RollbackTransactionFunc func(...) (...)
}

func (m *mockRDSDataAPI) ExecuteStatement(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
	return m.ExecuteStatementFunc(ctx, params, optFns...)
}

// newTestClient creates a Client with a mock RDS Data API for testing.
func newTestClient(mock *mockRDSDataAPI) *Client {
	return &Client{
		ClientConfig: ClientConfig{
			ClusterARN: "arn:aws:rds:us-east-1:123456789012:cluster:test-cluster",
			SecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-secret",
			Database:   "testdb",
		},
		rdsDataClient: mock,
		idGen:         NewIDGeneratorWithSeed(42),
	}
}

func TestNewClient(t *testing.T) {
	t.Run("returns error when config validation fails", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "", // missing required field
			SecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:test",
			Database:   "testdb",
		}

		client, err := NewClient(t.Context(), &cfg)

		assert.Nil(t, client)
		assert.ErrorContains(t, err, "invalid APISKA RDS config")
	})
}

func TestClient_NewQuery(t *testing.T) {
	t.Run("creates query with SQL and generated label", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})
		before := time.Now()

		query, err := client.NewQuery("SELECT * FROM bands")

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM bands", query.SQL)
		assert.NotEmpty(t, query.Label)
		assert.True(t, query.CreatedAt.After(before) || query.CreatedAt.Equal(before))
		assert.Nil(t, query.Result)
	})

	t.Run("generates unique labels for each query", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})

		q1, err1 := client.NewQuery("SELECT 1")
		q2, err2 := client.NewQuery("SELECT 2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, q1.Label, q2.Label)
	})

	t.Run("returns error when SQL is empty", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})

		query, err := client.NewQuery("")

		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryIsEmpty)
	})

	t.Run("returns error when SQL exceeds max size", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})
		longSQL := string(make([]byte, MaxQuerySize+1))

		query, err := client.NewQuery(longSQL)

		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryTooLong)
	})

	t.Run("accepts SQL at max size", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})
		maxSQL := string(make([]byte, MaxQuerySize))

		query, err := client.NewQuery(maxSQL)

		require.NoError(t, err)
		assert.NotNil(t, query)
	})
}

func TestClient_ExecuteQuery(t *testing.T) {
	t.Run("executes query and populates result", func(t *testing.T) {
		mock := &mockRDSDataAPI{
			ExecuteStatementFunc: func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
				return &rdsdata.ExecuteStatementOutput{
					ColumnMetadata: []types.ColumnMetadata{
						{Name: aws.String("id")},
						{Name: aws.String("name")},
					},
					Records: [][]types.Field{
						{&types.FieldMemberStringValue{Value: "1"}, &types.FieldMemberStringValue{Value: "rancid"}},
					},
				}, nil
			},
		}
		client := newTestClient(mock)
		query, _ := client.NewQuery("SELECT id, name FROM bands")

		err := client.ExecuteQuery(t.Context(), query)

		require.NoError(t, err)
		assert.NotNil(t, query.Result)
		assert.Equal(t, []string{"id", "name"}, query.Result.Columns)
		assert.Len(t, query.Result.Rows, 1)
	})

	t.Run("passes correct parameters to RDS Data API", func(t *testing.T) {
		var capturedParams *rdsdata.ExecuteStatementInput
		mock := &mockRDSDataAPI{
			ExecuteStatementFunc: func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
				capturedParams = params
				return &rdsdata.ExecuteStatementOutput{}, nil
			},
		}
		client := newTestClient(mock)
		query, _ := client.NewQuery("SELECT 1")

		_ = client.ExecuteQuery(t.Context(), query)

		require.NotNil(t, capturedParams)
		assert.Equal(t, "arn:aws:rds:us-east-1:123456789012:cluster:test-cluster", *capturedParams.ResourceArn)
		assert.Equal(t, "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-secret", *capturedParams.SecretArn)
		assert.Equal(t, "testdb", *capturedParams.Database)
		assert.Equal(t, "SELECT 1", *capturedParams.Sql)
		assert.True(t, capturedParams.IncludeResultMetadata)
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		mock := &mockRDSDataAPI{
			ExecuteStatementFunc: func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
				return nil, errors.New("connection refused")
			},
		}
		client := newTestClient(mock)
		query, _ := client.NewQuery("SELECT 1")

		err := client.ExecuteQuery(t.Context(), query)

		assert.ErrorContains(t, err, "query execution failed")
		assert.ErrorContains(t, err, "connection refused")
		assert.Nil(t, query.Result)
	})
}

func TestClient_ExecuteSQL(t *testing.T) {
	t.Run("creates and executes query in one call", func(t *testing.T) {
		mock := &mockRDSDataAPI{
			ExecuteStatementFunc: func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
				return &rdsdata.ExecuteStatementOutput{
					NumberOfRecordsUpdated: 5,
				}, nil
			},
		}
		client := newTestClient(mock)

		query, err := client.ExecuteSQL(t.Context(), "DELETE FROM genres WHERE name = 'reggaeton'")

		require.NoError(t, err)
		assert.Equal(t, "DELETE FROM genres WHERE name = 'reggaeton'", query.SQL)
		assert.NotEmpty(t, query.Label)
		assert.NotNil(t, query.Result)
		assert.Equal(t, int64(5), query.Result.NumberOfRowsUpdated)
	})

	t.Run("returns error when execution fails", func(t *testing.T) {
		mock := &mockRDSDataAPI{
			ExecuteStatementFunc: func(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error) {
				return nil, errors.New("syntax error")
			},
		}
		client := newTestClient(mock)

		query, err := client.ExecuteSQL(t.Context(), "SELEKT 1")

		assert.Error(t, err)
		assert.NotNil(t, query) // query is returned even on execution error
		assert.Nil(t, query.Result)
	})

	t.Run("returns error when SQL is empty", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})

		query, err := client.ExecuteSQL(t.Context(), "")

		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryIsEmpty)
	})

	t.Run("returns error when SQL exceeds max size", func(t *testing.T) {
		client := newTestClient(&mockRDSDataAPI{})
		longSQL := string(make([]byte, MaxQuerySize+1))

		query, err := client.ExecuteSQL(t.Context(), longSQL)

		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryTooLong)
	})
}
