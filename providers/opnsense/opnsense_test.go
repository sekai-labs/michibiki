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
