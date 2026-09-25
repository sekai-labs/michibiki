package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/ipam"
)

type IPAMData struct {
	Subnets       []ipam.SubnetUsage
	SelectedIndex int
	Error         error
}

func RenderIPAM(data IPAMData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load IPAM data: " + data.Error.Error()),
		)
	}

	if len(data.Subnets) == 0 {
		return theme.StyleCard.Width(width - 4).Render(
			theme.StyleMuted.Render("No subnets discovered or configured."),
		)
	}

	leftPaneWidth := width / 3
	if leftPaneWidth < 30 {
		leftPaneWidth = 30
	}
	rightPaneWidth := width - leftPaneWidth - 6
	if rightPaneWidth < 30 {
		rightPaneWidth = 30
	}

	leftHeader := theme.StyleTitle.Render("SUBNETS & VLANS") + "\n\n"
	var leftRows []string
	for i, sub := range data.Subnets {
		cidrStr := sub.CIDR.String()
		ifaceLabel := sub.InterfaceName
		if sub.VLANID > 0 {
			ifaceLabel = fmt.Sprintf("VLAN %d", sub.VLANID)
		}

		bar := theme.RenderProgressBar(sub.UtilizationPct, 10)
		rowText := fmt.Sprintf("%-16s %-10s\n%s %5.1f%% (%d/%d)",
			cidrStr, ifaceLabel, bar, sub.UtilizationPct, sub.UsedIPs, sub.UsableIPs)

		if i == data.SelectedIndex {
			leftRows = append(leftRows, theme.StyleCardActive.Width(leftPaneWidth-4).Render(rowText))
		} else {
			leftRows = append(leftRows, theme.StyleCard.Width(leftPaneWidth-4).Render(rowText))
		}
	}
	leftContent := leftHeader + strings.Join(leftRows, "\n")
	leftPane := lipgloss.NewStyle().Width(leftPaneWidth).Render(leftContent)

	selIdx := data.SelectedIndex
	if selIdx < 0 || selIdx >= len(data.Subnets) {
		selIdx = 0
	}
	currentSub := data.Subnets[selIdx]

	statsTitle := theme.StyleTitle.Render("SUBNET DETAILS: " + currentSub.CIDR.String())
	statsBody := fmt.Sprintf("%s %s    %s %s\n%s %s    %s %s\n%s %d    %s %d    %s %d    %s %5.1f%%",
		theme.StyleSubTitle.Render("Interface:   "), currentSub.InterfaceName,
		theme.StyleSubTitle.Render("Gateway:     "), currentSub.GatewayIP.String(),
		theme.StyleSubTitle.Render("Network:     "), currentSub.NetworkIP.String(),
		theme.StyleSubTitle.Render("Broadcast:   "), currentSub.BroadcastIP.String(),
		theme.StyleSubTitle.Render("Total IPs:   "), currentSub.TotalIPs,
		theme.StyleSubTitle.Render("Usable IPs:  "), currentSub.UsableIPs,
		theme.StyleSubTitle.Render("Free IPs:    "), currentSub.FreeIPs,
		theme.StyleSubTitle.Render("Utilization: "), currentSub.UtilizationPct,
	)
	statsCard := theme.StyleCard.Width(rightPaneWidth).Render(statsTitle + "\n\n" + statsBody)

	freeIPsTitle := theme.StyleTitle.Render("NEXT AVAILABLE FREE IPS (READY FOR PROVISIONING)")
	nextFree := currentSub.FindNextFree(8)
	var freeIPTags []string
	if len(nextFree) == 0 {
		freeIPTags = append(freeIPTags, theme.StyleMuted.Render("Subnet fully allocated (0 free addresses)"))
	} else {
		for _, ip := range nextFree {
			freeIPTags = append(freeIPTags, theme.StyleFreeIP.Render("  "+ip.String()+"  "))
		}
	}
	freeIPsBody := strings.Join(freeIPTags, "  ")
	freeCard := theme.StyleCard.Width(rightPaneWidth).Render(freeIPsTitle + "\n\n" + freeIPsBody)

	allocTitle := theme.StyleTitle.Render(fmt.Sprintf("ALLOCATED IPS (%d)", len(currentSub.Allocations)))
	allocHeader := fmt.Sprintf("%-16s %-18s %-20s %-16s %-10s", "IP ADDRESS", "MAC ADDRESS", "HOSTNAME", "SOURCE", "STATUS")
	allocHeaderRendered := theme.StyleTableHeader.Render(allocHeader)

	var allocRows []string
	for i, a := range currentSub.Allocations {
		hname := a.Hostname
		if hname == "" {
			hname = "-"
		}
		if len(hname) > 20 {
			hname = hname[:17] + "..."
		}
		macStr := a.MAC
		if macStr == "" {
			macStr = "-"
		}

		line := fmt.Sprintf("%-16s %-18s %-20s %-16s %-10s",
			a.IP.String(),
			macStr,
			hname,
			string(a.Source),
			string(a.Status),
		)
		if i%2 == 0 {
			allocRows = append(allocRows, theme.StyleTableRow.Render(line))
		} else {
			allocRows = append(allocRows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(allocRows) == 0 {
		allocRows = append(allocRows, theme.StyleMuted.Render("No active allocations found in this subnet."))
	}

	allocCard := theme.StyleCard.Width(rightPaneWidth).Render(allocTitle + "\n\n" + allocHeaderRendered + "\n" + strings.Join(allocRows, "\n"))

	rightContent := lipgloss.JoinVertical(lipgloss.Left, statsCard, freeCard, allocCard)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightContent)
}
