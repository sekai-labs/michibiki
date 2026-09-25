package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/plugin"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var (
	flagDevice   string
	flagURL      string
	flagOutput   string
	flagInsecure bool
	flagConfig   string
	flagNoColor  bool
)

var rootCmd = &cobra.Command{
	Use:   "michibiki",
	Short: "Universal Network CLI & TUI",
	Long:  "Michibiki (導き) - One terminal. Every network.\nVendor-neutral network management for OPNsense, RouterOS, OpenWrt, VyOS, pfSense, and FRRouting.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagDevice, "device", "d", "", "Device profile name")
	rootCmd.PersistentFlags().StringVarP(&flagURL, "url", "u", "", "Direct device URL or address")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "k", false, "TLS insecure skip verify")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "Custom configuration file path")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color escapes")

	rootCmd.AddCommand(systemCmd)
	rootCmd.AddCommand(interfaceCmd)
	rootCmd.AddCommand(routeCmd)
	rootCmd.AddCommand(gatewayCmd)
	rootCmd.AddCommand(arpCmd)
	rootCmd.AddCommand(dhcpCmd)
	rootCmd.AddCommand(firewallCmd)
	rootCmd.AddCommand(natCmd)
	rootCmd.AddCommand(bgpCmd)
	rootCmd.AddCommand(vpnCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(ipCmd)
	rootCmd.AddCommand(deviceCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(tuiCmd)
}

func loadAppConfig() (*config.Config, error) {
	cfgPath := flagConfig
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) {
			return &config.Config{
				Devices: make(map[string]config.DeviceProfile),
			}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func resolveProvider(cmd *cobra.Command) (provider.Provider, *config.DeviceProfile, error) {
	cfg, err := loadAppConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	var pluginDirs []string
	if cfg.PluginDir != "" {
		pluginDirs = append(pluginDirs, cfg.PluginDir)
	}
	if envPluginDir := os.Getenv("MICHIBIKI_PLUGIN_DIR"); envPluginDir != "" {
		pluginDirs = append(pluginDirs, envPluginDir)
	}
	discovered, _ := plugin.DiscoverPlugins(pluginDirs)
	if len(discovered) > 0 {
		plugin.RegisterDiscoveredPlugins(discovered)
	}

	deviceName := flagDevice
	if deviceName == "" && flagURL == "" {
		deviceName = cfg.DefaultDevice
	}

	var profile config.DeviceProfile
	if deviceName != "" {
		p, err := cfg.GetDevice(deviceName)
		if err == nil {
			profile = *p
		} else if flagURL == "" {
			return nil, nil, fmt.Errorf("device profile '%s' not found: %w", deviceName, err)
		}
	}

	if flagURL != "" {
		profile.Address = flagURL
		if profile.Name == "" {
			profile.Name = "ephemeral"
		}
		if profile.Provider == "" {
			return nil, nil, errors.New("provider must be specified when using direct URL")
		}
	}

	if flagInsecure {
		profile.Insecure = true
	}

	if profile.Provider == "" {
		return nil, nil, errors.New("no provider configured for device")
	}

	prov, err := provider.Create(profile.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create provider '%s': %w", profile.Provider, err)
	}

	var vault *credential.VaultStore
	vaultPass := os.Getenv("MICHIBIKI_VAULT_PASSPHRASE")
	if vaultPass != "" {
		vault = credential.NewVaultStore(config.DefaultVaultPath(), vaultPass)
	}
	resolver := credential.NewResolver(vault)

	creds, err := resolver.Resolve(profile.CredentialRef, profile.Name)
	if err != nil {
		creds = &credential.Credentials{}
	}

	options := make(map[string]string)
	for k, v := range profile.Options {
		options[k] = v
	}
	if profile.Insecure {
		options["insecure"] = "true"
	}
	if profile.Port > 0 {
		options["port"] = strconv.Itoa(profile.Port)
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	endpoint := profile.Address
	if profile.Port > 0 && !strings.Contains(endpoint, ":") {
		endpoint = netip.AddrPortFrom(netip.MustParseAddr(profile.Address), uint16(profile.Port)).String()
	}

	if err := prov.Connect(ctx, endpoint, creds, options); err != nil {
		return nil, nil, fmt.Errorf("connection to '%s' failed: %w", profile.Name, err)
	}

	return prov, &profile, nil
}

func formatOutput(cmd *cobra.Command, data any, printTable func()) error {
	outFormat := strings.ToLower(flagOutput)
	switch outFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml", "yml":
		enc := yaml.NewEncoder(os.Stdout)
		defer enc.Close()
		return enc.Encode(data)
	case "table", "":
		if printTable != nil {
			printTable()
			return nil
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		return fmt.Errorf("unsupported output format '%s' (use table, json, or yaml)", flagOutput)
	}
}
