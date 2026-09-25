package sdk_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/plugin"
	"github.com/sekai-labs/michibiki/pkg/plugin/sdk"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type customProvider struct{}

func (c *customProvider) ID() string   { return "test-sdk" }
func (c *customProvider) Name() string { return "Test SDK Provider" }
func (c *customProvider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	return nil
}
func (c *customProvider) Disconnect(ctx context.Context) error { return nil }
func (c *customProvider) Capabilities() provider.Capabilities {
	return provider.CapSystem | provider.CapInterfaces
}
func (c *customProvider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname: "sdk-switch-01",
		OS:       "CustomOS",
	}, nil
}
func (c *customProvider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return []model.Interface{
		{
			ID:            "1",
			Name:          "eth0",
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   model.AdminStatusUp,
			OperStatus:    model.OperStatusUp,
			IPv4Addresses: []string{"10.0.0.1/24"},
		},
	}, nil
}
func (c *customProvider) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	return nil, nil
}
func (c *customProvider) ListRoutes(ctx context.Context) ([]model.Route, error) {
	return nil, nil
}
func (c *customProvider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return nil, nil
}
func (c *customProvider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return nil, nil
}
func (c *customProvider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return nil, nil
}
func (c *customProvider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	return nil, nil
}
func (c *customProvider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return nil, nil
}
func (c *customProvider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return nil, nil
}
func (c *customProvider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return nil, nil
}
func (c *customProvider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return nil, nil
}
func (c *customProvider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	return nil, nil
}
func (c *customProvider) GetSubnetUsage(ctx context.Context, filter string) ([]ipam.SubnetUsage, error) {
	return nil, nil
}
func (c *customProvider) GetRunningConfig(ctx context.Context) (string, error) {
	return "hostname sdk-switch-01", nil
}
func (c *customProvider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return &model.ValidationResult{Valid: true}, nil
}
func (c *customProvider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return &model.ConfigApplyResult{Success: true}, nil
}
func (c *customProvider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return nil
}

func TestSDKServeAndClientCall(t *testing.T) {
	serverR, clientW := io.Pipe()
	clientR, serverW := io.Pipe()

	server := sdk.NewServerWithIO(&customProvider{}, serverR, serverW)
	go func() {
		_ = server.Run(context.Background())
	}()

	client := plugin.NewClient(clientR, clientW)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var handshake plugin.HandshakeResponse
	err := client.Call(ctx, plugin.MethodPluginHandshake, plugin.HandshakeParams{Version: "1.0"}, &handshake)
	if err != nil {
		t.Fatalf("handshake failed: %v", err)
	}
	if handshake.Name != "test-sdk" {
		t.Errorf("expected test-sdk, got %s", handshake.Name)
	}

	var sysInfo model.SystemInfo
	err = client.Call(ctx, plugin.MethodProviderGetSystem, nil, &sysInfo)
	if err != nil {
		t.Fatalf("getSystemInfo failed: %v", err)
	}
	if sysInfo.Hostname != "sdk-switch-01" {
		t.Errorf("expected sdk-switch-01, got %s", sysInfo.Hostname)
	}

	var ifaces []model.Interface
	err = client.Call(ctx, plugin.MethodProviderListIfaces, nil, &ifaces)
	if err != nil {
		t.Fatalf("listInterfaces failed: %v", err)
	}
	if len(ifaces) != 1 || ifaces[0].Name != "eth0" {
		t.Errorf("unexpected ifaces: %+v", ifaces)
	}
}
