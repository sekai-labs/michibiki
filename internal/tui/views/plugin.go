package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type PluginViewData struct {
	PluginName   string
	Version      string
	Capabilities []string
	BinaryPath   string
	SystemInfo   *model.SystemInfo
	Interfaces   []model.Interface
	Stats        []model.InterfaceStats
	Filter       string
}

func RenderPluginCustomUI(data PluginViewData, width, height int) string {
	cardWidth := (width - 6) / 2
	if cardWidth < 30 {
		cardWidth = 30
	}

	badge := theme.StyleBadgeOnline.Render("EXTERNAL PLUGIN PROVIDER")

	pluginCardContent := fmt.Sprintf("%s  %s\n\n", theme.StyleTitle.Render("PLUGIN INFORMATION"), badge)
	pluginCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Plugin ID:     "), data.PluginName)
	if data.Version != "" {
		pluginCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Version:       "), data.Version)
	}
	if data.BinaryPath != "" {
		pluginCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Binary Path:   "), data.BinaryPath)
	}

	capsStr := "None declared"
	if len(data.Capabilities) > 0 {
		var renderedCaps []string
		for _, capName := range data.Capabilities {
			renderedCaps = append(renderedCaps, theme.StyleBadgeInfo.Render(capName))
		}
		capsStr = strings.Join(renderedCaps, " ")
	}
	pluginCardContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Capabilities:  "), capsStr)
	pluginCard := theme.StyleCard.Width(cardWidth).Render(pluginCardContent)

	sysContent := fmt.Sprintf("%s\n\n", theme.StyleTitle.Render("HOST & RUNTIME TELEMETRY"))
	if data.SystemInfo != nil {
		sysContent += fmt.Sprintf("%s %s\n", theme.StyleSubTitle.Render("Hostname:     "), data.SystemInfo.Hostname)
		sysContent += fmt.Sprintf("%s %s %s (%s)\n", theme.StyleSubTitle.Render("OS / Arch:    "), data.SystemInfo.OS, data.SystemInfo.Version, data.SystemInfo.Architecture)
		sysContent += fmt.Sprintf("%s %d cores\n", theme.StyleSubTitle.Render("CPU Cores:    "), data.SystemInfo.CPUCount)
		sysContent += fmt.Sprintf("%s %.1f%%\n", theme.StyleSubTitle.Render("CPU Load:     "), data.SystemInfo.CPUUsagePct)
	} else {
		sysContent += theme.StyleMuted.Render("System telemetry awaiting initial sync...")
	}
	sysCard := theme.StyleCard.Width(cardWidth).Render(sysContent)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, pluginCard, sysCard)

	ifaceContent := fmt.Sprintf("%s\n\n", theme.StyleTitle.Render("DISCOVERED INTERFACES VIA PLUGIN"))
	ifaces := data.Interfaces
	if data.Filter != "" {
		q := strings.ToLower(data.Filter)
		var matched []model.Interface
		for _, iface := range ifaces {
			if strings.Contains(strings.ToLower(iface.Name), q) || strings.Contains(strings.ToLower(iface.MACAddress), q) {
				matched = append(matched, iface)
			}
		}
		ifaces = matched
	}
	if len(ifaces) == 0 {
		ifaceContent += theme.StyleMuted.Render("No active interfaces returned by plugin provider.")
	} else {
		for _, iface := range ifaces {
			statusBadge := theme.StyleBadgeOffline.Render("DOWN")
			if iface.OperStatus == model.OperStatusUp {
				statusBadge = theme.StyleBadgeOnline.Render(" UP ")
			}
			ipStr := "-"
			if len(iface.IPv4Addresses) > 0 {
				ipStr = strings.Join(iface.IPv4Addresses, ", ")
			}
			macStr := iface.MACAddress
			if macStr == "" {
				macStr = "-"
			}
			ifaceContent += fmt.Sprintf("  • %-16s %s %-20s MAC: %s\n", iface.Name, statusBadge, ipStr, macStr)
		}
	}
	bottomCard := theme.StyleCard.Width(width - 4).Render(ifaceContent)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomCard)
}
