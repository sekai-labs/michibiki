package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/pkg/plugin"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

var pluginCmd = &cobra.Command{
	Use:     "plugin",
	Aliases: []string{"plugins"},
	Short:   "Discover, inspect, and manage out-of-tree provider plugins",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List built-in and discovered external provider plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		var pluginDirs []string
		if cfg.PluginDir != "" {
			pluginDirs = append(pluginDirs, cfg.PluginDir)
		}
		if envPluginDir := os.Getenv("MICHIBIKI_PLUGIN_DIR"); envPluginDir != "" {
			pluginDirs = append(pluginDirs, envPluginDir)
		}

		discovered, err := plugin.DiscoverPlugins(pluginDirs)
		if err != nil {
			return fmt.Errorf("plugin discovery failed: %w", err)
		}

		type PluginEntry struct {
			Name         string   `json:"name" yaml:"name"`
			Type         string   `json:"type" yaml:"type"`
			Version      string   `json:"version" yaml:"version"`
			BinaryPath   string   `json:"binary_path,omitempty" yaml:"binary_path,omitempty"`
			Capabilities []string `json:"capabilities" yaml:"capabilities"`
		}

		var allPlugins []PluginEntry

		for _, regName := range provider.ListRegistered() {
			isDiscovered := false
			for _, dp := range discovered {
				if dp.Name == regName {
					isDiscovered = true
					break
				}
			}
			if !isDiscovered {
				p, err := provider.Create(regName)
				var capStrs []string
				if err == nil {
					capStrs = p.Capabilities().Strings()
				}
				allPlugins = append(allPlugins, PluginEntry{
					Name:         regName,
					Type:         "built-in",
					Version:      "embedded",
					Capabilities: capStrs,
				})
			}
		}

		for _, dp := range discovered {
			allPlugins = append(allPlugins, PluginEntry{
				Name:         dp.Name,
				Type:         "external",
				Version:      dp.Version,
				BinaryPath:   dp.BinaryPath,
				Capabilities: dp.Capabilities.Strings(),
			})
		}

		return formatOutput(cmd, allPlugins, func() {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tVERSION\tBINARY PATH\tCAPABILITIES\n")
			for _, p := range allPlugins {
				binPath := p.BinaryPath
				if binPath == "" {
					binPath = "(compiled-in)"
				}
				capsStr := strings.Join(p.Capabilities, ", ")
				if capsStr == "" {
					capsStr = "all"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					p.Name,
					p.Type,
					p.Version,
					binPath,
					capsStr,
				)
			}
			w.Flush()
		})
	},
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <name-or-binary>",
	Short: "Inspect metadata and capabilities of a provider plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		if _, err := os.Stat(target); err == nil {
			absPath, _ := filepath.Abs(target)
			dp, err := plugin.InspectPlugin(absPath)
			if err != nil {
				return fmt.Errorf("failed to inspect plugin binary '%s': %w", target, err)
			}

			return formatOutput(cmd, dp, func() {
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintf(w, "PROPERTY\tVALUE\n")
				fmt.Fprintf(w, "Name\t%s\n", dp.Name)
				fmt.Fprintf(w, "Type\tExternal Subprocess Plugin\n")
				fmt.Fprintf(w, "Version\t%s\n", dp.Version)
				fmt.Fprintf(w, "Binary Path\t%s\n", dp.BinaryPath)
				fmt.Fprintf(w, "Capabilities\t%s\n", strings.Join(dp.Capabilities.Strings(), ", "))
				w.Flush()
			})
		}

		cfg, err := loadAppConfig()
		if err != nil {
			return err
		}

		var pluginDirs []string
		if cfg.PluginDir != "" {
			pluginDirs = append(pluginDirs, cfg.PluginDir)
		}
		discovered, _ := plugin.DiscoverPlugins(pluginDirs)
		for _, dp := range discovered {
			if dp.Name == target {
				return formatOutput(cmd, dp, func() {
					w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
					fmt.Fprintf(w, "PROPERTY\tVALUE\n")
					fmt.Fprintf(w, "Name\t%s\n", dp.Name)
					fmt.Fprintf(w, "Type\tExternal Subprocess Plugin\n")
					fmt.Fprintf(w, "Version\t%s\n", dp.Version)
					fmt.Fprintf(w, "Binary Path\t%s\n", dp.BinaryPath)
					fmt.Fprintf(w, "Capabilities\t%s\n", strings.Join(dp.Capabilities.Strings(), ", "))
					w.Flush()
				})
			}
		}

		p, err := provider.Create(target)
		if err == nil {
			info := map[string]any{
				"name":         p.Name(),
				"id":           p.ID(),
				"type":         "built-in",
				"capabilities": p.Capabilities().Strings(),
			}
			return formatOutput(cmd, info, func() {
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintf(w, "PROPERTY\tVALUE\n")
				fmt.Fprintf(w, "Name\t%s\n", p.Name())
				fmt.Fprintf(w, "ID\t%s\n", p.ID())
				fmt.Fprintf(w, "Type\tBuilt-in Compiled Provider\n")
				fmt.Fprintf(w, "Capabilities\t%s\n", strings.Join(p.Capabilities().Strings(), ", "))
				w.Flush()
			})
		}

		return fmt.Errorf("plugin '%s' not found as built-in provider or discovered external binary", target)
	},
}
var (
	flagInstallTargetDir string
	flagInstallForce     bool
)

var pluginInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a community provider plugin from a Go package, URL, or local binary",
	Long:  "Install a community plugin by specifying a Go package (e.g. github.com/user/michibiki-provider-cisco@latest), a download URL (.tar.gz, .zip, or raw binary), or a local executable file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := strings.TrimSpace(args[0])
		opts := plugin.InstallOptions{
			TargetDir: flagInstallTargetDir,
			Force:     flagInstallForce,
		}

		fmt.Printf("Installing plugin from %s...\n", source)

		var installedPath string
		var err error

		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			installedPath, err = plugin.InstallFromURL(cmd.Context(), source, opts)
		} else if _, statErr := os.Stat(source); statErr == nil {
			installedPath, err = plugin.InstallFromLocalFile(source, opts)
		} else {
			installedPath, err = plugin.InstallFromGoModule(cmd.Context(), source, opts)
		}

		if err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}

		fmt.Printf("Successfully installed plugin to %s\n", installedPath)
		return nil
	},
}

func init() {
	pluginInstallCmd.Flags().StringVar(&flagInstallTargetDir, "dir", "", "Custom target plugin directory")
	pluginInstallCmd.Flags().BoolVarP(&flagInstallForce, "force", "f", false, "Overwrite existing plugin binary")

	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
}
