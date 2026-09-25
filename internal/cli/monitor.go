package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagMonitorInterval time.Duration
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Real-time streaming telemetry and performance monitoring",
}

var monitorTrafficCmd = &cobra.Command{
	Use:   "traffic",
	Short: "Stream real-time interface throughput and packet rates",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		ticker := time.NewTicker(flagMonitorInterval)
		defer ticker.Stop()

		for {
			stats, err := prov.GetInterfaceStats(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch interface stats: %w", err)
			}

			if stringsEqualFold(flagOutput, "json") || stringsEqualFold(flagOutput, "yaml") {
				return formatOutput(cmd, stats, nil)
			}

			fmt.Print("\033[H\033[2J")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "INTERFACE\tRX RATE\tTX RATE\tRX PACKETS\tTX PACKETS\tRX ERRORS\tTX ERRORS\n")
			for _, s := range stats {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%d\n",
					s.InterfaceName,
					formatBps(s.RxBps),
					formatBps(s.TxBps),
					s.RxPackets,
					s.TxPackets,
					s.RxErrors,
					s.TxErrors,
				)
			}
			w.Flush()
			fmt.Printf("\n[Refreshed at %s | Interval: %s | Press Ctrl+C to exit]\n",
				time.Now().Format("15:04:05"), flagMonitorInterval)

			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	},
}

var monitorInterfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Stream interface operational status and traffic counters",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		ticker := time.NewTicker(flagMonitorInterval)
		defer ticker.Stop()

		for {
			ifaces, err := prov.ListInterfaces(ctx)
			if err != nil {
				return fmt.Errorf("failed to list interfaces: %w", err)
			}

			if stringsEqualFold(flagOutput, "json") || stringsEqualFold(flagOutput, "yaml") {
				return formatOutput(cmd, ifaces, nil)
			}

			fmt.Print("\033[H\033[2J")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tOPER\tADMIN\tMTU\tSPEED\n")
			for _, iface := range ifaces {
				speedStr := "-"
				if iface.SpeedBps > 0 {
					if iface.SpeedBps >= 1_000_000_000 {
						speedStr = fmt.Sprintf("%d Gbps", iface.SpeedBps/1_000_000_000)
					} else {
						speedStr = fmt.Sprintf("%d Mbps", iface.SpeedBps/1_000_000)
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
					iface.Name,
					iface.Type,
					iface.OperStatus,
					iface.AdminStatus,
					iface.MTU,
					speedStr,
				)
			}
			w.Flush()
			fmt.Printf("\n[Refreshed at %s | Interval: %s | Press Ctrl+C to exit]\n",
				time.Now().Format("15:04:05"), flagMonitorInterval)

			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	},
}

var monitorSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Stream system resource utilization (CPU, memory, storage)",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, profile, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		ticker := time.NewTicker(flagMonitorInterval)
		defer ticker.Stop()

		for {
			info, err := prov.GetSystemInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to retrieve system info: %w", err)
			}

			if stringsEqualFold(flagOutput, "json") || stringsEqualFold(flagOutput, "yaml") {
				return formatOutput(cmd, info, nil)
			}

			fmt.Print("\033[H\033[2J")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "RESOURCE\tUSAGE\tDETAILS\n")
			fmt.Fprintf(w, "Device\t%s\t%s (%s)\n", profile.Name, info.Hostname, info.OS)
			fmt.Fprintf(w, "CPU Usage\t%.1f%%\t%d Cores\n", info.CPUUsagePct, info.CPUCount)

			memPct := 0.0
			if info.MemoryTotalBytes > 0 {
				memPct = (float64(info.MemoryUsedBytes) / float64(info.MemoryTotalBytes)) * 100.0
			}
			fmt.Fprintf(w, "Memory\t%.1f%%\t%s / %s\n",
				memPct, formatBytes(info.MemoryUsedBytes), formatBytes(info.MemoryTotalBytes))

			diskPct := 0.0
			if info.StorageTotal > 0 {
				diskPct = (float64(info.StorageUsed) / float64(info.StorageTotal)) * 100.0
			}
			fmt.Fprintf(w, "Storage\t%.1f%%\t%s / %s\n",
				diskPct, formatBytes(info.StorageUsed), formatBytes(info.StorageTotal))

			w.Flush()
			fmt.Printf("\n[Refreshed at %s | Interval: %s | Press Ctrl+C to exit]\n",
				time.Now().Format("15:04:05"), flagMonitorInterval)

			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	},
}

func formatBps(bps float64) string {
	if bps >= 1_000_000_000 {
		return fmt.Sprintf("%.2f Gbps", bps/1_000_000_000)
	}
	if bps >= 1_000_000 {
		return fmt.Sprintf("%.2f Mbps", bps/1_000_000)
	}
	if bps >= 1_000 {
		return fmt.Sprintf("%.2f Kbps", bps/1_000)
	}
	return fmt.Sprintf("%.0f bps", bps)
}

func stringsEqualFold(s1, s2 string) bool {
	return stringsEqual(s1, s2)
}

func stringsEqual(s1, s2 string) bool {
	return len(s1) == len(s2) && (s1 == s2 || (len(s1) > 0 && (s1[0]|0x20) == (s2[0]|0x20)))
}

func init() {
	monitorCmd.PersistentFlags().DurationVar(&flagMonitorInterval, "interval", 2*time.Second, "Polling interval")
	monitorCmd.AddCommand(monitorTrafficCmd)
	monitorCmd.AddCommand(monitorInterfacesCmd)
	monitorCmd.AddCommand(monitorSystemCmd)
}
