package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/plugin"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

type Server struct {
	provider provider.Provider
	reader   io.Reader
	writer   io.Writer
}

func NewServer(p provider.Provider) *Server {
	return &Server{
		provider: p,
		reader:   os.Stdin,
		writer:   os.Stdout,
	}
}

func NewServerWithIO(p provider.Provider, r io.Reader, w io.Writer) *Server {
	return &Server{
		provider: p,
		reader:   r,
		writer:   w,
	}
}

func Serve(p provider.Provider) error {
	server := NewServer(p)
	return server.Run(context.Background())
}

func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.reader)
	maxMsgSize := 16 * 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxMsgSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req plugin.Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(0, plugin.RPCParseError, err.Error())
			continue
		}

		s.handleRequest(ctx, req)
	}

	return scanner.Err()
}

func (s *Server) sendResponse(id uint64, result any) {
	data, err := json.Marshal(result)
	if err != nil {
		s.sendError(id, plugin.RPCInternalError, err.Error())
		return
	}

	resp := plugin.Response{
		JSONRPC: plugin.JSONRPCVersion,
		ID:      id,
		Result:  data,
	}
	encoded, _ := json.Marshal(resp)
	encoded = append(encoded, '\n')
	_, _ = s.writer.Write(encoded)
}

func (s *Server) sendError(id uint64, code int, msg string) {
	resp := plugin.Response{
		JSONRPC: plugin.JSONRPCVersion,
		ID:      id,
		Error: &plugin.RPCError{
			Code:    code,
			Message: msg,
		},
	}
	encoded, _ := json.Marshal(resp)
	encoded = append(encoded, '\n')
	_, _ = s.writer.Write(encoded)
}

func (s *Server) handleRequest(ctx context.Context, req plugin.Request) {
	p := s.provider

	switch req.Method {
	case plugin.MethodPluginHandshake:
		var params plugin.HandshakeParams
		_ = json.Unmarshal(req.Params, &params)
		resp := plugin.HandshakeResponse{
			Name:         p.ID(),
			Version:      "1.0.0",
			Capabilities: p.Capabilities().Strings(),
		}
		s.sendResponse(req.ID, resp)

	case plugin.MethodProviderConnect:
		var params plugin.ConnectParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, plugin.RPCInvalidParams, err.Error())
			return
		}
		if err := p.Connect(ctx, params.Endpoint, params.Creds, params.Options); err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, true)

	case plugin.MethodProviderDisconnect:
		if err := p.Disconnect(ctx); err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, true)

	case plugin.MethodProviderGetSystem:
		info, err := p.GetSystemInfo(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, info)

	case plugin.MethodProviderListIfaces:
		ifaces, err := p.ListInterfaces(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, ifaces)

	case plugin.MethodProviderGetIface:
		var params plugin.GetInterfaceParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, plugin.RPCInvalidParams, err.Error())
			return
		}
		iface, err := p.GetInterface(ctx, params.Name)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, iface)

	case plugin.MethodProviderListRoutes:
		routes, err := p.ListRoutes(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, routes)

	case plugin.MethodProviderListGateways:
		gateways, err := p.ListGateways(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, gateways)

	case plugin.MethodProviderListARP:
		entries, err := p.ListARPEntries(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, entries)

	case plugin.MethodProviderListDHCP:
		leases, err := p.ListDHCPLeases(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, leases)

	case plugin.MethodProviderListRules:
		rules, err := p.ListFirewallRules(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, rules)

	case plugin.MethodProviderListAliases:
		aliases, err := p.ListFirewallAliases(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, aliases)

	case plugin.MethodProviderListNAT:
		nats, err := p.ListNATRules(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, nats)

	case plugin.MethodProviderListBGP:
		bgps, err := p.ListBGPNeighbors(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, bgps)

	case plugin.MethodProviderListWG:
		peers, err := p.ListWireGuardPeers(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, peers)

	case plugin.MethodProviderGetStats:
		stats, err := p.GetInterfaceStats(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, stats)

	case plugin.MethodProviderGetConfig:
		cfg, err := p.GetRunningConfig(ctx)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, cfg)

	case plugin.MethodProviderValidateConfig:
		var params plugin.ValidateConfigParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, plugin.RPCInvalidParams, err.Error())
			return
		}
		res, err := p.ValidateConfig(ctx, params.Candidate)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, res)

	case plugin.MethodProviderApplyConfig:
		var params plugin.ApplyConfigParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, plugin.RPCInvalidParams, err.Error())
			return
		}
		res, err := p.ApplyConfig(ctx, params.Request)
		if err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, res)

	case plugin.MethodProviderRollback:
		var params plugin.RollbackConfigParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, plugin.RPCInvalidParams, err.Error())
			return
		}
		if err := p.RollbackConfig(ctx, params.RollbackID); err != nil {
			s.sendError(req.ID, plugin.RPCInternalError, err.Error())
			return
		}
		s.sendResponse(req.ID, true)

	default:
		s.sendError(req.ID, plugin.RPCMethodNotFound, fmt.Sprintf("method %s not found", req.Method))
	}
}

type BaseProvider struct{}

func (b *BaseProvider) ID() string   { return "" }
func (b *BaseProvider) Name() string { return "" }
func (b *BaseProvider) Connect(ctx context.Context, endpoint string, creds *model.SystemInfo, options map[string]string) error {
	return nil
}
func (b *BaseProvider) Disconnect(ctx context.Context) error { return nil }
func (b *BaseProvider) Capabilities() provider.Capabilities   { return 0 }
func (b *BaseProvider) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListInterfaces(ctx context.Context) ([]model.Interface, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) GetInterface(ctx context.Context, name string) (*model.Interface, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListRoutes(ctx context.Context) ([]model.Route, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListARPEntries(ctx context.Context) ([]model.ARPEntry, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListDHCPLeases(ctx context.Context) ([]model.DHCPLease, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListFirewallAliases(ctx context.Context) ([]model.FirewallAlias, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListNATRules(ctx context.Context) ([]model.NATRule, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListBGPNeighbors(ctx context.Context) ([]model.BGPNeighbor, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ListWireGuardPeers(ctx context.Context) ([]model.WireGuardPeer, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) GetInterfaceStats(ctx context.Context) ([]model.InterfaceStats, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) GetSubnetUsage(ctx context.Context, filter string) ([]model.SystemInfo, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) GetRunningConfig(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}
func (b *BaseProvider) ValidateConfig(ctx context.Context, candidate string) (*model.ValidationResult, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) ApplyConfig(ctx context.Context, req model.ConfigApplyRequest) (*model.ConfigApplyResult, error) {
	return nil, errors.New("not implemented")
}
func (b *BaseProvider) RollbackConfig(ctx context.Context, rollbackID string) error {
	return errors.New("not implemented")
}
