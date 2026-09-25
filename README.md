# Michibiki (導き) — Universal Network CLI & TUI

> **"One terminal. Every network."**

Michibiki is a vendor-neutral command-line interface (CLI) and interactive Terminal User Interface (TUI) for configuring, managing, troubleshooting, and monitoring network infrastructure across diverse platforms: **OPNsense**, **MikroTik RouterOS**, **OpenWrt**, **VyOS**, **pfSense**, and **FRRouting**.

---

## Table of Contents
1. [Key Features](#key-features)
2. [Installation](#installation)
3. [Quick Start & Device Configuration](#quick-start--device-configuration)
4. [Universal CLI Reference](#universal-cli-reference)
5. [IPAM & Free IP Finder (Subnet Allocation Matrix)](#ipam--free-ip-finder-subnet-allocation-matrix)
6. [Interactive Terminal UI (TUI)](#interactive-terminal-ui-tui)
7. [Configuration Safety & Commit-Confirm](#configuration-safety--commit-confirm)
8. [Plugin System & Community Providers](#plugin-system--community-providers)
   - [Installing Community Plugins](#installing-community-plugins)
   - [Developing Plugins with the Go SDK](#developing-plugins-with-the-go-sdk)
9. [Credential Vault](#credential-vault)
10. [Architecture & Protocol](#architecture--protocol)

---

## Key Features

- **Vendor-Neutral Primitives:** Normalize network constructs (Interfaces, VLANs, Routes, ARP, DHCP leases, Firewall rules, NAT, BGP sessions, WireGuard tunnels) across diverse router and firewall operating systems.
- **Built-in First-Class IPAM:** Subnet & VLAN IP Allocation Matrix that inspects configured CIDRs, identifies all active IPs (Interface IPs, DHCP leases, ARP/NDP tables, static bindings), and surfaces free unallocated IP ranges for safe VM/container/bare-metal provisioning.
- **Dual Mode (CLI + TUI):** Fast scriptable CLI with structured output formats (`table`, `json`, `yaml`) and a full-featured keyboard-navigated Terminal UI powered by Bubble Tea & Lip Gloss.
- **Pre-flight Safety & Commit-Confirm:** Prevents management lockout and dropped default routes with configurable countdown rollback timers (`--confirm <seconds>`).
- **Extensible Plugin Architecture:** Write out-of-tree plugins in any language using JSON-RPC 2.0 over standard I/O, or create Go plugins via the native `pkg/plugin/sdk`.
- **Encrypted Local Vault:** Secure storage for API keys, SSH credentials, and tokens encrypted with AES-256-GCM.

---

## Installation

### Method 1: Using `go install` (Recommended)

Requires Go 1.24 or later:

```bash
go install github.com/sekai-labs/michibiki/cmd/michibiki@v0.1.0
```

To install the bleeding-edge version from `main`:
```bash
go install github.com/sekai-labs/michibiki/cmd/michibiki@latest
```

Ensure `$GOPATH/bin` or `~/go/bin` is in your `$PATH`.

### Method 2: Building from Source

```bash
git clone https://codeberg.org/sekai-labs/michibiki.git
cd michibiki
go build -o michibiki ./cmd/michibiki
sudo mv michibiki /usr/local/bin/
```

Verify installation:
```bash
michibiki --help
```

---

## Quick Start & Device Configuration

### 1. Add Device Profiles

Save your routers or firewalls to your local profile config (`~/.config/michibiki/config.yaml`):

```bash
# Add an OPNsense firewall
michibiki device add core-fw --provider opnsense --address https://192.168.1.1 --insecure

# Add a MikroTik RouterOS v7 device
michibiki device add edge-router --provider routeros --address https://10.0.0.1 --insecure

# Add an OpenWrt router
michibiki device add ap-livingroom --provider openwrt --address http://192.168.1.2

# Add a VyOS router
michibiki device add vyos-gw --provider vyos --address https://172.16.0.1 --insecure
```

List configured profiles:
```bash
michibiki device list
```

Test connectivity:
```bash
michibiki device test core-fw
```

### 2. Ephemeral Direct Connections

Connect directly to an address without creating a profile using `-u/--url`:

```bash
michibiki -u https://192.168.1.1 --insecure system info
```

---

## Universal CLI Reference

Global flags available on all operational commands:
- `-d, --device <name>`: Target device profile name.
- `-u, --url <url>`: Direct device URL.
- `-o, --output <table|json|yaml>`: Output formatting (default: `table`).
- `-k, --insecure`: Allow self-signed TLS certificates.
- `--no-color`: Disable ANSI color escapes.

### System & Hardware
```bash
michibiki -d core-fw system info
```

### Interfaces & VLANs
```bash
# List all interfaces and addresses
michibiki -d core-fw interface list

# Inspect detailed interface status
michibiki -d core-fw interface show eth0
```

### Routing & Gateways
```bash
# Routing table
michibiki -d core-fw route list

# Gateway health, latency, and packet loss
michibiki -d core-fw gateway list

# ARP / NDP neighbor cache
michibiki -d core-fw arp list
```

### DHCP Leases
```bash
michibiki -d core-fw dhcp lease list
```

### Firewall & NAT
```bash
# Filter rules
michibiki -d core-fw firewall rule list

# Aliases & address lists
michibiki -d core-fw firewall alias list

# NAT rules (SNAT, DNAT, Masquerade)
michibiki -d core-fw nat list
```

### BGP & WireGuard
```bash
# BGP peering sessions
michibiki -d core-fw bgp neighbor list
michibiki -d core-fw bgp summary

# WireGuard tunnels and peer metrics
michibiki -d core-fw vpn peers
```

### Real-Time Telemetry Monitor
```bash
michibiki -d core-fw monitor traffic --interval 1s
```

---

## IPAM & Free IP Finder (Subnet Allocation Matrix)

Michibiki includes a universal IP address management engine that correlates interface subnets, active DHCP leases, and ARP/NDP tables to show exact subnet usage and find free IP addresses for provisioning new virtual machines, LXC containers, or physical hardware.

### 1. Subnet Usage & Allocation Matrix
```bash
# Inspect all subnets or filter by interface / VLAN / CIDR
michibiki -d core-fw ip usage vlan20
```

Sample output:
```text
SUBNET       192.168.20.0/24
INTERFACE    vlan20
VLAN ID      20
GATEWAY      192.168.20.1
TOTAL IPS    256
USABLE IPS   254
USED IPS     8
FREE IPS     246
UTILIZATION  3.1%  [░░░░░░░░░░░░░░░░░░░░░░░░]

FREE IP BLOCKS (4 blocks, 246 free addresses):
  • 192.168.20.2 - 192.168.20.9 (8 hosts)
  • 192.168.20.12 - 192.168.20.19 (8 hosts)
  • 192.168.20.23 - 192.168.20.49 (27 hosts)
  • 192.168.20.51 - 192.168.20.253 (203 hosts)

NEXT AVAILABLE STATIC IPS: 192.168.20.2, 192.168.20.3, 192.168.20.4

ALLOCATED IP ADDRESSES:
IP ADDRESS      MAC ADDRESS        HOSTNAME         SOURCE        STATUS    EXPIRES
192.168.20.1    52:54:00:12:34:02  vlan20           InterfaceIP   reserved  -
192.168.20.10   52:54:00:20:00:10  server01         DHCPStatic    active    720h0m0s
192.168.20.11   52:54:00:20:00:11  db01             DHCPStatic    active    720h0m0s
...
```

### 2. Next Free IP Discovery
Query the next $N$ free IP addresses formatted as JSON for CI/CD or automation scripts:
```bash
michibiki -d core-fw ip free vlan20 --count 5 -o json
```

### 3. IP Conflict Checker
Verify whether a candidate IP is available or already allocated before assigning it:
```bash
michibiki -d core-fw ip check 192.168.20.15
```

---

## Interactive Terminal UI (TUI)

Launch the full-screen terminal interface:
```bash
michibiki -d core-fw tui
```

### Navigation & Keybindings:
- `1` – `9`: Jump directly to a tab:
  - `[1] Dashboard`: System hardware, CPU/RAM/disk gauges, connectivity summary.
  - `[2] Interfaces`: Interfaces, MAC, MTU, IPv4/IPv6, operational badges.
  - `[3] Subnet & IPAM`: Subnet matrix, free IP blocks, allocation inspector.
  - `[4] Routing`: Kernel routes, gateway health, BGP sessions.
  - `[5] Firewall`: Security rules, pass/block badges, NAT table.
  - `[6] DHCP Leases`: Active DHCP leases, client hostnames, MACs, lease expiration countdowns.
  - `[7] VPN`: WireGuard peers, handshake status, transfer metrics.
  - `[8] Monitor`: Real-time streaming traffic charts.
  - `[9] Config`: Running configuration viewer with commit-confirm status.
- `Tab` / `Shift+Tab`: Cycle through tabs.
- `r`: Refresh data immediately.
- `q` / `Ctrl+C`: Exit TUI.

---

## Configuration Safety & Commit-Confirm

Michibiki features pre-flight safety inspection to prevent administrative lockouts:

```bash
# View running configuration
michibiki -d core-fw config show

# Run pre-flight inspection on candidate configuration
michibiki -d core-fw config validate candidate.conf

# Commit with a 60-second automatic rollback safety timer
michibiki -d core-fw config commit candidate.conf --confirm 60

# Confirm within 60 seconds to make changes permanent
michibiki -d core-fw config rollback confirm <id>
```

If confirmation is not received within the timeout window, the configuration is automatically rolled back.

---

## Plugin System & Community Providers

Michibiki supports out-of-tree plugins that run as isolated subprocesses over JSON-RPC 2.0 (stdin/stdout).

### Installing Community Plugins

```bash
# Install directly from Go package / repository:
michibiki plugin install github.com/community/michibiki-provider-cisco@latest

# Install from a release archive (.tar.gz, .zip, or raw binary):
michibiki plugin install https://github.com/community/michibiki-provider-cisco/releases/download/v1.0.0/michibiki-provider-cisco-linux-amd64.tar.gz

# Install from a local executable:
michibiki plugin install ./dist/michibiki-provider-arista
```

List installed and built-in plugins:
```bash
michibiki plugin list
michibiki plugin info cisco
```

### Developing Plugins with the Go SDK

You can develop a plugin in Go using `pkg/plugin/sdk`:

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
	return nil
}

func (p *CiscoProvider) Disconnect(ctx context.Context) error {
	return nil
}

func (p *CiscoProvider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Hostname: "cisco-core-01",
		OS:       "Cisco IOS-XE",
		Version:  "17.9.4",
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
			IPv4Addresses: []string{"10.250.0.1/24"},
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

Build the plugin with the executable name prefix `michibiki-provider-<name>`:
```bash
go build -o michibiki-provider-cisco .
```

---

## Credential Vault

Michibiki never stores plaintext passwords in device profiles. Credentials are resolved in the following priority:

1. **Environment Variables:**
   - `MICHIBIKI_<DEVICE>_USERNAME`
   - `MICHIBIKI_<DEVICE>_PASSWORD`
   - `MICHIBIKI_<DEVICE>_API_KEY`
   - `MICHIBIKI_<DEVICE>_API_SECRET`
   - `MICHIBIKI_<DEVICE>_TOKEN`
   - `MICHIBIKI_<DEVICE>_SSH_KEY`
2. **Encrypted File Vault:**
   Stored at `~/.config/michibiki/vault.enc` using AES-256-GCM encryption with a master password. Referenced in profiles via `cred-ref: vault:<key>`.

---

## Architecture & Protocol

```text
┌─────────────────────────────────────────────────────────────┐
│                       Michibiki CLI / TUI                   │
│         (Cobra Commands / Bubble Tea Elm Architecture)      │
└──────────────────────────────┬──────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                │   Provider SDK & Registry   │
                └──────┬───────────────┬──────┘
                       │               │
        ┌──────────────┴──────┐ ┌──────┴──────────────────────┐
        │  Built-in Providers │ │  Out-of-Tree Plugins Engine │
        │  • OPNsense (REST)  │ │  (JSON-RPC 2.0 over Stdio)  │
        │  • RouterOS (REST)  │ └──────────────┬──────────────┘
        │  • OpenWrt (ubus)   │                │
        │  • VyOS (HTTP)      │    michibiki-provider-*
        │  • pfSense (REST)   │    (Cisco, Arista, etc.)
        │  • FRRouting (vtysh)│
        └─────────────────────┘
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
