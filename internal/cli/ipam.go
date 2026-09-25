package cli

import (
	"fmt"
	"net/netip"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/pkg/ipam"
)

var (
	flagFreeCount int
)

var ipCmd = &cobra.Command{
	Use:     "ip",
	Aliases: []string{"ipam"},
	Short:   "Subnet IP allocation matrix and free IP finder (IPAM)",
}

var ipUsageCmd = &cobra.Command{
	Use:   "usage [interface|vlan|cidr]",
	Short: "Display subnet allocation matrix, utilization progress, and used IPs",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		usages, err := svc.GetIPAMAllocation(cmd.Context(), query)
		if err != nil {
			return fmt.Errorf("failed to retrieve subnet usage: %w", err)
		}

		if len(usages) == 0 {
			if query != "" {
				return fmt.Errorf("no subnet found matching query '%s'", query)
			}
			return fmt.Errorf("no active IPv4 subnets discovered on device")
		}

		return formatOutput(cmd, usages, func() {
			for idx, su := range usages {
				if idx > 0 {
					fmt.Println()
					fmt.Println(strings.Repeat("-", 60))
					fmt.Println()
				}

				progressBar := renderBar(su.UtilizationPct, 24)

				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintf(w, "SUBNET\t%s\n", su.CIDR.String())
				if su.InterfaceName != "" {
					fmt.Fprintf(w, "INTERFACE\t%s\n", su.InterfaceName)
				}
				if su.VLANID > 0 {
					fmt.Fprintf(w, "VLAN ID\t%d\n", su.VLANID)
				}
				if su.GatewayIP.IsValid() {
					fmt.Fprintf(w, "GATEWAY\t%s\n", su.GatewayIP.String())
				}
				fmt.Fprintf(w, "TOTAL IPS\t%d\n", su.TotalIPs)
				fmt.Fprintf(w, "USABLE IPS\t%d\n", su.UsableIPs)
				fmt.Fprintf(w, "USED IPS\t%d\n", su.UsedIPs)
				fmt.Fprintf(w, "FREE IPS\t%d\n", su.FreeIPs)
				fmt.Fprintf(w, "UTILIZATION\t%.1f%%  [%s]\n", su.UtilizationPct, progressBar)
				w.Flush()

				if len(su.FreeRanges) > 0 {
					fmt.Println()
					fmt.Printf("FREE IP BLOCKS (%d blocks, %d free addresses):\n", len(su.FreeRanges), su.FreeIPs)
					for _, r := range su.FreeRanges {
						fmt.Printf("  • %s - %s (%d hosts)\n", r.Start.String(), r.End.String(), r.Count)
					}
				}

				nextFree := su.FindNextFree(3)
				if len(nextFree) > 0 {
					var nfStrs []string
					for _, a := range nextFree {
						nfStrs = append(nfStrs, a.String())
					}
					fmt.Println()
					fmt.Printf("NEXT AVAILABLE STATIC IPS: %s\n", strings.Join(nfStrs, ", "))
				}

				if len(su.Allocations) > 0 {
					fmt.Println()
					fmt.Println("ALLOCATED IP ADDRESSES:")
					tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
					fmt.Fprintf(tw, "IP ADDRESS\tMAC ADDRESS\tHOSTNAME\tSOURCE\tSTATUS\tEXPIRES\n")
					now := time.Now()
					for _, a := range su.Allocations {
						mac := a.MAC
						if mac == "" {
							mac = "-"
						}
						host := a.Hostname
						if host == "" {
							host = "-"
						}
						expStr := "-"
						if !a.ExpiresAt.IsZero() {
							if a.ExpiresAt.Before(now) {
								expStr = "expired"
							} else {
								expStr = a.ExpiresAt.Sub(now).Round(time.Second).String()
							}
						}
						fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
							a.IP.String(),
							mac,
							host,
							a.Source,
							a.Status,
							expStr,
						)
					}
					tw.Flush()
				}
			}
		})
	},
}

