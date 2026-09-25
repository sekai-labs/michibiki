package domain

import (
	"errors"

	"github.com/sekai-labs/michibiki/pkg/model"
)

var (
	ErrUnsupportedCapability = errors.New("unsupported capability for target device")
)

type DeviceOverview struct {
	SystemInfo       *model.SystemInfo
	TotalInterfaces  int
	UpInterfaces     int
	TotalGateways    int
	OnlineGateways   int
	DegradedGateways int
	HealthScore      int
	StatusSummary    string
}

func CalculateHealthScore(info *model.SystemInfo, gateways []model.Gateway) int {
	score := 100

	if info != nil {
		if info.CPUUsagePct > 90 {
			score -= 30
		} else if info.CPUUsagePct > 75 {
			score -= 15
		}

		if info.MemoryTotalBytes > 0 {
			memPct := (float64(info.MemoryUsedBytes) / float64(info.MemoryTotalBytes)) * 100
			if memPct > 90 {
				score -= 30
			} else if memPct > 75 {
				score -= 15
			}
		}
	}

	for _, gw := range gateways {
		if gw.Status == model.GatewayOffline {
			score -= 25
		} else if gw.Status == model.GatewayDegraded {
			score -= 10
		}
		if gw.PacketLossPct > 20 {
			score -= 10
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}
