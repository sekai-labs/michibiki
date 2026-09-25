package routeros

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestRouterOSRegistration(t *testing.T) {
	p, err := provider.Create("routeros")
	if err != nil {
		t.Fatalf("expected routeros registered: %v", err)
	}
	if p.ID() != "routeros" {
		t.Errorf("expected ID routeros, got %s", p.ID())
	}
}
