package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type DHCPData struct {
	Leases []model.DHCPLease
	Filter string
	Error  error
}

func RenderDHCP(data DHCPData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load DHCP leases: " + data.Error.Error()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	title := theme.StyleTitle.Render(fmt.Sprintf("DHCP ACTIVE LEASES (%d LEASES)", len(data.Leases)))
	header := fmt.Sprintf("%-16s %-18s %-22s %-16s %-10s %-10s %-14s",
		"IP ADDRESS", "MAC ADDRESS", "CLIENT HOSTNAME", "SUBNET", "IFACE", "STATE", "EXPIRES IN")
	headerRendered := theme.StyleTableHeader.Render(header)

	var rows []string
	now := time.Now()
	q := strings.ToLower(data.Filter)
	for i, l := range data.Leases {
		if q != "" && !strings.Contains(strings.ToLower(l.IPAddress), q) && !strings.Contains(strings.ToLower(l.MACAddress), q) && !strings.Contains(strings.ToLower(l.ClientHostname), q) && !strings.Contains(strings.ToLower(l.Interface), q) {
			continue
		}
		stateBadge := theme.StyleBadgeOnline.Render("ACTIVE")
		if l.State == model.DHCPStatic {
			stateBadge = theme.StyleBadgeInfo.Render("STATIC")
		} else if l.State == model.DHCPExpired {
			stateBadge = theme.StyleBadgeOffline.Render("EXPIRED")
		}

		expiresIn := "-"
		if l.State == model.DHCPStatic {
			expiresIn = "never"
		} else if !l.Ends.IsZero() {
			if l.Ends.Before(now) {
				expiresIn = "expired"
			} else {
				remaining := l.Ends.Sub(now).Round(time.Second)
				expiresIn = remaining.String()
			}
		}

		hname := l.ClientHostname
		if hname == "" {
			hname = "-"
		}
		if len(hname) > 22 {
			hname = hname[:19] + "..."
		}

		mac := l.MACAddress
		if mac == "" {
			mac = "-"
		}

		line := fmt.Sprintf("%-16s %-18s %-22s %-16s %-10s %-10s %-14s",
			l.IPAddress,
			mac,
			hname,
			l.SubnetCIDR,
			l.Interface,
			stateBadge,
			expiresIn,
		)
		if i%2 == 0 {
			rows = append(rows, theme.StyleTableRow.Render(line))
		} else {
			rows = append(rows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, theme.StyleMuted.Render("No DHCP leases found."))
	}

	content := title + "\n\n" + headerRendered + "\n" + strings.Join(rows, "\n")
	return theme.StyleCard.Width(cardWidth).Render(content)
}
