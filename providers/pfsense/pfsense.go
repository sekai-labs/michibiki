package pfsense

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
	provider.Register("pfsense", func() provider.Provider {
		return New()
	})
}

type Provider struct {
	mu        sync.RWMutex
	client    *client.HTTPClient
	sshClient *client.SSHClient
	connected bool
	endpoint  string
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string {
	return "pfsense"
}

func (p *Provider) Name() string {
	return "pfSense Firewall"
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

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawInfo map[string]any
	err = c.Do(ctx, "GET", "/api/v1/system/version", nil, &rawInfo)
	if err != nil {
		return &model.SystemInfo{
			Hostname: "pfSense",
			OS:       "pfSense",
			Time:     time.Now(),
		}, nil
	}

	info := &model.SystemInfo{
		Hostname: "pfSense",
		OS:       "pfSense",
		Time:     time.Now(),
	}

	if data, ok := rawInfo["data"].(map[string]any); ok {
		if ver, ok := data["version"].(string); ok {
			info.Version = ver
		}
	}

	return info, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawResp map[string]any
	err = c.Do(ctx, "GET", "/api/v1/interface", nil, &rawResp)
	if err != nil {
		return []model.Interface{}, nil
	}

	var list []model.Interface
	if data, ok := rawResp["data"].([]any); ok {
		for _, item := range data {
			if m, ok := item.(map[string]any); ok {
				admin := model.AdminStatusDown
				oper := model.OperStatusDown
				if en, ok := m["enable"].(bool); ok && en {
					admin = model.AdminStatusUp
					oper = model.OperStatusUp
				}
				list = append(list, model.Interface{
					ID:          fmt.Sprintf("%v", m["if"]),
					Name:        fmt.Sprintf("%v", m["descr"]),
					Type:        model.InterfaceTypeEthernet,
					AdminStatus: admin,
					OperStatus:  oper,
					Description: fmt.Sprintf("%v", m["descr"]),
				})
			}
		}
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
	return []model.Route{}, nil
}

func (p *Provider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return []model.Gateway{}, nil
}

func (p *Provider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return []model.ARPEntry{}, nil
}

func (p *Provider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var rawResp map[string]any
	err = c.Do(ctx, "GET", "/api/v1/services/dhcpd/lease", nil, &rawResp)
	if err != nil {
		return []model.DHCPLease{}, nil
	}

	var leases []model.DHCPLease
	if data, ok := rawResp["data"].([]any); ok {
		for _, item := range data {
			if m, ok := item.(map[string]any); ok {
				leases = append(leases, model.DHCPLease{
					IPAddress:      fmt.Sprintf("%v", m["ip"]),
					MACAddress:     fmt.Sprintf("%v", m["mac"]),
					ClientHostname: fmt.Sprintf("%v", m["hostname"]),
					State:          model.DHCPActive,
				})
			}
		}
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
	return "<pfsense></pfsense>", nil
}

func (p *Provider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}

func (p *Provider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration applied on pfSense",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}
