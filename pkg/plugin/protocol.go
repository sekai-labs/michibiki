package plugin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sekai-labs/michibiki/pkg/credential"
	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

const (
	JSONRPCVersion = "2.0"

	RPCParseError     = -32700
	RPCInvalidRequest = -32600
	RPCMethodNotFound = -32601
	RPCInvalidParams  = -32602
	RPCInternalError  = -32603

	MethodPluginHandshake        = "plugin.handshake"
	MethodProviderConnect        = "provider.connect"
	MethodProviderDisconnect     = "provider.disconnect"
	MethodProviderGetSystem      = "provider.getSystemInfo"
	MethodProviderListIfaces     = "provider.listInterfaces"
	MethodProviderGetIface       = "provider.getInterface"
	MethodProviderListRoutes     = "provider.listRoutes"
	MethodProviderListGateways   = "provider.listGateways"
	MethodProviderListARP        = "provider.listARPEntries"
	MethodProviderListDHCP       = "provider.listDHCPLeases"
	MethodProviderListRules      = "provider.listFirewallRules"
	MethodProviderListAliases    = "provider.listFirewallAliases"
	MethodProviderListNAT        = "provider.listNATRules"
	MethodProviderListBGP        = "provider.listBGPNeighbors"
	MethodProviderListWG         = "provider.listWireGuardPeers"
	MethodProviderGetStats       = "provider.getInterfaceStats"
	MethodProviderGetConfig      = "provider.getRunningConfig"
	MethodProviderValidateConfig = "provider.validateConfig"
	MethodProviderApplyConfig    = "provider.applyConfig"
	MethodProviderRollback       = "provider.rollbackConfig"
)

const CapAll provider.Capabilities = provider.CapSystem |
	provider.CapInterfaces |
	provider.CapRouting |
	provider.CapFirewall |
	provider.CapNAT |
	provider.CapDHCP |
	provider.CapDNS |
	provider.CapWireGuard |
	provider.CapBGP |
	provider.CapOSPF |
	provider.CapMonitoring |
	provider.CapConfig |
	provider.CapIPAM

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

type HandshakeParams struct {
	Version string `json:"version"`
}

type HandshakeResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type ConnectParams struct {
	Endpoint string                  `json:"endpoint"`
	Creds    *credential.Credentials `json:"creds,omitempty"`
	Options  map[string]string       `json:"options,omitempty"`
}

type GetInterfaceParams struct {
	Name string `json:"name"`
}

type ValidateConfigParams struct {
	Candidate string `json:"candidate"`
}

type ApplyConfigParams struct {
	Request model.ConfigApplyRequest `json:"request"`
}

type RollbackConfigParams struct {
	RollbackID string `json:"rollback_id"`
}

func ParseCapabilities(caps []string) provider.Capabilities {
	var c provider.Capabilities
	for _, s := range caps {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "all", "*":
			return CapAll
		case "system":
			c |= provider.CapSystem
		case "interfaces":
			c |= provider.CapInterfaces
		case "routing":
			c |= provider.CapRouting
		case "firewall":
			c |= provider.CapFirewall
		case "nat":
			c |= provider.CapNAT
		case "dhcp":
			c |= provider.CapDHCP
		case "dns":
			c |= provider.CapDNS
		case "wireguard":
			c |= provider.CapWireGuard
		case "bgp":
			c |= provider.CapBGP
		case "ospf":
			c |= provider.CapOSPF
		case "monitoring":
			c |= provider.CapMonitoring
		case "config":
			c |= provider.CapConfig
		case "ipam":
			c |= provider.CapIPAM
		}
	}
	return c
}
