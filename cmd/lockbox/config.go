package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lockbox/internal/config"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Manage configuration settings",
	Example: `  lockbox config show` + "\n" + `  lockbox config set vault_path ~/my-vault.lock`,
}

var configShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show resolved configuration values",
	Example: `  lockbox config show`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("📁 Config file: %s\n", config.DefaultConfigPath())
		fmt.Printf("vault_path: %s\n", cfg.VaultPath)
		fmt.Printf("kdf_iterations: %d\n", cfg.KDFIterations)
		fmt.Printf("default_gen_length: %d\n", cfg.DefaultGenLen)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:     "set <key> <value>",
	Short:   "Set a configuration value",
	Long:    "Set a configuration value. Keys: vault_path, kdf_iterations, default_gen_length.",
	Example: `  lockbox config set vault_path ~/my-vault.lock` + "\n" + `  lockbox config set kdf_iterations 100000`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		normalized, err := config.Set(key, value)
		if err != nil {
			return fmt.Errorf("🚨 %w", err)
		}
		fmt.Printf("✅ %s = %s\n", key, normalized)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}
