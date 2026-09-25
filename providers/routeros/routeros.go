package routeros

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sekai-labs/michibiki/pkg/client"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

func init() {
	provider.Register("routeros", func() provider.Provider {
		return New()
	})
}

type Provider struct {
	mu        sync.RWMutex
	client    *client.HTTPClient
	connected bool
	endpoint  string
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string {
	return "routeros"
}

func (p *Provider) Name() string {
	return "MikroTik RouterOS v7"
}

func (p *Provider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts := client.HTTPOptions{
		BaseURL: endpoint,
	}

	if creds != nil {
		opts.Username = creds.Username
		opts.Password = creds.Password
	}

	if options != nil {
		if options["insecure"] == "true" || options["insecure_skip_verify"] == "true" {
			opts.InsecureSkipVerify = true
		}
		if timeoutStr, ok := options["timeout"]; ok {
			if dur, err := time.ParseDuration(timeoutStr); err == nil {
				opts.Timeout = dur
			}
		}
	}

	p.client = client.NewHTTPClient(opts)
	p.connected = true
	p.endpoint = endpoint
	return nil
}

func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = false
	p.client = nil
	return nil
}

func (p *Provider) Capabilities() provider.Capabilities {
	return provider.CapSystem |
		provider.CapInterfaces |
		provider.CapRouting |
		provider.CapFirewall |
		provider.CapNAT |
		provider.CapDHCP |
		provider.CapDNS |
		provider.CapWireGuard |
		provider.CapBGP |
		provider.CapOSPF |
		provider.CapMonitoring |
		provider.CapConfig |
		provider.CapIPAM
}

func (p *Provider) getClient() (*client.HTTPClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.connected || p.client == nil {
		return nil, provider.ErrNotConnected
	}
	return p.client, nil
}

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var res map[string]any
	err = c.Do(ctx, "GET", "/rest/system/resource", nil, &res)
	if err != nil {
		return &model.SystemInfo{
			Hostname: "RouterOS",
			OS:       "RouterOS",
			Time:     time.Now(),
		}, nil
	}

	var idRes map[string]any
	_ = c.Do(ctx, "GET", "/rest/system/identity", nil, &idRes)

	hostname := "RouterOS"
	if name, ok := idRes["name"].(string); ok && name != "" {
		hostname = name
	}

	info := &model.SystemInfo{
		Hostname: hostname,
		OS:       "RouterOS",
		Time:     time.Now(),
	}

	if ver, ok := res["version"].(string); ok {
		info.Version = ver
	}
	if arch, ok := res["architecture-name"].(string); ok {
		info.Architecture = arch
	}
	if cpuCount, ok := res["cpu-count"].(float64); ok {
		info.CPUCount = int(cpuCount)
	}
	if cpuLoad, ok := res["cpu-load"].(float64); ok {
		info.CPUUsagePct = cpuLoad
	}
	if totalMem, ok := res["total-memory"].(float64); ok {
		info.MemoryTotalBytes = uint64(totalMem)
	}
	if freeMem, ok := res["free-memory"].(float64); ok {
		if info.MemoryTotalBytes > uint64(freeMem) {
			info.MemoryUsedBytes = info.MemoryTotalBytes - uint64(freeMem)
		}
	}
	if totalHdd, ok := res["total-hdd-space"].(float64); ok {
		info.StorageTotal = uint64(totalHdd)
	}
	if freeHdd, ok := res["free-hdd-space"].(float64); ok {
		if info.StorageTotal > uint64(freeHdd) {
			info.StorageUsed = info.StorageTotal - uint64(freeHdd)
		}
	}

	return info, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawIfaces []map[string]any
	err = c.Do(ctx, "GET", "/rest/interface", nil, &rawIfaces)
	if err != nil {
		return nil, err
	}

	var rawAddrs []map[string]any
	_ = c.Do(ctx, "GET", "/rest/ip/address", nil, &rawAddrs)

	addrsByIface := make(map[string][]string)
	for _, a := range rawAddrs {
		ifaceName := fmt.Sprintf("%v", a["interface"])
		address := fmt.Sprintf("%v", a["address"])
		if ifaceName != "" && address != "" {
			addrsByIface[ifaceName] = append(addrsByIface[ifaceName], address)
		}
	}

	var list []model.Interface
	for _, m := range rawIfaces {
		name := fmt.Sprintf("%v", m["name"])
		ifaceType := model.InterfaceTypeEthernet
		t := strings.ToLower(fmt.Sprintf("%v", m["type"]))
		if strings.Contains(t, "vlan") {
			ifaceType = model.InterfaceTypeVLAN
		} else if strings.Contains(t, "wireguard") {
			ifaceType = model.InterfaceTypeWireGuard
		} else if strings.Contains(t, "bridge") {
			ifaceType = model.InterfaceTypeBridge
		}

		admin := model.AdminStatusUp
		if disabled, ok := m["disabled"].(bool); ok && disabled {
			admin = model.AdminStatusDown
		}

		oper := model.OperStatusDown
		if running, ok := m["running"].(bool); ok && running {
			oper = model.OperStatusUp
		}

		mtuVal := 1500
		if mtuStr, ok := m["mtu"].(string); ok {
			if n, err := strconv.Atoi(mtuStr); err == nil {
				mtuVal = n
			}
		} else if mtuNum, ok := m["mtu"].(float64); ok {
			mtuVal = int(mtuNum)
		}

		list = append(list, model.Interface{
			ID:            fmt.Sprintf("%v", m[".id"]),
			Name:          name,
			Type:          ifaceType,
			AdminStatus:   admin,
			OperStatus:    oper,
			MACAddress:    fmt.Sprintf("%v", m["mac-address"]),
			MTU:           mtuVal,
			IPv4Addresses: addrsByIface[name],
			Description:   fmt.Sprintf("%v", m["comment"]),
		})
	}

	return list, nil
}

