package tui_test

import (
	"context"
	"net/netip"
	"strings"
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
		Hostname:         "gateway.local",
		OS:               "OPNsense",
		Version:          "24.1",
		Architecture:     "amd64",
		UptimeSeconds:    123456,
		CPUCount:         4,
		CPUUsagePct:      15.5,
		MemoryTotalBytes: 8589934592,
		MemoryUsedBytes:  2147483648,
		StorageTotal:     64000000000,
		StorageUsed:      16000000000,
	}, nil
}
func (d *dummyProvider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return []model.Interface{
		{
			Name:          "vtnet0",
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   model.AdminStatusUp,
			OperStatus:    model.OperStatusUp,
			MACAddress:    "00:11:22:33:44:55",
			IPv4Addresses: []string{"192.168.1.1/24"},
			SpeedBps:      1000000000,
		},
		{
			Name:          "vtnet1",
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   model.AdminStatusUp,
			OperStatus:    model.OperStatusUp,
			MACAddress:    "00:11:22:33:44:56",
			IPv4Addresses: []string{"10.0.0.1/24"},
			SpeedBps:      1000000000,
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
			Metric:      1,
		},
	}, nil
}
func (d *dummyProvider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return []model.Gateway{
		{
			Name:          "WAN_GW",
			Address:       "192.168.1.254",
			Interface:     "vtnet0",
			Status:        model.GatewayOnline,
			LatencyMs:     2.5,
			PacketLossPct: 0.0,
			IsDefault:     true,
		},
	}, nil
}
func (d *dummyProvider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return []model.ARPEntry{
		{
			IPAddress:  "192.168.1.100",
			MACAddress: "aa:bb:cc:dd:ee:ff",
			Interface:  "vtnet0",
		},
	}, nil
}
func (d *dummyProvider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return []model.DHCPLease{
		{
			IPAddress:      "192.168.1.101",
			MACAddress:     "aa:bb:cc:dd:ee:01",
			ClientHostname: "laptop",
			SubnetCIDR:     "192.168.1.0/24",
			State:          model.DHCPActive,
			Ends:           time.Now().Add(1 * time.Hour),
		},
	}, nil
}
func (d *dummyProvider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	return []model.FirewallRule{
		{
			Sequence:    1,
			Interface:   "vtnet0",
			Direction:   "in",
			Action:      model.FirewallPass,
			Protocol:    "tcp",
			Destination: "192.168.1.1",
			Description: "Allow SSH",
		},
	}, nil
}
func (d *dummyProvider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return nil, nil
}
func (d *dummyProvider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return []model.NATRule{
		{
			Interface:   "vtnet0",
			Protocol:    "tcp",
			Source:      "any",
			Destination: "192.168.1.100",
			Target:      "192.168.1.100",
			TargetPort:  "80",
			Description: "Port Forward HTTP",
		},
	}, nil
}
func (d *dummyProvider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return []model.BGPNeighbor{
		{
			PeerAddress:      "192.168.1.2",
			RemoteAS:         65001,
			LocalAS:          65000,
			State:            model.BGPEstablished,
			PrefixesReceived: 5,
			PrefixesAccepted: 5,
		},
	}, nil
}
func (d *dummyProvider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return []model.WireGuardPeer{
		{
			PublicKey:       "peer1publickeyabcdef123456=",
			Endpoint:        "203.0.113.5:51820",
			AllowedIPs:      []string{"10.10.0.2/32"},
			LatestHandshake: time.Now().Add(-2 * time.Minute),
			TransferRxBytes: 1024000,
			TransferTxBytes: 2048000,
		},
	}, nil
}
func (d *dummyProvider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	return []model.InterfaceStats{
		{
			InterfaceName: "vtnet0",
			RxBps:         5000000,
			TxBps:         2000000,
			RxBytes:       100000000,
			TxBytes:       50000000,
		},
	}, nil
}
func (d *dummyProvider) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	prefix := netip.MustParsePrefix("192.168.1.0/24")
	gw := netip.MustParseAddr("192.168.1.1")
	return []ipam.SubnetUsage{
		{
			CIDR:           prefix,
			InterfaceName:  "vtnet0",
			GatewayIP:      gw,
			TotalIPs:       256,
			UsableIPs:      254,
			UsedIPs:        10,
			FreeIPs:        244,
			UtilizationPct: 3.9,
			Allocations: []ipam.IPAllocation{
				{
					IP:     gw,
					Source: ipam.SourceInterfaceIP,
					Status: ipam.StatusActive,
				},
			},
		},
		{
			CIDR:           netip.MustParsePrefix("10.0.0.0/24"),
			InterfaceName:  "vtnet1",
			GatewayIP:      netip.MustParseAddr("10.0.0.1"),
			TotalIPs:       256,
			UsableIPs:      254,
			UsedIPs:        5,
			FreeIPs:        249,
			UtilizationPct: 2.0,
		},
	}, nil
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

	for panel := 0; panel < 4; panel++ {
		updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(rune('1' + panel)))})
		currModel = updated.(tui.Model)
		if currModel.ActivePanel != panel {
			t.Fatalf("expected ActivePanel %d, got %d", panel, currModel.ActivePanel)
		}
		if currModel.ActiveTab != panel {
			t.Fatalf("expected ActiveTab %d, got %d", panel, currModel.ActiveTab)
		}
		viewStr := currModel.View()
		if len(viewStr) == 0 {
			t.Fatalf("view string for panel %d is empty", panel)
		}
	}

	for i := 0; i < 4; i++ {
		updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyTab})
		currModel = updated.(tui.Model)
		expected := (3 + 1 + i) % 4
		if currModel.ActivePanel != expected {
			t.Fatalf("Tab cycle: expected panel %d, got %d", expected, currModel.ActivePanel)
		}
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	currModel = updated.(tui.Model)
	if currModel.ActivePanel != 2 {
		t.Fatalf("Shift+Tab: expected panel 2, got %d", currModel.ActivePanel)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	currModel = updated.(tui.Model)
	if currModel.ActivePanel != 1 {
		t.Fatalf("Shift+Tab: expected panel 1, got %d", currModel.ActivePanel)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	currModel = updated.(tui.Model)
	if currModel.ActivePanel != 0 {
		t.Fatalf("Shift+Tab: expected panel 0, got %d", currModel.ActivePanel)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	currModel = updated.(tui.Model)
	if currModel.ActivePanel != 3 {
		t.Fatalf("Shift+Tab wrap: expected panel 3, got %d", currModel.ActivePanel)
	}
	if currModel.FocusedPane != tui.PaneDock {
		t.Fatalf("expected initial focus to be PaneDock, got %v", currModel.FocusedPane)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected focus to switch to PaneDetail on 'l', got %v", currModel.FocusedPane)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDock {
		t.Fatalf("expected focus to return to PaneDock on 'h', got %v", currModel.FocusedPane)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected focus to switch to PaneDetail on Enter, got %v", currModel.FocusedPane)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	currModel = updated.(tui.Model)

	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected initial DetailSubTab 0, got %d", currModel.DetailSubTab)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 1 {
		t.Fatalf("expected DetailSubTab 1 on ']', got %d", currModel.DetailSubTab)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected DetailSubTab 0 on '[', got %d", currModel.DetailSubTab)
	}

	if currModel.ShowHelpModal {
		t.Fatalf("expected ShowHelpModal to be initially false")
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	currModel = updated.(tui.Model)
	if !currModel.ShowHelpModal {
		t.Fatalf("expected ShowHelpModal to be true after '?'")
	}
	helpView := currModel.View()
	if !strings.Contains(helpView, "MICHIBIKI DOCK NAVIGATION") {
		t.Fatalf("help view does not contain navigation title")
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	currModel = updated.(tui.Model)
	if currModel.ShowHelpModal {
		t.Fatalf("expected ShowHelpModal to close on Esc")
	}
}

func TestTUIResponsiveLayout(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	wideModel := updated.(tui.Model)
	wideView := wideModel.View()
	if len(wideView) == 0 {
		t.Fatalf("wide view is empty")
	}
	if !strings.Contains(wideView, "Status & Health") {
		t.Fatalf("wide view should render Status & Health panel")
	}
	if !strings.Contains(wideView, "Detail Inspector") {
		t.Fatalf("wide view should render Detail Inspector pane")
	}

	updated, _ = m.Update(tea.WindowSizeMsg{Width: 90, Height: 40})
	stackedModel := updated.(tui.Model)
	stackedView := stackedModel.View()
	if len(stackedView) == 0 {
		t.Fatalf("stacked view is empty")
	}
	if !strings.Contains(stackedView, "Status & Health") {
		t.Fatalf("stacked dock view should render dock panels")
	}

	updated, _ = stackedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	stackedDetailModel := updated.(tui.Model)
	stackedDetailView := stackedDetailModel.View()
	if !strings.Contains(stackedDetailView, "Detail Inspector") {
		t.Fatalf("stacked detail view should render Detail Inspector pane")
	}
}

func TestTUIListNavigationAndYank(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")

	info, _ := prov.GetSystemInfo(context.Background())
	ifaces, _ := prov.ListInterfaces(context.Background())
	gws, _ := prov.ListGateways(context.Background())
	routes, _ := prov.ListRoutes(context.Background())

	updated, _ := m.Update(tui.DataFetchedMsg{
		SysInfo:  info,
		Ifaces:   ifaces,
		Gateways: gws,
		Routes:   routes,
	})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	currModel = updated.(tui.Model)
	if currModel.ActivePanel != 1 {
		t.Fatalf("expected ActivePanel 1, got %d", currModel.ActivePanel)
	}

	if currModel.SelectedIfaceIndex != 0 {
		t.Fatalf("expected initial SelectedIfaceIndex 0, got %d", currModel.SelectedIfaceIndex)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	currModel = updated.(tui.Model)
	if currModel.SelectedIfaceIndex != 1 {
		t.Fatalf("expected SelectedIfaceIndex 1 after 'j', got %d", currModel.SelectedIfaceIndex)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	currModel = updated.(tui.Model)
	if currModel.SelectedIfaceIndex != 0 {
		t.Fatalf("expected SelectedIfaceIndex 0 after 'k', got %d", currModel.SelectedIfaceIndex)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	currModel = updated.(tui.Model)
	if currModel.SelectedIfaceIndex != 1 {
		t.Fatalf("expected SelectedIfaceIndex 1 after 'G', got %d", currModel.SelectedIfaceIndex)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	currModel = updated.(tui.Model)
	if currModel.SelectedIfaceIndex != 0 {
		t.Fatalf("expected SelectedIfaceIndex 0 after 'g', got %d", currModel.SelectedIfaceIndex)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	currModel = updated.(tui.Model)
	if !strings.Contains(currModel.StatusFlash, "Yanked '192.168.1.1/24'") {
		t.Fatalf("expected StatusFlash to contain yanked IP, got %q", currModel.StatusFlash)
	}
}

func TestTUISearchFilter(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	currModel = updated.(tui.Model)
	if !currModel.FilterActive {
		t.Fatalf("expected FilterActive to be true after '/'")
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	currModel = updated.(tui.Model)
	if currModel.FilterInput != "wan" {
		t.Fatalf("expected FilterInput 'wan', got %q", currModel.FilterInput)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	currModel = updated.(tui.Model)
	if currModel.FilterActive {
		t.Fatalf("expected FilterActive to be false after Enter")
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	currModel = updated.(tui.Model)
	if currModel.FilterActive {
		t.Fatalf("expected FilterActive to be false after Esc")
	}
	if currModel.FilterInput != "" {
		t.Fatalf("expected FilterInput to be cleared on Esc, got %q", currModel.FilterInput)
	}
}

func TestTUIViewportScrollingAndHalfPage(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected focus in detail pane")
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	currModel = updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	currModel = updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	currModel = updated.(tui.Model)
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	currModel = updated.(tui.Model)

	_, cmd := currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatalf("expected non-nil tea.Quit cmd on 'q'")
	}
}

func TestTUIDetailSubTabsKeymaps(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(">")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 1 {
		t.Fatalf("expected DetailSubTab 1 on '>', got %d", currModel.DetailSubTab)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("<")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected DetailSubTab 0 on '<', got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 1 {
		t.Fatalf("expected DetailSubTab 1 on '.', got %d", currModel.DetailSubTab)
	}
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(",")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected DetailSubTab 0 on ',', got %d", currModel.DetailSubTab)
	}

	for st := 0; st < 4; st++ {
		currModel.DetailSubTab = st
		viewStr := currModel.View()
		if len(viewStr) == 0 {
			t.Fatalf("subtab %d view is empty", st)
		}
	}
}

func TestTUIDetailDirectSubTabKeymaps(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected FocusedPane PaneDetail on 't', got %v", currModel.FocusedPane)
	}
	if currModel.DetailSubTab != 1 {
		t.Fatalf("expected DetailSubTab 1 (Live Traffic) on 't', got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 2 {
		t.Fatalf("expected DetailSubTab 2 (IPAM Matrix) on 'i', got %d", currModel.DetailSubTab)
	}

	currModel.DetailSubTab = 0
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 2 {
		t.Fatalf("expected DetailSubTab 2 (IPAM Matrix) on 'm', got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 3 {
		t.Fatalf("expected DetailSubTab 3 (Running Config) on 'c', got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected DetailSubTab 0 (Overview) on 'o', got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected FocusedPane to stay PaneDetail on '2', got %v", currModel.FocusedPane)
	}
	if currModel.DetailSubTab != 1 {
		t.Fatalf("expected DetailSubTab 1 on '2' while in PaneDetail, got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 2 {
		t.Fatalf("expected DetailSubTab 2 on '3' while in PaneDetail, got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 3 {
		t.Fatalf("expected DetailSubTab 3 on '4' while in PaneDetail, got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 0 {
		t.Fatalf("expected DetailSubTab 0 on '1' while in PaneDetail, got %d", currModel.DetailSubTab)
	}
}

type dummyPluginProvider struct {
	dummyProvider
	id      string
	version string
}

func (d *dummyPluginProvider) ID() string      { return d.id }
func (d *dummyPluginProvider) Version() string { return d.version }
func (d *dummyPluginProvider) Name() string    { return "traefik-plugin" }

func TestTUIPluginCustomUI(t *testing.T) {
	prov := &dummyPluginProvider{
		id:      "traefik-plugin",
		version: "1.2.3",
	}
	m := tui.NewModel(prov, "plugin-box")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)
	viewStr := currModel.View()
	if !strings.Contains(viewStr, "EXTERNAL") || !strings.Contains(viewStr, "PLUGIN PROVIDER") {
		t.Fatalf("expected view to contain EXTERNAL PLUGIN PROVIDER, got:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "traefik-plugin") {
		t.Fatalf("expected view to contain plugin name, got:\n%s", viewStr)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected FocusedPane PaneDetail on 'p', got %v", currModel.FocusedPane)
	}
	if currModel.DetailSubTab != 4 {
		t.Fatalf("expected DetailSubTab 4 on 'p', got %d", currModel.DetailSubTab)
	}

	currModel.DetailSubTab = 0
	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	currModel = updated.(tui.Model)
	if currModel.DetailSubTab != 4 {
		t.Fatalf("expected DetailSubTab 4 on '5' while in PaneDetail, got %d", currModel.DetailSubTab)
	}
}

func TestTUIIPAMSubnetNavigation(t *testing.T) {
	prov := &dummyProvider{}
	m := tui.NewModel(prov, "test-box")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	currModel := updated.(tui.Model)

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	currModel = updated.(tui.Model)
	if currModel.FocusedPane != tui.PaneDetail {
		t.Fatalf("expected FocusedPane PaneDetail on 'i', got %v", currModel.FocusedPane)
	}
	if currModel.DetailSubTab != 2 {
		t.Fatalf("expected DetailSubTab 2, got %d", currModel.DetailSubTab)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	currModel = updated.(tui.Model)
	if len(currModel.LastFetched.Subnets) > 1 && currModel.IPAMSelected != 1 {
		t.Fatalf("expected IPAMSelected to advance to 1 on 'j', got %d", currModel.IPAMSelected)
	}

	updated, _ = currModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	currModel = updated.(tui.Model)
	if currModel.IPAMSelected != 0 {
		t.Fatalf("expected IPAMSelected to return to 0 on 'k', got %d", currModel.IPAMSelected)
	}
}
