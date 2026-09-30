package auth

import (
	"os"
	"strings"

	"github.com/sekai-labs/michibiki/pkg/credential"
)

func (a *AuthHandler) SessionStatus() (*SessionStatusInfo, error) {
	info, err := os.Stat(a.sessionFilePath)
	if os.IsNotExist(err) {
		return &SessionStatusInfo{
			Active: false,
			Path:   a.sessionFilePath,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	creds, err := a.LoadTempEncryptedToken("")
	if err != nil {
		return &SessionStatusInfo{
			Active: false,
			Path:   a.sessionFilePath,
		}, nil
	}

	hint := ObfuscateCredential(creds)
	return &SessionStatusInfo{
		Active:    true,
		TokenHint: hint,
		Path:      a.sessionFilePath,
		CreatedAt: info.ModTime(),
	}, nil
}

func ObfuscateCredential(creds *credential.Credentials) string {
	if creds == nil {
		return ""
	}
	raw := creds.Token
	if raw == "" {
		if creds.APIKey != "" {
			raw = creds.APIKey
		} else if creds.Password != "" {
			raw = creds.Password
		}
	}
	if raw == "" {
		return "configured"
	}
	if len(raw) <= 8 {
		return "••••••••"
	}
	var b strings.Builder
	b.WriteString(raw[:4])
	b.WriteString("...")
	b.WriteString(raw[len(raw)-4:])
	return b.String()
}
