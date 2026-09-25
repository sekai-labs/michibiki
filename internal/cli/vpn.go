package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var vpnCmd = &cobra.Command{
	Use:   "vpn",
	Short: "VPN tunnels and WireGuard peers",
}

var vpnPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "List WireGuard peers and tunnel status",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		peers, err := prov.ListWireGuardPeers(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list WireGuard peers: %w", err)
		}

		return formatOutput(cmd, peers, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "PUBLIC KEY\tENDPOINT\tALLOWED IPS\tLATEST HANDSHAKE\tRX\tTX\n")
			now := time.Now()
			for _, p := range peers {
				pubKey := p.PublicKey
				if len(pubKey) > 16 {
					pubKey = pubKey[:16] + "..."
				}
				endpoint := p.Endpoint
				if endpoint == "" {
					endpoint = "(none)"
				}
				allowedIPs := strings.Join(p.AllowedIPs, ", ")
				handshakeStr := "never"
				if !p.LatestHandshake.IsZero() {
					dur := now.Sub(p.LatestHandshake).Round(time.Second)
					handshakeStr = fmt.Sprintf("%s ago", dur)
				}
				rxStr := formatBytes(p.TransferRxBytes)
				txStr := formatBytes(p.TransferTxBytes)

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					pubKey,
					endpoint,
					allowedIPs,
					handshakeStr,
					rxStr,
					txStr,
				)
			}
			w.Flush()
		})
	},
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

func init() {
	vpnCmd.AddCommand(vpnPeersCmd)
}
