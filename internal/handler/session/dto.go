package session

import "github.com/sekai-labs/michibiki/pkg/config"

type SessionConnectRequest struct {
	DeviceName    string
	DirectURL     string
	TokenFilePath string
	Insecure      bool
}

type SessionResult struct {
	Profile *config.DeviceProfile
}
