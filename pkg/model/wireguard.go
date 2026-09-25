package model

import "time"

type WireGuardPeer struct {
	PublicKey       string    `json:"public_key" yaml:"public_key"`
	Endpoint        string    `json:"endpoint" yaml:"endpoint"`
	AllowedIPs      []string  `json:"allowed_ips" yaml:"allowed_ips"`
	LatestHandshake time.Time `json:"latest_handshake" yaml:"latest_handshake"`
	TransferRxBytes uint64    `json:"transfer_rx_bytes" yaml:"transfer_rx_bytes"`
	TransferTxBytes uint64    `json:"transfer_tx_bytes" yaml:"transfer_tx_bytes"`
	KeepaliveSec    int       `json:"keepalive_sec" yaml:"keepalive_sec"`
}
