package cli

import (
	"bytes"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/ipam"
	"github.com/sekai-labs/michibiki/pkg/model"
)

func TestCobraCLICommands(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	testCases := [][]string{
		{"device", "list", "-o", "json"},
		{"plugin", "list", "-o", "json"},
	}

	for _, tc := range testCases {
		rootCmd.SetArgs(tc)
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("failed executing %v: %v", tc, err)
		}
	}
}

func TestIPAMLogic(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.20.0/24")
	ifaces := []model.Interface{
		{
			Name:          "vlan20",
			VLANID:        20,
			IPv4Addresses: []string{"192.168.20.1/24"},
			MACAddress:    "52:54:00:12:34:02",
		},
	}
	leases := []model.DHCPLease{
		{
			IPAddress:      "192.168.20.10",
			MACAddress:     "52:54:00:20:00:10",
			ClientHostname: "server01",
			State:          model.DHCPStatic,
		},
	}
	usage := ipam.CalculateSubnetUsage(prefix, "vlan20", 20, ifaces, leases, nil, nil)
	if usage.FreeIPs <= 0 {
		t.Errorf("expected free IPs in vlan20, got %d", usage.FreeIPs)
	}

	freeIPs := usage.FindNextFree(3)
	if len(freeIPs) != 3 {
		t.Errorf("expected 3 free IPs, got %d", len(freeIPs))
	}

	for _, ip := range freeIPs {
		avail, _ := usage.ValidateIPAvailable(ip)
		if !avail {
			t.Errorf("ip %s should be free", ip.String())
		}
	}

	collidingIP := netip.MustParseAddr("192.168.20.1")
	avail, _ := usage.ValidateIPAvailable(collidingIP)
	if avail {
		t.Errorf("gateway IP should not be reported as available")
	}
}

func TestConfigDeviceLookup(t *testing.T) {
	cfg := &config.Config{
		Devices: map[string]config.DeviceProfile{
			"gw": {
				Name:     "gw",
				Provider: "opnsense",
				Address:  "https://192.168.1.1",
			},
		},
	}
	dev, err := cfg.GetDevice("gw")
	if err != nil {
		t.Fatalf("expected device gw: %v", err)
	}
	if dev.Provider != "opnsense" {
		t.Errorf("expected provider opnsense, got %s", dev.Provider)
	}
}

func TestAuthTokenCLICommands(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "test_token.txt")
	if err := os.WriteFile(tokenFile, []byte("tok_sample_test_credential_9999\n"), 0600); err != nil {
		t.Fatalf("failed to write test token file: %v", err)
	}
	rootCmd.SetArgs([]string{"auth", "token", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth token --help failed: %v", err)
	}

	rootCmd.SetArgs([]string{"auth", "token", "import", "-f", tokenFile, "--device", "router-01"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth token import failed: %v", err)
	}

	rootCmd.SetArgs([]string{"auth", "token", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth token status failed: %v", err)
	}

	rootCmd.SetArgs([]string{"auth", "token", "clear"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth token clear failed: %v", err)
	}

	rootCmd.SetArgs([]string{"auth", "token", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth token status failed after clear: %v", err)
	}
}
