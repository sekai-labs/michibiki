package auth

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
	"time"

	"github.com/sekai-labs/michibiki/internal/port"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"golang.org/x/crypto/hkdf"
)

var (
	ErrNoActiveSession = errors.New("no active encrypted token session found")
)

type AuthHandler struct {
	sessionFilePath string
	sessionKeyPath  string
	salt            string
}

func getSessionKeyPath(sessionFilePath string) string {
	dir := filepath.Dir(sessionFilePath)
	return filepath.Join(dir, "session_key")
}

func NewAuthHandler() *AuthHandler {
	var sessionPath string
	if cacheDir, err := os.UserCacheDir(); err == nil && cacheDir != "" {
		sessionPath = filepath.Join(cacheDir, "michibiki", "session_token.enc")
	} else {
		sessionPath = filepath.Join(os.TempDir(), fmt.Sprintf(".michibiki_session_%d.enc", os.Getuid()))
	}
	return &AuthHandler{
		sessionFilePath: sessionPath,
		sessionKeyPath:  getSessionKeyPath(sessionPath),
		salt:            "michibiki-secret-session-salt-v1",
	}
}

func NewAuthHandlerWithPath(sessionPath string) *AuthHandler {
	return &AuthHandler{
		sessionFilePath: sessionPath,
		sessionKeyPath:  getSessionKeyPath(sessionPath),
		salt:            "michibiki-secret-session-salt-v1",
	}
}

func (a *AuthHandler) GetTempSessionPath() string {
	return a.sessionFilePath
}

func (a *AuthHandler) getOrCreateSessionKeySecret() ([]byte, error) {
	keyDir := filepath.Dir(a.sessionKeyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session key directory: %w", err)
	}

	secret, err := os.ReadFile(a.sessionKeyPath)
	if err == nil && len(secret) >= 32 {
		return secret, nil
	}
	newSecret := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newSecret); err != nil {
		return nil, fmt.Errorf("failed to generate random session secret: %w", err)
	}

	if err := os.WriteFile(a.sessionKeyPath, newSecret, 0600); err != nil {
		return nil, fmt.Errorf("failed to persist session secret file: %w", err)
	}

	return newSecret, nil
}

func (a *AuthHandler) deriveMachineKey() ([]byte, error) {
	secret, err := a.getOrCreateSessionKeySecret()
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	uid := os.Getuid()
	info := fmt.Sprintf("michibiki-session:%s:%d", hostname, uid)

	h := hkdf.New(sha256.New, secret, []byte(a.salt), []byte(info))
	derived := make([]byte, 32)
	if _, err := io.ReadFull(h, derived); err != nil {
		return nil, fmt.Errorf("failed to derive session encryption key: %w", err)
	}
	return derived, nil
}

func (a *AuthHandler) StoreTempEncryptedToken(deviceName string, creds *credential.Credentials) error {
	if creds == nil {
		return errors.New("cannot store nil credentials")
	}

	payload := SessionPayload{
		Device:      deviceName,
		Credentials: *creds,
		CreatedAt:   time.Now(),
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode session payload: %w", err)
	}

	key, err := a.deriveMachineKey()
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	dir := filepath.Dir(a.sessionFilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	if err := os.WriteFile(a.sessionFilePath, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted session token: %w", err)
	}

	return nil
}

func (a *AuthHandler) LoadTempEncryptedToken(deviceName string) (*credential.Credentials, error) {
	if _, err := os.Stat(a.sessionFilePath); os.IsNotExist(err) {
		return nil, ErrNoActiveSession
	}

	data, err := os.ReadFile(a.sessionFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session token file: %w", err)
	}

	key, err := a.deriveMachineKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("corrupted session token file")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session token: %w", err)
	}

	var payload SessionPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode session token payload: %w", err)
	}

	if deviceName != "" && payload.Device != "" && payload.Device != deviceName {
		return nil, fmt.Errorf("session token is for device '%s', not '%s'", payload.Device, deviceName)
	}

	return &payload.Credentials, nil
}

func (a *AuthHandler) ClearTempEncryptedTokens() error {
	if _, err := os.Stat(a.sessionFilePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(a.sessionFilePath)
}

var _ port.TokenStorePort = (*AuthHandler)(nil)

func (a *AuthHandler) ReadTokenFile(path string) (*credential.Credentials, error) {
	return a.ParseTokenFile(path)
}

func (a *AuthHandler) SaveTempSessionToken(key string, creds *credential.Credentials) error {
	return a.StoreTempEncryptedToken(key, creds)
}

func (a *AuthHandler) LoadTempSessionToken(key string) (*credential.Credentials, error) {
	return a.LoadTempEncryptedToken(key)
}

func (a *AuthHandler) ClearTempSessionTokens() error {
	return a.ClearTempEncryptedTokens()
}
