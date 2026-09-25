# IPAM & Free IP Finder (Subnet Allocation Matrix)

Michibiki incorporates a universal IP address management (IPAM) engine that solves a critical problem for network engineers: knowing exactly which IPs are used across a subnet or VLAN, and safely finding unallocated IPs to provision new servers, VMs, or physical gear without collision.

## How it works

The engine aggregates data across 4 distinct sources:
1. **Interface IP bindings** (router IP, broadcast, virtual IPs)
2. **DHCP active and static leases** (client hostnames, leases)
3. **ARP / NDP discovery tables** (unmanaged or static devices)
4. **WireGuard peers** (configured tunnel IPs)

It calculates the usable host range for `/32` down to `/16` prefixes, merges multiple observations of the same host, and identifies contiguous runs of unallocated IP addresses.

---

## 1. Inspecting Subnet Usage

```bash
# Inspect all configured subnets
michibiki -d core-fw ip usage

# Filter by interface name, VLAN ID, or CIDR
michibiki -d core-fw ip usage vlan20
michibiki -d core-fw ip usage 192.168.20.0/24
```

### Sample Output:
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

---

## 2. Finding Next Available Free IPs

Query the next $N$ free static IPs for immediate assignment:

```bash
# Print next 5 free IPs
michibiki -d core-fw ip free vlan20 --count 5

# JSON format for automated provisioning scripts (Terraform, Ansible, scripts)
michibiki -d core-fw ip free vlan20 --count 3 -o json
```

Example JSON response:
```json
[
  {
    "subnet": "192.168.20.0/24",
    "interface": "vlan20",
    "vlan_id": 20,
    "free_count": 246,
    "next_free": [
      "192.168.20.2",
      "192.168.20.3",
      "192.168.20.4"
    ]
  }
]
```

---

## 3. Validating an IP Before Provisioning

Check if an IP address is safe to use:

```bash
michibiki -d core-fw ip check 192.168.20.15
```

Returns:
- `SUCCESS: IP 192.168.20.15 is AVAILABLE for provisioning`
- Or `CONFLICT: IP ... is NOT AVAILABLE (Reason: already allocated by DHCPStatic / ARP)`
