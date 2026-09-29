package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listShowPassword bool
	listSearch       string
	listUser         string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Decrypt and display all stored credentials",
	Long: "Decrypt and display all stored credentials with their ID, account,\n" +
		"username, password, and save date. Passwords are hidden by default.",
	Example: `  lockbox list
  lockbox list --show-password
  lockbox list --search github
  lockbox list --user alice`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := unlockVault(false)
		if err != nil {
			return err
		}
		if len(s.vault) == 0 {
			fmt.Println("📭 Your cryptographic vault is empty.")
			return nil
		}
		filtered := s.vault.Filter(listSearch, listUser)
		if len(filtered) == 0 {
			fmt.Println("📭 No credentials matched.")
			return nil
		}
		fmt.Println("\n🔐 --- SECURE DECRYPTED ACCOUNTS ---")
		for _, c := range filtered {
			pass := "********"
			if listShowPassword {
				pass = c.Password
			}
			fmt.Printf("[%d] ID: %s | Account: %s | User: %s | Pass: %s | Saved: %s\n",
				displayPosition(s.vault, c.ID), c.ID, c.Account, c.Username, pass,
				c.SavedAt.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listShowPassword, "show-password", "p", false, "show passwords")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "search account, username, or password")
	listCmd.Flags().StringVarP(&listUser, "user", "u", "", "filter by username")
}
