package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	EnvFile         = ".env"
	EnvKeyUsername  = "KCDEV_USERNAME"
	EnvKeyPassword  = "KCDEV_PASSWORD"
)

type EnvConfig struct {
	Username string
	Password string
}

func LoadEnv(projectDir string) (*EnvConfig, error) {
	envPath := filepath.Join(projectDir, EnvFile)

	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return nil, err
		}
	}

	return &EnvConfig{
		Username: os.Getenv(EnvKeyUsername),
		Password: os.Getenv(EnvKeyPassword),
	}, nil
}

func (e *EnvConfig) HasAuth() bool {
	return e.Username != "" && e.Password != ""
}

func EnvExists(projectDir string) bool {
	_, err := os.Stat(filepath.Join(projectDir, EnvFile))
	return err == nil
}
