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
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load VPN peers: " + data.Error.Error()),
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
			pubKey = pubKey[:10] + "..." + pubKey[len(pubKey)-6:]
		}

		endpoint := p.Endpoint
		if endpoint == "" {
			endpoint = "(none)"
		}

		allowed := strings.Join(p.AllowedIPs, ", ")
		if len(allowed) > 20 {
			allowed = allowed[:17] + "..."
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

	content := title + "\n\n" + headerRendered + "\n" + strings.Join(rows, "\n")
	return theme.StyleCard.Width(cardWidth).Render(content)
}
