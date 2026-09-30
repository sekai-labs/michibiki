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
func TestEscapeShellArg(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "simple command",
			input:    "show running-config",
			expected: "'show running-config'",
		},
		{
			name:     "command with single quotes",
			input:    "show bgp 'test'",
			expected: "'show bgp '\\''test'\\'''",
		},
		{
			name:     "command substitution attempt dollar paren",
			input:    "show $(reboot)",
			expected: "'show $(reboot)'",
		},
		{
			name:     "command substitution attempt backticks",
			input:    "show `reboot`",
			expected: "'show `reboot`'",
		},
		{
			name:     "command with double quotes and variable",
			input:    `show "hello $USER"`,
			expected: `'show "hello $USER"'`,
		},
		{
			name:      "command with null byte rejected",
			input:     "show\x00reboot",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := escapeShellArg(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, res)
			}
		})
	}
}

func TestFRRConnect_HostKeyVerification(t *testing.T) {
	p := New()
	err := p.Connect(t.Context(), "127.0.0.1:54321", nil, nil)
	if err == nil {
		t.Fatal("expected Connect to fail")
	}
}
