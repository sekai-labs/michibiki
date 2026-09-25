package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/views"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type TickMsg time.Time

type DataFetchedMsg struct {
	SysInfo   *model.SystemInfo
	Ifaces    []model.Interface
	Stats     []model.InterfaceStats
	Gateways  []model.Gateway
	Routes    []model.Route
	DHCP      []model.DHCPLease
	ARP       []model.ARPEntry
	Firewall  []model.FirewallRule
	NAT       []model.NATRule
	VPN       []model.WireGuardPeer
	BGP       []model.BGPNeighbor
	Subnets   []ipam.SubnetUsage
	Config    string
	FetchTime time.Time
	Err       error
}

type Model struct {
	ActiveTab      int
	Provider       provider.Provider
	DeviceName     string
	Width          int
	Height         int
	FilterInput    string
	FilterActive   bool
	IPAMSelected   int
	TrafficHistory map[string][]float64
	LastFetched    DataFetchedMsg
	Loading        bool
	Err            error
}

var tabNames = []string{
	"1: Dashboard",
	"2: Interfaces",
	"3: IPAM & Free IPs",
	"4: Routing",
	"5: Firewall",
	"6: DHCP",
	"7: VPN",
	"8: Traffic Monitor",
	"9: Config & Safety",
}

func NewModel(prov provider.Provider, deviceName string) Model {
	return Model{
		ActiveTab:      0,
		Provider:       prov,
		DeviceName:     deviceName,
		Width:          100,
		Height:         30,
		TrafficHistory: make(map[string][]float64),
		Loading:        true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchDataCmd(m.Provider),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func fetchDataCmd(prov provider.Provider) tea.Cmd {
	return func() tea.Msg {
		if prov == nil {
			return DataFetchedMsg{FetchTime: time.Now()}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var msg DataFetchedMsg
		msg.FetchTime = time.Now()

		if info, err := prov.GetSystemInfo(ctx); err == nil {
			msg.SysInfo = info
		}
		if ifaces, err := prov.ListInterfaces(ctx); err == nil {
			msg.Ifaces = ifaces
		}
		if stats, err := prov.GetInterfaceStats(ctx); err == nil {
			msg.Stats = stats
		}
		if gws, err := prov.ListGateways(ctx); err == nil {
			msg.Gateways = gws
		}
		if rts, err := prov.ListRoutes(ctx); err == nil {
			msg.Routes = rts
		}
		if dhcp, err := prov.ListDHCPLeases(ctx); err == nil {
			msg.DHCP = dhcp
		}
		if arp, err := prov.ListARPEntries(ctx); err == nil {
			msg.ARP = arp
		}
		if fw, err := prov.ListFirewallRules(ctx); err == nil {
			msg.Firewall = fw
		}
		if nat, err := prov.ListNATRules(ctx); err == nil {
			msg.NAT = nat
		}
		if vpn, err := prov.ListWireGuardPeers(ctx); err == nil {
			msg.VPN = vpn
		}
		if bgp, err := prov.ListBGPNeighbors(ctx); err == nil {
			msg.BGP = bgp
		}
		if subnets, err := prov.GetSubnetUsage(ctx, ""); err == nil {
			msg.Subnets = subnets
		}
		if cfg, err := prov.GetRunningConfig(ctx); err == nil {
			msg.Config = cfg
		}

		return msg
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case TickMsg:
		var cmd tea.Cmd
		if m.ActiveTab == 7 {
			cmd = fetchDataCmd(m.Provider)
		}
		return m, tea.Batch(cmd, tickCmd())

	case DataFetchedMsg:
		m.Loading = false
		m.LastFetched = msg
		m.Err = msg.Err

		for _, stat := range msg.Stats {
			totalRate := stat.RxBps + stat.TxBps
			history := m.TrafficHistory[stat.InterfaceName]
			history = append(history, totalRate)
			if len(history) > 20 {
				history = history[len(history)-20:]
			}
			m.TrafficHistory[stat.InterfaceName] = history
		}
		return m, nil

	case tea.KeyMsg:
		if m.FilterActive {
			switch msg.String() {
			case "esc", "enter":
				m.FilterActive = false
				return m, nil
			case "backspace":
				if len(m.FilterInput) > 0 {
					m.FilterInput = m.FilterInput[:len(m.FilterInput)-1]
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.FilterInput += msg.String()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			m.Loading = true
			return m, fetchDataCmd(m.Provider)

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % len(tabNames)
			return m, nil

		case "shift+tab":
			m.ActiveTab = (m.ActiveTab - 1 + len(tabNames)) % len(tabNames)
			return m, nil

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.ActiveTab = int(msg.String()[0] - '1')
			return m, nil

		case "/":
			m.FilterActive = true
			return m, nil

		case "esc":
			m.FilterActive = false
			m.FilterInput = ""
			return m, nil

		case "up", "k":
			if m.ActiveTab == 2 && m.IPAMSelected > 0 {
				m.IPAMSelected--
			}
			return m, nil

		case "down", "j":
			if m.ActiveTab == 2 && m.IPAMSelected < len(m.LastFetched.Subnets)-1 {
				m.IPAMSelected++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.Width < 20 || m.Height < 10 {
		return "Terminal window too small."
	}

	header := m.renderHeader()
	tabs := m.renderTabs()
	content := m.renderContent()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, tabs, content, footer)
}

func (m Model) renderHeader() string {
	dev := m.DeviceName
	if dev == "" {
		dev = "michibiki-node"
	}

	vendor := "UNKNOWN"
	if m.Provider != nil {
		vendor = m.Provider.Name()
	}

	statusDot := StyleBadgeOnline.Render("● CONNECTED")
	if m.Provider == nil {
		statusDot = StyleBadgeOffline.Render("○ DISCONNECTED")
	}

	titleText := StyleHeader.Render(fmt.Sprintf(" 導き MICHIBIKI  [%s] ", dev))
	vendorText := StyleHeaderSub.Render(fmt.Sprintf(" %s ", vendor))
	timeStr := time.Now().Format("15:04:05")
	clockText := StyleHeaderSub.Render(fmt.Sprintf(" %s ", timeStr))

	leftHeader := lipgloss.JoinHorizontal(lipgloss.Top, titleText, vendorText)
	rightHeader := lipgloss.JoinHorizontal(lipgloss.Top, statusDot, clockText)

	gap := m.Width - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader)
	if gap < 0 {
		gap = 0
	}
	gapStr := strings.Repeat(" ", gap)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftHeader, gapStr, rightHeader)
}

func (m Model) renderTabs() string {
	var renderedTabs []string
	for i, tab := range tabNames {
		if i == m.ActiveTab {
			renderedTabs = append(renderedTabs, StyleTabActive.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, StyleTabInactive.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m Model) renderContent() string {
	contentHeight := m.Height - 5
	if contentHeight < 5 {
		contentHeight = 5
	}

	switch m.ActiveTab {
	case 0:
		return views.RenderDashboard(views.DashboardData{
			SystemInfo: m.LastFetched.SysInfo,
			Interfaces: m.LastFetched.Ifaces,
			Gateways:   m.LastFetched.Gateways,
			Error:      m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 1:
		statsMap := make(map[string]model.InterfaceStats)
		for _, s := range m.LastFetched.Stats {
			statsMap[s.InterfaceName] = s
		}
		return views.RenderInterfaces(views.InterfacesData{
			Interfaces: m.LastFetched.Ifaces,
			Stats:      statsMap,
			Error:      m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 2:
		return views.RenderIPAM(views.IPAMData{
			Subnets:       m.LastFetched.Subnets,
			SelectedIndex: m.IPAMSelected,
			Error:         m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 3:
		return views.RenderRouting(views.RoutingData{
			Routes:    m.LastFetched.Routes,
			Gateways:  m.LastFetched.Gateways,
			Neighbors: m.LastFetched.BGP,
			Error:     m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 4:
		return views.RenderFirewall(views.FirewallData{
			Rules: m.LastFetched.Firewall,
			NAT:   m.LastFetched.NAT,
			Error: m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 5:
		return views.RenderDHCP(views.DHCPData{
			Leases: m.LastFetched.DHCP,
			Error:  m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 6:
		return views.RenderVPN(views.VPNData{
			Peers: m.LastFetched.VPN,
			Error: m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 7:
		return views.RenderMonitor(views.MonitorData{
			Stats:   m.LastFetched.Stats,
			History: m.TrafficHistory,
			Error:   m.LastFetched.Err,
		}, m.Width, contentHeight)

	case 8:
		return views.RenderConfig(views.ConfigData{
			RunningConfig: m.LastFetched.Config,
			SafetyStatus:  "Safety Engine Active — rollback tracking ready",
			Error:         m.LastFetched.Err,
		}, m.Width, contentHeight)
	}

	return ""
}

func (m Model) renderFooter() string {
	var keys []string
	keys = append(keys, StyleStatusKey.Render("1-9")+" Switch Tab")
	keys = append(keys, StyleStatusKey.Render("Tab")+" Next")
	keys = append(keys, StyleStatusKey.Render("r")+" Refresh")
	keys = append(keys, StyleStatusKey.Render("/")+" Filter")
	keys = append(keys, StyleStatusKey.Render("q")+" Quit")

	if m.FilterActive {
		filterBar := StyleStatusKey.Render("FILTER: ") + m.FilterInput + " █ (esc/enter to finish)"
		return StyleStatusBar.Width(m.Width).Render(filterBar)
	}

	footerText := strings.Join(keys, "  •  ")
	if m.Loading {
		footerText += "  " + StyleProgressFilledWarn.Render("[FETCHING DATA...]")
	}
	return StyleStatusBar.Width(m.Width).Render(footerText)
}
