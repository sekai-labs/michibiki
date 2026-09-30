package opnsense

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestOpnsenseRegistration(t *testing.T) {
	p, err := provider.Create("opnsense")
	if err != nil {
		t.Fatalf("expected opnsense registered: %v", err)
	}
	if p.ID() != "opnsense" {
		t.Errorf("expected ID opnsense, got %s", p.ID())
	}
}
func TestRollbackConfig_Validation(t *testing.T) {
	p := New()
	badIDs := []string{
		"../foo",
		"id?param=value",
		"id#frag",
		"id/something",
		"id;rm -rf",
		"id with space",
		"",
	}

	for _, badID := range badIDs {
		err := p.RollbackConfig(t.Context(), badID)
		if err == nil {
			t.Errorf("expected error for bad rollback ID %q, got nil", badID)
		}
	}
}
