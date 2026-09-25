package auth

import (
	"time"

	"github.com/sekai-labs/michibiki/pkg/credential"
)

type SessionPayload struct {
	Device      string                 `json:"device"`
	Credentials credential.Credentials `json:"credentials"`
	CreatedAt   time.Time              `json:"created_at"`
}

type SessionStatusInfo struct {
	Active    bool      `json:"active"`
	Device    string    `json:"device"`
	TokenHint string    `json:"token_hint"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}
