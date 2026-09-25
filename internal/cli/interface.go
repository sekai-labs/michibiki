package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var interfaceCmd = &cobra.Command{
	Use:     "interface",
	Aliases: []string{"interfaces", "iface", "if"},
	Short:   "Network interfaces and VLANs",
}

var interfaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all network interfaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		ifaces, err := prov.ListInterfaces(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list interfaces: %w", err)
		}

		return formatOutput(cmd, ifaces, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tADMIN\tOPER\tMAC\tIPV4\tIPV6\tSPEED\n")
			for _, iface := range ifaces {
				ipv4Str := strings.Join(iface.IPv4Addresses, ", ")
				if ipv4Str == "" {
					ipv4Str = "-"
				}
				ipv6Str := strings.Join(iface.IPv6Addresses, ", ")
				if ipv6Str == "" {
					ipv6Str = "-"
				}
				macStr := iface.MACAddress
				if macStr == "" {
					macStr = "-"
				}
				speedStr := "-"
				if iface.SpeedBps > 0 {
					if iface.SpeedBps >= 1_000_000_000 {
						speedStr = fmt.Sprintf("%d Gbps", iface.SpeedBps/1_000_000_000)
					} else if iface.SpeedBps >= 1_000_000 {
						speedStr = fmt.Sprintf("%d Mbps", iface.SpeedBps/1_000_000)
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					iface.Name,
					iface.Type,
					iface.AdminStatus,
					iface.OperStatus,
					macStr,
					ipv4Str,
					ipv6Str,
					speedStr,
				)
			}
			w.Flush()
		})
	},
}

var interfaceShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show detailed information for a single interface",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		iface, err := prov.GetInterface(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to get interface %s: %w", name, err)
		}

		return formatOutput(cmd, iface, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "PROPERTY\tVALUE\n")
			fmt.Fprintf(w, "ID\t%s\n", iface.ID)
			fmt.Fprintf(w, "Name\t%s\n", iface.Name)
			fmt.Fprintf(w, "Type\t%s\n", iface.Type)
			fmt.Fprintf(w, "Admin Status\t%s\n", iface.AdminStatus)
			fmt.Fprintf(w, "Operational Status\t%s\n", iface.OperStatus)
			fmt.Fprintf(w, "MAC Address\t%s\n", iface.MACAddress)
			fmt.Fprintf(w, "MTU\t%d\n", iface.MTU)
			fmt.Fprintf(w, "IPv4 Addresses\t%s\n", strings.Join(iface.IPv4Addresses, ", "))
			fmt.Fprintf(w, "IPv6 Addresses\t%s\n", strings.Join(iface.IPv6Addresses, ", "))
			if iface.VLANID > 0 {
				fmt.Fprintf(w, "VLAN ID\t%d\n", iface.VLANID)
			}
			if iface.ParentInterface != "" {
				fmt.Fprintf(w, "Parent Interface\t%s\n", iface.ParentInterface)
			}
			if iface.SpeedBps > 0 {
				fmt.Fprintf(w, "Speed (bps)\t%d\n", iface.SpeedBps)
			}
			if iface.Duplex != "" {
				fmt.Fprintf(w, "Duplex\t%s\n", iface.Duplex)
			}
			w.Flush()
		})
	},
}

func init() {
	interfaceCmd.AddCommand(interfaceListCmd)
	interfaceCmd.AddCommand(interfaceShowCmd)
}
