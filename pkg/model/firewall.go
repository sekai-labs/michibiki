package model

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
