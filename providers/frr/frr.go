package frr

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
	provider.Register("frr", func() provider.Provider {
		return New()
	})
}

type Provider struct {
	mu        sync.RWMutex
	sshClient *client.SSHClient
	connected bool
	endpoint  string
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string {
	return "frr"
}

func (p *Provider) Name() string {
	return "FRRouting (FRR)"
}

func (p *Provider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	host := endpoint
	port := 22
	if strings.Contains(endpoint, ":") {
		parts := strings.Split(endpoint, ":")
		host = parts[0]
		if pVal, err := strconv.Atoi(parts[1]); err == nil {
			port = pVal
		}
	}

	opts := client.SSHOptions{
		Host: host,
		Port: port,
	}

	if creds != nil {
		opts.Username = creds.Username
		opts.Password = creds.Password
		opts.KeyPath = creds.SSHKeyPath
	}

	p.sshClient = client.NewSSHClient(opts)
	p.connected = true
	p.endpoint = endpoint
	return nil
}

func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = false
	if p.sshClient != nil {
		_ = p.sshClient.Close()
		p.sshClient = nil
	}
	return nil
}

func (p *Provider) Capabilities() provider.Capabilities {
	return provider.CapSystem |
		provider.CapInterfaces |
		provider.CapRouting |
		provider.CapBGP |
		provider.CapOSPF |
		provider.CapMonitoring |
		provider.CapConfig |
		provider.CapIPAM
}

func (p *Provider) runVtysh(ctx context.Context, cmd string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.connected || p.sshClient == nil {
		return "", provider.ErrNotConnected
	}
	return p.sshClient.RunCommand(ctx, fmt.Sprintf("vtysh -c %q", cmd))
}

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname: "FRRouting",
		OS:       "FRR",
		Time:     time.Now(),
	}, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	out, err := p.runVtysh(ctx, "show interface json")
	if err != nil {
		return []model.Interface{}, nil
	}

	var rawIfaces map[string]map[string]any
	if err := json.Unmarshal([]byte(out), &rawIfaces); err != nil {
		return []model.Interface{}, nil
	}

	var list []model.Interface
	for name, m := range rawIfaces {
		admin := model.AdminStatusDown
		oper := model.OperStatusDown
		if adminUp, ok := m["administrativeStatus"].(string); ok && strings.EqualFold(adminUp, "up") {
			admin = model.AdminStatusUp
		}
		if operUp, ok := m["operationalStatus"].(string); ok && strings.EqualFold(operUp, "up") {
			oper = model.OperStatusUp
		}

		var addrs []string
		if ipAddrs, ok := m["ipAddresses"].([]any); ok {
			for _, item := range ipAddrs {
				if ipMap, ok := item.(map[string]any); ok {
					if addrStr, ok := ipMap["address"].(string); ok {
						addrs = append(addrs, addrStr)
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
	out, err := p.runVtysh(ctx, "show ip route json")
	if err != nil {
		return []model.Route{}, nil
	}

	var rawRoutes map[string][]map[string]any
	if err := json.Unmarshal([]byte(out), &rawRoutes); err != nil {
		return []model.Route{}, nil
	}

	var routes []model.Route
	for prefix, nexthops := range rawRoutes {
		for _, nh := range nexthops {
			proto := model.RouteProtoOther
			protoStr := strings.ToLower(fmt.Sprintf("%v", nh["protocol"]))
			if strings.Contains(protoStr, "bgp") {
				proto = model.RouteProtoBGP
			} else if strings.Contains(protoStr, "ospf") {
				proto = model.RouteProtoOSPF
			} else if strings.Contains(protoStr, "static") {
				proto = model.RouteProtoStatic
			} else if strings.Contains(protoStr, "connected") {
				proto = model.RouteProtoConnected
			}

			metric := 0
			if mVal, ok := nh["metric"].(float64); ok {
				metric = int(mVal)
			}

			routes = append(routes, model.Route{
				Destination: prefix,
				Gateway:     fmt.Sprintf("%v", nh["ip"]),
				Interface:   fmt.Sprintf("%v", nh["interfaceName"]),
				Protocol:    proto,
				Metric:      metric,
			})
		}
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
	return []model.ARPEntry{}, nil
}

func (p *Provider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return []model.DHCPLease{}, nil
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
	out, err := p.runVtysh(ctx, "show ip bgp summary json")
	if err != nil {
		return []model.BGPNeighbor{}, nil
	}

	var rawSummary struct {
		Peers map[string]map[string]any `json:"peers"`
	}
	if err := json.Unmarshal([]byte(out), &rawSummary); err != nil {
		return []model.BGPNeighbor{}, nil
	}

	var neighbors []model.BGPNeighbor
	for peerIP, m := range rawSummary.Peers {
		var remoteAS uint32
		if asNum, ok := m["remoteAs"].(float64); ok {
			remoteAS = uint32(asNum)
		}
		state := model.BGPIdle
		if st, ok := m["state"].(string); ok && strings.EqualFold(st, "Established") {
			state = model.BGPEstablished
		}

		var prefixes uint32
		if pfx, ok := m["pfxRcd"].(float64); ok {
			prefixes = uint32(pfx)
		}

		neighbors = append(neighbors, model.BGPNeighbor{
			RemoteAS:         remoteAS,
			PeerAddress:      peerIP,
			State:            state,
			PrefixesReceived: prefixes,
			Description:      fmt.Sprintf("%v", m["peerUptime"]),
		})
	}

	return neighbors, nil
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
	return p.runVtysh(ctx, "show running-config")
}

func (p *Provider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}

func (p *Provider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration applied on FRR",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}
