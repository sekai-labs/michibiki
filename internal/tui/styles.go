package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/michibiki/internal/tui/theme"
)

var (
	ColorEmerald   = theme.ColorEmerald
	ColorCrimson   = theme.ColorCrimson
	ColorAmber     = theme.ColorAmber
	ColorCyan      = theme.ColorCyan
	ColorPurple    = theme.ColorPurple
	ColorGray      = theme.ColorGray
	ColorDarkGray  = theme.ColorDarkGray
	ColorLightGray = theme.ColorLightGray
	ColorWhite     = theme.ColorWhite
	ColorBg        = theme.ColorBg
	ColorCardBg    = theme.ColorCardBg

	StyleHeader               = theme.StyleHeader
	StyleHeaderSub            = theme.StyleHeaderSub
	StyleBadgeOnline          = theme.StyleBadgeOnline
	StyleBadgeOffline         = theme.StyleBadgeOffline
	StyleBadgeWarning         = theme.StyleBadgeWarning
	StyleBadgeInfo            = theme.StyleBadgeInfo
	StyleTabActive            = theme.StyleTabActive
	StyleTabInactive          = theme.StyleTabInactive
	StyleCard                 = theme.StyleCard
	StyleCardActive           = theme.StyleCardActive
	StyleCardAlert            = theme.StyleCardAlert
	StyleTitle                = theme.StyleTitle
	StyleSubTitle             = theme.StyleSubTitle
	StyleTableHeader          = theme.StyleTableHeader
	StyleTableRow             = theme.StyleTableRow
	StyleTableRowAlt          = theme.StyleTableRowAlt
	StyleStatusBar            = theme.StyleStatusBar
	StyleStatusKey            = theme.StyleStatusKey
	StyleProgressFilled       = theme.StyleProgressFilled
	StyleProgressFilledWarn   = theme.StyleProgressFilledWarn
	StyleProgressFilledDanger = theme.StyleProgressFilledDanger
	StyleProgressEmpty        = theme.StyleProgressEmpty
	StyleFreeIP               = theme.StyleFreeIP
	StyleMuted                = theme.StyleMuted
	StyleError                = theme.StyleError
)

func RenderProgressBar(pct float64, width int) string {
	return theme.RenderProgressBar(pct, width)
}

func init() {
	_ = lipgloss.Color("")
}
