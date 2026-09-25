package network

import (
	"context"

	"github.com/sekai-labs/michibiki/internal/domain"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

func (s *NetworkHandler) GetRoutingOverview(ctx context.Context) (*domain.RoutingOverview, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapRouting) {
		return nil, domain.ErrUnsupportedCapability
	}

	routes, err := s.prov.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}

	gateways, _ := s.prov.ListGateways(ctx)

	var bgpNeighbors []model.BGPNeighbor
	if s.prov.Capabilities().Has(provider.CapBGP) {
		if bgp, err := s.prov.ListBGPNeighbors(ctx); err == nil {
			bgpNeighbors = bgp
		}
	}

	var defaultGw *model.Gateway
	for i := range gateways {
		if gateways[i].IsDefault {
			defaultGw = &gateways[i]
			break
		}
	}

	activePeers := 0
	for _, peer := range bgpNeighbors {
		if peer.State == model.BGPEstablished {
			activePeers++
		}
	}

	return &domain.RoutingOverview{
		Routes:         routes,
		Gateways:       gateways,
		BGPNeighbors:   bgpNeighbors,
		DefaultGateway: defaultGw,
		TotalRoutes:    len(routes),
		ActivePeers:    activePeers,
	}, nil
}

func (s *NetworkHandler) GetSecurityOverview(ctx context.Context) (*domain.SecurityOverview, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapFirewall) {
		return nil, domain.ErrUnsupportedCapability
	}

	rules, err := s.prov.ListFirewallRules(ctx)
	if err != nil {
		return nil, err
	}

	aliases, _ := s.prov.ListFirewallAliases(ctx)

	var natRules []model.NATRule
	if s.prov.Capabilities().Has(provider.CapNAT) {
		if nat, err := s.prov.ListNATRules(ctx); err == nil {
			natRules = nat
		}
	}

	enabled := 0
	for _, r := range rules {
		if r.Enabled {
			enabled++
		}
	}

	return &domain.SecurityOverview{
		FirewallRules: rules,
		Aliases:       aliases,
		NATRules:      natRules,
		TotalRules:    len(rules),
		EnabledRules:  enabled,
		TotalNATRules: len(natRules),
	}, nil
}

func (s *NetworkHandler) GetServicesOverview(ctx context.Context) (*domain.ServicesOverview, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}

	var dhcpLeases []model.DHCPLease
	if s.prov.Capabilities().Has(provider.CapDHCP) {
		if leases, err := s.prov.ListDHCPLeases(ctx); err == nil {
			dhcpLeases = leases
		}
	}

	var vpnPeers []model.WireGuardPeer
	if s.prov.Capabilities().Has(provider.CapWireGuard) {
		if peers, err := s.prov.ListWireGuardPeers(ctx); err == nil {
			vpnPeers = peers
		}
	}

	var arpEntries []model.ARPEntry
	if arp, err := s.prov.ListARPEntries(ctx); err == nil {
		arpEntries = arp
	}

	activeLeases := 0
	for _, l := range dhcpLeases {
		if l.State == model.DHCPActive || l.State == model.DHCPStatic {
			activeLeases++
		}
	}

	return &domain.ServicesOverview{
		DHCPLeases:     dhcpLeases,
		WireGuardPeers: vpnPeers,
		ARPEntries:     arpEntries,
		ActiveLeases:   activeLeases,
		ActiveVPNPeers: len(vpnPeers),
	}, nil
}

func (s *NetworkHandler) GetIPAMAllocation(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapIPAM) {
		return nil, domain.ErrUnsupportedCapability
	}
	return s.prov.GetSubnetUsage(ctx, filter)
}

func (s *NetworkHandler) GetConfigState(ctx context.Context) (*domain.ConfigState, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapConfig) {
		return nil, domain.ErrUnsupportedCapability
	}

	cfg, err := s.prov.GetRunningConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.ConfigState{
		RunningConfig: cfg,
	}, nil
}

func (s *NetworkHandler) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapConfig) {
		return nil, domain.ErrUnsupportedCapability
	}
	return s.prov.ValidateConfig(ctx, candidate)
}

func (s *NetworkHandler) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapConfig) {
		return nil, domain.ErrUnsupportedCapability
	}
	return s.prov.ApplyConfig(ctx, req)
}

func (s *NetworkHandler) RollbackConfig(ctx context.Context, rollbackID string) error {
	if s.prov == nil {
		return provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapConfig) {
		return domain.ErrUnsupportedCapability
	}
	return s.prov.RollbackConfig(ctx, rollbackID)
}
