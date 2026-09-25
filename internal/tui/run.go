package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/michibiki/pkg/provider"
)

func Run(prov provider.Provider, deviceName string) error {
	m := NewModel(prov, deviceName)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
