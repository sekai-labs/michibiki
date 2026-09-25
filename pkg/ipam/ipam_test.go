package ipam

import (
	"net/netip"
	"testing"
	"time"

	"github.com/sekai-labs/michibiki/pkg/model"
)

func TestCalculateSubnetUsageSlash24(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.10.0/24")

	ifaces := []model.Interface{
		{
			Name:          "vlan10",
			VLANID:        10,
			IPv4Addresses: []string{"192.168.10.1/24"},
			MACAddress:    "00:11:22:33:44:01",
		},
	}

	leases := []model.DHCPLease{
		{
			IPAddress:      "192.168.10.50",
			MACAddress:     "aa:bb:cc:dd:ee:01",
			ClientHostname: "server01",
			State:          model.DHCPActive,
			Ends:           time.Now().Add(12 * time.Hour),
		},
		{
			IPAddress:      "192.168.10.51",
			MACAddress:     "aa:bb:cc:dd:ee:02",
			ClientHostname: "desktop01",
			State:          model.DHCPActive,
			Ends:           time.Now().Add(6 * time.Hour),
		},
	}

	arps := []model.ARPEntry{
		{
			IPAddress:  "192.168.10.50",
			MACAddress: "aa:bb:cc:dd:ee:01",
			Hostname:   "server01.local",
			Status:     model.ARPReachable,
		},
		{
			IPAddress:  "192.168.10.100",
			MACAddress: "aa:bb:cc:dd:ee:99",
			Hostname:   "printer",
			Status:     model.ARPReachable,
		},
	}

	wg := []model.WireGuardPeer{
		{
			Endpoint:   "198.51.100.1:51820",
			AllowedIPs: []string{"192.168.10.200/32"},
		},
	}

	usage := CalculateSubnetUsage(prefix, "vlan10", 10, ifaces, leases, arps, wg)

	if usage.TotalIPs != 256 {
		t.Fatalf("expected total 256, got %d", usage.TotalIPs)
	}

	if usage.UsableIPs != 254 {
		t.Fatalf("expected usable 254, got %d", usage.UsableIPs)
	}

	if usage.UsedIPs != 5 {
		t.Fatalf("expected 5 used IPs (1 gateway, 2 leases, 1 arp, 1 wg), got %d", usage.UsedIPs)
	}

	if usage.FreeIPs != 249 {
		t.Fatalf("expected 249 free IPs, got %d", usage.FreeIPs)
	}

	nextFree := usage.FindNextFree(3)
	if len(nextFree) != 3 {
		t.Fatalf("expected 3 free IPs, got %d", len(nextFree))
	}
	if nextFree[0].String() != "192.168.10.2" {
		t.Errorf("expected 192.168.10.2, got %s", nextFree[0])
	}
	if nextFree[1].String() != "192.168.10.3" {
		t.Errorf("expected 192.168.10.3, got %s", nextFree[1])
	}
	if nextFree[2].String() != "192.168.10.4" {
		t.Errorf("expected 192.168.10.4, got %s", nextFree[2])
	}

	avail, reason := usage.ValidateIPAvailable(netip.MustParseAddr("192.168.10.50"))
	if avail {
		t.Errorf("expected 192.168.10.50 to be unavailable, reason: %s", reason)
	}

	availGateway, _ := usage.ValidateIPAvailable(netip.MustParseAddr("192.168.10.1"))
	if availGateway {
		t.Errorf("expected gateway 192.168.10.1 to be unavailable")
	}

	availFree, _ := usage.ValidateIPAvailable(netip.MustParseAddr("192.168.10.15"))
	if !availFree {
		t.Errorf("expected 192.168.10.15 to be available")
	}
}

func TestSubnetSlash30And31(t *testing.T) {
	p30 := netip.MustParsePrefix("10.0.0.0/30")
	u30 := CalculateSubnetUsage(p30, "eth1", 0, nil, nil, nil, nil)
	if u30.TotalIPs != 4 || u30.UsableIPs != 2 {
		t.Errorf("unexpected /30: total %d usable %d", u30.TotalIPs, u30.UsableIPs)
	}

	p31 := netip.MustParsePrefix("10.0.0.0/31")
	u31 := CalculateSubnetUsage(p31, "eth2", 0, nil, nil, nil, nil)
	if u31.TotalIPs != 2 || u31.UsableIPs != 2 {
		t.Errorf("unexpected /31: total %d usable %d", u31.TotalIPs, u31.UsableIPs)
	}
}

func TestMatchSubnet(t *testing.T) {
	p1 := netip.MustParsePrefix("192.168.1.0/24")
	p2 := netip.MustParsePrefix("10.20.0.0/16")

	usages := []SubnetUsage{
		{
			CIDR:          p1,
			InterfaceName: "lan",
			VLANID:        1,
		},
		{
			CIDR:          p2,
			InterfaceName: "dmz",
			VLANID:        20,
		},
	}

	m1 := MatchSubnet(usages, "lan")
	if m1 == nil || m1.InterfaceName != "lan" {
		t.Fatalf("match failed for 'lan'")
	}

	m2 := MatchSubnet(usages, "vlan20")
	if m2 == nil || m2.VLANID != 20 {
		t.Fatalf("match failed for 'vlan20'")
	}

	m3 := MatchSubnet(usages, "10.20.15.4")
	if m3 == nil || m3.InterfaceName != "dmz" {
		t.Fatalf("match failed for IP '10.20.15.4'")
	}
}
