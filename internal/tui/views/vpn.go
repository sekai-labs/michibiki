package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type VPNData struct {
	Peers []model.WireGuardPeer
	Error error
}

func RenderVPN(data VPNData, width, height int) string {
	if data.Error != nil {
		var eb strings.Builder
		eb.WriteString("Failed to load VPN peers: ")
		eb.WriteString(data.Error.Error())
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render(eb.String()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	title := theme.StyleTitle.Render(fmt.Sprintf("WIREGUARD PEERS (%d PEERS)", len(data.Peers)))
	header := fmt.Sprintf("%-24s %-20s %-20s %-16s %-12s %-12s",
		"PUBLIC KEY", "ENDPOINT", "ALLOWED IPS", "LAST HANDSHAKE", "TRANSFER RX", "TRANSFER TX")
	headerRendered := theme.StyleTableHeader.Render(header)

	var rows []string
	now := time.Now()
	for i, p := range data.Peers {
		pubKey := p.PublicKey
		if len(pubKey) > 20 {
			var pkb strings.Builder
			pkb.WriteString(pubKey[:10])
			pkb.WriteString("...")
			pkb.WriteString(pubKey[len(pubKey)-6:])
			pubKey = pkb.String()
		}

		endpoint := p.Endpoint
		if endpoint == "" {
			endpoint = "(none)"
		}

		var alb strings.Builder
		for idx, ip := range p.AllowedIPs {
			if idx > 0 {
				alb.WriteString(", ")
			}
			alb.WriteString(ip)
		}
		allowed := alb.String()
		if len(allowed) > 20 {
			var tb strings.Builder
			tb.WriteString(allowed[:17])
			tb.WriteString("...")
			allowed = tb.String()
		}

		handshakeStr := "never"
		if !p.LatestHandshake.IsZero() {
			ago := now.Sub(p.LatestHandshake).Round(time.Second)
			if ago < 3*time.Minute {
				handshakeStr = theme.StyleBadgeOnline.Render(fmt.Sprintf("%s ago", ago))
			} else {
				handshakeStr = fmt.Sprintf("%s ago", ago)
			}
		}

		line := fmt.Sprintf("%-24s %-20s %-20s %-16s %-12s %-12s",
			pubKey,
			endpoint,
			allowed,
			handshakeStr,
			formatBytes(p.TransferRxBytes),
			formatBytes(p.TransferTxBytes),
		)
		if i%2 == 0 {
			rows = append(rows, theme.StyleTableRow.Render(line))
		} else {
			rows = append(rows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, theme.StyleMuted.Render("No WireGuard peers configured."))
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(headerRendered)
	b.WriteByte('\n')
	for i, r := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(r)
	}
	return theme.StyleCard.Width(cardWidth).Render(b.String())
}
