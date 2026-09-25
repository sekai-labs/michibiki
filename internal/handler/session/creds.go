package session

import (
	"fmt"
	"os"

	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/credential"
)

func (m *SessionHandler) ResolveCredentials(profile *config.DeviceProfile, tokenFilePath string) (*credential.Credentials, error) {
	if tokenFilePath != "" {
		creds, err := m.authHdl.ParseTokenFile(tokenFilePath)
		if err == nil {
			return creds, nil
		}
		return nil, fmt.Errorf("failed to parse specified token file '%s': %w", tokenFilePath, err)
	}

	targetName := ""
	if profile != nil {
		targetName = profile.Name
	}
	if sessionCreds, err := m.authHdl.LoadTempEncryptedToken(targetName); err == nil {
		return sessionCreds, nil
	}

	var vault *credential.VaultStore
	vaultPass := os.Getenv("MICHIBIKI_VAULT_PASSPHRASE")
	if vaultPass != "" {
		vault = credential.NewVaultStore(config.DefaultVaultPath(), vaultPass)
	}
	resolver := credential.NewResolver(vault)

	ref := ""
	name := ""
	if profile != nil {
		ref = profile.CredentialRef
		name = profile.Name
	}

	return resolver.Resolve(ref, name)
}
