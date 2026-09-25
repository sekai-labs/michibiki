package port

import (
	"context"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type SecurityPort interface {
	ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error)
	ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error)
	ListNATRules(ctx context.Context) ([]model.NATRule, error)
}

type ServicesPort interface {
	ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error)
	ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error)
	ListARPEntries(ctx context.Context) ([]model.ARPEntry, error)
}

type ConfigPort interface {
	GetRunningConfig(ctx context.Context) (string, error)
	ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error)
	ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error)
	RollbackConfig(ctx context.Context, rollbackID string) error
}

type IPAMPort interface {
	GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error)
}

type ProviderPort interface {
	DevicePort
	InterfacePort
	RoutingPort
	SecurityPort
	ServicesPort
	ConfigPort
	IPAMPort
}

type TokenStorePort interface {
	ReadTokenFile(path string) (*credential.Credentials, error)
	SaveTempSessionToken(key string, creds *credential.Credentials) error
	LoadTempSessionToken(key string) (*credential.Credentials, error)
	ClearTempSessionTokens() error
}