var ipFreeCmd = &cobra.Command{
	Use:   "free [interface|vlan|cidr]",
	Short: "Find next unallocated static IP addresses and free IP ranges",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		usages, err := prov.GetSubnetUsage(cmd.Context(), query)
		if err != nil {
			return fmt.Errorf("failed to retrieve subnet usage: %w", err)
		}

		if len(usages) == 0 {
			if query != "" {
				return fmt.Errorf("no subnet found matching query '%s'", query)
			}
			return fmt.Errorf("no active IPv4 subnets discovered")
		}

		type FreeReport struct {
			Subnet     string        `json:"subnet" yaml:"subnet"`
			Interface  string        `json:"interface,omitempty" yaml:"interface,omitempty"`
			VLANID     int           `json:"vlan_id,omitempty" yaml:"vlan_id,omitempty"`
			FreeCount  int           `json:"free_count" yaml:"free_count"`
			NextFree   []string      `json:"next_free" yaml:"next_free"`
			FreeRanges []ipam.IPRange `json:"free_ranges" yaml:"free_ranges"`
		}

		var reports []FreeReport
		for _, su := range usages {
			addrs := su.FindNextFree(flagFreeCount)
			var addrStrs []string
			for _, a := range addrs {
				addrStrs = append(addrStrs, a.String())
			}
			reports = append(reports, FreeReport{
				Subnet:     su.CIDR.String(),
				Interface:  su.InterfaceName,
				VLANID:     su.VLANID,
				FreeCount:  su.FreeIPs,
				NextFree:   addrStrs,
				FreeRanges: su.FreeRanges,
			})
		}

		return formatOutput(cmd, reports, func() {
			for _, r := range reports {
				fmt.Printf("Subnet: %s", r.Subnet)
				if r.Interface != "" {
					fmt.Printf(" (Interface: %s)", r.Interface)
				}
				if r.VLANID > 0 {
					fmt.Printf(" (VLAN: %d)", r.VLANID)
				}
				fmt.Printf(" — %d Total Free IPs\n", r.FreeCount)

				if len(r.NextFree) > 0 {
					fmt.Printf("  Next %d Free IPs: %s\n", len(r.NextFree), strings.Join(r.NextFree, ", "))
				} else {
					fmt.Println("  Subnet is fully exhausted! No free static IPs available.")
				}

				if len(r.FreeRanges) > 0 {
					fmt.Println("  Available Free Ranges:")
					for _, fr := range r.FreeRanges {
						fmt.Printf("    • %s to %s (%d hosts)\n", fr.Start.String(), fr.End.String(), fr.Count)
					}
				}
				fmt.Println()
			}
		})
	},
}

var ipCheckCmd = &cobra.Command{
	Use:   "check <ip> [subnet-or-interface]",
	Short: "Check whether a candidate IP address is free or in conflict",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		candStr := args[0]
		candIP, err := netip.ParseAddr(candStr)
		if err != nil {
			return fmt.Errorf("invalid IP address '%s': %w", candStr, err)
		}

		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		query := ""
		if len(args) > 1 {
			query = args[1]
		}

		usages, err := prov.GetSubnetUsage(cmd.Context(), query)
		if err != nil {
			return fmt.Errorf("failed to retrieve subnet usage: %w", err)
		}

		var targetSubnet *ipam.SubnetUsage
		for i := range usages {
			if usages[i].CIDR.Contains(candIP) {
				targetSubnet = &usages[i]
				break
			}
		}

		if targetSubnet == nil {
			if len(usages) == 1 {
				targetSubnet = &usages[0]
			} else {
				return fmt.Errorf("ip %s does not belong to any discovered subnet", candStr)
			}
		}

		available, reason := targetSubnet.ValidateIPAvailable(candIP)

		result := map[string]any{
			"candidate_ip": candStr,
			"subnet":       targetSubnet.CIDR.String(),
			"available":    available,
			"reason":       reason,
		}

		return formatOutput(cmd, result, func() {
			if available {
				fmt.Printf("SUCCESS: IP %s is AVAILABLE for provisioning in subnet %s\n", candStr, targetSubnet.CIDR.String())
				if reason != "" {
					fmt.Printf("Details: %s\n", reason)
				}
			} else {
				fmt.Printf("CONFLICT: IP %s is NOT AVAILABLE in subnet %s\n", candStr, targetSubnet.CIDR.String())
				fmt.Printf("Reason: %s\n", reason)
			}
		})
	},
}

func renderBar(pct float64, width int) string {
	if width < 5 {
		width = 10
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filledLen := int((pct / 100.0) * float64(width))
	if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	var b strings.Builder
	for i := 0; i < filledLen; i++ {
		b.WriteString("█")
	}
	for i := 0; i < emptyLen; i++ {
		b.WriteString("░")
	}
	return b.String()
}

func init() {
	ipFreeCmd.Flags().IntVarP(&flagFreeCount, "count", "n", 5, "Number of next free static IPs to find")
	ipCmd.AddCommand(ipUsageCmd)
	ipCmd.AddCommand(ipFreeCmd)
	ipCmd.AddCommand(ipCheckCmd)
}
