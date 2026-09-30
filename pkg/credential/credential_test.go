package credential

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
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
func TestVault_SaltAndFormat(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath1 := filepath.Join(tempDir, "vault1.enc")
	vaultPath2 := filepath.Join(tempDir, "vault2.enc")
	passphrase := "identical-passphrase"

	store1 := NewVaultStore(vaultPath1, passphrase)
	store2 := NewVaultStore(vaultPath2, passphrase)

	creds := Credentials{Username: "admin", Password: "secret"}
	if err := store1.Set("dev1", creds); err != nil {
		t.Fatalf("failed to set in store1: %v", err)
	}
	if err := store2.Set("dev1", creds); err != nil {
		t.Fatalf("failed to set in store2: %v", err)
	}

	data1, err := os.ReadFile(vaultPath1)
	if err != nil {
		t.Fatalf("failed to read vault1: %v", err)
	}
	data2, err := os.ReadFile(vaultPath2)
	if err != nil {
		t.Fatalf("failed to read vault2: %v", err)
	}

	if len(data1) < 5 || string(data1[:4]) != "MBKV" || data1[4] != 0x02 {
		t.Fatalf("vault1 missing MBKV header or version 2")
	}

	if bytes.Equal(data1, data2) {
		t.Fatalf("identical passphrases and content resulted in identical ciphertext; random salt is not functioning properly")
	}

	salt1 := data1[5 : 5+16]
	salt2 := data2[5 : 5+16]
	if bytes.Equal(salt1, salt2) {
		t.Fatalf("salts must not match across vaults")
	}
}

func TestVault_LegacyV1BackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "legacy_vault.enc")
	passphrase := "legacy-passphrase"

	h := sha256.Sum256([]byte(passphrase))
	key := h[:]
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("failed to create gcm: %v", err)
	}

	plaintext := []byte(`{"router1":{"username":"legacy_user","password":"legacy_password"}}`)
	nonce := make([]byte, gcm.NonceSize())
	for i := range nonce {
		nonce[i] = byte(i)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	if err := os.WriteFile(vaultPath, ciphertext, 0600); err != nil {
		t.Fatalf("failed to write legacy vault: %v", err)
	}

	store := NewVaultStore(vaultPath, passphrase)
	creds, err := store.Get("router1")
	if err != nil {
		t.Fatalf("failed to read legacy vault: %v", err)
	}
	if creds.Username != "legacy_user" || creds.Password != "legacy_password" {
		t.Fatalf("unexpected legacy credentials: %+v", creds)
	}

	if err := store.Set("router2", Credentials{Username: "new_user"}); err != nil {
		t.Fatalf("failed to write upgraded vault: %v", err)
	}

	upgradedData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("failed to read upgraded vault: %v", err)
	}
	if len(upgradedData) < 5 || string(upgradedData[:4]) != "MBKV" {
		t.Fatalf("expected upgraded vault to use MBKV format")
	}
	c1, err := store.Get("router1")
	if err != nil || c1.Username != "legacy_user" {
		t.Fatalf("failed to read router1 from upgraded vault: %v", err)
	}
	c2, err := store.Get("router2")
	if err != nil || c2.Username != "new_user" {
		t.Fatalf("failed to read router2 from upgraded vault: %v", err)
	}
}
func TestCredentials_RedactedAndString(t *testing.T) {
	c := Credentials{
		Username:   "admin",
		Password:   "plain_secret_password",
		APIKey:     "plain_api_key",
		APISecret:  "plain_api_secret",
		Token:      "plain_token",
		SSHKeyPath: "/home/user/.ssh/id_ed25519",
		Custom:     map[string]string{"secret_header": "sensitive_data"},
	}

	str := c.String()
	if strings.Contains(str, "plain_secret_password") ||
		strings.Contains(str, "plain_api_key") ||
		strings.Contains(str, "plain_api_secret") ||
		strings.Contains(str, "plain_token") ||
		strings.Contains(str, "sensitive_data") {
		t.Fatalf("Credentials.String() leaked secrets: %s", str)
	}

	redacted := c.Redacted()
	if redacted.Password == c.Password || redacted.APIKey == c.APIKey || redacted.Token == c.Token {
		t.Fatalf("Credentials.Redacted() did not obscure secrets")
	}
	if redacted.Username != "admin" || redacted.SSHKeyPath != "/home/user/.ssh/id_ed25519" {
		t.Fatalf("Credentials.Redacted() lost non-secret fields")
	}
}