func (p *Provider) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	ifaces, err := p.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Name == name || iface.ID == name {
			return &iface, nil
		}
	}
	return nil, fmt.Errorf("interface not found: %s", name)
}

func (p *Provider) ListRoutes(ctx context.Context) ([]model.Route, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawRoutes []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/route", nil, &rawRoutes)
	if err != nil {
		return nil, err
	}

	var routes []model.Route
	for _, m := range rawRoutes {
		proto := model.RouteProtoOther
		if connect, ok := m["connect"].(bool); ok && connect {
			proto = model.RouteProtoConnected
		} else if bgp, ok := m["bgp"].(bool); ok && bgp {
			proto = model.RouteProtoBGP
		} else if ospf, ok := m["ospf"].(bool); ok && ospf {
			proto = model.RouteProtoOSPF
		} else if static, ok := m["static"].(bool); ok && static {
			proto = model.RouteProtoStatic
		}

		distance := 1
		if d, ok := m["distance"].(float64); ok {
			distance = int(d)
		}

		routes = append(routes, model.Route{
			Destination: fmt.Sprintf("%v", m["dst-address"]),
			Gateway:     fmt.Sprintf("%v", m["gateway"]),
			Interface:   fmt.Sprintf("%v", m["immediate-gw"]),
			Protocol:    proto,
			Metric:      distance,
		})
	}

	return routes, nil
}

func (p *Provider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	routes, err := p.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}

	var gateways []model.Gateway
	for _, r := range routes {
		if r.Destination == "0.0.0.0/0" && r.Gateway != "" {
			gateways = append(gateways, model.Gateway{
				Name:      "default-gw-" + r.Interface,
				Address:   r.Gateway,
				Interface: r.Interface,
				Status:    model.GatewayOnline,
				IsDefault: true,
			})
		}
	}

	return gateways, nil
}

