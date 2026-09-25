package model

import "time"

type InterfaceType string

const (
	InterfaceTypeEthernet  InterfaceType = "ethernet"
	InterfaceTypeVLAN      InterfaceType = "vlan"
	InterfaceTypeBridge    InterfaceType = "bridge"
	InterfaceTypeLoopback  InterfaceType = "loopback"
	InterfaceTypeWireGuard InterfaceType = "wireguard"
	InterfaceTypePPPoE     InterfaceType = "pppoe"
	InterfaceTypeBond      InterfaceType = "bond"
	InterfaceTypeOther     InterfaceType = "other"
)

type AdminStatus string

const (
	AdminStatusUp   AdminStatus = "up"
	AdminStatusDown AdminStatus = "down"
)

type OperStatus string

const (
	OperStatusUp      OperStatus = "up"
	OperStatusDown    OperStatus = "down"
	OperStatusDormant OperStatus = "dormant"
	OperStatusUnknown OperStatus = "unknown"
)

type SystemInfo struct {
	Hostname         string    `json:"hostname" yaml:"hostname"`
	OS               string    `json:"os" yaml:"os"`
	Version          string    `json:"version" yaml:"version"`
	Architecture     string    `json:"architecture" yaml:"architecture"`
	UptimeSeconds    uint64    `json:"uptime_seconds" yaml:"uptime_seconds"`
	CPUCount         int       `json:"cpu_count" yaml:"cpu_count"`
	CPUUsagePct      float64   `json:"cpu_usage_pct" yaml:"cpu_usage_pct"`
	MemoryTotalBytes uint64    `json:"memory_total_bytes" yaml:"memory_total_bytes"`
	MemoryUsedBytes  uint64    `json:"memory_used_bytes" yaml:"memory_used_bytes"`
	StorageTotal     uint64    `json:"storage_total_bytes" yaml:"storage_total_bytes"`
	StorageUsed      uint64    `json:"storage_used_bytes" yaml:"storage_used_bytes"`
	SerialNumber     string    `json:"serial_number" yaml:"serial_number"`
	Time             time.Time `json:"time" yaml:"time"`
}

type Interface struct {
	ID              string        `json:"id" yaml:"id"`
	Name            string        `json:"name" yaml:"name"`
	Type            InterfaceType `json:"type" yaml:"type"`
	AdminStatus     AdminStatus   `json:"admin_status" yaml:"admin_status"`
	OperStatus      OperStatus    `json:"oper_status" yaml:"oper_status"`
	MACAddress      string        `json:"mac_address" yaml:"mac_address"`
	MTU             int           `json:"mtu" yaml:"mtu"`
	IPv4Addresses   []string      `json:"ipv4_addresses" yaml:"ipv4_addresses"`
	IPv6Addresses   []string      `json:"ipv6_addresses" yaml:"ipv6_addresses"`
	VLANID          int           `json:"vlan_id,omitempty" yaml:"vlan_id,omitempty"`
	ParentInterface string        `json:"parent_interface,omitempty" yaml:"parent_interface,omitempty"`
	SpeedBps        uint64        `json:"speed_bps" yaml:"speed_bps"`
	Duplex          string        `json:"duplex" yaml:"duplex"`
	Description     string        `json:"description" yaml:"description"`
}

type VLAN struct {
	ID              int      `json:"id" yaml:"id"`
	Name            string   `json:"name" yaml:"name"`
	ParentInterface string   `json:"parent_interface" yaml:"parent_interface"`
	Description     string   `json:"description" yaml:"description"`
	Addresses       []string `json:"addresses" yaml:"addresses"`
}

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

type DHCPState string

const (
	DHCPActive  DHCPState = "active"
	DHCPStatic  DHCPState = "static"
	DHCPExpired DHCPState = "expired"
)

type DHCPLease struct {
	IPAddress      string    `json:"ip_address" yaml:"ip_address"`
	MACAddress     string    `json:"mac_address" yaml:"mac_address"`
	ClientHostname string    `json:"client_hostname" yaml:"client_hostname"`
	SubnetCIDR     string    `json:"subnet_cidr" yaml:"subnet_cidr"`
	Interface      string    `json:"interface" yaml:"interface"`
	State          DHCPState `json:"state" yaml:"state"`
	Starts         time.Time `json:"starts" yaml:"starts"`
	Ends           time.Time `json:"ends" yaml:"ends"`
}

type FirewallAction string

const (
	FirewallPass   FirewallAction = "pass"
	FirewallBlock  FirewallAction = "block"
	FirewallReject FirewallAction = "reject"
)

