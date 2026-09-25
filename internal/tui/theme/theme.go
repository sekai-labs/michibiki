package theme

import "github.com/charmbracelet/lipgloss"

var (
	ColorEmerald   = lipgloss.Color("#10B981")
	ColorCrimson   = lipgloss.Color("#EF4444")
	ColorAmber     = lipgloss.Color("#F59E0B")
	ColorCyan      = lipgloss.Color("#06B6D4")
	ColorPurple    = lipgloss.Color("#8B5CF6")
	ColorGray      = lipgloss.Color("#6B7280")
	ColorDarkGray  = lipgloss.Color("#374151")
	ColorLightGray = lipgloss.Color("#D1D5DB")
	ColorWhite     = lipgloss.Color("#F9FAFB")
	ColorBg        = lipgloss.Color("#111827")
	ColorCardBg    = lipgloss.Color("#1F2937")

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPurple).
			Padding(0, 1)

	StyleHeaderSub = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			Background(ColorDarkGray).
			Padding(0, 1)

	StyleBadgeOnline = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#064E3B")).
				Background(ColorEmerald).
				Padding(0, 1)

	StyleBadgeOffline = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7F1D1D")).
				Background(ColorCrimson).
				Padding(0, 1)

	StyleBadgeWarning = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#78350F")).
				Background(ColorAmber).
				Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#164E63")).
			Background(ColorCyan).
			Padding(0, 1)

	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPurple).
			Padding(0, 1)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColorLightGray).
				Background(ColorDarkGray).
				Padding(0, 1)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkGray).
			Padding(0, 1)

	StyleCardActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(0, 1)

	StyleCardAlert = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCrimson).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	StyleSubTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorLightGray)

	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorDarkGray)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorWhite)

	StyleTableRowAlt = lipgloss.NewStyle().
				Foreground(ColorLightGray)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			Background(ColorDarkGray).
			Padding(0, 1)

	StyleStatusKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	StyleProgressFilled = lipgloss.NewStyle().
				Foreground(ColorEmerald)

	StyleProgressFilledWarn = lipgloss.NewStyle().
				Foreground(ColorAmber)

	StyleProgressFilledDanger = lipgloss.NewStyle().
					Foreground(ColorCrimson)

	StyleProgressEmpty = lipgloss.NewStyle().
				Foreground(ColorDarkGray)

	StyleFreeIP = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorEmerald).
			Background(ColorDarkGray).
			Padding(0, 1)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCrimson)
)

func RenderProgressBar(pct float64, width int) string {
	if width < 3 {
		width = 10
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filledLen := int((pct / 100.0) * float64(width))
	if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	filledStr := ""
	for i := 0; i < filledLen; i++ {
		filledStr += "█"
	}
	emptyStr := ""
	for i := 0; i < emptyLen; i++ {
		emptyStr += "░"
	}

	filledStyle := StyleProgressFilled
	if pct >= 90 {
		filledStyle = StyleProgressFilledDanger
	} else if pct >= 75 {
		filledStyle = StyleProgressFilledWarn
	}

	return filledStyle.Render(filledStr) + StyleProgressEmpty.Render(emptyStr)
}
