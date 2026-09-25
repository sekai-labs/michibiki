# Michibiki (導き) — Universal Network CLI & TUI

> **"One terminal. Every network."**

Michibiki is an open-source, vendor-neutral CLI and interactive TUI for configuring, managing, troubleshooting, and monitoring network infrastructure across **OPNsense**, **MikroTik RouterOS**, **OpenWrt**, **VyOS**, **pfSense**, and **FRRouting**.

---

## Quick Install

```bash
go install github.com/sekai-labs/michibiki/cmd/michibiki@latest
```

Or build from source:
```bash
git clone https://codeberg.org/sekai-labs/michibiki.git
cd michibiki
go build -o michibiki ./cmd/michibiki
```

---

## Quick Start

```bash
# Add a device profile
michibiki device add core-fw --provider opnsense --address https://192.168.1.1 --insecure

# Launch interactive TUI
michibiki -d core-fw tui

# Or use CLI
michibiki -d core-fw system info
michibiki -d core-fw interface list
```

---

## Key Highlights

- **Vendor-Neutral:** Single unified CLI and TUI for 6+ major network platforms.
- **Subnet IPAM & Free IP Finder:** View real-time IP allocation matrices and discover next available unassigned static IPs for VM provisioning.
- **Safety First:** Commit-confirm with automated countdown rollback (`--confirm 60`) prevents remote lockouts.
- **Community Plugins:** Extend to any platform (Cisco, Arista, etc.) using `michibiki plugin install` or build plugins in Go using `pkg/plugin/sdk`.
- **Zero Plaintext Credentials:** AES-256-GCM encrypted vault or environment variable resolution.

---

## Documentation & Wiki

Detailed guides, full CLI reference, and developer documentation are available in the [Wiki](https://github.com/sekai-labs/michibiki/wiki):

- [Installation Guide](https://github.com/sekai-labs/michibiki/wiki/Installation)
- [Device Profiles & Vault](https://github.com/sekai-labs/michibiki/wiki/Device-Profiles)
- [CLI Command Reference](https://github.com/sekai-labs/michibiki/wiki/CLI-Reference)
- [IPAM & Free IP Finder](https://github.com/sekai-labs/michibiki/wiki/IPAM)
- [Interactive TUI Guide](https://github.com/sekai-labs/michibiki/wiki/TUI-Guide)
- [Plugin Development Guide](https://github.com/sekai-labs/michibiki/wiki/Plugin-Development)
- [Safety & Commit-Confirm](https://github.com/sekai-labs/michibiki/wiki/Safety-Engine)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
