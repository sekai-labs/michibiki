package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	BgBase    = lipgloss.Color("#0F172A")
	BgCard    = lipgloss.Color("#1E293B")
	BgOverlay = lipgloss.Color("#0A0F1D")
	BgActive  = lipgloss.Color("#334155")

	BorderInactive = lipgloss.Color("#334155")
	BorderActive   = lipgloss.Color("#38BDF8")
	BorderMuted    = lipgloss.Color("#1E293B")

	TextPrimary   = lipgloss.Color("#F8FAFC")
	TextSecondary = lipgloss.Color("#94A3B8")
	TextMuted     = lipgloss.Color("#64748B")
	TextDisabled  = lipgloss.Color("#475569")

	AccentCyan   = lipgloss.Color("#38BDF8")
	AccentViolet = lipgloss.Color("#A855F7")
	AccentGreen  = lipgloss.Color("#34D399")
	AccentAmber  = lipgloss.Color("#FBBF24")
	AccentRose   = lipgloss.Color("#F43F5E")

	ColorEmerald   = AccentGreen
	ColorCrimson   = AccentRose
	ColorAmber     = AccentAmber
	ColorCyan      = AccentCyan
	ColorPurple    = AccentViolet
	ColorGray      = TextMuted
	ColorDarkGray  = BorderInactive
	ColorLightGray = TextSecondary
	ColorWhite     = TextPrimary
	ColorBg        = BgBase
	ColorCardBg    = BgCard
	ColorAccent    = AccentCyan

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(BgBase).
			Background(AccentCyan).
			Padding(0, 1)

	StyleHeaderSub = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextSecondary).
			Background(BgCard).
			Padding(0, 1)

	StyleBadgeOnline = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#064E3B")).
				Background(AccentGreen).
				Padding(0, 1)

	StyleBadgeOffline = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFF1F2")).
				Background(AccentRose).
				Padding(0, 1)

	StyleBadgeWarning = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#78350F")).
				Background(AccentAmber).
				Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#082F49")).
			Background(AccentCyan).
			Padding(0, 1)

	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(BgBase).
			Background(AccentCyan).
			Padding(0, 1)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(TextSecondary).
				Background(BgCard).
				Padding(0, 1)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderInactive).
			Padding(0, 1)

	StyleCardActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderActive).
			Padding(0, 1)

	StyleCardAlert = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentRose).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan)

	StyleSubTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextSecondary)

	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(AccentCyan).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(BorderInactive)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(TextPrimary)

	StyleTableRowAlt = lipgloss.NewStyle().
				Foreground(TextSecondary)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Background(BgCard).
			Padding(0, 1)

	StyleStatusKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan)

	StyleProgressFilled = lipgloss.NewStyle().
				Foreground(AccentGreen)

	StyleProgressFilledWarn = lipgloss.NewStyle().
				Foreground(AccentAmber)

	StyleProgressFilledDanger = lipgloss.NewStyle().
					Foreground(AccentRose)

	StyleProgressEmpty = lipgloss.NewStyle().
				Foreground(BorderInactive)

	StyleFreeIP = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentGreen).
			Background(BgActive).
			Padding(0, 1)

	StyleMuted = lipgloss.NewStyle().
			Foreground(TextMuted)

	StyleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentRose)

	StyleDockSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(TextPrimary).
				Background(BgActive)

	StyleSubTabActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(BgBase).
				Background(AccentCyan).
				Padding(0, 1)

	StyleSubTabInactive = lipgloss.NewStyle().
				Foreground(TextSecondary).
				Background(BgCard).
				Padding(0, 1)

	StyleCursor = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan)

	StyleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderActive).
			Background(BgCard).
			Padding(1, 2)
)

func RenderProgressBar(pct float64, width int) string {
	if width < 4 {
		width = 4
	}
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
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

	var b strings.Builder
	b.WriteString(filledStyle.Render(filledStr))
	b.WriteString(StyleProgressEmpty.Render(emptyStr))
	return b.String()
}
