package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type DashboardData struct {
	SystemInfo *model.SystemInfo
	Interfaces []model.Interface
	Gateways   []model.Gateway
	Error      error
}

func RenderDashboard(data DashboardData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load dashboard: " + data.Error.Error()),
		)
	}

	sys := data.SystemInfo
	if sys == nil {
		return theme.StyleCard.Width(width - 4).Render(
			theme.StyleMuted.Render("Loading dashboard information..."),
		)
	}

	cardWidth := (width - 6) / 2
	if cardWidth < 30 {
		cardWidth = 30
	}

	var alerts []string
	for _, gw := range data.Gateways {
		if gw.Status != model.GatewayOnline {
			alerts = append(alerts, fmt.Sprintf("Gateway %s (%s) is %s", gw.Name, gw.Address, string(gw.Status)))
		}
	}
	downIfaces := 0
	for _, iface := range data.Interfaces {
		if iface.OperStatus == model.OperStatusDown && iface.AdminStatus == model.AdminStatusUp {
			downIfaces++
		}
	}
	if downIfaces > 0 {
		alerts = append(alerts, fmt.Sprintf("%d interface(s) admin up but link is down", downIfaces))
	}

	var alertBanner string
	if len(alerts) > 0 {
		alertContent := theme.StyleError.Render("ALERTS:") + "\n"
		for _, a := range alerts {
			alertContent += " • " + a + "\n"
		}
		alertBanner = theme.StyleCardAlert.Width(width - 4).Render(strings.TrimSpace(alertContent))
	}

	uptimeStr := formatUptime(sys.UptimeSeconds)
	deviceCardContent := fmt.Sprintf("%s\n\n", theme.StyleTitle.Render("SYSTEM OVERVIEW"))
	deviceCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Hostname:     "), sys.Hostname)
	deviceCardContent += fmt.Sprintf("%s %s %s (%s)\n", theme.StyleSubTitle.Render("OS / Version: "), sys.OS, sys.Version, sys.Architecture)
	deviceCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Uptime:       "), uptimeStr)
	deviceCardContent += fmt.Sprintf("%s %d cores\n", theme.StyleSubTitle.Render("CPU Cores:    "), sys.CPUCount)
	if sys.SerialNumber != "" {
		deviceCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Serial Number:"), sys.SerialNumber)
	}
	deviceCard := theme.StyleCard.Width(cardWidth).Render(deviceCardContent)

	gaugeWidth := cardWidth - 18
	if gaugeWidth < 8 {
		gaugeWidth = 8
	}

	cpuGauge := theme.RenderProgressBar(sys.CPUUsagePct, gaugeWidth)
	var memPct float64
	if sys.MemoryTotalBytes > 0 {
		memPct = (float64(sys.MemoryUsedBytes) / float64(sys.MemoryTotalBytes)) * 100.0
	}
	memGauge := theme.RenderProgressBar(memPct, gaugeWidth)

	var diskPct float64
	if sys.StorageTotal > 0 {
		diskPct = (float64(sys.StorageUsed) / float64(sys.StorageTotal)) * 100.0
	}
	diskGauge := theme.RenderProgressBar(diskPct, gaugeWidth)

	resourcesContent := fmt.Sprintf("%s\n\n", theme.StyleTitle.Render("RESOURCE UTILIZATION"))
	resourcesContent += fmt.Sprintf("%s %s %5.1f%%\n", theme.StyleSubTitle.Render("CPU Usage:"), cpuGauge, sys.CPUUsagePct)
	resourcesContent += fmt.Sprintf("%s %s %5.1f%% (%s / %s)\n",
		theme.StyleSubTitle.Render("Memory:   "),
		memGauge,
		memPct,
		formatBytes(sys.MemoryUsedBytes),
		formatBytes(sys.MemoryTotalBytes),
	)
	resourcesContent += fmt.Sprintf("%s %s %5.1f%% (%s / %s)\n",
		theme.StyleSubTitle.Render("Storage:  "),
		diskGauge,
		diskPct,
		formatBytes(sys.StorageUsed),
		formatBytes(sys.StorageTotal),
	)
	resourcesCard := theme.StyleCard.Width(cardWidth).Render(resourcesContent)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, deviceCard, resourcesCard)

	ifaceSummaryContent := fmt.Sprintf("%s\n\n", theme.StyleTitle.Render("INTERFACES & GATEWAYS SUMMARY"))
	upCount := 0
	for _, iface := range data.Interfaces {
		if iface.OperStatus == model.OperStatusUp {
			upCount++
		}
	}
	ifaceSummaryContent += fmt.Sprintf("Interfaces: %d total, %s online, %s offline\n",
		len(data.Interfaces),
		theme.StyleBadgeOnline.Render(fmt.Sprintf("%d UP", upCount)),
		theme.StyleBadgeOffline.Render(fmt.Sprintf("%d DOWN", len(data.Interfaces)-upCount)),
	)
	ifaceSummaryContent += "\nGateways:\n"
	if len(data.Gateways) == 0 {
		ifaceSummaryContent += "  No gateways reported\n"
	}
	for _, gw := range data.Gateways {
		badge := theme.StyleBadgeOnline.Render("ONLINE")
		if gw.Status != model.GatewayOnline {
			badge = theme.StyleBadgeOffline.Render(strings.ToUpper(string(gw.Status)))
		}
		defStr := ""
		if gw.IsDefault {
			defStr = " [Default]"
		}
		ifaceSummaryContent += fmt.Sprintf("  • %-16s %-16s %s %6.1fms latency %4.1f%% loss%s\n",
			gw.Name,
			gw.Address,
			badge,
			gw.LatencyMs,
			gw.PacketLossPct,
			defStr,
		)
	}
	bottomCard := theme.StyleCard.Width(width - 4).Render(ifaceSummaryContent)

	var fullView string
	if alertBanner != "" {
		fullView = lipgloss.JoinVertical(lipgloss.Left, alertBanner, topRow, bottomCard)
	} else {
		fullView = lipgloss.JoinVertical(lipgloss.Left, topRow, bottomCard)
	}

	return fullView
}

func formatUptime(seconds uint64) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %dm", d, h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
