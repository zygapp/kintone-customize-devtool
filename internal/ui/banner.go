package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var asciiArt = []string{
	` ██╗  ██╗  ██████╗ ██████╗  ███████╗ ██╗   ██╗`,
	` ██║ ██╔╝ ██╔════╝ ██╔══██╗ ██╔════╝ ██║   ██║`,
	` █████╔╝  ██║      ██║  ██║ █████╗   ██║   ██║`,
	` ██╔═██╗  ██║      ██║  ██║ ██╔══╝   ╚██╗ ██╔╝`,
	` ██║  ██╗ ╚██████╗ ██████╔╝ ███████╗  ╚████╔╝ `,
	` ╚═╝  ╚═╝  ╚═════╝ ╚═════╝  ╚══════╝   ╚═══╝  `,
}

var bannerGradient = []string{
	"#FFD54F",
	"#FFCA28",
	"#FFBF00",
	"#FFB300",
	"#FFA000",
	"#FF8F00",
}

var bannerTags = []string{
	"kintone customize devtool",
	"adaptive setup engine",
	"ready.",
}

// Banner はASCIIアートバナーを表示
func Banner() {
	// AA行の最大表示幅を取得
	maxWidth := 0
	for _, line := range asciiArt {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}

	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFBF00"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	tagOffset := len(asciiArt) - len(bannerTags)

	fmt.Println()
	for i, line := range asciiArt {
		artStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(bannerGradient[i]))
		rendered := artStyle.Render(line)

		if tagIdx := i - tagOffset; tagIdx >= 0 && tagIdx < len(bannerTags) {
			pad := strings.Repeat(" ", maxWidth-lipgloss.Width(line))
			tag := fmt.Sprintf("  %s %s", accent.Render("::"), muted.Render(bannerTags[tagIdx]))
			fmt.Println(rendered + pad + tag)
		} else {
			fmt.Println(rendered)
		}
	}
	fmt.Println()
}
