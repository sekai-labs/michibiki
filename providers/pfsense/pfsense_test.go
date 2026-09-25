package pfsense

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestPfSenseRegistration(t *testing.T) {
	p, err := provider.Create("pfsense")
	if err != nil {
		t.Fatalf("expected pfsense registered: %v", err)
	}
	if p.ID() != "pfsense" {
		t.Errorf("expected ID pfsense, got %s", p.ID())
	}
}
