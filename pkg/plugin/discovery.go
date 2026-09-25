package plugin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

type DiscoveredPlugin struct {
	Name         string                `json:"name"`
	BinaryPath   string                `json:"binary_path"`
	Version      string                `json:"version"`
	Capabilities provider.Capabilities `json:"capabilities"`
}

func DiscoverPlugins(dirs []string) ([]DiscoveredPlugin, error) {
	searchDirs := make([]string, 0, len(dirs)+8)
	searchDirs = append(searchDirs, dirs...)

	searchDirs = append(searchDirs, "./plugins")

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		searchDirs = append(searchDirs, filepath.Join(home, ".config", "michibiki", "plugins"))
	}

	if pathEnv := os.Getenv("PATH"); pathEnv != "" {
		for _, p := range filepath.SplitList(pathEnv) {
			if p != "" {
				searchDirs = append(searchDirs, p)
			}
		}
	}

	seenPaths := make(map[string]bool)
	seenNames := make(map[string]bool)
	var discovered []DiscoveredPlugin

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fileName := entry.Name()
			if !strings.HasPrefix(fileName, "michibiki-provider-") {
				continue
			}

			fullPath := filepath.Join(dir, fileName)
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				absPath = fullPath
			}

			if seenPaths[absPath] {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 == 0 {
				continue
			}

			seenPaths[absPath] = true

			pluginInfo, err := InspectPlugin(absPath)
			if err != nil {
				name := strings.TrimPrefix(fileName, "michibiki-provider-")
				name = strings.TrimSuffix(name, filepath.Ext(name))
				if seenNames[name] {
					continue
				}
				seenNames[name] = true
				discovered = append(discovered, DiscoveredPlugin{
					Name:         name,
					BinaryPath:   absPath,
					Version:      "unknown",
					Capabilities: 0,
				})
				continue
			}

			if seenNames[pluginInfo.Name] {
				continue
			}
			seenNames[pluginInfo.Name] = true
			discovered = append(discovered, *pluginInfo)
		}
	}

	return discovered, nil
}

func InspectPlugin(binaryPath string) (*DiscoveredPlugin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, err
	}

	client := NewClient(stdout, stdin)
	defer func() {
		_ = client.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	var handshakeResp HandshakeResponse
	if err := client.Call(ctx, MethodPluginHandshake, HandshakeParams{Version: "1.0"}, &handshakeResp); err != nil {
		return nil, err
	}

	name := handshakeResp.Name
	if name == "" {
		base := filepath.Base(binaryPath)
		name = strings.TrimPrefix(base, "michibiki-provider-")
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}

	return &DiscoveredPlugin{
		Name:         name,
		BinaryPath:   binaryPath,
		Version:      handshakeResp.Version,
		Capabilities: ParseCapabilities(handshakeResp.Capabilities),
	}, nil
}

func RegisterDiscoveredPlugins(plugins []DiscoveredPlugin) {
	for _, p := range plugins {
		plug := p
		provider.Register(plug.Name, func() provider.Provider {
			return NewPlugin(plug.BinaryPath)
		})
	}
}

func RegisterDiscoveredPlugin(p DiscoveredPlugin) {
	provider.Register(p.Name, func() provider.Provider {
		return NewPlugin(p.BinaryPath)
	})
}
