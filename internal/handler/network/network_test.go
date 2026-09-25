package network_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sekai-labs/michibiki/internal/domain"
	"github.com/sekai-labs/michibiki/internal/handler/network"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)


func TestNetworkHandler_GetDeviceOverview(t *testing.T) {
	ctx := context.Background()

	provNoCap := &mockProviderPort{caps: 0}
	svcNoCap := network.NewNetworkHandler(provNoCap)
	_, err := svcNoCap.GetDeviceOverview(ctx)
	if !errors.Is(err, domain.ErrUnsupportedCapability) {
		t.Fatalf("expected ErrUnsupportedCapability, got: %v", err)
	}

	prov := &mockProviderPort{
		caps: provider.CapSystem | provider.CapInterfaces | provider.CapRouting,
		sysInfo: &model.SystemInfo{
			Hostname:         "edge-gw-01",
			CPUUsagePct:      15.0,
			MemoryTotalBytes: 1000000,
			MemoryUsedBytes:  200000,
		},
		ifaces: []model.Interface{
			{Name: "eth0", OperStatus: model.OperStatusUp},
			{Name: "eth1", OperStatus: model.OperStatusDown},
		},
		gateways: []model.Gateway{
			{Name: "WAN_GW", Status: model.GatewayOnline, IsDefault: true},
		},
	}

	svc := network.NewNetworkHandler(prov)
	overview, err := svc.GetDeviceOverview(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.TotalInterfaces != 2 || overview.UpInterfaces != 1 {
		t.Fatalf("unexpected interfaces: total=%d, up=%d", overview.TotalInterfaces, overview.UpInterfaces)
	}
	if overview.HealthScore < 90 {
		t.Fatalf("expected health score >= 90, got %d", overview.HealthScore)
	}
	if overview.StatusSummary != "HEALTHY" {
		t.Fatalf("expected HEALTHY status summary, got %s", overview.StatusSummary)
	}
}

func TestNetworkHandler_ListInterfacesWithStats(t *testing.T) {
	ctx := context.Background()
	prov := &mockProviderPort{
		caps: provider.CapInterfaces | provider.CapMonitoring,
		ifaces: []model.Interface{
			{Name: "eth0", SpeedBps: 1_000_000_000, OperStatus: model.OperStatusUp},
		},
		stats: []model.InterfaceStats{
			{InterfaceName: "eth0", RxBps: 200_000_000, TxBps: 100_000_000},
		},
	}

	svc := network.NewNetworkHandler(prov)
	items, err := svc.ListInterfacesWithStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 interface item, got %d", len(items))
	}
	if items[0].RxUsagePct != 20.0 {
		t.Fatalf("expected 20.0%% RxUsage, got %.1f%%", items[0].RxUsagePct)
	}
	if items[0].TxUsagePct != 10.0 {
		t.Fatalf("expected 10.0%% TxUsage, got %.1f%%", items[0].TxUsagePct)
	}
}
