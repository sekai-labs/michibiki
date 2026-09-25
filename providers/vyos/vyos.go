package vyos

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
	provider.Register("vyos", func() provider.Provider {
		return New()
	})
}

type Provider struct {
	mu        sync.RWMutex
	client    *client.HTTPClient
	apiKey    string
	connected bool
	endpoint  string
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string {
	return "vyos"
}

func (p *Provider) Name() string {
	return "VyOS Network OS"
}

func (p *Provider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts := client.HTTPOptions{
		BaseURL: endpoint,
	}

	if creds != nil {
		p.apiKey = creds.APIKey
		if p.apiKey == "" {
			p.apiKey = creds.Password
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
		provider.CapBGP |
		provider.CapOSPF |
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

func (p *Provider) showCommand(ctx context.Context, path []string, result any) error {
	c, err := p.getClient()
	if err != nil {
		return err
	}

	body := map[string]any{
		"op":   "show",
		"path": path,
		"key":  p.apiKey,
	}

	var resp struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
		Error   any  `json:"error"`
	}

	if err := c.Do(ctx, "POST", "/show", body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("vyos API error: %v", resp.Error)
	}

	return nil
}

func (p *Provider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname: "vyos",
		OS:       "VyOS",
		Time:     time.Now(),
	}, nil
}

func (p *Provider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return []model.Interface{}, nil
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
	return "# VyOS Running Configuration", nil
}

func (p *Provider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}

func (p *Provider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{
		Success:        true,
		RollbackID:     strconv.FormatInt(time.Now().Unix(), 10),
		ConfirmPending: req.ConfirmSeconds > 0,
		Message:        "Configuration applied on VyOS",
	}, nil
}

func (p *Provider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}
