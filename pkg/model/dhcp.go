package model

import "time"

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
