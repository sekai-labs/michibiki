package opnsense

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
	provider.Register("opnsense", func() provider.Provider {
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
	return "opnsense"
}

func (p *Provider) Name() string {
	return "OPNsense Firewall"
}

func (p *Provider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts := client.HTTPOptions{
		BaseURL: endpoint,
	}

	if creds != nil {
		opts.APIKey = creds.APIKey
		opts.APISecret = creds.APISecret
		if opts.APIKey == "" && creds.Username != "" {
			opts.Username = creds.Username
			opts.Password = creds.Password
		}
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

type opnsenseSystemStatus struct {
	Status string `json:"status"`
}

type opnsenseSystemInfo struct {
	Hostname string `json:"name"`
	Versions []struct {
		ProductVersion string `json:"product_version"`
		ProductSeries  string `json:"product_series"`
		ProductArch    string `json:"product_arch"`
		ProductName    string `json:"product_name"`
	} `json:"versions"`
	Uptime      string `json:"uptime"`
	CPUModel    string `json:"cpu_model"`
	CPUs        int    `json:"cpus"`
	CPUUsage    string `json:"cpu_usage"`
	MemoryUsage string `json:"memory_usage"`
}

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawInfo map[string]any
	err = c.Do(ctx, "GET", "/api/diagnostics/system/systemInformation", nil, &rawInfo)
	if err != nil {
		return &model.SystemInfo{
			Hostname: "opnsense",
			OS:       "OPNsense",
			Time:     time.Now(),
		}, nil
	}

	info := &model.SystemInfo{
		OS:   "OPNsense",
		Time: time.Now(),
	}

	if val, ok := rawInfo["name"].(string); ok {
		info.Hostname = val
	}
	if val, ok := rawInfo["cpu_count"].(float64); ok {
		info.CPUCount = int(val)
	}

	return info, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawInterfaces map[string]any
	err = c.Do(ctx, "GET", "/api/interfaces/overview/interfacesInfo", nil, &rawInterfaces)
	if err != nil {
		return nil, err
	}

	var list []model.Interface

	parseIface := func(name string, data map[string]any) model.Interface {
		iface := model.Interface{
			ID:          name,
			Name:        name,
			Type:        model.InterfaceTypeEthernet,
			AdminStatus: model.AdminStatusDown,
			OperStatus:  model.OperStatusDown,
		}

		if status, ok := data["status"].(string); ok {
			if strings.EqualFold(status, "up") {
				iface.OperStatus = model.OperStatusUp
			}
		}
		if enabled, ok := data["enabled"].(bool); ok && enabled {
			iface.AdminStatus = model.AdminStatusUp
		} else if iface.OperStatus == model.OperStatusUp {
			iface.AdminStatus = model.AdminStatusUp
		}

		if mac, ok := data["macaddr"].(string); ok {
			iface.MACAddress = mac
		}
		if descr, ok := data["description"].(string); ok {
			iface.Description = descr
		}
		if mtu, ok := data["mtu"].(float64); ok {
			iface.MTU = int(mtu)
		}

		if ipv4, ok := data["ipv4"].([]any); ok {
			for _, item := range ipv4 {
				if m, ok := item.(map[string]any); ok {
					ip := fmt.Sprintf("%v", m["ipaddr"])
					mask := fmt.Sprintf("%v", m["subnetbits"])
					if ip != "" && mask != "" {
						iface.IPv4Addresses = append(iface.IPv4Addresses, fmt.Sprintf("%s/%s", ip, mask))
					}
				}
			}
		}
		return iface
	}

	if rows, ok := rawInterfaces["rows"].([]any); ok {
		for _, r := range rows {
			if data, ok := r.(map[string]any); ok {
				name := fmt.Sprintf("%v", data["identifier"])
				if name == "" || name == "<nil>" {
					name = fmt.Sprintf("%v", data["name"])
				}
				if name != "" && name != "<nil>" {
					list = append(list, parseIface(name, data))
				}
			}
		}
		return list, nil
	}

	for name, val := range rawInterfaces {
		data, ok := val.(map[string]any)
		if !ok {
			continue
		}
		list = append(list, parseIface(name, data))
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
	err = c.Do(ctx, "GET", "/api/diagnostics/interface/getRoutes", nil, &rawRoutes)
	if err != nil {
		return nil, err
	}

	var routes []model.Route
	for _, item := range rawRoutes {
		proto := model.RouteProtoOther
		protoStr := fmt.Sprintf("%v", item["proto"])
		if strings.Contains(strings.ToLower(protoStr), "static") {
			proto = model.RouteProtoStatic
		} else if strings.Contains(strings.ToLower(protoStr), "bgp") {
			proto = model.RouteProtoBGP
		}

		routes = append(routes, model.Route{
			Destination: fmt.Sprintf("%v", item["destination"]),
			Gateway:     fmt.Sprintf("%v", item["gateway"]),
			Interface:   fmt.Sprintf("%v", item["interface"]),
			Protocol:    proto,
		})
	}

	return routes, nil
}

func (p *Provider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawGateways map[string]any
	err = c.Do(ctx, "GET", "/api/routes/gateway/status", nil, &rawGateways)
	if err != nil {
		return nil, err
	}

	var gateways []model.Gateway
	if items, ok := rawGateways["items"].([]any); ok {
		for _, rawItem := range items {
			if item, ok := rawItem.(map[string]any); ok {
				status := model.GatewayOnline
				if s, ok := item["status"].(string); ok {
					if strings.Contains(strings.ToLower(s), "down") {
						status = model.GatewayOffline
					}
				}
				gateways = append(gateways, model.Gateway{
					Name:      fmt.Sprintf("%v", item["name"]),
					Address:   fmt.Sprintf("%v", item["address"]),
					Interface: fmt.Sprintf("%v", item["interface"]),
					Status:    status,
				})
			}
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
	err = c.Do(ctx, "GET", "/api/diagnostics/interface/getArp", nil, &rawARP)
	if err != nil {
		return nil, err
	}

	var entries []model.ARPEntry
	for _, item := range rawARP {
		entries = append(entries, model.ARPEntry{
			IPAddress:  fmt.Sprintf("%v", item["ip"]),
			MACAddress: fmt.Sprintf("%v", item["mac"]),
			Interface:  fmt.Sprintf("%v", item["intf"]),
			Hostname:   fmt.Sprintf("%v", item["hostname"]),
			Status:     model.ARPReachable,
		})
	}

	return entries, nil
}

func (p *Provider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawLeases map[string]any
	err = c.Do(ctx, "GET", "/api/kea/leases/search", nil, &rawLeases)
	if err != nil {
		err = c.Do(ctx, "GET", "/api/dhcpv4/leases/searchLease", nil, &rawLeases)
		if err != nil {
			return nil, err
		}
	}

	var leases []model.DHCPLease
	if rows, ok := rawLeases["rows"].([]any); ok {
		for _, row := range rows {
			if m, ok := row.(map[string]any); ok {
				leases = append(leases, model.DHCPLease{
					IPAddress:      fmt.Sprintf("%v", m["address"]),
					MACAddress:     fmt.Sprintf("%v", m["hw_address"]),
					ClientHostname: fmt.Sprintf("%v", m["hostname"]),
					Interface:      fmt.Sprintf("%v", m["interface"]),
					State:          model.DHCPActive,
				})
			}
		}
	}

	return leases, nil
}

func (p *Provider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawRules map[string]any
	err = c.Do(ctx, "GET", "/api/firewall/filter/searchRule", nil, &rawRules)
	if err != nil {
		return nil, err
	}

	var rules []model.FirewallRule
	if rows, ok := rawRules["rows"].([]any); ok {
		for i, row := range rows {
			if m, ok := row.(map[string]any); ok {
				action := model.FirewallPass
				if act, ok := m["action"].(string); ok && strings.EqualFold(act, "block") {
					action = model.FirewallBlock
				}
				rules = append(rules, model.FirewallRule{
					ID:          fmt.Sprintf("%v", m["uuid"]),
					Sequence:    i + 1,
					Interface:   fmt.Sprintf("%v", m["interface"]),
					Action:      action,
					Protocol:    fmt.Sprintf("%v", m["protocol"]),
					Source:      fmt.Sprintf("%v", m["source_net"]),
					Destination: fmt.Sprintf("%v", m["destination_net"]),
					Description: fmt.Sprintf("%v", m["description"]),
					Enabled:     fmt.Sprintf("%v", m["enabled"]) == "1",
				})
			}
		}
	}

	return rules, nil
}

func (p *Provider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawAliases map[string]any
	err = c.Do(ctx, "GET", "/api/firewall/alias/searchItem", nil, &rawAliases)
	if err != nil {
		return nil, err
	}

	var aliases []model.FirewallAlias
	if rows, ok := rawAliases["rows"].([]any); ok {
		for _, row := range rows {
			if m, ok := row.(map[string]any); ok {
				aliases = append(aliases, model.FirewallAlias{
					Name:        fmt.Sprintf("%v", m["name"]),
					Type:        fmt.Sprintf("%v", m["type"]),
					Description: fmt.Sprintf("%v", m["description"]),
					Content:     strings.Split(fmt.Sprintf("%v", m["content"]), "\n"),
				})
			}
		}
	}

	return aliases, nil
}

func (p *Provider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return []model.NATRule{}, nil
}

func (p *Provider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return []model.BGPNeighbor{}, nil
}

func (p *Provider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawPeers map[string]any
	err = c.Do(ctx, "GET", "/api/wireguard/client/searchClient", nil, &rawPeers)
	if err != nil {
		return []model.WireGuardPeer{}, nil
	}

	var peers []model.WireGuardPeer
	if rows, ok := rawPeers["rows"].([]any); ok {
		for _, row := range rows {
			if m, ok := row.(map[string]any); ok {
				peers = append(peers, model.WireGuardPeer{
					PublicKey:  fmt.Sprintf("%v", m["pubkey"]),
					Endpoint:   fmt.Sprintf("%v", m["endpoint"]),
					AllowedIPs: strings.Split(fmt.Sprintf("%v", m["tunneladdress"]), ","),
				})
			}
		}
	}

	return peers, nil
}

func (p *Provider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawStats map[string]any
	err = c.Do(ctx, "GET", "/api/diagnostics/interface/getInterfaceStatistics", nil, &rawStats)
	if err != nil {
		return []model.InterfaceStats{}, nil
	}

	var stats []model.InterfaceStats
	now := time.Now()
	for name, val := range rawStats {
		m, ok := val.(map[string]any)
		if !ok {
			continue
		}
		s := model.InterfaceStats{
			InterfaceName: name,
			Timestamp:     now,
		}
		if bytesIn, ok := m["bytes received"].(float64); ok {
			s.RxBytes = uint64(bytesIn)
		}
		if bytesOut, ok := m["bytes transmitted"].(float64); ok {
			s.TxBytes = uint64(bytesOut)
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

	var result string
	err = c.Do(ctx, "GET", "/api/core/backup/download/this", nil, &result)
	if err != nil {
		return "", err
	}
	return result, nil
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
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var resp map[string]any
	body := map[string]string{"config": req.CandidateConfig}
	err = c.Do(ctx, "POST", "/api/core/backup/restore", body, &resp)
	if err != nil {
		return nil, err
	}

	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration restored/applied on OPNsense",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	c, err := p.getClient()
	if err != nil {
		return err
	}
	var resp map[string]any
	return c.Do(ctx, "POST", "/api/core/backup/rollback/"+rollbackID, nil, &resp)
}

func ParseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
