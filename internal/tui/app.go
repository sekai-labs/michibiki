package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/internal/tui/views"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type TickMsg time.Time
type ClearFlashMsg struct{}

type FocusedPane int

const (
	PaneDock FocusedPane = iota
	PaneDetail
)

type SubTab int

const (
	SubTabOverview SubTab = iota
	SubTabLiveTraffic
	SubTabIPAMMatrix
	SubTabRunningConfig
	SubTabPlugin
)

var subTabNames = []string{
	"Detail",
	"[t] Traffic",
	"[m] IPAM",
	"[c] Config",
}

func (m Model) isPluginProvider() bool {
	if m.Provider == nil {
		return false
	}
	name := m.Provider.Name()
	if name == "opnsense" || name == "pfsense" || name == "routeros" || name == "openwrt" || name == "vyos" || name == "frr" || name == "mock" {
		return false
	}
	return true
}

func (m Model) getSubTabNames() []string {
	var inspectorTitle string
	switch m.ActivePanel {
	case 0:
		inspectorTitle = "1 Status Detail"
	case 1:
		inspectorTitle = "2 Interfaces & VLANs"
	case 2:
		inspectorTitle = "3 Routes & Gateways"
	case 3:
		inspectorTitle = "4 Security & Services"
	default:
		inspectorTitle = "Detail"
	}

	names := []string{
		inspectorTitle,
		"[t] Traffic",
		"[m] IPAM",
		"[c] Config",
	}
	if m.isPluginProvider() {
		names = append(names, "[p] Plugin UI")
	}
	return names
}

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
	FocusedPane          FocusedPane
	ActivePanel          int
	DetailSubTab         int
	SelectedIfaceIndex   int
	SelectedRouteIndex   int
	SelectedServiceIndex int
	ShowHelpModal        bool
	StatusFlash          string

	ActiveTab int

	Provider       provider.Provider
	DeviceName     string
	Width          int
	Height         int
	FilterInput    string
	FilterActive   bool
	IPAMSelected   int
	Viewport       viewport.Model
	ViewportReady  bool
	TrafficHistory map[string][]float64
	LastFetched    DataFetchedMsg
	Loading        bool
	Err            error
}

