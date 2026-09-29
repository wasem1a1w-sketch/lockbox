package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var reorderCmd = &cobra.Command{
	Use:     `reorder "old_index:new_index"`,
	Short:   "Move a credential to a new display position",
	Long:    "Move a credential to a new display position. Other credentials shift to fill the gap.",
	Example: `  lockbox reorder "3:1"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf(`🚨 Formatting error! Please match: reorder "old_index:new_index"`)
		}
		oldIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("🚨 Old index must be a number, got %q", parts[0])
		}
		newIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("🚨 New index must be a number, got %q", parts[1])
		}
		s, err := unlockVault(false)
		if err != nil {
			return err
		}
		if len(s.vault) == 0 {
			fmt.Println("📭 Your cryptographic vault is empty.")
			return nil
		}
		if err := s.vault.ReOrder(oldIndex, newIndex); err != nil {
			return fmt.Errorf("🚨 %w", err)
		}
		if err := s.save(); err != nil {
			return fmt.Errorf("🚨 Failed to commit changes to file: %w", err)
		}
		fmt.Printf("🛡️ Credential moved to position %d\n", newIndex)
		return nil
	},
}