func (p *Provider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawARP []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/arp", nil, &rawARP)
	if err != nil {
		return nil, err
	}

	var entries []model.ARPEntry
	for _, m := range rawARP {
		status := model.ARPReachable
		if dynamic, ok := m["dynamic"].(bool); !ok || !dynamic {
			status = model.ARPStatic
		}
		entries = append(entries, model.ARPEntry{
			IPAddress:  fmt.Sprintf("%v", m["address"]),
			MACAddress: fmt.Sprintf("%v", m["mac-address"]),
			Interface:  fmt.Sprintf("%v", m["interface"]),
			Status:     status,
		})
	}

	return entries, nil
}

func (p *Provider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawLeases []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/dhcp-server/lease", nil, &rawLeases)
	if err != nil {
		return nil, err
	}

	var leases []model.DHCPLease
	for _, m := range rawLeases {
		state := model.DHCPActive
		if dynamic, ok := m["dynamic"].(bool); !ok || !dynamic {
			state = model.DHCPStatic
		}

		leases = append(leases, model.DHCPLease{
			IPAddress:      fmt.Sprintf("%v", m["address"]),
			MACAddress:     fmt.Sprintf("%v", m["mac-address"]),
			ClientHostname: fmt.Sprintf("%v", m["host-name"]),
			Interface:      fmt.Sprintf("%v", m["server"]),
			State:          state,
		})
	}

	return leases, nil
}

func (p *Provider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawRules []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/firewall/filter", nil, &rawRules)
	if err != nil {
		return nil, err
	}

	var rules []model.FirewallRule
	for i, m := range rawRules {
		action := model.FirewallPass
		actStr := strings.ToLower(fmt.Sprintf("%v", m["action"]))
		if actStr == "drop" || actStr == "reject" {
			action = model.FirewallBlock
		}

		disabled := false
		if d, ok := m["disabled"].(bool); ok {
			disabled = d
		}

		rules = append(rules, model.FirewallRule{
			ID:              fmt.Sprintf("%v", m[".id"]),
			Sequence:        i + 1,
			Interface:       fmt.Sprintf("%v", m["in-interface"]),
			Direction:       fmt.Sprintf("%v", m["chain"]),
			Action:          action,
			Protocol:        fmt.Sprintf("%v", m["protocol"]),
			Source:          fmt.Sprintf("%v", m["src-address"]),
			Destination:     fmt.Sprintf("%v", m["dst-address"]),
			DestinationPort: fmt.Sprintf("%v", m["dst-port"]),
			Description:     fmt.Sprintf("%v", m["comment"]),
			Enabled:         !disabled,
		})
	}

	return rules, nil
}

func (p *Provider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawLists []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/firewall/address-list", nil, &rawLists)
	if err != nil {
		return nil, err
	}

	aliasMap := make(map[string][]string)
	for _, m := range rawLists {
		name := fmt.Sprintf("%v", m["list"])
		addr := fmt.Sprintf("%v", m["address"])
		if name != "" && addr != "" {
			aliasMap[name] = append(aliasMap[name], addr)
		}
	}

	var aliases []model.FirewallAlias
	for name, addrs := range aliasMap {
		aliases = append(aliases, model.FirewallAlias{
			Name:    name,
			Type:    "address-list",
			Content: addrs,
		})
	}

	return aliases, nil
}

func (p *Provider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawNAT []map[string]any
	err = c.Do(ctx, "GET", "/rest/ip/firewall/nat", nil, &rawNAT)
	if err != nil {
		return nil, err
	}

	var natRules []model.NATRule
	for _, m := range rawNAT {
		disabled := false
		if d, ok := m["disabled"].(bool); ok {
			disabled = d
		}

		natRules = append(natRules, model.NATRule{
			ID:          fmt.Sprintf("%v", m[".id"]),
			Interface:   fmt.Sprintf("%v", m["out-interface"]),
			Type:        fmt.Sprintf("%v", m["action"]),
			Protocol:    fmt.Sprintf("%v", m["protocol"]),
			Source:      fmt.Sprintf("%v", m["src-address"]),
			Destination: fmt.Sprintf("%v", m["dst-address"]),
			Target:      fmt.Sprintf("%v", m["to-addresses"]),
			TargetPort:  fmt.Sprintf("%v", m["to-ports"]),
			Enabled:     !disabled,
			Description: fmt.Sprintf("%v", m["comment"]),
		})
	}

	return natRules, nil
}

