package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "Border Gateway Protocol (BGP) status and peering",
}

var bgpNeighborCmd = &cobra.Command{
	Use:   "neighbor",
	Short: "BGP neighbors",
}

var bgpNeighborListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all BGP peers and neighbor sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		neighbors, err := prov.ListBGPNeighbors(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list BGP neighbors: %w", err)
		}

		return formatOutput(cmd, neighbors, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "PEER ADDRESS\tREMOTE AS\tLOCAL AS\tSTATE\tUPTIME\tRCVD\tACCEPTED\tDESCRIPTION\n")
			for _, n := range neighbors {
				desc := n.Description
				if desc == "" {
					desc = "-"
				}
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%d\t%d\t%s\n",
					n.PeerAddress,
					n.RemoteAS,
					n.LocalAS,
					n.State,
					(time.Duration(n.UptimeSeconds) * time.Second).String(),
					n.PrefixesReceived,
					n.PrefixesAccepted,
					desc,
				)
			}
			w.Flush()
		})
	},
}

var bgpSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Display BGP session summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		neighbors, err := prov.ListBGPNeighbors(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list BGP neighbors: %w", err)
		}

		total := len(neighbors)
		established := 0
		var totalPrefixes uint32
		for _, n := range neighbors {
			if n.State == "Established" {
				established++
			}
			totalPrefixes += n.PrefixesAccepted
		}

		summaryData := map[string]any{
			"total_peers":       total,
			"established_peers": established,
			"accepted_prefixes": totalPrefixes,
			"peers":             neighbors,
		}

		return formatOutput(cmd, summaryData, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "METRIC\tVALUE\n")
			fmt.Fprintf(w, "Total BGP Peers\t%d\n", total)
			fmt.Fprintf(w, "Established Peers\t%d\n", established)
			fmt.Fprintf(w, "Total Accepted Prefixes\t%d\n", totalPrefixes)
			w.Flush()
			fmt.Println()

			w2 := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w2, "PEER\tREMOTE AS\tSTATE\tUPTIME\tPREFIXES\n")
			for _, n := range neighbors {
				fmt.Fprintf(w2, "%s\t%d\t%s\t%s\t%d\n",
					n.PeerAddress,
					n.RemoteAS,
					n.State,
					(time.Duration(n.UptimeSeconds) * time.Second).String(),
					n.PrefixesAccepted,
				)
			}
			w2.Flush()
		})
	},
}

func init() {
	bgpNeighborCmd.AddCommand(bgpNeighborListCmd)
	bgpCmd.AddCommand(bgpNeighborCmd)
	bgpCmd.AddCommand(bgpSummaryCmd)
}
