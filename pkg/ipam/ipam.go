package ipam

import (
	"cmp"
	"fmt"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/sekai-labs/michibiki/pkg/model"
)

type AllocationSource string

const (
	SourceInterfaceIP   AllocationSource = "InterfaceIP"
	SourceDHCPDynamic   AllocationSource = "DHCPDynamic"
	SourceDHCPStatic    AllocationSource = "DHCPStatic"
	SourceARPDiscovery  AllocationSource = "ARPDiscovery"
	SourceStaticRoute   AllocationSource = "StaticRoute"
	SourceWireGuardPeer AllocationSource = "WireGuardPeer"
	SourceReserved      AllocationSource = "Reserved"
)

type AllocationStatus string

const (
	StatusActive   AllocationStatus = "active"
	StatusOnline   AllocationStatus = "online"
	StatusStale    AllocationStatus = "stale"
	StatusReserved AllocationStatus = "reserved"
)

type IPAllocation struct {
	IP        netip.Addr       `json:"ip" yaml:"ip"`
	MAC       string           `json:"mac" yaml:"mac"`
	Hostname  string           `json:"hostname" yaml:"hostname"`
	Source    AllocationSource `json:"source" yaml:"source"`
	Status    AllocationStatus `json:"status" yaml:"status"`
	ExpiresAt time.Time        `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
}

type IPRange struct {
	Start netip.Addr `json:"start" yaml:"start"`
	End   netip.Addr `json:"end" yaml:"end"`
	Count int        `json:"count" yaml:"count"`
}

type SubnetUsage struct {
	CIDR           netip.Prefix   `json:"cidr" yaml:"cidr"`
	InterfaceName  string         `json:"interface_name" yaml:"interface_name"`
	VLANID         int            `json:"vlan_id,omitempty" yaml:"vlan_id,omitempty"`
	NetworkIP      netip.Addr     `json:"network_ip" yaml:"network_ip"`
	BroadcastIP    netip.Addr     `json:"broadcast_ip,omitempty" yaml:"broadcast_ip,omitempty"`
	GatewayIP      netip.Addr     `json:"gateway_ip,omitempty" yaml:"gateway_ip,omitempty"`
	TotalIPs       int            `json:"total_ips" yaml:"total_ips"`
	UsableIPs      int            `json:"usable_ips" yaml:"usable_ips"`
	UsedIPs        int            `json:"used_ips" yaml:"used_ips"`
	FreeIPs        int            `json:"free_ips" yaml:"free_ips"`
	UtilizationPct float64        `json:"utilization_pct" yaml:"utilization_pct"`
	Allocations    []IPAllocation `json:"allocations" yaml:"allocations"`
	FreeRanges     []IPRange      `json:"free_ranges" yaml:"free_ranges"`
}

func CalculateSubnetUsage(
	prefix netip.Prefix,
	ifaceName string,
	vlanID int,
	interfaces []model.Interface,
	leases []model.DHCPLease,
	arpEntries []model.ARPEntry,
	wgPeers []model.WireGuardPeer,
) SubnetUsage {
	prefix = prefix.Masked()
	isIPv4 := prefix.Addr().Is4()
	totalCount := prefixTotalIPs(prefix)
	usableCount, networkIP, broadcastIP := prefixUsableBounds(prefix)

	allocMap := make(map[netip.Addr]IPAllocation)

	var gatewayIP netip.Addr
	for _, iface := range interfaces {
		allIPs := append([]string{}, iface.IPv4Addresses...)
		allIPs = append(allIPs, iface.IPv6Addresses...)
		for _, addrStr := range allIPs {
			parsedPrefix, err := netip.ParsePrefix(addrStr)
			if err != nil {
				parsedAddr, err2 := netip.ParseAddr(addrStr)
				if err2 != nil {
					continue
				}
				if prefix.Contains(parsedAddr) {
					if !gatewayIP.IsValid() || iface.Name == ifaceName {
						gatewayIP = parsedAddr
					}
					allocMap[parsedAddr] = IPAllocation{
						IP:       parsedAddr,
						MAC:      iface.MACAddress,
						Hostname: iface.Name,
						Source:   SourceInterfaceIP,
						Status:   StatusReserved,
					}
				}
				continue
			}

			if prefix.Contains(parsedPrefix.Addr()) {
				if !gatewayIP.IsValid() || iface.Name == ifaceName {
					gatewayIP = parsedPrefix.Addr()
				}
				allocMap[parsedPrefix.Addr()] = IPAllocation{
					IP:       parsedPrefix.Addr(),
					MAC:      iface.MACAddress,
					Hostname: iface.Name,
					Source:   SourceInterfaceIP,
					Status:   StatusReserved,
				}
			}
		}
	}

	for _, lease := range leases {
		parsedAddr, err := netip.ParseAddr(lease.IPAddress)
		if err != nil || !prefix.Contains(parsedAddr) {
			continue
		}

		source := SourceDHCPDynamic
		if lease.State == model.DHCPStatic {
			source = SourceDHCPStatic
		}

		existing, exists := allocMap[parsedAddr]
		if exists {
			if existing.MAC == "" {
				existing.MAC = lease.MACAddress
			}
			if existing.Hostname == "" {
				existing.Hostname = lease.ClientHostname
			}
			existing.ExpiresAt = lease.Ends
			allocMap[parsedAddr] = existing
		} else {
			allocMap[parsedAddr] = IPAllocation{
				IP:        parsedAddr,
				MAC:       lease.MACAddress,
				Hostname:  lease.ClientHostname,
				Source:    source,
				Status:    StatusActive,
				ExpiresAt: lease.Ends,
			}
		}
	}

	for _, arp := range arpEntries {
		parsedAddr, err := netip.ParseAddr(arp.IPAddress)
		if err != nil || !prefix.Contains(parsedAddr) {
			continue
		}

		existing, exists := allocMap[parsedAddr]
		if exists {
			if existing.MAC == "" {
				existing.MAC = arp.MACAddress
			}
			if existing.Hostname == "" && arp.Hostname != "" {
				existing.Hostname = arp.Hostname
			}
			if existing.Status == StatusStale && arp.Status == model.ARPReachable {
				existing.Status = StatusActive
			}
			allocMap[parsedAddr] = existing
		} else {
			status := StatusActive
			if arp.Status == model.ARPStale {
				status = StatusStale
			}
			allocMap[parsedAddr] = IPAllocation{
				IP:       parsedAddr,
				MAC:      arp.MACAddress,
				Hostname: arp.Hostname,
				Source:   SourceARPDiscovery,
				Status:   status,
			}
		}
	}

	for _, peer := range wgPeers {
		for _, allowed := range peer.AllowedIPs {
			pfx, err := netip.ParsePrefix(allowed)
			if err != nil {
				parsedAddr, err2 := netip.ParseAddr(allowed)
				if err2 == nil && prefix.Contains(parsedAddr) {
					if _, exists := allocMap[parsedAddr]; !exists {
						allocMap[parsedAddr] = IPAllocation{
							IP:       parsedAddr,
							Source:   SourceWireGuardPeer,
							Status:   StatusActive,
							Hostname: peer.Endpoint,
						}
					}
				}
				continue
			}
			if prefix.Contains(pfx.Addr()) {
				if _, exists := allocMap[pfx.Addr()]; !exists {
					allocMap[pfx.Addr()] = IPAllocation{
						IP:       pfx.Addr(),
						Source:   SourceWireGuardPeer,
						Status:   StatusActive,
						Hostname: peer.Endpoint,
					}
				}
			}
		}
	}

	allocations := make([]IPAllocation, 0, len(allocMap))
	for _, alloc := range allocMap {
		if isIPv4 && (alloc.IP == networkIP || alloc.IP == broadcastIP) && prefix.Bits() < 31 {
			continue
		}
		allocations = append(allocations, alloc)
	}

	slices.SortFunc(allocations, func(a, b IPAllocation) int {
		return a.IP.Compare(b.IP)
	})

	freeRanges := calculateFreeRanges(prefix, allocMap, networkIP, broadcastIP)

	usedCount := len(allocations)
	freeCount := usableCount - usedCount
	if freeCount < 0 {
		freeCount = 0
	}

	var utilPct float64
	if usableCount > 0 {
		utilPct = (float64(usedCount) / float64(usableCount)) * 100.0
	}

	return SubnetUsage{
		CIDR:           prefix,
		InterfaceName:  ifaceName,
		VLANID:         vlanID,
		NetworkIP:      networkIP,
		BroadcastIP:    broadcastIP,
		GatewayIP:      gatewayIP,
		TotalIPs:       totalCount,
		UsableIPs:      usableCount,
		UsedIPs:        usedCount,
		FreeIPs:        freeCount,
		UtilizationPct: utilPct,
		Allocations:    allocations,
		FreeRanges:     freeRanges,
	}
}

func (su *SubnetUsage) FindNextFree(count int) []netip.Addr {
	if count <= 0 {
		return nil
	}

	result := make([]netip.Addr, 0, count)
	for _, r := range su.FreeRanges {
		curr := r.Start
		for {
			result = append(result, curr)
			if len(result) == count {
				return result
			}
			if curr == r.End {
				break
			}
			curr = curr.Next()
		}
	}
	return result
}

func (su *SubnetUsage) ValidateIPAvailable(candidate netip.Addr) (bool, string) {
	if !su.CIDR.Contains(candidate) {
		return false, fmt.Sprintf("IP %s does not belong to subnet %s", candidate, su.CIDR)
	}

	if su.CIDR.Addr().Is4() && su.CIDR.Bits() < 31 {
		if candidate == su.NetworkIP {
			return false, fmt.Sprintf("IP %s is the network address", candidate)
		}
		if candidate == su.BroadcastIP {
			return false, fmt.Sprintf("IP %s is the broadcast address", candidate)
		}
	}

	for _, alloc := range su.Allocations {
		if alloc.IP == candidate {
			return false, fmt.Sprintf("IP %s is already allocated (%s, %s)", candidate, alloc.Source, alloc.Hostname)
		}
	}

	return true, ""
}

func prefixTotalIPs(prefix netip.Prefix) int {
	if !prefix.Addr().Is4() {
		return 65536
	}
	bits := prefix.Bits()
	if bits >= 32 {
		return 1
	}
	return 1 << (32 - bits)
}

func prefixUsableBounds(prefix netip.Prefix) (int, netip.Addr, netip.Addr) {
	if !prefix.Addr().Is4() {
		netAddr := prefix.Masked().Addr()
		return 65536, netAddr, netip.Addr{}
	}

	bits := prefix.Bits()
	total := 1 << (32 - bits)
	netAddr := prefix.Masked().Addr()
	b := netAddr.As4()

	bcastNum := (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
	bcastNum += uint32(total - 1)

	bcastBytes := [4]byte{
		byte(bcastNum >> 24),
		byte(bcastNum >> 16),
		byte(bcastNum >> 8),
		byte(bcastNum),
	}
	broadcastAddr := netip.AddrFrom4(bcastBytes)

	if bits == 32 {
		return 1, netAddr, netAddr
	}
	if bits == 31 {
		return 2, netAddr, broadcastAddr
	}

	return total - 2, netAddr, broadcastAddr
}

func calculateFreeRanges(prefix netip.Prefix, allocMap map[netip.Addr]IPAllocation, netIP, bcastIP netip.Addr) []IPRange {
	if !prefix.Addr().Is4() {
		return nil
	}

	bits := prefix.Bits()
	if bits < 16 {
		return nil
	}

	var startAddr netip.Addr
	var endAddr netip.Addr

	if bits == 32 {
		startAddr = netIP
		endAddr = netIP
	} else if bits == 31 {
		startAddr = netIP
		endAddr = bcastIP
	} else {
		startAddr = netIP.Next()
		endAddr = bcastIP.Prev()
	}

	var ranges []IPRange
	var currentStart netip.Addr
	inFree := false
	count := 0

	curr := startAddr
	for {
		_, allocated := allocMap[curr]
		if !allocated {
			if !inFree {
				inFree = true
				currentStart = curr
				count = 1
			} else {
				count++
			}
		} else {
			if inFree {
				ranges = append(ranges, IPRange{
					Start: currentStart,
					End:   curr.Prev(),
					Count: count,
				})
				inFree = false
				count = 0
			}
		}

		if curr == endAddr {
			if inFree {
				ranges = append(ranges, IPRange{
					Start: currentStart,
					End:   curr,
					Count: count,
				})
			}
			break
		}
		curr = curr.Next()
	}

	return ranges
}

func DiscoverSubnets(
	interfaces []model.Interface,
	leases []model.DHCPLease,
	arpEntries []model.ARPEntry,
	wgPeers []model.WireGuardPeer,
) []SubnetUsage {
	seenPrefixes := make(map[netip.Prefix]bool)
	var results []SubnetUsage

	for _, iface := range interfaces {
		for _, addrStr := range iface.IPv4Addresses {
			prefix, err := netip.ParsePrefix(addrStr)
			if err != nil {
				continue
			}
			prefix = prefix.Masked()
			if seenPrefixes[prefix] {
				continue
			}
			seenPrefixes[prefix] = true

			usage := CalculateSubnetUsage(prefix, iface.Name, iface.VLANID, interfaces, leases, arpEntries, wgPeers)
			results = append(results, usage)
		}
	}

	slices.SortFunc(results, func(a, b SubnetUsage) int {
		return cmp.Compare(a.CIDR.String(), b.CIDR.String())
	})

	return results
}

func MatchSubnet(usages []SubnetUsage, query string) *SubnetUsage {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	for i := range usages {
		if strings.EqualFold(usages[i].InterfaceName, query) {
			return &usages[i]
		}
		if usages[i].CIDR.String() == query {
			return &usages[i]
		}
		if fmt.Sprintf("vlan%d", usages[i].VLANID) == strings.ToLower(query) {
			return &usages[i]
		}
		if fmt.Sprintf("%d", usages[i].VLANID) == query && usages[i].VLANID > 0 {
			return &usages[i]
		}
	}

	parsedAddr, err := netip.ParseAddr(query)
	if err == nil {
		for i := range usages {
			if usages[i].CIDR.Contains(parsedAddr) {
				return &usages[i]
			}
		}
	}

	return nil
}
