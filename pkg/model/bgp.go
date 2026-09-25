package model

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
