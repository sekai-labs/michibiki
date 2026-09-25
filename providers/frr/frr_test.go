package frr

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestFRRRegistration(t *testing.T) {
	p, err := provider.Create("frr")
	if err != nil {
		t.Fatalf("expected frr registered: %v", err)
	}
	if p.ID() != "frr" {
		t.Errorf("expected ID frr, got %s", p.ID())
	}
}
