package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var routeCmd = &cobra.Command{
	Use:     "route",
	Aliases: []string{"routes"},
	Short:   "Routing table and static routes",
}

var routeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all routes",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		routes, err := prov.ListRoutes(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list routes: %w", err)
		}

		return formatOutput(cmd, routes, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "DESTINATION\tGATEWAY\tINTERFACE\tPROTOCOL\tMETRIC\tSCOPE\n")
			for _, r := range routes {
				gw := r.Gateway
				if gw == "" {
					gw = "*"
				}
				iface := r.Interface
				if iface == "" {
					iface = "-"
				}
				scope := r.Scope
				if scope == "" {
					scope = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
					r.Destination,
					gw,
					iface,
					r.Protocol,
					r.Metric,
					scope,
				)
			}
			w.Flush()
		})
	},
}

var gatewayCmd = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gateways", "gw"},
	Short:   "Default and custom gateways",
}

var gatewayListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all network gateways",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		gws, err := prov.ListGateways(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list gateways: %w", err)
		}

		return formatOutput(cmd, gws, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tADDRESS\tINTERFACE\tSTATUS\tDEFAULT\tLATENCY\tLOSS\n")
			for _, g := range gws {
				defStr := "no"
				if g.IsDefault {
					defStr = "yes"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.2f ms\t%.1f%%\n",
					g.Name,
					g.Address,
					g.Interface,
					g.Status,
					defStr,
					g.LatencyMs,
					g.PacketLossPct,
				)
			}
			w.Flush()
		})
	},
}

var arpCmd = &cobra.Command{
	Use:     "arp",
	Aliases: []string{"ndp"},
	Short:   "Address resolution protocol (ARP / NDP) cache",
}

var arpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ARP entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		arps, err := prov.ListARPEntries(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list ARP entries: %w", err)
		}

		return formatOutput(cmd, arps, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "IP ADDRESS\tMAC ADDRESS\tINTERFACE\tHOSTNAME\tSTATUS\n")
			for _, a := range arps {
				host := a.Hostname
				if host == "" {
					host = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					a.IPAddress,
					a.MACAddress,
					a.Interface,
					host,
					a.Status,
				)
			}
			w.Flush()
		})
	},
}

func init() {
	routeCmd.AddCommand(routeListCmd)
	gatewayCmd.AddCommand(gatewayListCmd)
	arpCmd.AddCommand(arpListCmd)
}
