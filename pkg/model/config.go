package model

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
