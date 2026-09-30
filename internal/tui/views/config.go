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
	Filter        string
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
	var scb strings.Builder
	scb.WriteString(theme.StyleBadgeOnline.Render("SAFETY VERIFIED"))
	scb.WriteByte('\n')
	scb.WriteString(safetyStatusText)

	var sb strings.Builder
	sb.WriteString(safetyTitle)
	sb.WriteString("\n\n")
	sb.WriteString(scb.String())
	safetyCard := theme.StyleCard.Width(cardWidth).Render(sb.String())

	cfgTitle := theme.StyleTitle.Render("RUNNING CONFIGURATION")
	cfgBody := data.RunningConfig
	if cfgBody == "" {
		cfgBody = theme.StyleMuted.Render("Configuration empty or not retrievable from provider.")
	} else {
		lines := strings.Split(cfgBody, "\n")
		if data.Filter != "" {
			q := strings.ToLower(data.Filter)
			var matched []string
			for _, l := range lines {
				if strings.Contains(strings.ToLower(l), q) {
					matched = append(matched, l)
				}
			}
			lines = matched
		}
		maxLines := height - 12
		if maxLines > 0 && len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, fmt.Sprintf("... (%d more lines truncated in view) ...", len(strings.Split(cfgBody, "\n"))-maxLines))
		}
		var linesB strings.Builder
		for i, l := range lines {
			if i > 0 {
				linesB.WriteByte('\n')
			}
			linesB.WriteString(l)
		}
		cfgBody = linesB.String()
	}

	var cb strings.Builder
	cb.WriteString(cfgTitle)
	cb.WriteString("\n\n")
	cb.WriteString(cfgBody)
	cfgCard := theme.StyleCard.Width(cardWidth).Render(cb.String())
	return lipgloss.JoinVertical(lipgloss.Left, safetyCard, cfgCard)
}
