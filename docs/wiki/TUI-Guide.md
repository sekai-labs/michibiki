# Interactive Terminal UI (TUI) Guide

Michibiki features a modern, responsive Terminal User Interface built with Bubble Tea and Lip Gloss.

## Launching the TUI

```bash
michibiki -d core-fw tui
```

---

## Navigation & Controls

| Key | Action |
| :--- | :--- |
| `1` – `9` | Jump directly to tab 1 through 9 |
| `Tab` | Next tab |
| `Shift + Tab` | Previous tab |
| `r` | Manually refresh current view data |
| `q` or `Ctrl+C` | Exit TUI |

---

## 9 Interactive Views

### 1. Dashboard (`[1]`)
Overview of system resources:
- Hostname, OS version, hardware serial number, uptime.
- Real-time CPU, RAM, and Storage gauges.
- Network interface operational health indicators.

### 2. Interfaces & VLANs (`[2]`)
Complete inventory of interfaces:
- Operational and administrative status badges (emerald green for Up, crimson red for Down).
- Configured IPv4 and IPv6 CIDRs, MAC addresses, link speeds, MTU.

### 3. Subnet & IPAM Matrix (`[3]`)
- Left pane: Subnet and VLAN list with visual utilization progress bars (`[████████░░] 78%`).
- Right pane:
  - Subnet summary (CIDR, Gateway, Usable Range, Total, Used, Free count).
  - Next Free Static IPs highlighted for instant copy/provisioning.
  - Complete allocation table (IP, MAC, Hostname, Source, Status, Expiry).

### 4. Routing & Gateways (`[4]`)
- Routing table with destination prefixes, next-hop gateways, and metrics.
- Default and static gateways with live latency (ms) and packet loss (%).
- Active BGP peering sessions.

### 5. Firewall & NAT (`[5]`)
- Ordered rule sequence with Action (`PASS`, `BLOCK`, `REJECT`), Protocol, Interface, and Direction.
- NAT rules (DNAT, SNAT, Masquerade) and destination port mappings.

### 6. DHCP Leases (`[6]`)
- Active and static DHCP lease bindings.
- Hostnames, MAC addresses, assigned IP addresses, and lease expiration countdown timers.

### 7. VPN & WireGuard (`[7]`)
- WireGuard peer status, endpoints, allowed IP prefixes.
- Latest handshake indicators and RX/TX traffic transfer counters.

### 8. Real-time Traffic Monitor (`[8]`)
- Streaming throughput meters and live ASCII sparklines for interface RX and TX data rates.

### 9. Configuration & Safety (`[9]`)
- Running configuration viewer.
- Active commit-confirm rollback session status.
