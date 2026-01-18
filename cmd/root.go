// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ixti/apiska/internal/rds"
)

var rootCmd = &cobra.Command{
	Use:   "apiska",
	Short: "A skankin' TUI for the AWS RDS Data API.",
	Long: "A human-friendly interface to query your RDS cluster without " +
		"leaving the terminal. One step beyond the aws rds-data CLI.",
	RunE: run,
}

// SetVersionInfo configures version information for the CLI.
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(
		fmt.Sprintf("{{.Name}} %s (commit: %s, built: %s)\n", version, commit, date),
	)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().String("cluster-arn", "", "RDS cluster ARN")
	rootCmd.Flags().String("secret-arn", "", "Secrets Manager ARN for database credentials")
	rootCmd.Flags().String("database", "", "Database name")

	rootCmd.MarkFlagRequired("cluster-arn")
	rootCmd.MarkFlagRequired("secret-arn")
	rootCmd.MarkFlagRequired("database")

	viper.BindPFlag("cluster-arn", rootCmd.Flags().Lookup("cluster-arn"))
	viper.BindPFlag("secret-arn", rootCmd.Flags().Lookup("secret-arn"))
	viper.BindPFlag("database", rootCmd.Flags().Lookup("database"))
}

func initConfig() {
	viper.SetEnvPrefix("APISKA")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_, err := rds.NewClient(ctx, &rds.ClientConfig{
		ClusterARN: viper.GetString("cluster-arn"),
		SecretARN:  viper.GetString("secret-arn"),
		Database:   viper.GetString("database"),
	})

	if err != nil {
		return fmt.Errorf("failed to initialize AWS RDS client: %w", err)
	}

	// HIC SUNT DRACONES

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
