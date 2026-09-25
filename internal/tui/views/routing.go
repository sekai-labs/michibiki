package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type RoutingData struct {
	Routes    []model.Route
	Gateways  []model.Gateway
	Neighbors []model.BGPNeighbor
	Error     error
}

func RenderRouting(data RoutingData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load routing data: " + data.Error.Error()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	gwTitle := theme.StyleTitle.Render("GATEWAYS")
	gwHeader := fmt.Sprintf("%-16s %-16s %-12s %-10s %-12s %-10s %-10s",
		"NAME", "ADDRESS", "INTERFACE", "STATUS", "LATENCY", "LOSS", "DEFAULT")
	gwHeaderRendered := theme.StyleTableHeader.Render(gwHeader)

	var gwRows []string
	for i, gw := range data.Gateways {
		badge := theme.StyleBadgeOnline.Render("ONLINE")
		if gw.Status != model.GatewayOnline {
			badge = theme.StyleBadgeOffline.Render(strings.ToUpper(string(gw.Status)))
		}

		defBadge := "-"
		if gw.IsDefault {
			defBadge = theme.StyleBadgeInfo.Render("YES")
		}

		line := fmt.Sprintf("%-16s %-16s %-12s %-10s %-12s %-10s %-10s",
			gw.Name,
			gw.Address,
			gw.Interface,
			badge,
			fmt.Sprintf("%.1f ms", gw.LatencyMs),
			fmt.Sprintf("%.1f%%", gw.PacketLossPct),
			defBadge,
		)
		if i%2 == 0 {
			gwRows = append(gwRows, theme.StyleTableRow.Render(line))
		} else {
			gwRows = append(gwRows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(gwRows) == 0 {
		gwRows = append(gwRows, theme.StyleMuted.Render("No gateways configured."))
	}
	gwCard := theme.StyleCard.Width(cardWidth).Render(gwTitle + "\n\n" + gwHeaderRendered + "\n" + strings.Join(gwRows, "\n"))

	routeTitle := theme.StyleTitle.Render(fmt.Sprintf("ROUTING TABLE (%d ROUTES)", len(data.Routes)))
	routeHeader := fmt.Sprintf("%-24s %-18s %-14s %-12s %-8s %-10s",
		"DESTINATION", "GATEWAY", "INTERFACE", "PROTOCOL", "METRIC", "SCOPE")
	routeHeaderRendered := theme.StyleTableHeader.Render(routeHeader)

	var routeRows []string
	for i, r := range data.Routes {
		gwStr := r.Gateway
		if gwStr == "" {
			gwStr = "*"
		}
		ifaceStr := r.Interface
		if ifaceStr == "" {
			ifaceStr = "*"
		}

		line := fmt.Sprintf("%-24s %-18s %-14s %-12s %-8d %-10s",
			r.Destination,
			gwStr,
			ifaceStr,
			string(r.Protocol),
			r.Metric,
			r.Scope,
		)
		if i%2 == 0 {
			routeRows = append(routeRows, theme.StyleTableRow.Render(line))
		} else {
			routeRows = append(routeRows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(routeRows) == 0 {
		routeRows = append(routeRows, theme.StyleMuted.Render("No routes found."))
	}
	routeCard := theme.StyleCard.Width(cardWidth).Render(routeTitle + "\n\n" + routeHeaderRendered + "\n" + strings.Join(routeRows, "\n"))

	var bgpCard string
	if len(data.Neighbors) > 0 {
		bgpTitle := theme.StyleTitle.Render(fmt.Sprintf("BGP NEIGHBORS (%d)", len(data.Neighbors)))
		bgpHeader := fmt.Sprintf("%-16s %-10s %-10s %-14s %-12s %-12s",
			"PEER ADDRESS", "REMOTE AS", "LOCAL AS", "STATE", "PFX RCVD", "PFX ACCEPT")
		bgpHeaderRendered := theme.StyleTableHeader.Render(bgpHeader)

		var bgpRows []string
		for i, n := range data.Neighbors {
			stateBadge := theme.StyleBadgeOnline.Render(string(n.State))
			if n.State != model.BGPEstablished {
				stateBadge = theme.StyleBadgeWarning.Render(string(n.State))
			}

			line := fmt.Sprintf("%-16s %-10d %-10d %-14s %-12d %-12d",
				n.PeerAddress,
				n.RemoteAS,
				n.LocalAS,
				stateBadge,
				n.PrefixesReceived,
				n.PrefixesAccepted,
			)
			if i%2 == 0 {
				bgpRows = append(bgpRows, theme.StyleTableRow.Render(line))
			} else {
				bgpRows = append(bgpRows, theme.StyleTableRowAlt.Render(line))
			}
		}
		bgpCard = theme.StyleCard.Width(cardWidth).Render(bgpTitle + "\n\n" + bgpHeaderRendered + "\n" + strings.Join(bgpRows, "\n"))
	}

	if bgpCard != "" {
		return lipgloss.JoinVertical(lipgloss.Left, gwCard, routeCard, bgpCard)
	}
	return lipgloss.JoinVertical(lipgloss.Left, gwCard, routeCard)
}
