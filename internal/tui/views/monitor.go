package views

import (
	"fmt"
	"strings"

	"github.com/sekai-labs/michibiki/internal/tui/theme"
	"github.com/sekai-labs/michibiki/pkg/model"
)

type MonitorData struct {
	Stats   []model.InterfaceStats
	History map[string][]float64
	Filter  string
	Error   error
}

var sparkChars = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func renderSparkline(values []float64, maxVal float64) string {
	if len(values) == 0 {
		return ""
	}
	if maxVal <= 0 {
		for _, v := range values {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	var sb strings.Builder
	for _, v := range values {
		ratio := v / maxVal
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		idx := int(ratio * float64(len(sparkChars)-1))
		sb.WriteRune(sparkChars[idx])
	}
	return sb.String()
}

func RenderMonitor(data MonitorData, width, height int) string {
	if data.Error != nil {
		return theme.StyleCardAlert.Width(width - 4).Render(
			theme.StyleError.Render("Failed to load traffic monitor: " + data.Error.Error()),
		)
	}

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	title := theme.StyleTitle.Render("REAL-TIME STREAMING TRAFFIC MONITOR (1s TICKER)")
	header := fmt.Sprintf("%-14s %-12s %-12s %-14s %-14s %-20s",
		"INTERFACE", "RX RATE", "TX RATE", "TOTAL RX", "TOTAL TX", "ACTIVITY SPARKLINE")
	headerRendered := theme.StyleTableHeader.Render(header)

	var rows []string
	q := strings.ToLower(data.Filter)
	for i, s := range data.Stats {
		if q != "" && !strings.Contains(strings.ToLower(s.InterfaceName), q) {
			continue
		}
		rxRate := formatRate(s.RxBps)
		txRate := formatRate(s.TxBps)
		rxTotal := formatBytes(s.RxBytes)
		txTotal := formatBytes(s.TxBytes)

		spark := ""
		if data.History != nil {
			hist := data.History[s.InterfaceName]
			if len(hist) > 0 {
				spark = renderSparkline(hist, 0)
			}
		}
		if spark == "" {
			spark = theme.StyleMuted.Render("collecting...")
		} else {
			spark = theme.StyleProgressFilled.Render(spark)
		}

		line := fmt.Sprintf("%-14s %-12s %-12s %-14s %-14s %-20s",
			s.InterfaceName,
			rxRate,
			txRate,
			rxTotal,
			txTotal,
			spark,
		)
		if i%2 == 0 {
			rows = append(rows, theme.StyleTableRow.Render(line))
		} else {
			rows = append(rows, theme.StyleTableRowAlt.Render(line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, theme.StyleMuted.Render("No interface statistics available."))
	}

	content := title + "\n\n" + headerRendered + "\n" + strings.Join(rows, "\n")
	return theme.StyleCard.Width(cardWidth).Render(content)
}

func formatRate(bps float64) string {
	if bps >= 1000000000 {
		return fmt.Sprintf("%.2f Gbps", bps/1000000000)
	}
	if bps >= 1000000 {
		return fmt.Sprintf("%.2f Mbps", bps/1000000)
	}
	if bps >= 1000 {
		return fmt.Sprintf("%.1f Kbps", bps/1000)
	}
	return fmt.Sprintf("%.0f bps", bps)
}
