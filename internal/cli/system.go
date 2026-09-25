package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System information and status",
}

var systemInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display system overview and hardware resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, profile, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		overview, err := svc.GetDeviceOverview(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to retrieve system overview: %w", err)
		}
		info := overview.SystemInfo
		return formatOutput(cmd, info, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "PROPERTY\tVALUE\n")
			fmt.Fprintf(w, "Device Name\t%s\n", profile.Name)
			fmt.Fprintf(w, "Hostname\t%s\n", info.Hostname)
			fmt.Fprintf(w, "OS\t%s\n", info.OS)
			fmt.Fprintf(w, "Version\t%s\n", info.Version)
			fmt.Fprintf(w, "Architecture\t%s\n", info.Architecture)
			fmt.Fprintf(w, "Serial Number\t%s\n", info.SerialNumber)

			uptimeDur := time.Duration(info.UptimeSeconds) * time.Second
			days := int(uptimeDur.Hours()) / 24
			hours := int(uptimeDur.Hours()) % 24
			mins := int(uptimeDur.Minutes()) % 60
			fmt.Fprintf(w, "Uptime\t%dd %dh %dm\n", days, hours, mins)

			fmt.Fprintf(w, "CPU Cores\t%d\n", info.CPUCount)
			fmt.Fprintf(w, "CPU Usage\t%.1f%%\n", info.CPUUsagePct)

			memUsedGB := float64(info.MemoryUsedBytes) / (1024 * 1024 * 1024)
			memTotalGB := float64(info.MemoryTotalBytes) / (1024 * 1024 * 1024)
			memPct := 0.0
			if info.MemoryTotalBytes > 0 {
				memPct = (float64(info.MemoryUsedBytes) / float64(info.MemoryTotalBytes)) * 100.0
			}
			fmt.Fprintf(w, "Memory\t%.2f GB / %.2f GB (%.1f%%)\n", memUsedGB, memTotalGB, memPct)

			diskUsedGB := float64(info.StorageUsed) / (1024 * 1024 * 1024)
			diskTotalGB := float64(info.StorageTotal) / (1024 * 1024 * 1024)
			diskPct := 0.0
			if info.StorageTotal > 0 {
				diskPct = (float64(info.StorageUsed) / float64(info.StorageTotal)) * 100.0
			}
			fmt.Fprintf(w, "Storage\t%.2f GB / %.2f GB (%.1f%%)\n", diskUsedGB, diskTotalGB, diskPct)
			w.Flush()
		})
	},
}

func init() {
	systemCmd.AddCommand(systemInfoCmd)
}
