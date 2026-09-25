package domain

import "github.com/sekai-labs/michibiki/pkg/model"

type SecurityOverview struct {
	FirewallRules []model.FirewallRule
	Aliases       []model.FirewallAlias
	NATRules      []model.NATRule
	TotalRules    int
	EnabledRules  int
	TotalNATRules int
}

type ServicesOverview struct {
	DHCPLeases     []model.DHCPLease
	WireGuardPeers []model.WireGuardPeer
	ARPEntries     []model.ARPEntry
	ActiveLeases   int
	ActiveVPNPeers int
}

type ConfigState struct {
	RunningConfig  string
	ConfirmPending bool
	RollbackID     string
	Message        string
}
