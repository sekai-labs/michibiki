package provider

import (
	"context"
	"errors"
	"sync"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
)

var (
	ErrCapabilityNotSupported = errors.New("capability not supported by provider")
	ErrNotConnected           = errors.New("provider not connected")
	ErrProviderNotFound       = errors.New("provider not found")
)

type Capabilities uint32

const (
	CapSystem Capabilities = 1 << iota
	CapInterfaces
	CapRouting
	CapFirewall
	CapNAT
	CapDHCP
	CapDNS
	CapWireGuard
	CapBGP
	CapOSPF
	CapMonitoring
	CapConfig
	CapIPAM
)

func (c Capabilities) Has(cap Capabilities) bool {
	return (c & cap) == cap
}

func (c Capabilities) Strings() []string {
	var list []string
	if c.Has(CapSystem) {
		list = append(list, "system")
	}
	if c.Has(CapInterfaces) {
		list = append(list, "interfaces")
	}
	if c.Has(CapRouting) {
		list = append(list, "routing")
	}
	if c.Has(CapFirewall) {
		list = append(list, "firewall")
	}
	if c.Has(CapNAT) {
		list = append(list, "nat")
	}
	if c.Has(CapDHCP) {
		list = append(list, "dhcp")
	}
	if c.Has(CapDNS) {
		list = append(list, "dns")
	}
	if c.Has(CapWireGuard) {
		list = append(list, "wireguard")
	}
	if c.Has(CapBGP) {
		list = append(list, "bgp")
	}
	if c.Has(CapOSPF) {
		list = append(list, "ospf")
	}
	if c.Has(CapMonitoring) {
		list = append(list, "monitoring")
	}
	if c.Has(CapConfig) {
		list = append(list, "config")
	}
	if c.Has(CapIPAM) {
		list = append(list, "ipam")
	}
	return list
}

type Provider interface {
	ID() string
	Name() string
	Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error
	Disconnect(ctx context.Context) error
	Capabilities() Capabilities

	GetSystemInfo(ctx context.Context) (*model.SystemInfo, error)
	ListInterfaces(ctx context.Context) ([]model.Interface, error)
	GetInterface(ctx context.Context, name string) (*model.Interface, error)
	ListRoutes(ctx context.Context) ([]model.Route, error)
	ListGateways(ctx context.Context) ([]model.Gateway, error)
	ListARPEntries(ctx context.Context) ([]model.ARPEntry, error)
	ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error)
	ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error)
	ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error)
	ListNATRules(ctx context.Context) ([]model.NATRule, error)
	ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error)
	ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error)
	GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error)

	GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error)

	GetRunningConfig(ctx context.Context) (string, error)
	ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error)
	ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error)
	RollbackConfig(ctx context.Context, rollbackID string) error
}

type Factory func() Provider

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Factory)
)

func Register(id string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[id] = factory
}

func Create(id string) (Provider, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, exists := registry[id]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return factory(), nil
}

func ListRegistered() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	list := make([]string, 0, len(registry))
	for id := range registry {
		list = append(list, id)
	}
	return list
}
