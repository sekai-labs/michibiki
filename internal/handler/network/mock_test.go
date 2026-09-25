package network_test

import (
	"context"

	"github.com/sekai-labs/michibiki/internal/port"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var _ port.ProviderPort = (*mockProviderPort)(nil)

type mockProviderPort struct {
	caps           provider.Capabilities
	sysInfo        *model.SystemInfo
	ifaces         []model.Interface
	stats          []model.InterfaceStats
	routes         []model.Route
	gateways       []model.Gateway
	bgpNeighbors   []model.BGPNeighbor
	fwRules        []model.FirewallRule
	fwAliases      []model.FirewallAlias
	natRules       []model.NATRule
	dhcpLeases     []model.DHCPLease
	wgPeers        []model.WireGuardPeer
	arpEntries     []model.ARPEntry
	runningConfig  string
	validateResult *model.ValidationResult
	applyResult    *model.ConfigApplyResult
}

func (m *mockProviderPort) ID() string   { return "mock" }
func (m *mockProviderPort) Name() string { return "Mock Provider" }
func (m *mockProviderPort) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	return nil
}
func (m *mockProviderPort) Disconnect(ctx context.Context) error { return nil }
func (m *mockProviderPort) Capabilities() provider.Capabilities   { return m.caps }

func (m *mockProviderPort) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	if !m.caps.Has(provider.CapSystem) {
		return nil, provider.ErrCapabilityNotSupported
	}
	return m.sysInfo, nil
}

func (m *mockProviderPort) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	if !m.caps.Has(provider.CapInterfaces) {
		return nil, provider.ErrCapabilityNotSupported
	}
	return m.ifaces, nil
}

func (m *mockProviderPort) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	for _, iface := range m.ifaces {
		if iface.Name == name {
			return &iface, nil
		}
	}
	return nil, nil
}

func (m *mockProviderPort) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	if !m.caps.Has(provider.CapMonitoring) {
		return nil, provider.ErrCapabilityNotSupported
	}
	return m.stats, nil
}

func (m *mockProviderPort) ListRoutes(ctx context.Context) ([]model.Route, error) {
	if !m.caps.Has(provider.CapRouting) {
		return nil, provider.ErrCapabilityNotSupported
	}
	return m.routes, nil
}

func (m *mockProviderPort) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return m.gateways, nil
}

func (m *mockProviderPort) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return m.bgpNeighbors, nil
}

func (m *mockProviderPort) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	if !m.caps.Has(provider.CapFirewall) {
		return nil, provider.ErrCapabilityNotSupported
	}
	return m.fwRules, nil
}

func (m *mockProviderPort) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return m.fwAliases, nil
}

func (m *mockProviderPort) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return m.natRules, nil
}

func (m *mockProviderPort) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return m.dhcpLeases, nil
}

func (m *mockProviderPort) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return m.wgPeers, nil
}

func (m *mockProviderPort) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return m.arpEntries, nil
}

func (m *mockProviderPort) GetRunningConfig(ctx context.Context) (string, error) {
	if !m.caps.Has(provider.CapConfig) {
		return "", provider.ErrCapabilityNotSupported
	}
	return m.runningConfig, nil
}

func (m *mockProviderPort) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return m.validateResult, nil
}

func (m *mockProviderPort) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return m.applyResult, nil
}

func (m *mockProviderPort) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}

func (m *mockProviderPort) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	return nil, nil
}
