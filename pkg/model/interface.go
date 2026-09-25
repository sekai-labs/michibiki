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
