package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	VaultPath     string `mapstructure:"vault_path"`
	KDFIterations int    `mapstructure:"kdf_iterations"`
	DefaultGenLen int    `mapstructure:"default_gen_length"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "lockbox")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	v.SetDefault("vault_path", filepath.Join(homeDir, ".local", "share", "lockbox", "vault.lock"))
	v.SetDefault("kdf_iterations", 100000)
	v.SetDefault("default_gen_length", 12)

	v.SetEnvPrefix("LOCKBOX")
	v.AutomaticEnv()
	v.BindEnv("vault_path", "LOCKBOX_VAULT_PATH")
	v.BindEnv("kdf_iterations", "LOCKBOX_KDF_ITERATIONS")
	v.BindEnv("default_gen_length", "LOCKBOX_DEFAULT_GEN_LENGTH")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if len(cfg.VaultPath) >= 2 && cfg.VaultPath[:2] == "~/" {
		cfg.VaultPath = filepath.Join(homeDir, cfg.VaultPath[2:])
	}

	return &cfg, nil
}

func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(homeDir, ".config", "lockbox")
	return os.MkdirAll(configDir, 0700)
}

func DefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lockbox", "config.yaml")
}

func LoadViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "lockbox")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	v.SetDefault("vault_path", filepath.Join(homeDir, ".local", "share", "lockbox", "vault.lock"))
	v.SetDefault("kdf_iterations", 100000)
	v.SetDefault("default_gen_length", 12)

	v.SetEnvPrefix("LOCKBOX")
	v.AutomaticEnv()
	v.BindEnv("vault_path", "LOCKBOX_VAULT_PATH")
	v.BindEnv("kdf_iterations", "LOCKBOX_KDF_ITERATIONS")
	v.BindEnv("default_gen_length", "LOCKBOX_DEFAULT_GEN_LENGTH")

	v.ReadInConfig() // ignore error, file may not exist

	return v
}

// Set validates key/value and writes the config file (0600).
// Returns the normalized value (e.g. "~/x" expanded).
func Set(key, value string) (string, error) {
	switch key {
	case "vault_path":
		value = expandHome(value)
		if value == "" {
			return "", fmt.Errorf("vault_path cannot be empty")
		}
	case "kdf_iterations", "default_gen_length":
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return "", fmt.Errorf("%s must be a positive integer, got %q", key, value)
		}
	default:
		return "", fmt.Errorf("unknown config key %q. Valid keys: vault_path, kdf_iterations, default_gen_length", key)
	}

	if err := EnsureConfigDir(); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	v := LoadViper()
	v.SetConfigFile(DefaultConfigPath())
	v.Set(key, value)
	if err := v.WriteConfig(); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}
	if err := os.Chmod(DefaultConfigPath(), 0600); err != nil {
		return "", fmt.Errorf("failed to set config permissions: %w", err)
	}
	return value, nil
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
