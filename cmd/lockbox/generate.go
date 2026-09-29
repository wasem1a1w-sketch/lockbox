package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"lockbox/internal/vault"
)

var generateCmd = &cobra.Command{
	Use:     "generate [length]",
	Short:   "Generate a cryptographically strong random password",
	Long:    "Generate a cryptographically strong random password (upper/lower/digits/symbols). Default length comes from config (default_gen_length).",
	Example: `  lockbox generate 20`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		length := cfg.DefaultGenLen
		if len(args) == 1 {
			n, err := strconv.Atoi(args[0])
			if err != nil || n <= 0 {
				return fmt.Errorf("🚨 Invalid length %q: must be a positive integer", args[0])
			}
			length = n
		}
		password, err := vault.GeneratePassword(length)
		if err != nil {
			return fmt.Errorf("🚨 Failed to generate password: %w", err)
		}
		fmt.Printf("🔐 Generated Password: %s\n", password)
		return nil
	},
}
