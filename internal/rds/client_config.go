// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	rdsClusterARNPattern     = regexp.MustCompile(`^arn:aws:rds:[a-z0-9-]+:[0-9]{12}:cluster:[a-zA-Z][a-zA-Z0-9-]*$`)
	secretsManagerARNPattern = regexp.MustCompile(`^arn:aws:secretsmanager:[a-z0-9-]+:[0-9]{12}:secret:[a-zA-Z0-9/_+=.@-]+$`)
)

// ValidateClusterARN checks if your cluster ARN is legit -- empty gets a pass,
// but a bogus venue means the gig's off before it starts.
func ValidateClusterARN(arn string) error {
	if arn == "" {
		return nil
	}

	if !rdsClusterARNPattern.MatchString(arn) {
		return fmt.Errorf("invalid RDS cluster ARN format: %s\nExpected format: arn:aws:rds:<region>:<account-id>:cluster:<cluster-name>", arn)
	}

	return nil
}

// ValidateSecretARN checks if your secret ARN passes muster -- empty is fine,
// but fake creds will get you slapped hard.
func ValidateSecretARN(arn string) error {
	if arn == "" {
		return nil
	}

	if !secretsManagerARNPattern.MatchString(arn) {
		return fmt.Errorf("invalid Secrets Manager ARN format: %s\nExpected format: arn:aws:secretsmanager:<region>:<account-id>:secret:<secret-name>", arn)
	}

	return nil
}

// ConnectionInfo is your rider for the gig.
//
//   - ClusterARN is the venue
//   - SecretARN gets you through the door
//   - Database is which stage you're playing
type ClientConfig struct {
	ClusterARN string
	SecretARN  string
	Database   string
}

// Validate checks if you've got venue, credentials, and stage sorted --
// can't start the show without all three.
func (c *ClientConfig) Validate() error {
	if c.ClusterARN == "" {
		return errors.New("missing ClusterARN")
	}

	if err := ValidateClusterARN(c.ClusterARN); err != nil {
		return err
	}

	if c.SecretARN == "" {
		return errors.New("missing SecretARN")
	}

	if err := ValidateSecretARN(c.SecretARN); err != nil {
		return err
	}

	if c.Database == "" {
		return errors.New("missing Database")
	}

	return nil
}
