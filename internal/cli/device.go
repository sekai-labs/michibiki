package cli

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var (
	flagDeviceProvider string
	flagDeviceAddress  string
	flagDevicePort     int
	flagDeviceInsecure bool
	flagDeviceCredRef  string
)

var deviceCmd = &cobra.Command{
	Use:     "device",
	Aliases: []string{"devices"},
	Short:   "Manage saved device profiles and connections",
}

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured device profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		type DeviceRow struct {
			Name      string `json:"name" yaml:"name"`
			Provider  string `json:"provider" yaml:"provider"`
			Address   string `json:"address" yaml:"address"`
			Port      int    `json:"port,omitempty" yaml:"port,omitempty"`
			Insecure  bool   `json:"insecure" yaml:"insecure"`
			IsDefault bool   `json:"is_default" yaml:"is_default"`
		}

		var rows []DeviceRow
		for name, dev := range cfg.Devices {
			rows = append(rows, DeviceRow{
				Name:      name,
				Provider:  dev.Provider,
				Address:   dev.Address,
				Port:      dev.Port,
				Insecure:  dev.Insecure,
				IsDefault: name == cfg.DefaultDevice,
			})
		}

		return formatOutput(cmd, rows, func() {
			if len(rows) == 0 {
				fmt.Println("No device profiles configured. Use 'michibiki device add <name> --provider <p> --address <a>' to add one.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tPROVIDER\tADDRESS\tPORT\tINSECURE\tDEFAULT\n")
			for _, r := range rows {
				portStr := "-"
				if r.Port > 0 {
					portStr = strconv.Itoa(r.Port)
				}
				defStr := ""
				if r.IsDefault {
					defStr = "✓ (default)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%s\n",
					r.Name,
					r.Provider,
					r.Address,
					portStr,
					r.Insecure,
					defStr,
				)
			}
			w.Flush()
		})
	},
}

var deviceAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new device profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if flagDeviceProvider == "" {
			return fmt.Errorf("--provider is required")
		}
		if flagDeviceAddress == "" {
			return fmt.Errorf("--address is required")
		}

		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		if cfg.Devices == nil {
			cfg.Devices = make(map[string]config.DeviceProfile)
		}

		profile := config.DeviceProfile{
			Name:          name,
			Provider:      flagDeviceProvider,
			Address:       flagDeviceAddress,
			Port:          flagDevicePort,
			Insecure:      flagDeviceInsecure,
			CredentialRef: flagDeviceCredRef,
		}

		cfg.Devices[name] = profile
		if cfg.DefaultDevice == "" {
			cfg.DefaultDevice = name
		}

		cfgPath := flagConfig
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}

		if err := config.SaveConfig(cfgPath, cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Device '%s' successfully added to %s\n", name, cfgPath)
		return nil
	},
}

var deviceRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a device profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		if _, exists := cfg.Devices[name]; !exists {
			return fmt.Errorf("device profile '%s' not found", name)
		}

		delete(cfg.Devices, name)
		if cfg.DefaultDevice == name {
			cfg.DefaultDevice = ""
			for remaining := range cfg.Devices {
				cfg.DefaultDevice = remaining
				break
			}
		}

		cfgPath := flagConfig
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}

		if err := config.SaveConfig(cfgPath, cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Device '%s' removed.\n", name)
		return nil
	},
}

var deviceTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test connectivity and credentials for a device profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		var profile config.DeviceProfile
		if p, err := cfg.GetDevice(name); err == nil {
			profile = *p
		} else {
			return fmt.Errorf("device profile '%s' not found: %w", name, err)
		}

		prov, err := provider.Create(profile.Provider)
		if err != nil {
			return fmt.Errorf("failed to create provider '%s': %w", profile.Provider, err)
		}

		var vault *credential.VaultStore
		vaultPass := os.Getenv("MICHIBIKI_VAULT_PASSPHRASE")
		if vaultPass != "" {
			vault = credential.NewVaultStore(config.DefaultVaultPath(), vaultPass)
		}
		resolver := credential.NewResolver(vault)
		creds, _ := resolver.Resolve(profile.CredentialRef, profile.Name)

		options := make(map[string]string)
		for k, v := range profile.Options {
			options[k] = v
		}
		if profile.Insecure {
			options["insecure"] = "true"
		}

		endpoint := profile.Address
		if profile.Port > 0 && !strings.Contains(endpoint, ":") {
			endpoint = netip.AddrPortFrom(netip.MustParseAddr(profile.Address), uint16(profile.Port)).String()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		if err := prov.Connect(ctx, endpoint, creds, options); err != nil {
			return fmt.Errorf("CONNECTION FAILED to %s (%s): %w", name, endpoint, err)
		}
		defer prov.Disconnect(ctx)
		latency := time.Since(start)

		sys, err := prov.GetSystemInfo(ctx)
		if err != nil {
			return fmt.Errorf("connected but failed to fetch system info: %w", err)
		}

		testResult := map[string]any{
			"device":     name,
			"provider":   profile.Provider,
			"endpoint":   endpoint,
			"status":     "connected",
			"latency_ms": latency.Milliseconds(),
			"hostname":   sys.Hostname,
			"os":         sys.OS,
			"version":    sys.Version,
		}

		return formatOutput(cmd, testResult, func() {
			fmt.Printf("✓ CONNECTED successfully to '%s' (%s) in %d ms\n", name, endpoint, latency.Milliseconds())
			fmt.Printf("  Hostname: %s\n", sys.Hostname)
			fmt.Printf("  OS:       %s %s\n", sys.OS, sys.Version)
			fmt.Printf("  Arch:     %s\n", sys.Architecture)
		})
	},
}

func init() {
	deviceAddCmd.Flags().StringVarP(&flagDeviceProvider, "provider", "p", "", "Provider type (opnsense, routeros, openwrt, vyos, pfsense, frr)")
	deviceAddCmd.Flags().StringVarP(&flagDeviceAddress, "address", "a", "", "IP address or hostname of the network device")
	deviceAddCmd.Flags().IntVar(&flagDevicePort, "port", 0, "Custom management port (optional)")
	deviceAddCmd.Flags().BoolVarP(&flagDeviceInsecure, "insecure", "k", false, "Allow self-signed TLS certificates")
	deviceAddCmd.Flags().StringVar(&flagDeviceCredRef, "cred-ref", "", "Credential reference (e.g. vault:name, env:VAR)")

	deviceCmd.AddCommand(deviceListCmd)
	deviceCmd.AddCommand(deviceAddCmd)
	deviceCmd.AddCommand(deviceRemoveCmd)
	deviceCmd.AddCommand(deviceTestCmd)
}
