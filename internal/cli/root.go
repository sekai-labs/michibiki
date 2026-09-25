package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/sekai-labs/michibiki/internal/handler/auth"
	"github.com/sekai-labs/michibiki/internal/handler/network"
	"github.com/sekai-labs/michibiki/internal/handler/session"
	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var (
	flagDevice   string
	flagURL      string
	flagOutput   string
	flagInsecure bool
	flagConfig    string
	flagNoColor   bool
	flagTokenFile string
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
	rootCmd.PersistentFlags().StringVarP(&flagTokenFile, "token-file", "f", "", "Path to token file (raw token, key:secret, or JSON credentials)")
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
	rootCmd.AddCommand(authCmd)
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

func getSessionHandler() (*session.SessionHandler, error) {
	cfg, err := loadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return session.NewSessionHandler(cfg, auth.NewAuthHandler()), nil
}

func resolveProvider(cmd *cobra.Command) (provider.Provider, *config.DeviceProfile, error) {
	sm, err := getSessionHandler()
	if err != nil {
		return nil, nil, err
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	prov, profile, err := sm.ResolveAndConnect(ctx, flagDevice, flagURL, flagTokenFile, flagInsecure)
	if err != nil {
		return nil, nil, err
	}
	return prov.(provider.Provider), profile, nil
}

func resolveNetworkHandler(cmd *cobra.Command) (*network.NetworkHandler, *config.DeviceProfile, error) {
	sm, err := getSessionHandler()
	if err != nil {
		return nil, nil, err
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return sm.CreateNetworkHandler(ctx, flagDevice, flagURL, flagTokenFile, flagInsecure)
}

func resolveNetworkService(cmd *cobra.Command) (*network.NetworkHandler, *config.DeviceProfile, error) {
	return resolveNetworkHandler(cmd)
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
