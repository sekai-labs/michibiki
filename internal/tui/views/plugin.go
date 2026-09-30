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

	var pcb strings.Builder
	pcb.WriteString(theme.StyleTitle.Render("PLUGIN INFORMATION"))
	pcb.WriteString("  ")
	pcb.WriteString(badge)
	pcb.WriteString("\n\n")
	fmt.Fprintf(&pcb, "%s %s\n", theme.StyleSubTitle.Render("Plugin ID:     "), data.PluginName)
	if data.Version != "" {
		fmt.Fprintf(&pcb, "%s %s\n", theme.StyleSubTitle.Render("Version:       "), data.Version)
	}
	if data.BinaryPath != "" {
		fmt.Fprintf(&pcb, "%s %s\n", theme.StyleSubTitle.Render("Binary Path:   "), data.BinaryPath)
	}

	capsStr := "None declared"
	if len(data.Capabilities) > 0 {
		var rcb strings.Builder
		for i, capName := range data.Capabilities {
			if i > 0 {
				rcb.WriteByte(' ')
			}
			rcb.WriteString(theme.StyleBadgeInfo.Render(capName))
		}
		capsStr = rcb.String()
	}
	fmt.Fprintf(&pcb, "%s %s\n", theme.StyleSubTitle.Render("Capabilities:  "), capsStr)
	pluginCard := theme.StyleCard.Width(cardWidth).Render(pcb.String())

	var scb strings.Builder
	scb.WriteString(theme.StyleTitle.Render("HOST & RUNTIME TELEMETRY"))
	scb.WriteString("\n\n")
	if data.SystemInfo != nil {
		fmt.Fprintf(&scb, "%s %s\n", theme.StyleSubTitle.Render("Hostname:     "), data.SystemInfo.Hostname)
		fmt.Fprintf(&scb, "%s %s %s (%s)\n", theme.StyleSubTitle.Render("OS / Arch:    "), data.SystemInfo.OS, data.SystemInfo.Version, data.SystemInfo.Architecture)
		fmt.Fprintf(&scb, "%s %d cores\n", theme.StyleSubTitle.Render("CPU Cores:    "), data.SystemInfo.CPUCount)
		fmt.Fprintf(&scb, "%s %.1f%%\n", theme.StyleSubTitle.Render("CPU Load:     "), data.SystemInfo.CPUUsagePct)
	} else {
		scb.WriteString(theme.StyleMuted.Render("System telemetry awaiting initial sync..."))
	}
	sysCard := theme.StyleCard.Width(cardWidth).Render(scb.String())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, pluginCard, sysCard)

	var icb strings.Builder
	icb.WriteString(theme.StyleTitle.Render("DISCOVERED INTERFACES VIA PLUGIN"))
	icb.WriteString("\n\n")
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
		icb.WriteString(theme.StyleMuted.Render("No active interfaces returned by plugin provider."))
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
			fmt.Fprintf(&icb, "  • %-16s %s %-20s MAC: %s\n", iface.Name, statusBadge, ipStr, macStr)
		}
	}
	bottomCard := theme.StyleCard.Width(width - 4).Render(icb.String())

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomCard)
}
