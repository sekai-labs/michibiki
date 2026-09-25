package auth_test

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/sekai-labs/michibiki/internal/handler/auth"
	"github.com/sekai-labs/michibiki/internal/port"
	"github.com/sekai-labs/michibiki/pkg/credential"
)

func TestAuthHandler_ParseTokenFile(t *testing.T) {
	tempDir := t.TempDir()
	authHdlr := auth.NewAuthHandler()

	rawFile := filepath.Join(tempDir, "raw_token.txt")
	if err := os.WriteFile(rawFile, []byte("tok_secret_value_12345\n"), 0600); err != nil {
		t.Fatalf("failed to create raw token file: %v", err)
	}
	creds, err := authHdlr.ParseTokenFile(rawFile)
	if err != nil {
		t.Fatalf("ParseTokenFile failed for raw token: %v", err)
	}
	if creds.Token != "tok_secret_value_12345" {
		t.Fatalf("expected token 'tok_secret_value_12345', got '%s'", creds.Token)
	}

	pairFile := filepath.Join(tempDir, "pair_token.txt")
	if err := os.WriteFile(pairFile, []byte("my_api_key:my_secret_key\n"), 0600); err != nil {
		t.Fatalf("failed to create pair file: %v", err)
	}
	creds, err = authHdlr.ParseTokenFile(pairFile)
	if err != nil {
		t.Fatalf("ParseTokenFile failed for pair: %v", err)
	}
	if creds.APIKey != "my_api_key" || creds.APISecret != "my_secret_key" {
		t.Fatalf("expected key 'my_api_key' and secret 'my_secret_key', got key='%s', secret='%s'", creds.APIKey, creds.APISecret)
	}

	jsonFile := filepath.Join(tempDir, "creds.json")
	jsonContent := `{"token": "json_secret_token", "username": "admin"}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("failed to create json file: %v", err)
	}
	creds, err = authHdlr.ParseTokenFile(jsonFile)
	if err != nil {
		t.Fatalf("ParseTokenFile failed for json: %v", err)
	}
	if creds.Token != "json_secret_token" || creds.Username != "admin" {
		t.Fatalf("unexpected json credentials: %+v", creds)
	}

	opnsenseFile := filepath.Join(tempDir, "opnsense_apikey.txt")
	opnContent := "key=oW42JfTI50LPfki3abbnLmQVxYseY7A9zP8CrW6\nsecret=YwfzeAjqWAONzlE9vGq7zUvRFhEhPmkeBbeV4VK4\n"
	if err := os.WriteFile(opnsenseFile, []byte(opnContent), 0600); err != nil {
		t.Fatalf("failed to create opnsense file: %v", err)
	}
	creds, err = authHdlr.ParseTokenFile(opnsenseFile)
	if err != nil {
		t.Fatalf("ParseTokenFile failed for opnsense file: %v", err)
	}
	if creds.APIKey != "oW42JfTI50LPfki3abbnLmQVxYseY7A9zP8CrW6" || creds.APISecret != "YwfzeAjqWAONzlE9vGq7zUvRFhEhPmkeBbeV4VK4" {
		t.Fatalf("unexpected opnsense credentials: %+v", creds)
	}
}

func TestAuthHandler_EncryptedSessionLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	sessionFile := filepath.Join(tempDir, "michibiki_test_session.enc")
	authHdlr := auth.NewAuthHandlerWithPath(sessionFile)

	status, err := authHdlr.SessionStatus()
	if err != nil {
		t.Fatalf("SessionStatus returned error: %v", err)
	}
	if status.Active {
		t.Fatalf("expected inactive session, got active")
	}

	testCreds := &credential.Credentials{
		Token: "super-secret-production-token",
	}
	if err := authHdlr.StoreTempEncryptedToken("core-gw", testCreds); err != nil {
		t.Fatalf("StoreTempEncryptedToken failed: %v", err)
	}

	info, err := os.Stat(sessionFile)
	if err != nil {
		t.Fatalf("session file not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected permissions 0600, got %o", perm)
	}

	status, err = authHdlr.SessionStatus()
	if err != nil {
		t.Fatalf("SessionStatus failed: %v", err)
	}
	if !status.Active {
		t.Fatalf("expected session to be active")
	}
	if status.TokenHint != "supe...oken" {
		t.Fatalf("expected hint 'supe...oken', got '%s'", status.TokenHint)
	}

	loaded, err := authHdlr.LoadTempEncryptedToken("core-gw")
	if err != nil {
		t.Fatalf("LoadTempEncryptedToken failed: %v", err)
	}
	if loaded.Token != testCreds.Token {
		t.Fatalf("expected token '%s', got '%s'", testCreds.Token, loaded.Token)
	}

	_, err = authHdlr.LoadTempEncryptedToken("other-router")
	if err == nil {
		t.Fatalf("expected error when loading session token for mismatched device, got nil")
	}

	var tokenStore port.TokenStorePort = authHdlr
	_, err = tokenStore.LoadTempSessionToken("core-gw")
	if err != nil {
		t.Fatalf("TokenStorePort interface load failed: %v", err)
	}

	if err := authHdlr.ClearTempEncryptedTokens(); err != nil {
		t.Fatalf("ClearTempEncryptedTokens failed: %v", err)
	}

	status, err = authHdlr.SessionStatus()
	if err != nil {
		t.Fatalf("SessionStatus failed after clear: %v", err)
	}
	if status.Active {
		t.Fatalf("expected session to be inactive after clear")
	}
}
