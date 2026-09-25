package auth

import "github.com/sekai-labs/michibiki/pkg/credential"

type TokenParseRequest struct {
	FilePath string
}

type TokenStoreRequest struct {
	Device      string
	Credentials *credential.Credentials
}

type TokenLoadRequest struct {
	Device string
}
