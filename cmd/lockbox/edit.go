package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     `edit "id_or_index:newpassword"`,
	Short:   "Update the password of an existing credential",
	Example: `  lockbox edit "2:MyN3wP@ss!"` + "\n" + `  lockbox edit "a1b2c3d4:MyN3wP@ss!"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], ":", 2)
		if len(parts) != 2 || parts[1] == "" {
			return fmt.Errorf(`🚨 Formatting error! Please match: edit "index:newpassword"`)
		}
		s, err := unlockVault(false)
		if err != nil {
			return err
		}
		if len(s.vault) == 0 {
			fmt.Println("📭 Your cryptographic vault is empty.")
			return nil
		}
		cred := resolveID(s.vault, parts[0])
		if cred == nil {
			return fmt.Errorf("🚨 Credential %q not found", parts[0])
		}
		if err := s.vault.EditPasswordByID(cred.ID, parts[1]); err != nil {
			return fmt.Errorf("🚨 %w", err)
		}
		if err := s.save(); err != nil {
			return fmt.Errorf("🚨 Failed to commit changes to file: %w", err)
		}
		fmt.Printf("🛡️ Password updated for credential [%d] %s\n",
			displayPosition(s.vault, cred.ID), cred.Account)
		return nil
	},
}
