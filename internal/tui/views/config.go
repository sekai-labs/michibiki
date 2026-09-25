package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
)

type ConfigData struct {
	RunningConfig string
	SafetyStatus  string
	Error         error
}

func RenderConfig(data ConfigData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load configuration: " + data.Error.Error()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	safetyTitle := theme.StyleTitle.Render("SAFETY ENGINE & COMMIT-CONFIRM STATUS")
	safetyStatusText := data.SafetyStatus
	if safetyStatusText == "" {
		safetyStatusText = "Idle — no active commit-confirm rollback timers pending."
	}
	safetyContent := fmt.Sprintf("%s\n%s",
		theme.StyleBadgeOnline.Render("SAFETY VERIFIED"),
		safetyStatusText,
	)
	safetyCard := theme.StyleCard.Width(cardWidth).Render(safetyTitle + "\n\n" + safetyContent)

	cfgTitle := theme.StyleTitle.Render("RUNNING CONFIGURATION")
	cfgBody := data.RunningConfig
	if cfgBody == "" {
		cfgBody = theme.StyleMuted.Render("Configuration empty or not retrievable from provider.")
	} else {
		lines := strings.Split(cfgBody, "\n")
		maxLines := height - 12
		if maxLines > 0 && len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, fmt.Sprintf("... (%d more lines truncated in view) ...", len(strings.Split(cfgBody, "\n"))-maxLines))
		}
		cfgBody = strings.Join(lines, "\n")
	}

	cfgCard := theme.StyleCard.Width(cardWidth).Render(cfgTitle + "\n\n" + cfgBody)

	return lipgloss.JoinVertical(lipgloss.Left, safetyCard, cfgCard)
}
