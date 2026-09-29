package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     `add "domain:username:password"`,
	Short:   "Add a new credential",
	Example: `  lockbox add "example.com:alice:myP@ss!"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		account, user, pass, err := parseCredential(args[0])
		if err != nil {
			return err
		}
		s, err := unlockVault(true)
		if err != nil {
			return err
		}
		s.vault.Add(account, user, pass)
		if err := s.save(); err != nil {
			return fmt.Errorf("🚨 Failed to commit changes to file: %w", err)
		}
		fmt.Printf("🛡️ Account profile successfully encrypted for: %s\n", account)
		return nil
	},
}

func parseCredential(raw string) (string, string, string, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf(`🚨 Formatting error! Please match: add "service.com:user:pass"`)
	}
	return parts[0], parts[1], parts[2], nil
}
