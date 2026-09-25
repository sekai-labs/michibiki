package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DeviceProfile struct {
	Name          string            `json:"name" yaml:"name"`
	Provider      string            `json:"provider" yaml:"provider"`
	Address       string            `json:"address" yaml:"address"`
	Port          int               `json:"port,omitempty" yaml:"port,omitempty"`
	Insecure      bool              `json:"insecure,omitempty" yaml:"insecure,omitempty"`
	CredentialRef string            `json:"credential_ref,omitempty" yaml:"credential_ref,omitempty"`
	Options       map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

type Config struct {
	DefaultDevice string                   `json:"default_device,omitempty" yaml:"default_device,omitempty"`
	Devices       map[string]DeviceProfile `json:"devices" yaml:"devices"`
	PluginDir     string                   `json:"plugin_dir,omitempty" yaml:"plugin_dir,omitempty"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "michibiki.yaml"
	}
	return filepath.Join(home, ".config", "michibiki", "config.yaml")
}

func DefaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "michibiki.vault"
	}
	return filepath.Join(home, ".config", "michibiki", "vault.enc")
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Devices: make(map[string]DeviceProfile),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Devices == nil {
		cfg.Devices = make(map[string]DeviceProfile)
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) GetDevice(name string) (*DeviceProfile, error) {
	if name == "" {
		name = c.DefaultDevice
	}
	if name == "" {
		return nil, errors.New("no device specified and no default device configured")
	}

	d, exists := c.Devices[name]
	if !exists {
		return nil, errors.New("device profile not found: " + name)
	}
	return &d, nil
}
