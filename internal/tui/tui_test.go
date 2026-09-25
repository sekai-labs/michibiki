package tui_test

import (
	"context"
	"net/netip"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/michibiki/internal/tui"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type dummyProvider struct{}

func (d *dummyProvider) ID() string   { return "test" }
func (d *dummyProvider) Name() string { return "Test Device" }
func (d *dummyProvider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	return nil
}
func (d *dummyProvider) Disconnect(ctx context.Context) error { return nil }
func (d *dummyProvider) Capabilities() provider.Capabilities {
	return provider.CapSystem | provider.CapInterfaces | provider.CapIPAM
}
func (d *dummyProvider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname:         "router.local",
		OS:               "OPNsense",
		Version:          "24.1",
		Architecture:     "amd64",
		UptimeSeconds:    3600,
		CPUCount:         4,
		CPUUsagePct:      12.5,
		MemoryTotalBytes: 8589934592,
		MemoryUsedBytes:  2147483648,
		StorageTotal:     64424509440,
		StorageUsed:      12884901888,
	}, nil
}
func (d *dummyProvider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return []model.Interface{
		{
			ID:            "1",
			Name:          "vtnet0",
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   model.AdminStatusUp,
			OperStatus:    model.OperStatusUp,
			MACAddress:    "52:54:00:12:34:56",
			IPv4Addresses: []string{"192.168.1.1/24"},
		},
	}, nil
}
func (d *dummyProvider) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	return nil, nil
}
func (d *dummyProvider) ListRoutes(ctx context.Context) ([]model.Route, error) {
	return []model.Route{
		{
			Destination: "0.0.0.0/0",
			Gateway:     "192.168.1.254",
			Interface:   "vtnet0",
			Protocol:    model.RouteProtoStatic,
		},
	}, nil
}
func (d *dummyProvider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return []model.Gateway{
		{
			Name:      "WAN_GW",
			Address:   "192.168.1.254",
			Interface: "vtnet0",
			Status:    model.GatewayOnline,
			LatencyMs: 4.2,
			IsDefault: true,
		},
	}, nil
}
func (d *dummyProvider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return []model.ARPEntry{
		{
			IPAddress:  "192.168.1.10",
			MACAddress: "52:54:00:ab:cd:ef",
			Interface:  "vtnet0",
			Hostname:   "workstation",
			Status:     model.ARPReachable,
		},
	}, nil
}
func (d *dummyProvider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return []model.DHCPLease{
		{
			IPAddress:      "192.168.1.50",
			MACAddress:     "52:54:00:99:88:77",
			ClientHostname: "printer",
			SubnetCIDR:     "192.168.1.0/24",
			Interface:      "vtnet0",
			State:          model.DHCPActive,
			Ends:           time.Now().Add(2 * time.Hour),
		},
	}, nil
}
func (d *dummyProvider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	return []model.FirewallRule{
		{
			ID:          "rule1",
			Sequence:    1,
			Interface:   "vtnet0",
			Direction:   "in",
			Action:      model.FirewallPass,
			Protocol:    "tcp",
			Destination: "192.168.1.1",
			Description: "Allow SSH",
			Enabled:     true,
		},
	}, nil
}
func (d *dummyProvider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return nil, nil
}
func (d *dummyProvider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return []model.NATRule{
		{
			ID:          "nat1",
			Interface:   "vtnet0",
			Type:        "snat",
			Source:      "192.168.1.0/24",
			Destination: "any",
			Target:      "WAN_IP",
			Enabled:     true,
		},
	}, nil
}
func (d *dummyProvider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return []model.BGPNeighbor{
		{
			RemoteAS:         65001,
			LocalAS:          65000,
			PeerAddress:      "10.0.0.2",
			State:            model.BGPEstablished,
			PrefixesReceived: 5,
			PrefixesAccepted: 5,
		},
	}, nil
}
func (d *dummyProvider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return []model.WireGuardPeer{
		{
			PublicKey:       "yA8j2X+samplewireguardkey123456=",
			Endpoint:        "198.51.100.1:51820",
			AllowedIPs:      []string{"10.100.0.2/32"},
			LatestHandshake: time.Now().Add(-1 * time.Minute),
			TransferRxBytes: 1048576,
			TransferTxBytes: 2097152,
		},
	}, nil
}
func (d *dummyProvider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	return []model.InterfaceStats{
		{
			InterfaceName: "vtnet0",
			RxBytes:       10485760,
			TxBytes:       5242880,
			RxBps:         1024000,
			TxBps:         512000,
		},
	}, nil
}
func (d *dummyProvider) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	prefix := netip.MustParsePrefix("192.168.1.0/24")
	ifaces := []model.Interface{
		{
			Name:          "vtnet0",
			IPv4Addresses: []string{"192.168.1.1/24"},
		},
	}
	leases := []model.DHCPLease{
		{
			IPAddress:      "192.168.1.50",
			MACAddress:     "52:54:00:99:88:77",
			ClientHostname: "printer",
		},
	}
	arp := []model.ARPEntry{
		{
			IPAddress:  "192.168.1.10",
			MACAddress: "52:54:00:ab:cd:ef",
			Hostname:   "workstation",
		},
	}
	u := ipam.CalculateSubnetUsage(prefix, "vtnet0", 0, ifaces, leases, arp, nil)
	return []ipam.SubnetUsage{u}, nil
}
func (d *dummyProvider) GetRunningConfig(ctx context.Context) (string, error) {
	return "<opnsense>\n  <version>1</version>\n</opnsense>", nil
}
func (d *dummyProvider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}
func (d *dummyProvider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{Success: true}, nil
}
func (d *dummyProvider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}

func TestTUIModelNavigation(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")

	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	for tab := 0; tab < 9; tab++ {
		updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(rune('1' + tab)))})
		currModel = updated.(tui.Model)
		if currModel.ActiveTab != tab {
			t.Fatalf("expected tab %d, got %d", tab, currModel.ActiveTab)
		}
		viewStr := currModel.View()
		if len(viewStr) == 0 {
			t.Fatalf("view string for tab %d is empty", tab)
		}
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	currModel = updated.(tui.Model)
	if currModel.ActiveTab != 0 {
		t.Fatalf("expected cycle to tab 0, got %d", currModel.ActiveTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	currModel = updated.(tui.Model)
	if currModel.ActiveTab != 8 {
		t.Fatalf("expected cycle back to tab 8, got %d", currModel.ActiveTab)
	}
}
