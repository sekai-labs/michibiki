package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveConfig_Permissions(t *testing.T) {
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "subconfig")
	confPath := filepath.Join(confDir, "config.yaml")

	cfg := &Config{
		DefaultDevice: "router1",
		Devices: map[string]DeviceProfile{
			"router1": {
				Name:    "router1",
				Address: "192.168.1.1",
			},
		},
	}

	if err := SaveConfig(confPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	dirInfo, err := os.Stat(confDir)
	if err != nil {
		t.Fatalf("failed to stat config dir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Fatalf("expected config dir permission 0700, got %o", perm)
	}

	fileInfo, err := os.Stat(confPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected config file permission 0600, got %o", perm)
	}

	loaded, err := LoadConfig(confPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.DefaultDevice != "router1" {
		t.Fatalf("expected default device router1, got %s", loaded.DefaultDevice)
	}
}
