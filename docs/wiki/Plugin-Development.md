# Plugin Development & Community Installation

Michibiki's extensible architecture allows out-of-tree plugins to support any proprietary or open-source network device without rebuilding or modifying the core Michibiki binary.

---

## 1. Installing Community Plugins

Community plugins can be installed in a single command using `michibiki plugin install`:

### From a Go Package
```bash
michibiki plugin install github.com/community/michibiki-provider-cisco@latest
```
*Michibiki builds the binary using `go install` and places it in `~/.config/michibiki/plugins/michibiki-provider-cisco`.*

### From a Release Archive (.tar.gz, .zip, or raw binary)
```bash
michibiki plugin install https://github.com/community/michibiki-provider-cisco/releases/download/v1.0.0/michibiki-provider-cisco-linux-amd64.tar.gz
```

### From a Local File
```bash
michibiki plugin install ./dist/michibiki-provider-arista
```

### Listing Installed Plugins
```bash
michibiki plugin list
michibiki plugin info cisco
```

---

## 2. Developing Plugins with the Go SDK

The official SDK (`pkg/plugin/sdk`) abstracts all JSON-RPC 2.0 transport details, letting you implement simple Go methods.

### Step 1: Initialize Module
```bash
mkdir michibiki-provider-cisco
cd michibiki-provider-cisco
go mod init github.com/community/michibiki-provider-cisco
go get github.com/sekai-labs/michibiki@latest
```

### Step 2: Implement the Provider in `main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/plugin/sdk"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type CiscoProvider struct {
	sdk.BaseProvider
	host string
}

func (p *CiscoProvider) ID() string {
	return "cisco"
}

func (p *CiscoProvider) Name() string {
	return "Cisco IOS-XE Provider"
}

func (p *CiscoProvider) Capabilities() provider.Capabilities {
	return provider.CapSystem | provider.CapInterfaces | provider.CapRouting
}

func (p *CiscoProvider) Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error {
	p.host = endpoint
	return nil
}

func (p *CiscoProvider) Disconnect(ctx context.Context) error {
	return nil
}

func (p *CiscoProvider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname:     "cisco-edge-01",
		OS:           "Cisco IOS-XE",
		Version:      "17.9.4",
		Architecture: "x86_64",
	}, nil
}

func (p *CiscoProvider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return []model.Interface{
		{
			ID:            "Gi0/0/0",
			Name:          "GigabitEthernet0/0/0",
			Type:          model.InterfaceTypeEthernet,
			AdminStatus:   model.AdminStatusUp,
			OperStatus:    model.OperStatusUp,
			IPv4Addresses: []string{"10.0.0.1/24"},
		},
	}, nil
}

func main() {
	if err := sdk.Serve(&CiscoProvider{}); err != nil {
		fmt.Fprintf(os.Stderr, "Plugin error: %v\n", err)
		os.Exit(1)
	}
}
```

### Step 3: Build the Plugin
Binary naming must follow `michibiki-provider-<name>`:

```bash
go build -o michibiki-provider-cisco .
```

Place the binary in `./plugins/`, `~/.config/michibiki/plugins/`, or anywhere in `$PATH`. Michibiki will discover it automatically.

---

## 3. Protocol Specification (For non-Go Plugins)

Plugins written in Python, Rust, C, or any other language communicate with Michibiki over **JSON-RPC 2.0 across standard I/O (stdin/stdout)**:

### Handshake
- Request:
  ```json
  {"jsonrpc": "2.0", "id": 1, "method": "plugin.handshake", "params": {"version": "1.0"}}
  ```
- Response:
  ```json
  {
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
      "name": "cisco",
      "version": "1.0.0",
      "capabilities": ["system", "interfaces", "routing"]
    }
  }
  ```

### Key RPC Methods
- `provider.connect`
- `provider.disconnect`
- `provider.getSystemInfo`
- `provider.listInterfaces`
- `provider.getInterface`
- `provider.listRoutes`
- `provider.listGateways`
- `provider.listARPEntries`
- `provider.listDHCPLeases`
- `provider.listFirewallRules`
- `provider.listFirewallAliases`
- `provider.listNATRules`
- `provider.listBGPNeighbors`
- `provider.listWireGuardPeers`
- `provider.getInterfaceStats`
- `provider.getRunningConfig`
- `provider.validateConfig`
- `provider.applyConfig`
- `provider.rollbackConfig`
