# Welcome to the Michibiki Wiki

Michibiki (導き) is a vendor-neutral CLI and interactive TUI for configuring, managing, troubleshooting, and monitoring network infrastructure across:
- **OPNsense** (REST API)
- **MikroTik RouterOS v7** (REST API)
- **OpenWrt** (ubus / JSON-RPC)
- **VyOS** (HTTP API)
- **pfSense** (REST / SSH)
- **FRRouting** (vtysh)

## Documentation Pages

1. **[Installation](Installation.md)** — Install via `go install`, pre-built binaries, or source build.
2. **[Device Profiles & Credential Vault](Device-Profiles.md)** — Configuring devices, resolving secrets, and AES-256-GCM vault.
3. **[CLI Command Reference](CLI-Reference.md)** — Complete command syntax for interfaces, routing, firewall, BGP, DHCP, VPN, telemetry.
4. **[IPAM & Free IP Finder](IPAM.md)** — Subnet allocation matrix, free IP block computation, and conflict checks.
5. **[Interactive TUI Guide](TUI-Guide.md)** — Keyboard navigation, real-time monitors, and 9 dashboard views.
6. **[Safety Engine & Commit-Confirm](Safety-Engine.md)** — Lockout prevention, pre-flight safety checks, and rollback timers.
7. **[Plugin Development & Community Installation](Plugin-Development.md)** — Developing plugins with the Go SDK and installing community packages.
