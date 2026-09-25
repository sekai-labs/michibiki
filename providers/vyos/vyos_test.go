package vyos

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestVyOSRegistration(t *testing.T) {
	p, err := provider.Create("vyos")
	if err != nil {
		t.Fatalf("expected vyos registered: %v", err)
	}
	if p.ID() != "vyos" {
		t.Errorf("expected ID vyos, got %s", p.ID())
	}
}
