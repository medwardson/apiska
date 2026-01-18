// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateClusterARN(t *testing.T) {
	t.Run("empty string is valid", func(t *testing.T) {
		assert.NoError(t, ValidateClusterARN(""))
	})

	t.Run("valid ARNs", func(t *testing.T) {
		validARNs := []string{
			"arn:aws:rds:us-east-1:123456789012:cluster:my-cluster",
			"arn:aws:rds:eu-west-1:123456789012:cluster:my-aurora-cluster-01",
			"arn:aws:rds:ap-southeast-2:999999999999:cluster:Cluster1",
		}
		for _, arn := range validARNs {
			assert.NoError(t, ValidateClusterARN(arn), "expected %q to be valid", arn)
		}
	})

	t.Run("invalid ARNs", func(t *testing.T) {
		invalidARNs := []string{
			"aws:rds:us-east-1:123456789012:cluster:my-cluster",      // missing arn prefix
			"arn:aws:ec2:us-east-1:123456789012:cluster:my-cluster",  // wrong service
			"arn:aws:rds:us-east-1:12345678901:cluster:my-cluster",   // account ID too short
			"arn:aws:rds:us-east-1:1234567890123:cluster:my-cluster", // account ID too long
			"arn:aws:rds:us-east-1:123456789012:cluster:1cluster",    // cluster name starts with number
			"arn:aws:rds:us-east-1:123456789012:cluster:my_cluster",  // cluster name with underscore
			"arn:aws:rds:us-east-1:123456789012:db:my-database",      // not a cluster resource
			"not-an-arn", // random string
		}
		for _, arn := range invalidARNs {
			assert.Error(t, ValidateClusterARN(arn), "expected %q to be invalid", arn)
		}
	})
}

func TestValidateSecretARN(t *testing.T) {
	t.Run("empty string is valid", func(t *testing.T) {
		assert.NoError(t, ValidateSecretARN(""))
	})

	t.Run("valid ARNs", func(t *testing.T) {
		validARNs := []string{
			"arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret",
			"arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-AbCdEf",
			"arn:aws:secretsmanager:eu-west-1:123456789012:secret:prod/db/credentials",
			"arn:aws:secretsmanager:us-east-1:123456789012:secret:my_secret+name=value@domain.com",
		}
		for _, arn := range validARNs {
			assert.NoError(t, ValidateSecretARN(arn), "expected %q to be valid", arn)
		}
	})

	t.Run("invalid ARNs", func(t *testing.T) {
		invalidARNs := []string{
			"aws:secretsmanager:us-east-1:123456789012:secret:my-secret", // missing arn prefix
			"arn:aws:ssm:us-east-1:123456789012:secret:my-secret",        // wrong service
			"arn:aws:secretsmanager:us-east-1:12345:secret:my-secret",    // account ID wrong length
			"not-an-arn", // random string
		}
		for _, arn := range invalidARNs {
			assert.Error(t, ValidateSecretARN(arn), "expected %q to be invalid", arn)
		}
	})
}

func TestClientConfig_Validate(t *testing.T) {
	t.Run("returns nil when all required fields are set", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "arn:aws:rds:us-east-1:123456789012:cluster:my-cluster",
			SecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret",
			Database:   "mydb",
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("returns error when cluster ARN is missing", func(t *testing.T) {
		cfg := ClientConfig{
			SecretARN: "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret",
			Database:  "mydb",
		}
		assert.ErrorContains(t, cfg.Validate(), "ClusterARN")
	})

	t.Run("returns error when cluster ARN format is invalid", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "not-a-valid-arn",
			SecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret",
			Database:   "mydb",
		}
		assert.ErrorContains(t, cfg.Validate(), "invalid RDS cluster ARN format")
	})

	t.Run("returns error when secret ARN is missing", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "arn:aws:rds:us-east-1:123456789012:cluster:my-cluster",
			Database:   "mydb",
		}
		assert.ErrorContains(t, cfg.Validate(), "SecretARN")
	})

	t.Run("returns error when secret ARN format is invalid", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "arn:aws:rds:us-east-1:123456789012:cluster:my-cluster",
			SecretARN:  "not-a-valid-arn",
			Database:   "mydb",
		}
		assert.ErrorContains(t, cfg.Validate(), "invalid Secrets Manager ARN format")
	})

	t.Run("returns error when database is missing", func(t *testing.T) {
		cfg := ClientConfig{
			ClusterARN: "arn:aws:rds:us-east-1:123456789012:cluster:my-cluster",
			SecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret",
		}
		assert.ErrorContains(t, cfg.Validate(), "Database")
	})

	t.Run("returns error when all fields are empty", func(t *testing.T) {
		cfg := ClientConfig{}
		assert.Error(t, cfg.Validate())
	})
}
