package network_test

import (
	"context"
	"testing"

	"github.com/sekai-labs/michibiki/internal/handler/network"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestNetworkHandler_GetRoutingOverview(t *testing.T) {
	ctx := context.Background()
	prov := &mockProviderPort{
		caps: provider.CapRouting | provider.CapBGP,
		routes: []model.Route{
			{Destination: "0.0.0.0/0", Gateway: "10.0.0.1", Interface: "eth0"},
		},
		gateways: []model.Gateway{
			{Name: "default_gw", Address: "10.0.0.1", IsDefault: true, Status: model.GatewayOnline},
		},
		bgpNeighbors: []model.BGPNeighbor{
			{PeerAddress: "10.0.0.2", State: model.BGPEstablished},
			{PeerAddress: "10.0.0.3", State: model.BGPActive},
		},
	}

	svc := network.NewNetworkHandler(prov)
	overview, err := svc.GetRoutingOverview(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.TotalRoutes != 1 {
		t.Fatalf("expected 1 route, got %d", overview.TotalRoutes)
	}
	if overview.DefaultGateway == nil || overview.DefaultGateway.Address != "10.0.0.1" {
		t.Fatalf("default gateway not resolved correctly")
	}
	if overview.ActivePeers != 1 {
		t.Fatalf("expected 1 active peer (Established), got %d", overview.ActivePeers)
	}
}
