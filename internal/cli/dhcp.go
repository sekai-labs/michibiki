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
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		overview, err := svc.GetServicesOverview(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list DHCP leases: %w", err)
		}
		leases := overview.DHCPLeases
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
