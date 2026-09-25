package model

type RouteProtocol string

const (
	RouteProtoKernel    RouteProtocol = "kernel"
	RouteProtoStatic    RouteProtocol = "static"
	RouteProtoBGP       RouteProtocol = "bgp"
	RouteProtoOSPF      RouteProtocol = "ospf"
	RouteProtoConnected RouteProtocol = "connected"
	RouteProtoOther     RouteProtocol = "other"
)

type Route struct {
	Destination string        `json:"destination" yaml:"destination"`
	Gateway     string        `json:"gateway" yaml:"gateway"`
	Interface   string        `json:"interface" yaml:"interface"`
	Protocol    RouteProtocol `json:"protocol" yaml:"protocol"`
	Metric      int           `json:"metric" yaml:"metric"`
	Scope       string        `json:"scope" yaml:"scope"`
}

type GatewayStatus string

const (
	GatewayOnline   GatewayStatus = "online"
	GatewayOffline  GatewayStatus = "offline"
	GatewayDegraded GatewayStatus = "degraded"
	GatewayUnknown  GatewayStatus = "unknown"
)

type Gateway struct {
	Name          string        `json:"name" yaml:"name"`
	Address       string        `json:"address" yaml:"address"`
	Interface     string        `json:"interface" yaml:"interface"`
	Status        GatewayStatus `json:"status" yaml:"status"`
	LatencyMs     float64       `json:"latency_ms" yaml:"latency_ms"`
	PacketLossPct float64       `json:"packet_loss_pct" yaml:"packet_loss_pct"`
	IsDefault     bool          `json:"is_default" yaml:"is_default"`
}

type ARPStatus string

const (
	ARPReachable ARPStatus = "reachable"
	ARPStale     ARPStatus = "stale"
	ARPStatic    ARPStatus = "static"
	ARPDelayed   ARPStatus = "delayed"
)

type ARPEntry struct {
	IPAddress  string    `json:"ip_address" yaml:"ip_address"`
	MACAddress string    `json:"mac_address" yaml:"mac_address"`
	Interface  string    `json:"interface" yaml:"interface"`
	Hostname   string    `json:"hostname" yaml:"hostname"`
	Status     ARPStatus `json:"status" yaml:"status"`
}
