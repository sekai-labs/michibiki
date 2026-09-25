package network

import (
	"context"
	"github.com/sekai-labs/michibiki/internal/domain"
	"github.com/sekai-labs/michibiki/internal/port"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)
type NetworkHandler struct {
	prov port.ProviderPort
}

func NewNetworkHandler(prov port.ProviderPort) *NetworkHandler {
	return &NetworkHandler{prov: prov}
}

func (s *NetworkHandler) Provider() port.ProviderPort {
	return s.prov
}

func (s *NetworkHandler) Capabilities() provider.Capabilities {
	if s.prov == nil {
		return 0
	}
	return s.prov.Capabilities()
}

func (s *NetworkHandler) GetDeviceOverview(ctx context.Context) (*domain.DeviceOverview, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapSystem) {
		return nil, domain.ErrUnsupportedCapability
	}

	sysInfo, err := s.prov.GetSystemInfo(ctx)
	if err != nil {
		return nil, err
	}

	var ifaces []model.Interface
	if s.prov.Capabilities().Has(provider.CapInterfaces) {
		if ifList, err := s.prov.ListInterfaces(ctx); err == nil {
			ifaces = ifList
		}
	}

	var gateways []model.Gateway
	if s.prov.Capabilities().Has(provider.CapRouting) {
		if gwList, err := s.prov.ListGateways(ctx); err == nil {
			gateways = gwList
		}
	}

	upCount := 0
	for _, iface := range ifaces {
		if iface.OperStatus == model.OperStatusUp {
			upCount++
		}
	}

	onlineGws := 0
	degradedGws := 0
	for _, gw := range gateways {
		if gw.Status == model.GatewayOnline {
			onlineGws++
		} else if gw.Status == model.GatewayDegraded {
			degradedGws++
		}
	}

	health := domain.CalculateHealthScore(sysInfo, gateways)

	statusSummary := "HEALTHY"
	if health < 50 {
		statusSummary = "CRITICAL"
	} else if health < 80 {
		statusSummary = "DEGRADED"
	}

	return &domain.DeviceOverview{
		SystemInfo:       sysInfo,
		TotalInterfaces:  len(ifaces),
		UpInterfaces:     upCount,
		TotalGateways:    len(gateways),
		OnlineGateways:   onlineGws,
		DegradedGateways: degradedGws,
		HealthScore:      health,
		StatusSummary:    statusSummary,
	}, nil
}

func (s *NetworkHandler) ListInterfacesWithStats(ctx context.Context) ([]domain.InterfaceOverview, error) {
	if s.prov == nil {
		return nil, provider.ErrNotConnected
	}
	if !s.prov.Capabilities().Has(provider.CapInterfaces) {
		return nil, domain.ErrUnsupportedCapability
	}

	ifaces, err := s.prov.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	statsMap := make(map[string]model.InterfaceStats)
	if s.prov.Capabilities().Has(provider.CapMonitoring) {
		if stats, err := s.prov.GetInterfaceStats(ctx); err == nil {
			for _, st := range stats {
				statsMap[st.InterfaceName] = st
			}
		}
	}

	overview := make([]domain.InterfaceOverview, 0, len(ifaces))
	for _, iface := range ifaces {
		item := domain.InterfaceOverview{
			Interface: iface,
		}
		if st, ok := statsMap[iface.Name]; ok {
			stCopy := st
			item.Stats = &stCopy
			if iface.SpeedBps > 0 {
				item.RxUsagePct = (st.RxBps / float64(iface.SpeedBps)) * 100.0
				item.TxUsagePct = (st.TxBps / float64(iface.SpeedBps)) * 100.0
			}
		}
		overview = append(overview, item)
	}

	return overview, nil
}
