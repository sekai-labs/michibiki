package views

import (
	"fmt"
	"strings"

	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type InterfacesData struct {
	Interfaces []model.Interface
	Stats      map[string]model.InterfaceStats
	Error      error
}

func RenderInterfaces(data InterfacesData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load interfaces: " + data.Error.Error()),
		)
	}

	if len(data.Interfaces) == 0 {
		return theme.StyleCard.Width(width - 4).Render(
			theme.StyleMuted.Render("No interfaces found."),
		)
	}

	header := fmt.Sprintf("%-12s %-10s %-8s %-8s %-20s %-18s %-10s %-12s %-12s",
		"NAME", "TYPE", "ADMIN", "OPER", "IP ADDRESSES", "MAC ADDRESS", "SPEED", "RX BYTES", "TX BYTES")
	headerRendered := theme.StyleTableHeader.Render(header)

	var rows []string
	for i, iface := range data.Interfaces {
		adminBadge := theme.StyleBadgeOnline.Render("UP")
		if iface.AdminStatus != model.AdminStatusUp {
			adminBadge = theme.StyleBadgeOffline.Render("DOWN")
		}

		operBadge := theme.StyleBadgeOnline.Render("UP")
		if iface.OperStatus != model.OperStatusUp {
			operBadge = theme.StyleBadgeOffline.Render(strings.ToUpper(string(iface.OperStatus)))
		}

		ips := strings.Join(iface.IPv4Addresses, ", ")
		if ips == "" && len(iface.IPv6Addresses) > 0 {
			ips = strings.Join(iface.IPv6Addresses, ", ")
		}
		if ips == "" {
			ips = "-"
		}
		if len(ips) > 20 {
			ips = ips[:17] + "..."
		}

		mac := iface.MACAddress
		if mac == "" {
			mac = "-"
		}

		speedStr := "-"
		if iface.SpeedBps > 0 {
			speedStr = formatSpeed(iface.SpeedBps)
		}

		rxBytesStr := "-"
		txBytesStr := "-"
		if stat, ok := data.Stats[iface.Name]; ok {
			rxBytesStr = formatBytes(stat.RxBytes)
			txBytesStr = formatBytes(stat.TxBytes)
		}

		line := fmt.Sprintf("%-12s %-10s %-8s %-8s %-20s %-18s %-10s %-12s %-12s",
			iface.Name,
			iface.Type,
			adminBadge,
			operBadge,
			ips,
			mac,
			speedStr,
			rxBytesStr,
			txBytesStr,
		)

		if i%2 == 0 {
			rows = append(rows, theme.StyleTableRow.Render(line))
		} else {
			rows = append(rows, theme.StyleTableRowAlt.Render(line))
		}
	}

	content := headerRendered + "\n" + strings.Join(rows, "\n")
	return theme.StyleCard.Width(width - 4).Render(content)
}

func formatSpeed(bps uint64) string {
	if bps >= 10000000000 {
		return fmt.Sprintf("%dG", bps/1000000000)
	}
	if bps >= 1000000000 {
		return fmt.Sprintf("%dG", bps/1000000000)
	}
	if bps >= 1000000 {
		return fmt.Sprintf("%dM", bps/1000000)
	}
	return fmt.Sprintf("%d bps", bps)
}
