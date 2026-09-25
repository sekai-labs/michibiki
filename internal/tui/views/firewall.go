package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type FirewallData struct {
	Rules   []model.FirewallRule
	NAT     []model.NATRule
	Aliases []model.FirewallAlias
	Error   error
}

func RenderFirewall(data FirewallData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load firewall data: " + data.Error.Error()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	rulesTitle := theme.StyleTitle.Render(fmt.Sprintf("FIREWALL FILTER RULES (%d RULES)", len(data.Rules)))
	rulesHeader := fmt.Sprintf("%-6s %-8s %-6s %-8s %-6s %-18s %-18s %-10s %-12s %-20s",
		"SEQ", "IFACE", "DIR", "ACTION", "PROTO", "SOURCE", "DESTINATION", "PACKETS", "BYTES", "DESCRIPTION")
	rulesHeaderRendered := theme.StyleTableHeader.Render(rulesHeader)

	var ruleRows []string
	for i, r := range data.Rules {
		actionBadge := theme.StyleBadgeOnline.Render("PASS")
		if r.Action == model.FirewallBlock {
			actionBadge = theme.StyleBadgeOffline.Render("BLOCK")
		} else if r.Action == model.FirewallReject {
			actionBadge = theme.StyleBadgeWarning.Render("REJECT")
		}

		src := r.Source
		if r.SourcePort != "" {
			src += ":" + r.SourcePort
		}
		if src == "" {
			src = "any"
		}
		if len(src) > 18 {
			src = src[:15] + "..."
		}

		dst := r.Destination
		if r.DestinationPort != "" {
			dst += ":" + r.DestinationPort
		}
		if dst == "" {
			dst = "any"
		}
		if len(dst) > 18 {
			dst = dst[:15] + "..."
		}

		desc := r.Description
		if len(desc) > 20 {
			desc = desc[:17] + "..."
		}

		line := fmt.Sprintf("%-6d %-8s %-6s %-8s %-6s %-18s %-18s %-10s %-12s %-20s",
			r.Sequence,
			r.Interface,
			r.Direction,
			actionBadge,
			r.Protocol,
			src,
			dst,
			fmt.Sprintf("%d", r.Packets),
			formatBytes(r.Bytes),
			desc,
		)
		if i%2 == 0 {
			ruleRows = append(ruleRows, theme.StyleTableRow.Render(line))
		} else {
			ruleRows = append(ruleRows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(ruleRows) == 0 {
		ruleRows = append(ruleRows, theme.StyleMuted.Render("No firewall filter rules found."))
	}
	rulesCard := theme.StyleCard.Width(cardWidth).Render(rulesTitle + "\n\n" + rulesHeaderRendered + "\n" + strings.Join(ruleRows, "\n"))

	var natCard string
	if len(data.NAT) > 0 {
		natTitle := theme.StyleTitle.Render(fmt.Sprintf("NAT RULES (%d RULES)", len(data.NAT)))
		natHeader := fmt.Sprintf("%-10s %-8s %-6s %-18s %-18s %-18s %-8s",
			"TYPE", "IFACE", "PROTO", "SOURCE", "DESTINATION", "TARGET", "STATUS")
		natHeaderRendered := theme.StyleTableHeader.Render(natHeader)

		var natRows []string
		for i, n := range data.NAT {
			statusBadge := theme.StyleBadgeOnline.Render("ON")
			if !n.Enabled {
				statusBadge = theme.StyleBadgeOffline.Render("OFF")
			}

			targetStr := n.Target
			if n.TargetPort != "" {
				targetStr += ":" + n.TargetPort
			}

			line := fmt.Sprintf("%-10s %-8s %-6s %-18s %-18s %-18s %-8s",
				n.Type,
				n.Interface,
				n.Protocol,
				n.Source,
				n.Destination,
				targetStr,
				statusBadge,
			)
			if i%2 == 0 {
				natRows = append(natRows, theme.StyleTableRow.Render(line))
			} else {
				natRows = append(natRows, theme.StyleTableRowAlt.Render(line))
			}
		}
		natCard = theme.StyleCard.Width(cardWidth).Render(natTitle + "\n\n" + natHeaderRendered + "\n" + strings.Join(natRows, "\n"))
	}

	if natCard != "" {
		return lipgloss.JoinVertical(lipgloss.Left, rulesCard, natCard)
	}
	return rulesCard
}
