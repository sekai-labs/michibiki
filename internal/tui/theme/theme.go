package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorEmerald   = lipgloss.Color("#73daca")
	ColorCrimson   = lipgloss.Color("#f7768e")
	ColorAmber     = lipgloss.Color("#e0af68")
	ColorCyan      = lipgloss.Color("#7dcfff")
	ColorPurple    = lipgloss.Color("#bb9af7")
	ColorGray      = lipgloss.Color("#565f89")
	ColorDarkGray  = lipgloss.Color("#24283b")
	ColorLightGray = lipgloss.Color("#c0caf5")
	ColorWhite     = lipgloss.Color("#ffffff")
	ColorBg        = lipgloss.Color("#1a1b26")
	ColorCardBg    = lipgloss.Color("#1f2335")
	ColorAccent    = lipgloss.Color("#7aa2f7")

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
			Padding(0, 1)

	StyleHeaderSub = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			Background(ColorDarkGray).
			Padding(0, 1)

	StyleBadgeOnline = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorEmerald).
				Padding(0, 1)

	StyleBadgeOffline = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorCrimson).
				Padding(0, 1)

	StyleBadgeWarning = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorAmber).
				Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorCyan).
			Padding(0, 1)

	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
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
			BorderForeground(ColorAccent).
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
				Foreground(ColorAccent).
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

	StyleDockSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorDarkGray)

	StyleSubTabActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorAccent).
				Padding(0, 1)

	StyleSubTabInactive = lipgloss.NewStyle().
				Foreground(ColorLightGray).
				Background(ColorDarkGray).
				Padding(0, 1)

	StyleCursor = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	StyleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Background(ColorBg).
			Padding(1, 2)
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

	filledStr := strings.Repeat("█", filledLen)
	emptyStr := strings.Repeat("░", emptyLen)

	filledStyle := StyleProgressFilled
	if pct >= 90 {
		filledStyle = StyleProgressFilledDanger
	} else if pct >= 75 {
		filledStyle = StyleProgressFilledWarn
	}

	return filledStyle.Render(filledStr) + StyleProgressEmpty.Render(emptyStr)
}
