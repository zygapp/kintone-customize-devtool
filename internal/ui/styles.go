package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

// カラー定義
var (
	ColorGreen  = lipgloss.Color("42")
	ColorRed    = lipgloss.Color("196")
	ColorYellow = lipgloss.Color("214")
	ColorCyan   = lipgloss.Color("39")
	ColorGray   = lipgloss.Color("245")
)

// スタイル定義
var (
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorRed)
	WarnStyle    = lipgloss.NewStyle().Foreground(ColorYellow)
	InfoStyle    = lipgloss.NewStyle().Foreground(ColorCyan)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorGray)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)
)

// アイコン
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarn    = "⚠"
	IconInfo    = "→"
)

// Success は成功メッセージを表示
func Success(msg string) {
	fmt.Println(SuccessStyle.Render(IconSuccess) + " " + msg)
}

// Error はエラーメッセージを表示
func Error(msg string) {
	fmt.Println(ErrorStyle.Render(IconError) + " " + msg)
}

// Warn は警告メッセージを表示
func Warn(msg string) {
	fmt.Println(WarnStyle.Render(IconWarn) + " " + msg)
}

// Info は情報メッセージを表示
func Info(msg string) {
	fmt.Println(InfoStyle.Render(IconInfo) + " " + msg)
}

// Title はタイトルを表示
func Title(msg string) {
	fmt.Println(TitleStyle.Render(msg))
}

// Box はボックスで囲んだコンテンツを表示
func Box(content string) {
	fmt.Println(BoxStyle.Render(content))
}

// Spinner はスピナーを実行
func Spinner(title string, action func()) error {
	return spinner.New().
		Title(title).
		Action(action).
		Run()
}

// SpinnerWithResult はスピナーを実行し、エラーを返す
func SpinnerWithResult(title string, action func() error) error {
	var actionErr error
	err := spinner.New().
		Title(title).
		Action(func() {
			actionErr = action()
		}).
		Run()
	if err != nil {
		return err
	}
	return actionErr
}

// NewForm はカスタムテーマ付きのフォームを作成
func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(huh.ThemeCatppuccin())
}
