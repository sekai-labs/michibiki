package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sekai-labs/michibiki/pkg/credential"
)

var (
	ErrInvalidTokenFile = errors.New("invalid or empty token file")
)

func (a *AuthHandler) ParseTokenFile(path string) (*credential.Credentials, error) {
	if strings.TrimSpace(path) == "" {
		return nil, ErrInvalidTokenFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file '%s': %w", path, err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, ErrInvalidTokenFile
	}

	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		var creds credential.Credentials
		if err := json.Unmarshal([]byte(content), &creds); err == nil {
			if creds.Token != "" || creds.APIKey != "" || creds.Password != "" || creds.Username != "" {
				return &creds, nil
			}
		}

		var m map[string]string
		if err := json.Unmarshal([]byte(content), &m); err == nil {
			c := &credential.Credentials{}
			if val, ok := m["token"]; ok {
				c.Token = val
			}
			if val, ok := m["api_key"]; ok {
				c.APIKey = val
			} else if val, ok := m["key"]; ok {
				c.APIKey = val
			}
			if val, ok := m["api_secret"]; ok {
				c.APISecret = val
			} else if val, ok := m["secret"]; ok {
				c.APISecret = val
			}
			if val, ok := m["username"]; ok {
				c.Username = val
			}
			if val, ok := m["password"]; ok {
				c.Password = val
			}
			if c.Token != "" || c.APIKey != "" || c.Password != "" || c.Username != "" {
				return c, nil
			}
		}
	}

	kvCreds := &credential.Credentials{}
	hasKV := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			k := strings.ToLower(strings.TrimSpace(parts[0]))
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			switch k {
			case "key", "api_key", "apikey":
				kvCreds.APIKey = v
				hasKV = true
			case "secret", "api_secret", "apisecret":
				kvCreds.APISecret = v
				hasKV = true
			case "token":
				kvCreds.Token = v
				hasKV = true
			case "username", "user":
				kvCreds.Username = v
				hasKV = true
			case "password", "pass":
				kvCreds.Password = v
				hasKV = true
			}
		}
	}
	if hasKV && (kvCreds.APIKey != "" || kvCreds.Token != "" || kvCreds.Username != "") {
		return kvCreds, nil
	}

	lines := strings.Split(content, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if strings.Contains(firstLine, ":") && !strings.Contains(firstLine, " ") {
		parts := strings.SplitN(firstLine, ":", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return &credential.Credentials{
				APIKey:    parts[0],
				APISecret: parts[1],
			}, nil
		}
	}

	return &credential.Credentials{
		Token: firstLine,
	}, nil
}
