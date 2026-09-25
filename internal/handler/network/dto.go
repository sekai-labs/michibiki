package network

type OverviewFilter struct {
	Query string
}

type ConfigApplyInput struct {
	CandidateConfig string
	ConfirmSeconds  int
	Description     string
}
