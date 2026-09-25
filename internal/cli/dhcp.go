package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var dhcpCmd = &cobra.Command{
	Use:   "dhcp",
	Short: "DHCP server management and lease tracking",
}

var dhcpLeaseCmd = &cobra.Command{
	Use:   "lease",
	Short: "DHCP leases",
}

var dhcpLeaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active and static DHCP leases",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		leases, err := prov.ListDHCPLeases(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list DHCP leases: %w", err)
		}

		return formatOutput(cmd, leases, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "IP ADDRESS\tMAC ADDRESS\tHOSTNAME\tSTATE\tINTERFACE\tEXPIRES\n")
			now := time.Now()
			for _, l := range leases {
				host := l.ClientHostname
				if host == "" {
					host = "-"
				}
				iface := l.Interface
				if iface == "" {
					iface = "-"
				}
				expiresStr := "-"
				if !l.Ends.IsZero() {
					if l.Ends.Before(now) {
						expiresStr = "expired"
					} else {
						dur := l.Ends.Sub(now).Round(time.Second)
						expiresStr = dur.String()
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					l.IPAddress,
					l.MACAddress,
					host,
					l.State,
					iface,
					expiresStr,
				)
			}
			w.Flush()
		})
	},
}

func init() {
	dhcpLeaseCmd.AddCommand(dhcpLeaseListCmd)
	dhcpCmd.AddCommand(dhcpLeaseCmd)
}
