package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <id_or_index>",
	Short:   "Delete a credential by its ID or display index",
	Example: `  lockbox delete 2` + "\n" + `  lockbox delete "a1b2c3d4"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := unlockVault(false)
		if err != nil {
			return err
		}
		if len(s.vault) == 0 {
			fmt.Println("📭 Your cryptographic vault is empty.")
			return nil
		}
		cred := resolveID(s.vault, args[0])
		if cred == nil {
			return fmt.Errorf("🚨 Credential %q not found", args[0])
		}
		account := cred.Account
		if err := s.vault.DeleteByID(cred.ID); err != nil {
			return fmt.Errorf("🚨 %w", err)
		}
		if err := s.save(); err != nil {
			return fmt.Errorf("🚨 Failed to commit changes to file: %w", err)
		}
		fmt.Printf("🗑️ Credential %s deleted successfully\n", account)
		return nil
	},
}
