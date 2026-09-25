package model

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestSystemInfoSerialization(t *testing.T) {
	info := SystemInfo{
		Hostname:         "gateway01",
		OS:               "OPNsense",
		Version:          "24.1",
		Architecture:     "amd64",
		UptimeSeconds:    86400,
		CPUCount:         4,
		CPUUsagePct:      12.5,
		MemoryTotalBytes: 16 * 1024 * 1024 * 1024,
		MemoryUsedBytes:  4 * 1024 * 1024 * 1024,
		StorageTotal:     120 * 1024 * 1024 * 1024,
		StorageUsed:      20 * 1024 * 1024 * 1024,
		SerialNumber:     "SN12345678",
		Time:             time.Now(),
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded SystemInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Hostname != info.Hostname {
		t.Errorf("expected hostname %s, got %s", info.Hostname, decoded.Hostname)
	}

	yamlData, err := yaml.Marshal(info)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	var yamlDecoded SystemInfo
	if err := yaml.Unmarshal(yamlData, &yamlDecoded); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if yamlDecoded.Hostname != info.Hostname {
		t.Errorf("expected hostname %s, got %s", info.Hostname, yamlDecoded.Hostname)
	}
}

func TestInterfaceModel(t *testing.T) {
	iface := Interface{
		ID:            "eth0",
		Name:          "wan0",
		Type:          InterfaceTypeEthernet,
		AdminStatus:   AdminStatusUp,
		OperStatus:    OperStatusUp,
		MACAddress:    "00:11:22:33:44:55",
		MTU:           1500,
		IPv4Addresses: []string{"192.168.1.100/24"},
		IPv6Addresses: []string{"2001:db8::1/64"},
		SpeedBps:      1000000000,
		Duplex:        "full",
		Description:   "Uplink to ISP",
	}

	if iface.Name != "wan0" || iface.AdminStatus != AdminStatusUp {
		t.Errorf("interface field mismatch: %+v", iface)
	}
}
