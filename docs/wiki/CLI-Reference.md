# CLI Command Reference

All operational commands accept the following global flags:
- `-d, --device <name>`: Target device profile.
- `-u, --url <url>`: Direct device endpoint.
- `-o, --output <table|json|yaml>`: Output formatting (default: `table`).
- `-k, --insecure`: Skip TLS certificate validation.
- `--no-color`: Disable ANSI color escapes.

---

## System Information
```bash
michibiki -d core-fw system info
```
Output includes hostname, OS version, CPU utilization %, memory usage, uptime, and storage counters.

---

## Interfaces & VLANs
```bash
# List all physical, virtual, and tunnel interfaces
michibiki -d core-fw interface list

# Detailed view for a specific interface
michibiki -d core-fw interface show eth0
```

---

## Routing & Gateways
```bash
# View complete routing table
michibiki -d core-fw route list

# View configured gateways with live latency and packet loss
michibiki -d core-fw gateway list

# View ARP / NDP neighbor cache
michibiki -d core-fw arp list
```

---

## DHCP Leases
```bash
michibiki -d core-fw dhcp lease list
```
Displays IP, MAC, client hostname, lease status (active/static), interface, and remaining lease time.

---

## Firewall & NAT
```bash
# List packet filter rules
michibiki -d core-fw firewall rule list

# List firewall aliases and port/host groups
michibiki -d core-fw firewall alias list

# List NAT rules (SNAT, DNAT, Masquerade)
michibiki -d core-fw nat list
```

---

## BGP & WireGuard VPN
```bash
# List BGP peering sessions
michibiki -d core-fw bgp neighbor list
michibiki -d core-fw bgp summary

# List WireGuard tunnels and peer transfer stats
michibiki -d core-fw vpn peers
```

---

## Real-Time Telemetry Monitor
```bash
# Stream interface traffic stats every 1s
michibiki -d core-fw monitor traffic --interval 1s

# Stream system telemetry
michibiki -d core-fw monitor system --interval 2s
```