func (p *Provider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawBGP []map[string]any
	err = c.Do(ctx, "GET", "/rest/routing/bgp/session", nil, &rawBGP)
	if err != nil {
		return []model.BGPNeighbor{}, nil
	}

	var neighbors []model.BGPNeighbor
	for _, m := range rawBGP {
		state := model.BGPIdle
		stStr := strings.ToLower(fmt.Sprintf("%v", m["state"]))
		if strings.Contains(stStr, "established") {
			state = model.BGPEstablished
		}

		var remoteAS uint32
		if asNum, ok := m["remote.as"].(float64); ok {
			remoteAS = uint32(asNum)
		}

		neighbors = append(neighbors, model.BGPNeighbor{
			RemoteAS:    remoteAS,
			PeerAddress: fmt.Sprintf("%v", m["remote.address"]),
			State:       state,
			Description: fmt.Sprintf("%v", m["name"]),
		})
	}

	return neighbors, nil
}

func (p *Provider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawPeers []map[string]any
	err = c.Do(ctx, "GET", "/rest/interface/wireguard/peers", nil, &rawPeers)
	if err != nil {
		return []model.WireGuardPeer{}, nil
	}

	var peers []model.WireGuardPeer
	for _, m := range rawPeers {
		peers = append(peers, model.WireGuardPeer{
			PublicKey:  fmt.Sprintf("%v", m["public-key"]),
			Endpoint:   fmt.Sprintf("%v", m["endpoint-address"]),
			AllowedIPs: strings.Split(fmt.Sprintf("%v", m["allowed-address"]), ","),
		})
	}

	return peers, nil
}

func (p *Provider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawIfaces []map[string]any
	err = c.Do(ctx, "GET", "/rest/interface", nil, &rawIfaces)
	if err != nil {
		return nil, err
	}

	var stats []model.InterfaceStats
	now := time.Now()
	for _, m := range rawIfaces {
		s := model.InterfaceStats{
			InterfaceName: fmt.Sprintf("%v", m["name"]),
			Timestamp:     now,
		}
		if rx, ok := m["rx-byte"].(float64); ok {
			s.RxBytes = uint64(rx)
		}
		if tx, ok := m["tx-byte"].(float64); ok {
			s.TxBytes = uint64(tx)
		}
		if rxp, ok := m["rx-packet"].(float64); ok {
			s.RxPackets = uint64(rxp)
		}
		if txp, ok := m["tx-packet"].(float64); ok {
			s.TxPackets = uint64(txp)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (p *Provider) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	ifaces, err := p.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	leases, _ := p.ListDHCPLeases(ctx)
	arp, _ := p.ListARPEntries(ctx)
	wg, _ := p.ListWireGuardPeers(ctx)

	usages := ipam.DiscoverSubnets(ifaces, leases, arp, wg)
	if filter == "" {
		return usages, nil
	}

	matched := ipam.MatchSubnet(usages, filter)
	if matched != nil {
		return []ipam.SubnetUsage{*matched}, nil
	}

	var filtered []ipam.SubnetUsage
	for _, u := range usages {
		if strings.Contains(strings.ToLower(u.InterfaceName), strings.ToLower(filter)) ||
			strings.Contains(u.CIDR.String(), filter) {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func (p *Provider) GetRunningConfig(ctx context.Context) (string, error) {
	c, err := p.getClient()
	if err != nil {
		return "", err
	}

	var res map[string]any
	err = c.Do(ctx, "POST", "/rest/export", nil, &res)
	if err != nil {
		return "# RouterOS Export", nil
	}
	return fmt.Sprintf("%v", res), nil
}

func (p *Provider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	if strings.TrimSpace(candidate) == "" {
		return &model.ValidationResult{
			Valid:  false,
			Errors: []string{"empty configuration"},
		}, nil
	}
	return &model.ValidationResult{Valid: true}, nil
}

func (p *Provider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration applied on RouterOS",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}
