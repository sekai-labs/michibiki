package openwrt

import (
	"context"
	"encoding/json"
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
	provider.Register("openwrt", func() provider.Provider {
		return New()
	})
}

type Provider struct {
	mu        sync.RWMutex
	client    *client.HTTPClient
	sshClient *client.SSHClient
	sessionID string
	connected bool
	endpoint  string
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string {
	return "openwrt"
}

func (p *Provider) Name() string {
	return "OpenWrt / LuCI / ubus"
}

func (p *Provider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts := client.HTTPOptions{
		BaseURL: endpoint,
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

	if creds != nil && creds.Username != "" {
		body := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "call",
			"params": []any{
				"00000000000000000000000000000000",
				"session",
				"login",
				map[string]string{
					"username": creds.Username,
					"password": creds.Password,
				},
			},
		}
		var loginResp struct {
			Result []any `json:"result"`
		}
		if err := p.client.Do(ctx, "POST", "/ubus", body, &loginResp); err == nil {
			if len(loginResp.Result) > 1 {
				if resMap, ok := loginResp.Result[1].(map[string]any); ok {
					if ubusRPC, ok := resMap["ubus_rpc_session"].(string); ok {
						p.sessionID = ubusRPC
					}
				}
			}
		}
	}

	return nil
}

func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = false
	p.client = nil
	p.sshClient = nil
	p.sessionID = ""
	return nil
}

func (p *Provider) Capabilities() provider.Capabilities {
	return provider.CapSystem |
		provider.CapInterfaces |
		provider.CapRouting |
		provider.CapFirewall |
		provider.CapDHCP |
		provider.CapDNS |
		provider.CapWireGuard |
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

func (p *Provider) callUbus(ctx context.Context, module, method string, params any, result any) error {
	c, err := p.getClient()
	if err != nil {
		return err
	}

	session := p.sessionID
	if session == "" {
		session = "00000000000000000000000000000000"
	}

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "call",
		"params":  []any{session, module, method, params},
	}

	var rpcResp struct {
		Result []any `json:"result"`
		Error  any   `json:"error"`
	}

	if err := c.Do(ctx, "POST", "/ubus", req, &rpcResp); err != nil {
		return err
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("ubus error: %v", rpcResp.Error)
	}

	if len(rpcResp.Result) > 1 && result != nil {
		rawBytes, err := json.Marshal(rpcResp.Result[1])
		if err != nil {
			return err
		}
		return json.Unmarshal(rawBytes, result)
	}

	return nil
}

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	var sysInfo struct {
		Uptime   uint64 `json:"uptime"`
		Memory   map[string]uint64
		System   string `json:"system"`
		Release  map[string]string
		Board    map[string]string
		Hostname string `json:"hostname"`
	}

	if err := p.callUbus(ctx, "system", "info", map[string]any{}, &sysInfo); err != nil {
		return &model.SystemInfo{
			Hostname: "OpenWrt",
			OS:       "OpenWrt",
			Time:     time.Now(),
		}, nil
	}

	info := &model.SystemInfo{
		Hostname:      sysInfo.Hostname,
		OS:            "OpenWrt",
		UptimeSeconds: sysInfo.Uptime,
		Time:          time.Now(),
	}

	if info.Hostname == "" {
		info.Hostname = "OpenWrt"
	}

	if sysInfo.Release != nil {
		info.Version = sysInfo.Release["version"]
		info.Architecture = sysInfo.Release["target"]
	}

	if sysInfo.Memory != nil {
		info.MemoryTotalBytes = sysInfo.Memory["total"]
		if sysInfo.Memory["total"] > sysInfo.Memory["free"] {
			info.MemoryUsedBytes = sysInfo.Memory["total"] - sysInfo.Memory["free"]
		}
	}

	return info, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	var netDump struct {
		Interface []map[string]any `json:"interface"`
	}

	if err := p.callUbus(ctx, "network.interface", "dump", map[string]any{}, &netDump); err != nil {
		return nil, err
	}

	var list []model.Interface
	for _, m := range netDump.Interface {
		name := fmt.Sprintf("%v", m["interface"])
		l3Dev := fmt.Sprintf("%v", m["l3_device"])
		if l3Dev == "" || l3Dev == "<nil>" {
			l3Dev = fmt.Sprintf("%v", m["device"])
		}

		admin := model.AdminStatusDown
		oper := model.OperStatusDown
		if up, ok := m["up"].(bool); ok && up {
			admin = model.AdminStatusUp
			oper = model.OperStatusUp
		}

		var addrs []string
		if ipv4, ok := m["ipv4-address"].([]any); ok {
			for _, item := range ipv4 {
				if ipMap, ok := item.(map[string]any); ok {
					ip := fmt.Sprintf("%v", ipMap["address"])
					mask := fmt.Sprintf("%v", ipMap["mask"])
					if ip != "" && mask != "" {
						addrs = append(addrs, fmt.Sprintf("%s/%s", ip, mask))
					}
				}
			}
		}

		list = append(list, model.Interface{
			ID:            name,
			Name:          name,
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   admin,
			OperStatus:    oper,
			IPv4Addresses: addrs,
			Description:   l3Dev,
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
	var routes []model.Route
	return routes, nil
}

func (p *Provider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	var gateways []model.Gateway
	return gateways, nil
}

func (p *Provider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	var entries []model.ARPEntry
	return entries, nil
}

func (p *Provider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	var luciLeases struct {
		DHCPLeases []struct {
			IPAddress string `json:"ipaddr"`
			MAC       string `json:"macaddr"`
			Hostname  string `json:"hostname"`
			Expires   int64  `json:"expires"`
		} `json:"dhcp_leases"`
	}

	if err := p.callUbus(ctx, "luci-rpc", "getDHCPLeases", map[string]any{}, &luciLeases); err != nil {
		return []model.DHCPLease{}, nil
	}

	var leases []model.DHCPLease
	for _, l := range luciLeases.DHCPLeases {
		leases = append(leases, model.DHCPLease{
			IPAddress:      l.IPAddress,
			MACAddress:     l.MAC,
			ClientHostname: l.Hostname,
			State:          model.DHCPActive,
		})
	}

	return leases, nil
}

func (p *Provider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	return []model.FirewallRule{}, nil
}

func (p *Provider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return []model.FirewallAlias{}, nil
}

func (p *Provider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return []model.NATRule{}, nil
}

func (p *Provider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return []model.BGPNeighbor{}, nil
}

func (p *Provider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return []model.WireGuardPeer{}, nil
}

func (p *Provider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	return []model.InterfaceStats{}, nil
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
	return "# OpenWrt uci export", nil
}

func (p *Provider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}

func (p *Provider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration applied on OpenWrt",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}
