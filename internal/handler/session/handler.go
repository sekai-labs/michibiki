package session

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/sekai-labs/michibiki/internal/port"
	"github.com/sekai-labs/michibiki/internal/handler/auth"
	"github.com/sekai-labs/michibiki/internal/handler/network"
	"github.com/sekai-labs/michibiki/pkg/config"
	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/plugin"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type SessionHandler struct {
	cfg     *config.Config
	authHdl *auth.AuthHandler
}

func NewSessionHandler(cfg *config.Config, authHdl *auth.AuthHandler) *SessionHandler {
	if authHdl == nil {
		authHdl = auth.NewAuthHandler()
	}
	return &SessionHandler{
		cfg:     cfg,
		authHdl: authHdl,
	}
}

func (m *SessionHandler) Config() *config.Config {
	return m.cfg
}

func (m *SessionHandler) AuthHandler() *auth.AuthHandler {
	return m.authHdl
}

func (m *SessionHandler) ResolveAndConnect(
	ctx context.Context,
	deviceName string,
	directURL string,
	tokenFilePath string,
	insecure bool,
) (port.ProviderPort, *config.DeviceProfile, error) {
	if m.cfg == nil {
		return nil, nil, errors.New("configuration not loaded")
	}

	var pluginDirs []string
	if m.cfg.PluginDir != "" {
		pluginDirs = append(pluginDirs, m.cfg.PluginDir)
	}
	if envPluginDir := os.Getenv("MICHIBIKI_PLUGIN_DIR"); envPluginDir != "" {
		pluginDirs = append(pluginDirs, envPluginDir)
	}
	discovered, _ := plugin.DiscoverPlugins(pluginDirs)
	if len(discovered) > 0 {
		plugin.RegisterDiscoveredPlugins(discovered)
	}

	if deviceName == "" && directURL == "" {
		deviceName = m.cfg.DefaultDevice
	}

	var profile config.DeviceProfile
	if deviceName != "" {
		p, err := m.cfg.GetDevice(deviceName)
		if err == nil {
			profile = *p
		} else if directURL == "" {
			return nil, nil, fmt.Errorf("device profile '%s' not found: %w", deviceName, err)
		}
	}

	if directURL != "" {
		profile.Address = directURL
		if profile.Name == "" {
			profile.Name = "ephemeral"
		}
		if profile.Provider == "" {
			return nil, nil, errors.New("provider must be specified when using direct URL")
		}
	}

	if insecure {
		profile.Insecure = true
	}

	if profile.Provider == "" {
		return nil, nil, errors.New("no provider configured for device")
	}

	prov, err := provider.Create(profile.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create provider '%s': %w", profile.Provider, err)
	}

	creds, err := m.ResolveCredentials(&profile, tokenFilePath)
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

	if ctx == nil {
		ctx = context.Background()
	}

	endpoint := profile.Address
	if profile.Port > 0 && !strings.Contains(endpoint, ":") {
		if addr, err := netip.ParseAddr(profile.Address); err == nil {
			endpoint = netip.AddrPortFrom(addr, uint16(profile.Port)).String()
		}
	}

	if err := prov.Connect(ctx, endpoint, creds, options); err != nil {
		return nil, nil, fmt.Errorf("connection to '%s' failed: %w", profile.Name, err)
	}

	return prov, &profile, nil
}

func (m *SessionHandler) CreateNetworkHandler(
	ctx context.Context,
	deviceName string,
	directURL string,
	tokenFilePath string,
	insecure bool,
) (*network.NetworkHandler, *config.DeviceProfile, error) {
	prov, profile, err := m.ResolveAndConnect(ctx, deviceName, directURL, tokenFilePath, insecure)
	if err != nil {
		return nil, nil, err
	}
	return network.NewNetworkHandler(prov), profile, nil
}
