package credential

import (
	"bytes"
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

	"golang.org/x/crypto/argon2"
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
func (c Credentials) String() string {
	var sb strings.Builder
	sb.WriteString("Credentials{")
	first := true
	addPart := func(part string) {
		if !first {
			sb.WriteString(", ")
		}
		first = false
		sb.WriteString(part)
	}

	if c.Username != "" {
		addPart("username=" + c.Username)
	}
	if c.Password != "" {
		addPart("password=REDACTED")
	}
	if c.APIKey != "" {
		addPart("api_key=REDACTED")
	}
	if c.APISecret != "" {
		addPart("api_secret=REDACTED")
	}
	if c.Token != "" {
		addPart("token=REDACTED")
	}
	if c.SSHKeyPath != "" {
		addPart("ssh_key_path=" + c.SSHKeyPath)
	}
	if len(c.Custom) > 0 {
		addPart(fmt.Sprintf("custom=[%d keys]", len(c.Custom)))
	}
	sb.WriteByte('}')
	return sb.String()
}
func (c Credentials) Redacted() Credentials {
	clone := c
	if clone.Password != "" {
		clone.Password = "••••••••"
	}
	if clone.APIKey != "" {
		clone.APIKey = "••••••••"
	}
	if clone.APISecret != "" {
		clone.APISecret = "••••••••"
	}
	if clone.Token != "" {
		clone.Token = "••••••••"
	}
	if clone.Custom != nil {
		clone.Custom = make(map[string]string)
		for k := range c.Custom {
			clone.Custom[k] = "••••••••"
		}
	}
	return clone
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

const (
	vaultMagicV2 = "MBKV"
	vaultVersion2 = 0x02
	saltSize     = 16
)

func deriveKeyArgon2id(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, 3, 64*1024, 2, 32)
}

func (v *VaultStore) deriveKeyLegacy() []byte {
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

	var key []byte
	var nonce []byte
	var ciphertext []byte

	if len(data) >= 4+1+saltSize && bytes.Equal(data[:4], []byte(vaultMagicV2)) && data[4] == vaultVersion2 {
		salt := data[5 : 5+saltSize]
		key = deriveKeyArgon2id(v.passphrase, salt)
		offset := 5 + saltSize

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		nonceSize := aesGCM.NonceSize()
		if len(data) < offset+nonceSize {
			return nil, ErrInvalidFormat
		}
		nonce = data[offset : offset+nonceSize]
		ciphertext = data[offset+nonceSize:]

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
	key = v.deriveKeyLegacy()
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

	nonce, ciphertext = data[:nonceSize], data[nonceSize:]
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

	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("failed to generate random salt: %w", err)
	}

	key := deriveKeyArgon2id(v.passphrase, salt)
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

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	var buf bytes.Buffer
	buf.WriteString(vaultMagicV2)
	buf.WriteByte(vaultVersion2)
	buf.Write(salt)
	buf.Write(nonce)
	buf.Write(ciphertext)

	return os.WriteFile(v.path, buf.Bytes(), 0600)
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
