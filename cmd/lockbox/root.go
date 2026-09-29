package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lockbox/internal/config"
)

var cfg *config.Config
var iterationsExplicit bool

var (
	vaultPathFlag  string
	iterationsFlag int
)

var rootCmd = &cobra.Command{
	Use:   "lockbox",
	Short: "A CLI password vault with AES-256-GCM encrypted storage",
	Long: "lockbox stores credentials in an AES-256-GCM encrypted file on disk.\n" +
		"Master password derives the encryption key via PBKDF2 (100,000 SHA-256 iterations).",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("🚨 Config load failed: %w", err)
		}
		if cmd.Flags().Changed("vault") {
			c.VaultPath = vaultPathFlag
		}
		if cmd.Flags().Changed("iterations") {
			c.KDFIterations = iterationsFlag
			iterationsExplicit = true
		}
		cfg = c
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultPathFlag, "vault", "", "override vault file location")
	rootCmd.PersistentFlags().IntVar(&iterationsFlag, "iterations", 0, "override PBKDF2 iterations")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(reorderCmd)
	rootCmd.AddCommand(changeMasterCmd)
	rootCmd.AddCommand(configCmd)
}
