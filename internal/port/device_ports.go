package port

import (
	"context"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type DevicePort interface {
	ID() string
	Name() string
	Connect(ctx context.Context, endpoint string, creds *credential.Credentials, options map[string]string) error
	Disconnect(ctx context.Context) error
	Capabilities() provider.Capabilities
	GetSystemInfo(ctx context.Context) (*model.SystemInfo, error)
}

type InterfacePort interface {
	ListInterfaces(ctx context.Context) ([]model.Interface, error)
	GetInterface(ctx context.Context, name string) (*model.Interface, error)
	GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error)
}

type RoutingPort interface {
	ListRoutes(ctx context.Context) ([]model.Route, error)
	ListGateways(ctx context.Context) ([]model.Gateway, error)
	ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error)
}
