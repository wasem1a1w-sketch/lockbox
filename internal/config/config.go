package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	VaultPath      string `mapstructure:"vault_path"`
	KDFIterations  int    `mapstructure:"kdf_iterations"`
	DefaultGenLen  int    `mapstructure:"default_gen_length"`
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
	v.SetDefault("default_gen_length", 20)

	v.SetEnvPrefix("LOCKBOX")
	v.AutomaticEnv()

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
	v.SetDefault("default_gen_length", 20)

	v.SetEnvPrefix("LOCKBOX")
	v.AutomaticEnv()

	v.ReadInConfig() // ignore error, file may not exist

	return v
}