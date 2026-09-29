package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lockbox/internal/crypto"
)

var changeMasterCmd = &cobra.Command{
	Use:     "change-master",
	Short:   "Change the master password for the entire vault",
	Long:    "Change the master password for the entire vault. Prompts for the current password, then a new password twice. The vault is re-encrypted with a fresh salt and a key derived from the new password.",
	Example: `  lockbox change-master`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(cfg.VaultPath); err != nil {
			return fmt.Errorf("🚨 No vault found at %s", cfg.VaultPath)
		}
		s, err := unlockVault(false)
		if err != nil {
			return err
		}
		newPassword, err := readNewMasterPassword()
		if err != nil {
			return err
		}
		newSalt, err := crypto.GenerateSalt()
		if err != nil {
			return fmt.Errorf("🚨 Failed to generate new salt: %w", err)
		}
		newKey := crypto.DeriveKey(newPassword, newSalt, s.iterations)
		if err := s.storage.Save(s.vault, newKey, newSalt, s.iterations); err != nil {
			return fmt.Errorf("🚨 Failed to commit changes to file: %w", err)
		}
		fmt.Println("🛡️ Master password changed. Vault re-encrypted with fresh salt.")
		return nil
	},
}
