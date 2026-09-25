package credential

import (
	"path/filepath"
	"testing"
)

func TestVaultEncryptDecrypt(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	passphrase := "secret-passphrase-123"

	store := NewVaultStore(vaultPath, passphrase)

	creds := Credentials{
		Username:  "admin",
		Password:  "P@ssw0rd1",
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	if err := store.Set("router1", creds); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	readCreds, err := store.Get("router1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if readCreds.Username != "admin" || readCreds.Password != "P@ssw0rd1" {
		t.Errorf("read credentials mismatch: %+v", readCreds)
	}

	badStore := NewVaultStore(vaultPath, "wrong-passphrase")
	_, err = badStore.Get("router1")
	if err == nil {
		t.Fatal("expected error with wrong passphrase, got nil")
	}

	keys, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(keys) != 1 || keys[0] != "router1" {
		t.Fatalf("unexpected keys list: %v", keys)
	}

	if err := store.Delete("router1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get("router1")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}