func NewModel(prov provider.Provider, deviceName string) Model {
	return Model{
		FocusedPane:    PaneDock,
		ActivePanel:    0,
		DetailSubTab:   0,
		ActiveTab:      0,
		Provider:       prov,
		DeviceName:     deviceName,
		Width:          120,
		Height:         40,
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

func clearFlashCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearFlashMsg{}
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
		m.updateViewportDimensions()
		m.updateViewportContent()
		return m, nil

	case TickMsg:
		var cmd tea.Cmd
		if m.DetailSubTab == int(SubTabLiveTraffic) {
			cmd = fetchDataCmd(m.Provider)
		}
		return m, tea.Batch(cmd, tickCmd())

	case ClearFlashMsg:
		m.StatusFlash = ""
		return m, nil

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
		m.clampSelections()
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		if m.FilterActive {
			switch msg.String() {
			case "esc":
				m.FilterActive = false
				m.FilterInput = ""
				m.updateViewportContent()
				return m, nil
			case "enter":
				m.FilterActive = false
				m.updateViewportContent()
				return m, nil
			case "backspace":
				if len(m.FilterInput) > 0 {
					m.FilterInput = m.FilterInput[:len(m.FilterInput)-1]
					m.updateViewportContent()
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.FilterInput += msg.String()
					m.updateViewportContent()
				}
				return m, nil
			}
		}

		if m.ShowHelpModal {
			switch msg.String() {
			case "?", "esc", "q", "enter":
				m.ShowHelpModal = false
				return m, nil
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.ShowHelpModal = true
			return m, nil

		case "r":
			m.Loading = true
			m.StatusFlash = "Refreshing data..."
			return m, tea.Batch(fetchDataCmd(m.Provider), clearFlashCmd())

		case "1":
			if m.FocusedPane == PaneDetail {
				m.DetailSubTab = int(SubTabOverview)
				m.Viewport.GotoTop()
				m.StatusFlash = "Opened Overview"
				m.updateViewportContent()
				return m, clearFlashCmd()
			}
			m.ActivePanel = 0
			m.ActiveTab = 0
			m.FocusedPane = PaneDock
			m.updateViewportContent()
			return m, nil
		case "2":
			if m.FocusedPane == PaneDetail {
				m.DetailSubTab = int(SubTabLiveTraffic)
				m.Viewport.GotoTop()
				m.StatusFlash = "Opened Live Traffic"
				m.updateViewportContent()
				return m, tea.Batch(fetchDataCmd(m.Provider), clearFlashCmd())
			}
			m.ActivePanel = 1
			m.ActiveTab = 1
			m.FocusedPane = PaneDock
			m.updateViewportContent()
			return m, nil
		case "3":
			if m.FocusedPane == PaneDetail {
				m.DetailSubTab = int(SubTabIPAMMatrix)
				m.Viewport.GotoTop()
				m.StatusFlash = "Opened IPAM Matrix"
				m.updateViewportContent()
				return m, clearFlashCmd()
			}
			m.ActivePanel = 2
			m.ActiveTab = 2
			m.FocusedPane = PaneDock
			m.updateViewportContent()
			return m, nil
		case "4":
			if m.FocusedPane == PaneDetail {
				m.DetailSubTab = int(SubTabRunningConfig)
				m.Viewport.GotoTop()
				m.StatusFlash = "Opened Config"
				m.updateViewportContent()
				return m, clearFlashCmd()
			}
			m.ActivePanel = 3
			m.ActiveTab = 3
			m.FocusedPane = PaneDock
			m.updateViewportContent()
			return m, nil
		case "5":
			if m.FocusedPane == PaneDetail && m.isPluginProvider() {
				m.DetailSubTab = int(SubTabPlugin)
				m.Viewport.GotoTop()
				m.StatusFlash = "Opened Plugin UI"
				m.updateViewportContent()
				return m, clearFlashCmd()
			}
			return m, nil

		case "tab":
			m.ActivePanel = (m.ActivePanel + 1) % 4
			m.ActiveTab = m.ActivePanel
			m.updateViewportContent()
			return m, nil
		case "shift+tab":
			m.ActivePanel = (m.ActivePanel - 1 + 4) % 4
			m.ActiveTab = m.ActivePanel
			m.updateViewportContent()
			return m, nil

		case "h":
			m.FocusedPane = PaneDock
			m.updateViewportContent()
			return m, nil
		case "l", "enter":
			m.FocusedPane = PaneDetail
			m.updateViewportContent()
			return m, nil
		case "[", "<", ",":
			tabsCount := len(m.getSubTabNames())
			m.DetailSubTab = (m.DetailSubTab - 1 + tabsCount) % tabsCount
			m.Viewport.GotoTop()
			m.updateViewportContent()
			return m, nil
		case "]", ">", ".":
			tabsCount := len(m.getSubTabNames())
			m.DetailSubTab = (m.DetailSubTab + 1) % tabsCount
			m.Viewport.GotoTop()
			m.updateViewportContent()
			return m, nil

		case "/":
			m.FilterActive = true
			return m, nil
		case "esc":
			m.FilterActive = false
			m.FilterInput = ""
			m.updateViewportContent()
			return m, nil

		case "y":
			yankText := m.getYankTarget()
			if yankText != "" {
				seq := osc52.New(yankText).String()
				fmt.Print(seq)
				m.StatusFlash = fmt.Sprintf("Yanked '%s' to clipboard", yankText)
			} else {
				m.StatusFlash = "Nothing to yank from selection"
			}
			return m, clearFlashCmd()

		case "o", "O":
			m.FocusedPane = PaneDetail
			m.DetailSubTab = int(SubTabOverview)
			m.Viewport.GotoTop()
			m.StatusFlash = "Opened Overview"
			m.updateViewportContent()
			return m, clearFlashCmd()

		case "t", "T":
			m.FocusedPane = PaneDetail
			m.DetailSubTab = int(SubTabLiveTraffic)
			m.Viewport.GotoTop()
			m.StatusFlash = "Opened Live Traffic"
			m.updateViewportContent()
			return m, tea.Batch(fetchDataCmd(m.Provider), clearFlashCmd())

		case "i", "I", "m", "M":
			m.FocusedPane = PaneDetail
			m.DetailSubTab = int(SubTabIPAMMatrix)
			m.Viewport.GotoTop()
			m.StatusFlash = "Opened IPAM Matrix"
			m.updateViewportContent()
			return m, clearFlashCmd()

		case "c", "C":
			m.FocusedPane = PaneDetail
			m.DetailSubTab = int(SubTabRunningConfig)
			m.Viewport.GotoTop()
			m.StatusFlash = "Opened Config"
			m.updateViewportContent()
			return m, clearFlashCmd()

		case "p", "P":
			m.FocusedPane = PaneDetail
			m.DetailSubTab = int(SubTabPlugin)
			m.Viewport.GotoTop()
			m.StatusFlash = "Opened Plugin UI"
			m.updateViewportContent()
			return m, clearFlashCmd()
		case "k", "up":
			if m.FocusedPane == PaneDock {
				m.moveCursor(-1)
			} else if SubTab(m.DetailSubTab) == SubTabIPAMMatrix {
				if m.IPAMSelected > 0 {
					m.IPAMSelected--
				}
			} else if SubTab(m.DetailSubTab) == SubTabOverview && m.ActivePanel == 1 {
				m.moveCursor(-1)
			} else {
				m.Viewport.LineUp(1)
			}
			m.updateViewportContent()
			return m, nil

		case "j", "down":
			if m.FocusedPane == PaneDock {
				m.moveCursor(1)
			} else if SubTab(m.DetailSubTab) == SubTabIPAMMatrix {
				if count := len(m.LastFetched.Subnets); count > 0 && m.IPAMSelected < count-1 {
					m.IPAMSelected++
				}
			} else if SubTab(m.DetailSubTab) == SubTabOverview && m.ActivePanel == 1 {
				m.moveCursor(1)
			} else {
				m.Viewport.LineDown(1)
			}
			m.updateViewportContent()
			return m, nil

		case "g":
			if m.FocusedPane == PaneDock {
				m.cursorToTop()
			} else if SubTab(m.DetailSubTab) == SubTabIPAMMatrix {
				m.IPAMSelected = 0
			} else {
				m.Viewport.GotoTop()
			}
			m.updateViewportContent()
			return m, nil

		case "G":
			if m.FocusedPane == PaneDock {
				m.cursorToBottom()
			} else if SubTab(m.DetailSubTab) == SubTabIPAMMatrix {
				if count := len(m.LastFetched.Subnets); count > 0 {
					m.IPAMSelected = count - 1
				}
			} else {
				m.Viewport.GotoBottom()
			}
			m.updateViewportContent()
			return m, nil

		case "ctrl+u":
			if m.FocusedPane == PaneDock {
				m.moveCursor(-5)
			} else {
				m.Viewport.HalfViewUp()
			}
			m.updateViewportContent()
			return m, nil

		case "ctrl+d":
			if m.FocusedPane == PaneDock {
				m.moveCursor(5)
			} else {
				m.Viewport.HalfViewDown()
			}
			m.updateViewportContent()
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.FocusedPane == PaneDetail {
		m.Viewport, cmd = m.Viewport.Update(msg)
	}
	return m, cmd
}

func (m *Model) clampSelections() {
	if count := len(m.LastFetched.Ifaces); count > 0 {
		if m.SelectedIfaceIndex >= count {
			m.SelectedIfaceIndex = count - 1
		}
	} else {
		m.SelectedIfaceIndex = 0
	}

	totalRoutes := len(m.LastFetched.Gateways) + len(m.LastFetched.Routes)
	if totalRoutes > 0 {
		if m.SelectedRouteIndex >= totalRoutes {
			m.SelectedRouteIndex = totalRoutes - 1
		}
	} else {
		m.SelectedRouteIndex = 0
	}

	totalServices := len(m.LastFetched.DHCP) + len(m.LastFetched.VPN) + len(m.LastFetched.Firewall)
	if totalServices > 0 {
		if m.SelectedServiceIndex >= totalServices {
			m.SelectedServiceIndex = totalServices - 1
		}
	} else {
		m.SelectedServiceIndex = 0
	}

	if count := len(m.LastFetched.Subnets); count > 0 {
		if m.IPAMSelected >= count {
			m.IPAMSelected = count - 1
		}
	} else {
		m.IPAMSelected = 0
	}
}

func (m *Model) moveCursor(delta int) {
	switch m.ActivePanel {
	case 1:
		count := len(m.LastFetched.Ifaces)
		if count > 0 {
			m.SelectedIfaceIndex += delta
			if m.SelectedIfaceIndex < 0 {
				m.SelectedIfaceIndex = 0
			}
			if m.SelectedIfaceIndex >= count {
				m.SelectedIfaceIndex = count - 1
			}
		}
	case 2:
		count := len(m.LastFetched.Gateways) + len(m.LastFetched.Routes)
		if count > 0 {
			m.SelectedRouteIndex += delta
			if m.SelectedRouteIndex < 0 {
				m.SelectedRouteIndex = 0
			}
			if m.SelectedRouteIndex >= count {
				m.SelectedRouteIndex = count - 1
			}
		}
	case 3:
		count := len(m.LastFetched.DHCP) + len(m.LastFetched.VPN) + len(m.LastFetched.Firewall)
		if count > 0 {
			m.SelectedServiceIndex += delta
			if m.SelectedServiceIndex < 0 {
				m.SelectedServiceIndex = 0
			}
			if m.SelectedServiceIndex >= count {
				m.SelectedServiceIndex = count - 1
			}
		}
	}
}

func (m *Model) cursorToTop() {
	switch m.ActivePanel {
	case 1:
		m.SelectedIfaceIndex = 0
	case 2:
		m.SelectedRouteIndex = 0
	case 3:
		m.SelectedServiceIndex = 0
	}
}

func (m *Model) cursorToBottom() {
	switch m.ActivePanel {
	case 1:
		if count := len(m.LastFetched.Ifaces); count > 0 {
			m.SelectedIfaceIndex = count - 1
		}
	case 2:
		if count := len(m.LastFetched.Gateways) + len(m.LastFetched.Routes); count > 0 {
			m.SelectedRouteIndex = count - 1
		}
	case 3:
		if count := len(m.LastFetched.DHCP) + len(m.LastFetched.VPN) + len(m.LastFetched.Firewall); count > 0 {
			m.SelectedServiceIndex = count - 1
		}
	}
}

func (m Model) getYankTarget() string {
	switch m.ActivePanel {
	case 1:
		if len(m.LastFetched.Ifaces) > m.SelectedIfaceIndex && m.SelectedIfaceIndex >= 0 {
			iface := m.LastFetched.Ifaces[m.SelectedIfaceIndex]
			if len(iface.IPv4Addresses) > 0 {
				return iface.IPv4Addresses[0]
			}
			if len(iface.IPv6Addresses) > 0 {
				return iface.IPv6Addresses[0]
			}
			if iface.MACAddress != "" {
				return iface.MACAddress
			}
			return iface.Name
		}
	case 2:
		gwCount := len(m.LastFetched.Gateways)
		if m.SelectedRouteIndex < gwCount && m.SelectedRouteIndex >= 0 {
			return m.LastFetched.Gateways[m.SelectedRouteIndex].Address
		}
		rtIdx := m.SelectedRouteIndex - gwCount
		if rtIdx >= 0 && rtIdx < len(m.LastFetched.Routes) {
			rt := m.LastFetched.Routes[rtIdx]
			if rt.Gateway != "" {
				return rt.Gateway
			}
			return rt.Destination
		}
	case 3:
		dhcpCount := len(m.LastFetched.DHCP)
		if m.SelectedServiceIndex < dhcpCount && m.SelectedServiceIndex >= 0 {
			return m.LastFetched.DHCP[m.SelectedServiceIndex].IPAddress
		}
		vpnIdx := m.SelectedServiceIndex - dhcpCount
		if vpnIdx >= 0 && vpnIdx < len(m.LastFetched.VPN) {
			peer := m.LastFetched.VPN[vpnIdx]
			if len(peer.AllowedIPs) > 0 {
				return peer.AllowedIPs[0]
			}
			return peer.PublicKey
		}
	case 0:
		if m.LastFetched.SysInfo != nil {
			return m.LastFetched.SysInfo.Hostname
		}
	}
	return ""
}

func (m *Model) updateViewportDimensions() {
	contentHeight := m.Height - 3
	if contentHeight < 6 {
		contentHeight = 6
	}

	detailWidth := m.Width
	if m.Width >= 110 {
		dockWidth := int(float64(m.Width) * 0.35)
		if dockWidth < 34 {
			dockWidth = 34
		}
		detailWidth = m.Width - dockWidth
	}

	innerDetailWidth := detailWidth - 2
	innerDetailHeight := contentHeight - 4
	if innerDetailWidth < 10 {
		innerDetailWidth = 10
	}
	if innerDetailHeight < 3 {
		innerDetailHeight = 3
	}

	if !m.ViewportReady {
		m.Viewport = viewport.New(innerDetailWidth, innerDetailHeight)
		m.ViewportReady = true
	} else {
		m.Viewport.Width = innerDetailWidth
		m.Viewport.Height = innerDetailHeight
	}
}

func (m *Model) updateViewportContent() {
	m.updateViewportDimensions()
	raw := m.renderDetailContent()
	m.Viewport.SetContent(raw)
}

func (m Model) View() string {
	if m.Width < 20 || m.Height < 10 {
		return "Terminal window too small."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	mainHeight := m.Height - 2
	if mainHeight < 6 {
		mainHeight = 6
	}

	var mainArea string
	if m.Width >= 110 {
		dockWidth := int(float64(m.Width) * 0.35)
		if dockWidth < 34 {
			dockWidth = 34
		}
		detailWidth := m.Width - dockWidth

		leftDock := m.renderLeftDock(dockWidth, mainHeight)
		rightDetail := m.renderRightDetail(detailWidth, mainHeight)
		mainArea = lipgloss.JoinHorizontal(lipgloss.Top, leftDock, rightDetail)
	} else {
		if m.FocusedPane == PaneDock {
			mainArea = m.renderLeftDock(m.Width, mainHeight)
		} else {
			mainArea = m.renderRightDetail(m.Width, mainHeight)
		}
	}

	rendered := lipgloss.JoinVertical(lipgloss.Left, header, mainArea, footer)
	lines := strings.Split(rendered, "\n")
	if len(lines) > m.Height && m.Height > 0 {
		lines = lines[:m.Height]
		rendered = strings.Join(lines, "\n")
	}

	if m.ShowHelpModal {
		rendered = m.overlayHelpModal(rendered)
	}

	return rendered
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

	titleText := StyleHeader.Render(fmt.Sprintf(" ▜▔ MICHIBIKI [%s] ", dev))
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

	res := lipgloss.JoinHorizontal(lipgloss.Top, leftHeader, gapStr, rightHeader)
	return lipgloss.NewStyle().MaxWidth(m.Width).MaxHeight(1).Render(res)
}

func (m Model) renderFooter() string {
	if m.FilterActive {
		filterBar := StyleStatusKey.Render("FILTER: ") + m.FilterInput + " █ (esc/enter to finish)"
		return StyleStatusBar.Width(m.Width).MaxWidth(m.Width).MaxHeight(1).Render(filterBar)
	}
	var keys []string
	keys = append(keys, StyleStatusKey.Render("1-4")+" panels/subtabs")
	keys = append(keys, StyleStatusKey.Render("h/l")+" dock/detail")
	keys = append(keys, StyleStatusKey.Render("j/k")+" nav")
	keys = append(keys, StyleStatusKey.Render("[]")+" tabs")
	keys = append(keys, StyleStatusKey.Render("t")+" traffic")
	keys = append(keys, StyleStatusKey.Render("i")+" ipam")
	keys = append(keys, StyleStatusKey.Render("c")+" config")
	if m.isPluginProvider() {
		keys = append(keys, StyleStatusKey.Render("p")+" plugin")
	}
	keys = append(keys, StyleStatusKey.Render("/")+" filter")
	keys = append(keys, StyleStatusKey.Render("y")+" yank")
	keys = append(keys, StyleStatusKey.Render("r")+" reload")
	keys = append(keys, StyleStatusKey.Render("?")+" help")
	keys = append(keys, StyleStatusKey.Render("q")+" quit")
	footerText := strings.Join(keys, " • ")
	if m.StatusFlash != "" {
		footerText = StyleBadgeWarning.Render(" "+m.StatusFlash+" ") + "  " + footerText
	} else if m.Loading {
		footerText += "  " + StyleProgressFilledWarn.Render("[FETCHING DATA...]")
	}

	return StyleStatusBar.Width(m.Width).MaxWidth(m.Width).MaxHeight(1).Render(footerText)
}

func (m Model) renderLeftDock(width, height int) string {
	h0 := height * 25 / 100
	h1 := height * 30 / 100
	h2 := height * 22 / 100
	h3 := height - h0 - h1 - h2

	if h0 < 4 {
		h0 = 4
	}
	if h1 < 4 {
		h1 = 4
	}
	if h2 < 4 {
		h2 = 4
	}
	if h3 < 4 {
		h3 = 4
	}

	p0Active := (m.FocusedPane == PaneDock && m.ActivePanel == 0)
	p1Active := (m.FocusedPane == PaneDock && m.ActivePanel == 1)
	p2Active := (m.FocusedPane == PaneDock && m.ActivePanel == 2)
	p3Active := (m.FocusedPane == PaneDock && m.ActivePanel == 3)

	panel0 := m.renderPanel0Status(width, h0, p0Active)
	panel1 := m.renderPanel1Interfaces(width, h1, p1Active)
	panel2 := m.renderPanel2Routing(width, h2, p2Active)
	panel3 := m.renderPanel3Services(width, h3, p3Active)

	return lipgloss.JoinVertical(lipgloss.Left, panel0, panel1, panel2, panel3)
}

func (m Model) renderPanel0Status(width, height int, active bool) string {
	title := "1 Status & Health"
	if active {
		title = "[1] Status & Health"
	}

	var lines []string
	sys := m.LastFetched.SysInfo
	if sys == nil {
		lines = append(lines, theme.StyleMuted.Render("Loading telemetry..."))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s", theme.StyleSubTitle.Render("Host:"), sys.Hostname))
		osStr := fmt.Sprintf("%s %s", sys.OS, sys.Version)
		if len(osStr) > width-8 {
			osStr = osStr[:width-8]
		}
		lines = append(lines, fmt.Sprintf("%s %s", theme.StyleSubTitle.Render("OS:  "), osStr))

		uptimeStr := viewsFormatUptime(sys.UptimeSeconds)
		lines = append(lines, fmt.Sprintf("%s %s", theme.StyleSubTitle.Render("Up:  "), uptimeStr))

		barWidth := width - 16
		if barWidth < 6 {
			barWidth = 6
		}
		cpuBar := RenderProgressBar(sys.CPUUsagePct, barWidth)
		lines = append(lines, fmt.Sprintf("CPU %s %3.0f%%", cpuBar, sys.CPUUsagePct))

		var memPct float64
		if sys.MemoryTotalBytes > 0 {
			memPct = (float64(sys.MemoryUsedBytes) / float64(sys.MemoryTotalBytes)) * 100.0
		}
		memBar := RenderProgressBar(memPct, barWidth)
		lines = append(lines, fmt.Sprintf("MEM %s %3.0f%%", memBar, memPct))
	}

	return renderDockBox(title, active, width, height, lines)
}

func (m Model) renderPanel1Interfaces(width, height int, active bool) string {
	ifaces := m.LastFetched.Ifaces
	title := fmt.Sprintf("2 Interfaces (%d)", len(ifaces))
	if active {
		title = fmt.Sprintf("[2] Interfaces (%d)", len(ifaces))
	}

	var lines []string
	if len(ifaces) == 0 {
		lines = append(lines, theme.StyleMuted.Render("No interfaces found."))
	} else {
		maxVisible := height - 2
		startIdx := 0
		if m.SelectedIfaceIndex >= maxVisible {
			startIdx = m.SelectedIfaceIndex - maxVisible + 1
		}

		for i := startIdx; i < len(ifaces) && len(lines) < maxVisible; i++ {
			iface := ifaces[i]
			isSel := (i == m.SelectedIfaceIndex)

			prefix := "  "
			if isSel && active {
				prefix = StyleCursor.Render("❯ ")
			} else if isSel {
				prefix = "> "
			}

			operBadge := StyleBadgeOnline.Render("UP")
			if iface.OperStatus != model.OperStatusUp {
				operBadge = StyleBadgeOffline.Render("DN")
			}

			ip := "-"
			if len(iface.IPv4Addresses) > 0 {
				ip = iface.IPv4Addresses[0]
			} else if len(iface.IPv6Addresses) > 0 {
				ip = iface.IPv6Addresses[0]
			}
			if len(ip) > 16 {
				ip = ip[:14] + ".."
			}

			name := iface.Name
			if len(name) > 8 {
				name = name[:7] + "…"
			}

			row := fmt.Sprintf("%s%-8s %s %s", prefix, name, operBadge, ip)
			if isSel {
				row = StyleDockSelected.Render(row)
			}
			lines = append(lines, row)
		}
	}

	return renderDockBox(title, active, width, height, lines)
}

func (m Model) renderPanel2Routing(width, height int, active bool) string {
	gws := m.LastFetched.Gateways
	routes := m.LastFetched.Routes
	title := fmt.Sprintf("3 Routing & Gateways (%d)", len(gws)+len(routes))
	if active {
		title = fmt.Sprintf("[3] Routing & Gateways (%d)", len(gws)+len(routes))
	}

	var lines []string
	totalItems := len(gws) + len(routes)
	if totalItems == 0 {
		lines = append(lines, theme.StyleMuted.Render("No routes / gateways."))
	} else {
		maxVisible := height - 2
		startIdx := 0
		if m.SelectedRouteIndex >= maxVisible {
			startIdx = m.SelectedRouteIndex - maxVisible + 1
		}

		idx := 0
		for _, gw := range gws {
			if idx >= startIdx && len(lines) < maxVisible {
				isSel := (idx == m.SelectedRouteIndex)
				prefix := "  "
				if isSel && active {
					prefix = StyleCursor.Render("❯ ")
				} else if isSel {
					prefix = "> "
				}

				badge := StyleBadgeOnline.Render("GW")
				if gw.Status != model.GatewayOnline {
					badge = StyleBadgeOffline.Render("GW")
				}

				row := fmt.Sprintf("%s%s %-12s %.0fms", prefix, badge, gw.Name, gw.LatencyMs)
				if isSel {
					row = StyleDockSelected.Render(row)
				}
				lines = append(lines, row)
			}
			idx++
		}

		for _, rt := range routes {
			if idx >= startIdx && len(lines) < maxVisible {
				isSel := (idx == m.SelectedRouteIndex)
				prefix := "  "
				if isSel && active {
					prefix = StyleCursor.Render("❯ ")
				} else if isSel {
					prefix = "> "
				}

				gwStr := rt.Gateway
				if gwStr == "" {
					gwStr = rt.Interface
				}
				if len(gwStr) > 10 {
					gwStr = gwStr[:8] + ".."
				}

				dst := rt.Destination
				if len(dst) > 14 {
					dst = dst[:12] + ".."
				}

				row := fmt.Sprintf("%s%-14s %s", prefix, dst, gwStr)
				if isSel {
					row = StyleDockSelected.Render(row)
				}
				lines = append(lines, row)
			}
			idx++
		}
	}

	return renderDockBox(title, active, width, height, lines)
}

func (m Model) renderPanel3Services(width, height int, active bool) string {
	fwCount := len(m.LastFetched.Firewall)
	dhcpCount := len(m.LastFetched.DHCP)
	vpnCount := len(m.LastFetched.VPN)

	title := "4 Services & Security"
	if active {
		title = "[4] Services & Security"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s %s %d rules",
		theme.StyleSubTitle.Render("Firewall: "),
		StyleBadgeInfo.Render("FW"),
		fwCount,
	))
	lines = append(lines, fmt.Sprintf("%s %s %d leases",
		theme.StyleSubTitle.Render("DHCP:     "),
		StyleBadgeOnline.Render("DHCP"),
		dhcpCount,
	))
	lines = append(lines, fmt.Sprintf("%s %s %d peers",
		theme.StyleSubTitle.Render("VPN:      "),
		StyleBadgeOnline.Render("WG"),
		vpnCount,
	))

	arpCount := len(m.LastFetched.ARP)
	lines = append(lines, fmt.Sprintf("%s %s %d entries",
		theme.StyleSubTitle.Render("ARP:      "),
		StyleBadgeInfo.Render("ARP"),
		arpCount,
	))

	return renderDockBox(title, active, width, height, lines)
}

func (m Model) renderRightDetail(width, height int) string {
	isDetailActive := (m.FocusedPane == PaneDetail)

	var tabStrs []string
	for idx, name := range m.getSubTabNames() {
		if idx == m.DetailSubTab {
			tabStrs = append(tabStrs, StyleSubTabActive.Render(name))
		} else {
			tabStrs = append(tabStrs, StyleSubTabInactive.Render(name))
		}
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabStrs...)

	borderColor := ColorDarkGray
	if isDetailActive {
		borderColor = ColorCyan
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	title := " Detail Inspector "
	if isDetailActive {
		title = " [Detail Inspector] "
	}
	titleRendered := theme.StyleTitle.Render(title)
	if !isDetailActive {
		titleRendered = theme.StyleSubTitle.Render(title)
	}
	remainTop := width - 2 - lipgloss.Width(titleRendered)
	if remainTop < 0 {
		remainTop = 0
	}
	topLine := borderStyle.Render("╭") + titleRendered + borderStyle.Render(strings.Repeat("─", remainTop)+"╮")

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	tabsPad := innerWidth - lipgloss.Width(tabsRow)
	if tabsPad < 0 {
		tabsPad = 0
	}
	subTabsLine := borderStyle.Render("│") + tabsRow + strings.Repeat(" ", tabsPad) + borderStyle.Render("│")

	divLine := borderStyle.Render("├" + strings.Repeat("─", innerWidth) + "┤")

	viewportView := m.Viewport.View()
	contentLines := strings.Split(viewportView, "\n")
	viewportHeight := height - 4
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	var res []string
	res = append(res, topLine, subTabsLine, divLine)

	for i := 0; i < viewportHeight; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		w := lipgloss.Width(line)
		pad := innerWidth - w
		if pad < 0 {
			line = lipgloss.NewStyle().MaxWidth(innerWidth).Render(line)
			pad = innerWidth - lipgloss.Width(line)
			if pad < 0 {
				pad = 0
			}
		}
		row := borderStyle.Render("│") + line + strings.Repeat(" ", pad) + borderStyle.Render("│")
		res = append(res, row)
	}

	bottomLine := borderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	res = append(res, bottomLine)

	return strings.Join(res, "\n")
}

func (m Model) renderDetailContent() string {
	contentHeight := m.Height - 6
	if contentHeight < 5 {
		contentHeight = 5
	}
	contentWidth := m.Viewport.Width
	if contentWidth < 40 {
		contentWidth = 40
	}

	switch SubTab(m.DetailSubTab) {
	case SubTabOverview:
		switch m.ActivePanel {
		case 0:
			if plug, ok := m.Provider.(interface {
				ID() string
				Version() string
				Capabilities() provider.Capabilities
			}); ok && m.Provider.Name() != "opnsense" && m.Provider.Name() != "pfsense" && m.Provider.Name() != "routeros" && m.Provider.Name() != "openwrt" && m.Provider.Name() != "vyos" && m.Provider.Name() != "frr" && m.Provider.Name() != "mock" {
				return views.RenderPluginCustomUI(views.PluginViewData{
					PluginName:   plug.ID(),
					Version:      plug.Version(),
					Capabilities: plug.Capabilities().Strings(),
					SystemInfo:   m.LastFetched.SysInfo,
					Interfaces:   m.LastFetched.Ifaces,
					Stats:        m.LastFetched.Stats,
					Filter:       m.FilterInput,
				}, contentWidth, contentHeight)
			}
			return views.RenderDashboard(views.DashboardData{
				SystemInfo: m.LastFetched.SysInfo,
				Interfaces: m.LastFetched.Ifaces,
				Gateways:   m.LastFetched.Gateways,
				Error:      m.LastFetched.Err,
			}, contentWidth, contentHeight)
		case 1:
			statsMap := make(map[string]model.InterfaceStats)
			for _, s := range m.LastFetched.Stats {
				statsMap[s.InterfaceName] = s
			}
			return views.RenderInterfaces(views.InterfacesData{
				Interfaces:    m.LastFetched.Ifaces,
				Stats:         statsMap,
				SelectedIndex: m.SelectedIfaceIndex,
				Filter:        m.FilterInput,
				Error:         m.LastFetched.Err,
			}, contentWidth, contentHeight)
		case 2:
			return views.RenderRouting(views.RoutingData{
				Routes:    m.LastFetched.Routes,
				Gateways:  m.LastFetched.Gateways,
				Neighbors: m.LastFetched.BGP,
				Filter:    m.FilterInput,
				Error:     m.LastFetched.Err,
			}, contentWidth, contentHeight)
		case 3:
			fw := views.RenderFirewall(views.FirewallData{
				Rules:  m.LastFetched.Firewall,
				NAT:    m.LastFetched.NAT,
				Filter: m.FilterInput,
				Error:  m.LastFetched.Err,
			}, contentWidth, contentHeight/2)
			dhcp := views.RenderDHCP(views.DHCPData{
				Leases: m.LastFetched.DHCP,
				Filter: m.FilterInput,
				Error:  m.LastFetched.Err,
			}, contentWidth, contentHeight/2)
			return lipgloss.JoinVertical(lipgloss.Left, fw, dhcp)
		default:
			return views.RenderDashboard(views.DashboardData{
				SystemInfo: m.LastFetched.SysInfo,
				Interfaces: m.LastFetched.Ifaces,
				Gateways:   m.LastFetched.Gateways,
				Error:      m.LastFetched.Err,
			}, contentWidth, contentHeight)
		}

	case SubTabLiveTraffic:
		return views.RenderMonitor(views.MonitorData{
			Stats:   m.LastFetched.Stats,
			History: m.TrafficHistory,
			Filter:  m.FilterInput,
			Error:   m.LastFetched.Err,
		}, contentWidth, contentHeight)

	case SubTabIPAMMatrix:
		return views.RenderIPAM(views.IPAMData{
			Subnets:       m.LastFetched.Subnets,
			SelectedIndex: m.IPAMSelected,
			Filter:        m.FilterInput,
			Error:         m.LastFetched.Err,
		}, contentWidth, contentHeight)

	case SubTabRunningConfig:
		return views.RenderConfig(views.ConfigData{
			RunningConfig: m.LastFetched.Config,
			SafetyStatus:  "Safety Engine Active — rollback tracking ready",
			Filter:        m.FilterInput,
			Error:         m.LastFetched.Err,
		}, contentWidth, contentHeight)

	case SubTabPlugin:
		var plugName, plugVer string
		var plugCaps []string
		if plug, ok := m.Provider.(interface {
			ID() string
			Version() string
			Capabilities() provider.Capabilities
		}); ok {
			plugName = plug.ID()
			plugVer = plug.Version()
			plugCaps = plug.Capabilities().Strings()
		} else if m.Provider != nil {
			plugName = m.Provider.Name()
			plugCaps = m.Provider.Capabilities().Strings()
		}
		return views.RenderPluginCustomUI(views.PluginViewData{
			PluginName:   plugName,
			Version:      plugVer,
			Capabilities: plugCaps,
			SystemInfo:   m.LastFetched.SysInfo,
			Interfaces:   m.LastFetched.Ifaces,
			Stats:        m.LastFetched.Stats,
			Filter:       m.FilterInput,
		}, contentWidth, contentHeight)
	}
	return ""
}

func (m Model) overlayHelpModal(baseView string) string {
	modalContent := fmt.Sprintf(`%s

%s
  1..4        Switch Dock Panels / Detail Sub-Tabs
  Tab         Cycle Panels Clockwise
  Shift+Tab   Cycle Panels Counter-Clockwise
  h           Focus Left Dock
  l / Enter   Focus Right Detail Inspector

%s
  j / Down    Next Item / Scroll Down
  k / Up      Previous Item / Scroll Up
  g / G       Top / Bottom of List
  ctrl+d      Half-Page Down
  ctrl+u      Half-Page Up

%s
  [ / ]       Previous / Next Detail Sub-Tab
  o           Open Overview Sub-Tab
  t           Open Live Traffic Monitor
  i / m       Open IPAM Matrix
  c / C       Jump to Running Config Inspector
  p / P       Open Plugin Custom UI (if plugin provider)
  /           Enter Filter / Search Mode
  Esc         Exit Search / Clear Filter
  y           Yank IP / MAC / Hostname (OSC 52)
  r           Force Telemetry Reload
  ?           Toggle this Help Modal
  q / ctrl+c  Quit Michibiki TUI
`,
		StyleTitle.Render("MICHIBIKI DOCK NAVIGATION & NEOVIM KEYMAPS"),
		StyleSubTitle.Render("PANEL & FOCUS SWITCHING"),
		StyleSubTitle.Render("NAVIGATION & SCROLLING"),
		StyleSubTitle.Render("ACTIONS & SUB-TABS"),
	)

	modalBox := StyleModal.Render(modalContent)

	bgLines := strings.Split(baseView, "\n")
	modalLines := strings.Split(modalBox, "\n")

	for len(bgLines) < m.Height {
		bgLines = append(bgLines, strings.Repeat(" ", m.Width))
	}

	modalH := len(modalLines)
	modalW := 0
	for _, l := range modalLines {
		if w := lipgloss.Width(l); w > modalW {
			modalW = w
		}
	}

	top := (m.Height - modalH) / 2
	if top < 0 {
		top = 0
	}
	left := (m.Width - modalW) / 2
	if left < 0 {
		left = 0
	}

	res := make([]string, len(bgLines))
	for y := 0; y < len(bgLines); y++ {
		if y >= top && y < top+modalH {
			mLine := modalLines[y-top]
			leftPad := strings.Repeat(" ", left)
			rightPadLen := m.Width - left - lipgloss.Width(mLine)
			if rightPadLen < 0 {
				rightPadLen = 0
			}
			rightPad := strings.Repeat(" ", rightPadLen)
			res[y] = leftPad + mLine + rightPad
		} else {
			res[y] = bgLines[y]
		}
	}

	return strings.Join(res, "\n")
}

func renderDockBox(title string, isActive bool, width int, height int, innerLines []string) string {
	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}

	borderColor := theme.ColorDarkGray
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorLightGray)
	if isActive {
		borderColor = theme.ColorAccent
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.ColorAccent)
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	titleRendered := " " + titleStyle.Render(title) + " "
	titleLen := lipgloss.Width(titleRendered)
	remain := width - 2 - 1 - titleLen
	if remain < 0 {
		remain = 0
	}
	top := borderStyle.Render("╭─") + titleRendered + borderStyle.Render(strings.Repeat("─", remain)+"╮")

	bodyHeight := height - 2
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	var res []string
	res = append(res, top)

	for i := 0; i < bodyHeight; i++ {
		line := ""
		if i < len(innerLines) {
			line = innerLines[i]
		}
		if lipgloss.Width(line) > innerWidth {
			line = lipgloss.NewStyle().MaxWidth(innerWidth).Render(line)
		}
		w := lipgloss.Width(line)
		pad := innerWidth - w
		if pad < 0 {
			pad = 0
		}
		middle := borderStyle.Render("│") + line + strings.Repeat(" ", pad) + borderStyle.Render("│")
		res = append(res, middle)
	}

	bottom := borderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	res = append(res, bottom)

	return strings.Join(res, "\n")
}

func viewsFormatUptime(seconds uint64) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %dm", d, h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}
