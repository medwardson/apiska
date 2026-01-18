// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
)

const MaxQuerySize = 65536

var (
	ErrQueryTooLong = fmt.Errorf("query size exceeds %d bytes", MaxQuerySize)
	ErrQueryIsEmpty = errors.New("query can't be empty")
)

// rdsDataAPI defines the RDS Data API operations we need -- keeps the
// soundsystem swappable for testing.
type rdsDataAPI interface {
	ExecuteStatement(ctx context.Context, params *rdsdata.ExecuteStatementInput, optFns ...func(*rdsdata.Options)) (*rdsdata.ExecuteStatementOutput, error)
}

// Client spins the decks -- your connection to the RDS Data API.
type Client struct {
	ClientConfig

	rdsDataClient rdsDataAPI
	idGen         *idGen
}

// NewClient fires up the soundsystem -- one thing that you can depend on.
// Validates your rider, plugs into AWS, and gets the decks ready to spin.
func NewClient(ctx context.Context, rider *ClientConfig) (*Client, error) {
	if err := rider.Validate(); err != nil {
		return nil, fmt.Errorf("invalid APISKA RDS config: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		ClientConfig:  *rider,
		rdsDataClient: rdsdata.NewFromConfig(cfg),
		idGen:         NewIDGenerator(),
	}, nil
}

// NewQuery cues up a fresh track without hitting play.
func (c *Client) NewQuery(sql string) (*Query, error) {
	if len(sql) > MaxQuerySize {
		return nil, ErrQueryTooLong
	}

	if sql == "" {
		return nil, ErrQueryIsEmpty
	}

	query := &Query{
		CreatedAt: time.Now(),
		Label:     c.idGen.Generate(),
		SQL:       sql,
	}

	return query, nil
}

// ExecuteQuery drops the needle and lets it rip.
func (c *Client) ExecuteQuery(ctx context.Context, query *Query) error {
	input := &rdsdata.ExecuteStatementInput{
		ResourceArn:           aws.String(c.ClusterARN),
		SecretArn:             aws.String(c.SecretARN),
		Database:              aws.String(c.Database),
		Sql:                   aws.String(query.SQL),
		IncludeResultMetadata: true,
	}

	output, err := c.rdsDataClient.ExecuteStatement(ctx, input)

	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	query.Result = BuildQueryResult(output)

	return nil
}

// ExecuteSQL is the one-shot single -- cue and play in one motion.
func (c *Client) ExecuteSQL(ctx context.Context, sql string) (*Query, error) {
	query, err := c.NewQuery(sql)

	if err != nil {
		return nil, err
	}

	return query, c.ExecuteQuery(ctx, query)
}
