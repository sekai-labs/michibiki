package plugin

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var (
	ErrSubprocessNotStarted = errors.New("plugin subprocess not started")
)

type Plugin struct {
	mu           sync.RWMutex
	binaryPath   string
	id           string
	name         string
	version      string
	capabilities provider.Capabilities

	cmd    *exec.Cmd
	client *Client
}

func NewPlugin(binaryPath string) *Plugin {
	base := filepath.Base(binaryPath)
	id := strings.TrimPrefix(base, "michibiki-provider-")
	id = strings.TrimSuffix(id, filepath.Ext(id))
	return &Plugin{
		binaryPath: binaryPath,
		id:         id,
		name:       id,
	}
}

func (p *Plugin) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *Plugin) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.name != "" {
		return p.name
	}
	return p.id
}

func (p *Plugin) Version() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.version
}

func (p *Plugin) Capabilities() provider.Capabilities {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.capabilities
}

func (p *Plugin) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		_ = p.client.Close()
		p.client = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}

	cmd := exec.CommandContext(ctx, p.binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return err
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return err
	}

	client := NewClient(stdout, stdin)
	p.cmd = cmd
	p.client = client

	var handshakeResp HandshakeResponse
	err = client.Call(ctx, MethodPluginHandshake, HandshakeParams{Version: "1.0"}, &handshakeResp)
	if err != nil {
		_ = client.Close()
		_ = cmd.Process.Kill()
		p.client = nil
		p.cmd = nil
		return err
	}

	if handshakeResp.Name != "" {
		p.name = handshakeResp.Name
	}
	p.version = handshakeResp.Version
	p.capabilities = ParseCapabilities(handshakeResp.Capabilities)

	connParams := ConnectParams{
		Endpoint: endpoint,
		Creds:    creds,
		Options:  options,
	}
	var connectResp any
	if err := client.Call(ctx, MethodProviderConnect, connParams, &connectResp); err != nil {
		_ = client.Close()
		_ = cmd.Process.Kill()
		p.client = nil
		p.cmd = nil
		return err
	}

	return nil
}

func (p *Plugin) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return nil
	}

	var dummy any
	_ = p.client.Call(ctx, MethodProviderDisconnect, nil, &dummy)
	_ = p.client.Close()
	p.client = nil

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
		p.cmd = nil
	}

	return nil
}

func (p *Plugin) getClient() (*Client, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return nil, provider.ErrNotConnected
	}
	return p.client, nil
}

func (p *Plugin) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res model.SystemInfo
	if err := client.Call(ctx, MethodProviderGetSystem, nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *Plugin) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.Interface
	if err := client.Call(ctx, MethodProviderListIfaces, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res model.Interface
	if err := client.Call(ctx, MethodProviderGetIface, GetInterfaceParams{Name: name}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *Plugin) ListRoutes(ctx context.Context) ([]model.Route, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.Route
	if err := client.Call(ctx, MethodProviderListRoutes, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.Gateway
	if err := client.Call(ctx, MethodProviderListGateways, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.ARPEntry
	if err := client.Call(ctx, MethodProviderListARP, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.DHCPLease
	if err := client.Call(ctx, MethodProviderListDHCP, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.FirewallRule
	if err := client.Call(ctx, MethodProviderListRules, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.FirewallAlias
	if err := client.Call(ctx, MethodProviderListAliases, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.NATRule
	if err := client.Call(ctx, MethodProviderListNAT, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.BGPNeighbor
	if err := client.Call(ctx, MethodProviderListBGP, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.WireGuardPeer
	if err := client.Call(ctx, MethodProviderListWG, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res []model.InterfaceStats
	if err := client.Call(ctx, MethodProviderGetStats, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Plugin) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	ifaces, err := p.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	leases, err := p.ListDHCPLeases(ctx)
	if err != nil {
		leases = nil
	}
	arpEntries, err := p.ListARPEntries(ctx)
	if err != nil {
		arpEntries = nil
	}
	wgPeers, err := p.ListWireGuardPeers(ctx)
	if err != nil {
		wgPeers = nil
	}

	usages := ipam.DiscoverSubnets(ifaces, leases, arpEntries, wgPeers)
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

func (p *Plugin) GetRunningConfig(ctx context.Context) (string, error) {
	client, err := p.getClient()
	if err != nil {
		return "", err
	}
	var res string
	if err := client.Call(ctx, MethodProviderGetConfig, nil, &res); err != nil {
		return "", err
	}
	return res, nil
}

func (p *Plugin) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res model.ValidationResult
	if err := client.Call(ctx, MethodProviderValidateConfig, ValidateConfigParams{Candidate: candidate}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *Plugin) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	var res model.ConfigApplyResult
	if err := client.Call(ctx, MethodProviderApplyConfig, ApplyConfigParams{Request: req}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *Plugin) RollbackConfig(ctx context.Context, rollbackID string) error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	var dummy any
	return client.Call(ctx, MethodProviderRollback, RollbackConfigParams{RollbackID: rollbackID}, &dummy)
}

