package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigDir  = ".kcdev"
	ConfigFile = "config.json"
)

type Config struct {
	Kintone KintoneConfig `json:"kintone"`
	Dev     DevConfig     `json:"dev"`
	Targets TargetsConfig `json:"targets"`
}

type TargetsConfig struct {
	Desktop bool `json:"desktop"`
	Mobile  bool `json:"mobile"`
}

type KintoneConfig struct {
	Domain string     `json:"domain"`
	AppID  int        `json:"appId"`
	Auth   AuthConfig `json:"auth,omitempty"`
}

type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type DevConfig struct {
	Origin string `json:"origin"`
	Entry  string `json:"entry"`
}

func DefaultConfig() *Config {
	return &Config{
		Dev: DevConfig{
			Origin: "https://localhost:3000",
			Entry:  "/src/main.tsx",
		},
		Targets: TargetsConfig{
			Desktop: true,
			Mobile:  false,
		},
	}
}

func ConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ConfigDir, ConfigFile)
}

func Load(projectDir string) (*Config, error) {
	path := ConfigPath(projectDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save(projectDir string) error {
	path := ConfigPath(projectDir)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Exists(projectDir string) bool {
	_, err := os.Stat(ConfigPath(projectDir))
	return err == nil
}
