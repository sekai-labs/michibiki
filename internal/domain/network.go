package domain

import "github.com/sekai-labs/michibiki/pkg/model"

type InterfaceOverview struct {
	Interface  model.Interface
	Stats      *model.InterfaceStats
	RxUsagePct float64
	TxUsagePct float64
}

type RoutingOverview struct {
	Routes         []model.Route
	Gateways       []model.Gateway
	BGPNeighbors   []model.BGPNeighbor
	DefaultGateway *model.Gateway
	TotalRoutes    int
	ActivePeers    int
}
