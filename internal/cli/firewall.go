package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var firewallCmd = &cobra.Command{
	Use:     "firewall",
	Aliases: []string{"fw"},
	Short:   "Firewall rules, aliases, and filter policies",
}

var firewallRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Firewall filter rules",
}

var firewallRuleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all firewall filter rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		overview, err := svc.GetSecurityOverview(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list firewall rules: %w", err)
		}
		rules := overview.FirewallRules
		return formatOutput(cmd, rules, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "SEQ\tIFACE\tDIR\tACTION\tPROTO\tSOURCE\tPORT\tDESTINATION\tPORT\tPACKETS\tBYTES\tDESCRIPTION\n")
			for _, r := range rules {
				dir := string(r.Direction)
				if dir == "" {
					dir = "in"
				}
				proto := r.Protocol
				if proto == "" {
					proto = "any"
				}
				src := r.Source
				if src == "" {
					src = "any"
				}
				srcPort := r.SourcePort
				if srcPort == "" {
					srcPort = "*"
				}
				dst := r.Destination
				if dst == "" {
					dst = "any"
				}
				dstPort := r.DestinationPort
				if dstPort == "" {
					dstPort = "*"
				}
				desc := r.Description
				if desc == "" {
					desc = "-"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
					r.Sequence,
					r.Interface,
					dir,
					r.Action,
					proto,
					src,
					srcPort,
					dst,
					dstPort,
					r.Packets,
					r.Bytes,
					desc,
				)
			}
			w.Flush()
		})
	},
}

var firewallAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Firewall aliases and address groups",
}

var firewallAliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all firewall aliases",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		overview, err := svc.GetSecurityOverview(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list firewall aliases: %w", err)
		}
		aliases := overview.Aliases
		return formatOutput(cmd, aliases, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tCONTENT\tDESCRIPTION\n")
			for _, a := range aliases {
				contentStr := strings.Join(a.Content, ", ")
				desc := a.Description
				if desc == "" {
					desc = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					a.Name,
					a.Type,
					contentStr,
					desc,
				)
			}
			w.Flush()
		})
	},
}

var natCmd = &cobra.Command{
	Use:   "nat",
	Short: "Network address translation (NAT) rules",
}

var natListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all NAT rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		overview, err := svc.GetSecurityOverview(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list NAT rules: %w", err)
		}
		nats := overview.NATRules
		return formatOutput(cmd, nats, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tTYPE\tIFACE\tPROTO\tSOURCE\tDESTINATION\tTARGET\tTARGET PORT\tENABLED\n")
			for _, n := range nats {
				iface := n.Interface
				if iface == "" {
					iface = "-"
				}
				proto := n.Protocol
				if proto == "" {
					proto = "any"
				}
				src := n.Source
				if src == "" {
					src = "any"
				}
				dst := n.Destination
				if dst == "" {
					dst = "any"
				}
				target := n.Target
				if target == "" {
					target = "-"
				}
				targetPort := n.TargetPort
				if targetPort == "" {
					targetPort = "-"
				}
				enabledStr := "yes"
				if !n.Enabled {
					enabledStr = "no"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					n.ID,
					n.Type,
					iface,
					proto,
					src,
					dst,
					target,
					targetPort,
					enabledStr,
				)
			}
			w.Flush()
		})
	},
}

func init() {
	firewallRuleCmd.AddCommand(firewallRuleListCmd)
	firewallAliasCmd.AddCommand(firewallAliasListCmd)
	firewallCmd.AddCommand(firewallRuleCmd)
	firewallCmd.AddCommand(firewallAliasCmd)
	natCmd.AddCommand(natListCmd)
}
