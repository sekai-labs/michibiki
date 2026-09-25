package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sekai-labs/michibiki/internal/handler/auth"
	"github.com/sekai-labs/michibiki/internal/handler/session"
	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/credential"
)

func TestSessionHandler_ResolveCredentialsHierarchy(t *testing.T) {
	tempDir := t.TempDir()
	authHdlr := auth.NewAuthHandlerWithPath(filepath.Join(tempDir, "session.enc"))
	cfg := &config.Config{
		Devices: map[string]config.DeviceProfile{
			"test-gw": {
				Name:     "test-gw",
				Provider: "opnsense",
				Address:  "https://10.0.0.1",
			},
		},
	}

	sm := session.NewSessionHandler(cfg, authHdlr)

	tokenFile := filepath.Join(tempDir, "explicit.txt")
	_ = os.WriteFile(tokenFile, []byte("tok_explicit_priority_one\n"), 0600)
	prof := cfg.Devices["test-gw"]

	creds, err := sm.ResolveCredentials(&prof, tokenFile)
	if err != nil {
		t.Fatalf("failed to resolve token file: %v", err)
	}
	if creds.Token != "tok_explicit_priority_one" {
		t.Fatalf("expected explicit token, got '%s'", creds.Token)
	}

	_ = authHdlr.StoreTempEncryptedToken("test-gw", &credential.Credentials{Token: "tok_session_priority_two"})
	creds, err = sm.ResolveCredentials(&prof, "")
	if err != nil {
		t.Fatalf("failed to resolve session token: %v", err)
	}
	if creds.Token != "tok_session_priority_two" {
		t.Fatalf("expected session token, got '%s'", creds.Token)
	}
}
