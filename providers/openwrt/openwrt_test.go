package openwrt

import (
	"testing"

	"github.com/sekai-labs/michibiki/pkg/provider"
)

func TestOpenWrtRegistration(t *testing.T) {
	p, err := provider.Create("openwrt")
	if err != nil {
		t.Fatalf("expected openwrt registered: %v", err)
	}
	if p.ID() != "openwrt" {
		t.Errorf("expected ID openwrt, got %s", p.ID())
	}
}
