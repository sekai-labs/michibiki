package credential

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound      = errors.New("credential not found")
	ErrVaultLocked   = errors.New("vault is locked or key invalid")
	ErrInvalidFormat = errors.New("invalid credential format")
)

type Credentials struct {
	Username   string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password   string            `json:"password,omitempty" yaml:"password,omitempty"`
	APIKey     string            `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	APISecret  string            `json:"api_secret,omitempty" yaml:"api_secret,omitempty"`
	Token      string            `json:"token,omitempty" yaml:"token,omitempty"`
	SSHKeyPath string            `json:"ssh_key_path,omitempty" yaml:"ssh_key_path,omitempty"`
	Custom     map[string]string `json:"custom,omitempty" yaml:"custom,omitempty"`
}

type Store interface {
	Get(key string) (*Credentials, error)
	Set(key string, creds Credentials) error
	Delete(key string) error
	List() ([]string, error)
}

type VaultStore struct {
	path       string
	passphrase string
}

func NewVaultStore(path string, passphrase string) *VaultStore {
	return &VaultStore{
		path:       path,
		passphrase: passphrase,
	}
}

func (v *VaultStore) deriveKey() []byte {
	h := sha256.Sum256([]byte(v.passphrase))
	return h[:]
}

func (v *VaultStore) readAll() (map[string]Credentials, error) {
	if _, err := os.Stat(v.path); os.IsNotExist(err) {
		return make(map[string]Credentials), nil
	}

	data, err := os.ReadFile(v.path)
	if err != nil {
		return nil, err
	}

	if len(data) < 12 {
		return nil, ErrInvalidFormat
	}

	key := v.deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidFormat
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrVaultLocked
	}

	var m map[string]Credentials
	if err := json.Unmarshal(plaintext, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (v *VaultStore) writeAll(m map[string]Credentials) error {
	dir := filepath.Dir(v.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	plaintext, err := json.Marshal(m)
	if err != nil {
		return err
	}

	key := v.deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return os.WriteFile(v.path, ciphertext, 0600)
}

func (v *VaultStore) Get(key string) (*Credentials, error) {
	m, err := v.readAll()
	if err != nil {
		return nil, err
	}
	creds, exists := m[key]
	if !exists {
		return nil, ErrNotFound
	}
	return &creds, nil
}

func (v *VaultStore) Set(key string, creds Credentials) error {
	m, err := v.readAll()
	if err != nil {
		return err
	}
	m[key] = creds
	return v.writeAll(m)
}

func (v *VaultStore) Delete(key string) error {
	m, err := v.readAll()
	if err != nil {
		return err
	}
	delete(m, key)
	return v.writeAll(m)
}

func (v *VaultStore) List() ([]string, error) {
	m, err := v.readAll()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys, nil
}

type Resolver struct {
	vault *VaultStore
}

func NewResolver(vault *VaultStore) *Resolver {
	return &Resolver{vault: vault}
}

func (r *Resolver) Resolve(ref string, deviceName string) (*Credentials, error) {
	if ref == "" {
		envPrefix := "MICHIBIKI_" + strings.ToUpper(strings.ReplaceAll(deviceName, "-", "_")) + "_"
		username := os.Getenv(envPrefix + "USERNAME")
		password := os.Getenv(envPrefix + "PASSWORD")
		apiKey := os.Getenv(envPrefix + "API_KEY")
		apiSecret := os.Getenv(envPrefix + "API_SECRET")
		token := os.Getenv(envPrefix + "TOKEN")
		sshKey := os.Getenv(envPrefix + "SSH_KEY")

		if username != "" || password != "" || apiKey != "" || token != "" || sshKey != "" {
			return &Credentials{
				Username:   username,
				Password:   password,
				APIKey:     apiKey,
				APISecret:  apiSecret,
				Token:      token,
				SSHKeyPath: sshKey,
			}, nil
		}
	}

	if strings.HasPrefix(ref, "env:") {
		varName := strings.TrimPrefix(ref, "env:")
		val := os.Getenv(varName)
		if val == "" {
			return nil, fmt.Errorf("env var %s not set", varName)
		}
		var creds Credentials
		if err := json.Unmarshal([]byte(val), &creds); err == nil {
			return &creds, nil
		}
		return &Credentials{Password: val, Token: val}, nil
	}

	if strings.HasPrefix(ref, "vault:") {
		vaultKey := strings.TrimPrefix(ref, "vault:")
		if r.vault == nil {
			return nil, errors.New("vault not configured")
		}
		return r.vault.Get(vaultKey)
	}

	if r.vault != nil {
		c, err := r.vault.Get(deviceName)
		if err == nil {
			return c, nil
		}
	}

	return &Credentials{}, nil
}
