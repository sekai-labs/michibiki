package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)


func TestClientCall(t *testing.T) {
	clientR, serverW := io.Pipe()
	serverR, clientW := io.Pipe()

	client := NewClient(clientR, clientW)
	defer client.Close()

	go func() {
		dec := json.NewDecoder(serverR)
		enc := json.NewEncoder(serverW)
		for {
			var req Request
			if err := dec.Decode(&req); err != nil {
				return
			}
			if req.Method == MethodPluginHandshake {
				respData, _ := json.Marshal(HandshakeResponse{
					Name:         "test-provider",
					Version:      "1.2.3",
					Capabilities: []string{"system", "interfaces", "ipam"},
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  respData,
				})
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp HandshakeResponse
	err := client.Call(ctx, MethodPluginHandshake, HandshakeParams{Version: "1.0"}, &resp)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if resp.Name != "test-provider" {
		t.Errorf("expected test-provider, got %s", resp.Name)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", resp.Version)
	}
}

func TestPluginMethods(t *testing.T) {
	clientR, serverW := io.Pipe()
	serverR, clientW := io.Pipe()

	client := NewClient(clientR, clientW)
	defer client.Close()

	plug := &Plugin{
		id:           "test",
		name:         "test",
		version:      "1.0.0",
		capabilities: provider.CapSystem | provider.CapInterfaces | provider.CapIPAM,
		client:       client,
	}

	go func() {
		dec := json.NewDecoder(serverR)
		enc := json.NewEncoder(serverW)
		for {
			var req Request
			if err := dec.Decode(&req); err != nil {
				return
			}
			switch req.Method {
			case MethodProviderGetSystem:
				data, _ := json.Marshal(model.SystemInfo{
					Hostname: "test-router",
					OS:       "RouterOS",
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderListIfaces:
				data, _ := json.Marshal([]model.Interface{
					{
						Name:          "eth0",
						IPv4Addresses: []string{"192.168.1.1/24"},
						MACAddress:    "aa:bb:cc:dd:ee:01",
					},
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderListDHCP:
				data, _ := json.Marshal([]model.DHCPLease{
					{
						IPAddress:      "192.168.1.50",
						MACAddress:     "aa:bb:cc:dd:ee:02",
						ClientHostname: "workstation",
					},
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderListARP:
				data, _ := json.Marshal([]model.ARPEntry{
					{
						IPAddress:  "192.168.1.51",
						MACAddress: "aa:bb:cc:dd:ee:03",
						Hostname:   "server",
					},
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderListWG:
				data, _ := json.Marshal([]model.WireGuardPeer{})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderValidateConfig:
				data, _ := json.Marshal(model.ValidationResult{
					Valid: true,
				})
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  data,
				})
			case MethodProviderRollback:
				_ = enc.Encode(Response{
					JSONRPC: JSONRPCVersion,
					ID:      req.ID,
					Result:  []byte("{}"),
				})
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sys, err := plug.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}
	if sys.Hostname != "test-router" {
		t.Errorf("expected test-router, got %s", sys.Hostname)
	}

	valRes, err := plug.ValidateConfig(ctx, "sample")
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	if !valRes.Valid {
		t.Errorf("expected valid true")
	}

	err = plug.RollbackConfig(ctx, "rollback-1")
	if err != nil {
		t.Fatalf("RollbackConfig failed: %v", err)
	}

	usages, err := plug.GetSubnetUsage(ctx, "")
	if err != nil {
		t.Fatalf("GetSubnetUsage failed: %v", err)
	}
	if len(usages) != 1 {
		t.Fatalf("expected 1 subnet usage, got %d", len(usages))
	}
	if usages[0].CIDR != netip.MustParsePrefix("192.168.1.0/24") {
		t.Errorf("unexpected cidr %s", usages[0].CIDR)
	}
	if usages[0].UsedIPs < 2 {
		t.Errorf("expected used ips >= 2, got %d", usages[0].UsedIPs)
	}

	filtered, err := plug.GetSubnetUsage(ctx, "eth0")
	if err != nil {
		t.Fatalf("GetSubnetUsage filtered failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 match for eth0, got %d", len(filtered))
	}
}

func TestDiscovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "michibiki-plugin-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "michibiki-provider-dummy")
	scriptContent := `#!/bin/sh
while IFS= read -r line; do
	case "$line" in
		*plugin.handshake*)
			echo '{"jsonrpc":"2.0","id":1,"result":{"name":"dummy","version":"0.1.0","capabilities":["system","interfaces"]}}'
			;;
		*provider.connect*)
			echo '{"jsonrpc":"2.0","id":2,"result":{}}'
			;;
		*provider.disconnect*)
			echo '{"jsonrpc":"2.0","id":3,"result":{}}'
			exit 0
			;;
	esac
done
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write dummy script: %v", err)
	}

	discovered, err := DiscoverPlugins([]string{tempDir})
	if err != nil {
		t.Fatalf("DiscoverPlugins failed: %v", err)
	}

	var found *DiscoveredPlugin
	for i := range discovered {
		if discovered[i].Name == "dummy" {
			found = &discovered[i]
			break
		}
	}

	if found == nil {
		t.Fatalf("expected dummy plugin to be discovered")
	}
	if found.Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", found.Version)
	}
	if !found.Capabilities.Has(provider.CapSystem) {
		t.Errorf("expected CapSystem capability")
	}

	RegisterDiscoveredPlugins(discovered)
	prov, err := provider.Create("dummy")
	if err != nil {
		t.Fatalf("failed to create registered dummy provider: %v", err)
	}
	if prov.ID() != "dummy" {
		t.Errorf("expected provider id dummy, got %s", prov.ID())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = prov.Connect(ctx, "http:dummy", &credential.Credentials{}, nil)
	if err != nil {
		t.Fatalf("failed to connect dummy provider: %v", err)
	}
	_ = prov.Disconnect(ctx)
}