type FirewallRule struct {
	ID              string         `json:"id" yaml:"id"`
	Sequence        int            `json:"sequence" yaml:"sequence"`
	Interface       string         `json:"interface" yaml:"interface"`
	Direction       string         `json:"direction" yaml:"direction"`
	Action          FirewallAction `json:"action" yaml:"action"`
	Protocol        string         `json:"protocol" yaml:"protocol"`
	Source          string         `json:"source" yaml:"source"`
	SourcePort      string         `json:"source_port" yaml:"source_port"`
	Destination     string         `json:"destination" yaml:"destination"`
	DestinationPort string         `json:"destination_port" yaml:"destination_port"`
	Description     string         `json:"description" yaml:"description"`
	Enabled         bool           `json:"enabled" yaml:"enabled"`
	Packets         uint64         `json:"packets" yaml:"packets"`
	Bytes           uint64         `json:"bytes" yaml:"bytes"`
}

type FirewallAlias struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"`
	Content     []string `json:"content" yaml:"content"`
	Description string   `json:"description" yaml:"description"`
}

type NATRule struct {
	ID          string `json:"id" yaml:"id"`
	Interface   string `json:"interface" yaml:"interface"`
	Type        string `json:"type" yaml:"type"`
	Protocol    string `json:"protocol" yaml:"protocol"`
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Target      string `json:"target" yaml:"target"`
	TargetPort  string `json:"target_port" yaml:"target_port"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Description string `json:"description" yaml:"description"`
}

type BGPPeerState string

const (
	BGPIdle        BGPPeerState = "Idle"
	BGPConnect     BGPPeerState = "Connect"
	BGPActive      BGPPeerState = "Active"
	BGPOpenSent    BGPPeerState = "OpenSent"
	BGPOpenConfirm BGPPeerState = "OpenConfirm"
	BGPEstablished BGPPeerState = "Established"
)

type BGPNeighbor struct {
	RemoteAS         uint32        `json:"remote_as" yaml:"remote_as"`
	LocalAS          uint32        `json:"local_as" yaml:"local_as"`
	PeerAddress      string        `json:"peer_address" yaml:"peer_address"`
	State            BGPPeerState  `json:"state" yaml:"state"`
	UptimeSeconds    uint64        `json:"uptime_seconds" yaml:"uptime_seconds"`
	PrefixesReceived uint32        `json:"prefixes_received" yaml:"prefixes_received"`
	PrefixesAccepted uint32        `json:"prefixes_accepted" yaml:"prefixes_accepted"`
	Description      string        `json:"description" yaml:"description"`
}

type WireGuardPeer struct {
	PublicKey       string    `json:"public_key" yaml:"public_key"`
	Endpoint        string    `json:"endpoint" yaml:"endpoint"`
	AllowedIPs      []string  `json:"allowed_ips" yaml:"allowed_ips"`
	LatestHandshake time.Time `json:"latest_handshake" yaml:"latest_handshake"`
	TransferRxBytes uint64    `json:"transfer_rx_bytes" yaml:"transfer_rx_bytes"`
	TransferTxBytes uint64    `json:"transfer_tx_bytes" yaml:"transfer_tx_bytes"`
	KeepaliveSec    int       `json:"keepalive_sec" yaml:"keepalive_sec"`
}

type InterfaceStats struct {
	InterfaceName string    `json:"interface_name" yaml:"interface_name"`
	RxBytes       uint64    `json:"rx_bytes" yaml:"rx_bytes"`
	TxBytes       uint64    `json:"tx_bytes" yaml:"tx_bytes"`
	RxPackets     uint64    `json:"rx_packets" yaml:"rx_packets"`
	TxPackets     uint64    `json:"tx_packets" yaml:"tx_packets"`
	RxErrors      uint64    `json:"rx_errors" yaml:"rx_errors"`
	TxErrors      uint64    `json:"tx_errors" yaml:"tx_errors"`
	RxDrops       uint64    `json:"rx_drops" yaml:"rx_drops"`
	TxDrops       uint64    `json:"tx_drops" yaml:"tx_drops"`
	RxBps         float64   `json:"rx_bps" yaml:"rx_bps"`
	TxBps         float64   `json:"tx_bps" yaml:"tx_bps"`
	Timestamp     time.Time `json:"timestamp" yaml:"timestamp"`
}

type MetricSample struct {
	MetricName string            `json:"metric_name" yaml:"metric_name"`
	Timestamp  time.Time         `json:"timestamp" yaml:"timestamp"`
	Value      float64           `json:"value" yaml:"value"`
	Labels     map[string]string `json:"labels" yaml:"labels"`
}

type ValidationResult struct {
	Valid    bool     `json:"valid" yaml:"valid"`
	Errors   []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type ConfigApplyRequest struct {
	CandidateConfig string `json:"candidate_config" yaml:"candidate_config"`
	ConfirmSeconds  int    `json:"confirm_seconds" yaml:"confirm_seconds"`
	Description     string `json:"description" yaml:"description"`
}

type ConfigApplyResult struct {
	Success        bool   `json:"success" yaml:"success"`
	RollbackID     string `json:"rollback_id,omitempty" yaml:"rollback_id,omitempty"`
	ConfirmPending bool   `json:"confirm_pending" yaml:"confirm_pending"`
	Message        string `json:"message" yaml:"message"`
}
