package prompt

import (
	"errors"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var errRequired = errors.New("入力必須です")

type Framework string

const (
	FrameworkReact   Framework = "react"
	FrameworkVue     Framework = "vue"
	FrameworkSvelte  Framework = "svelte"
	FrameworkVanilla Framework = "vanilla"
)

type Language string

const (
	LanguageTypeScript Language = "typescript"
	LanguageJavaScript Language = "javascript"
)

type PackageManager string

const (
	PackageManagerNpm  PackageManager = "npm"
	PackageManagerPnpm PackageManager = "pnpm"
	PackageManagerYarn PackageManager = "yarn"
	PackageManagerBun  PackageManager = "bun"
)

type Scope string

const (
	ScopeAll   Scope = "ALL"
	ScopeAdmin Scope = "ADMIN"
	ScopeNone  Scope = "NONE"
)

type InitAnswers struct {
	ProjectName    string
	CreateDir      bool
	Domain         string
	AppID          int
	Framework      Framework
	Language       Language
	Username       string
	Password       string
	PackageManager PackageManager
	TargetDesktop  bool
	TargetMobile   bool
	Scope          Scope
	Output         string
}

// カラー定義
var (
	colorCyan   = lipgloss.Color("39")
	colorGreen  = lipgloss.Color("42")
	colorYellow = lipgloss.Color("214")
	colorRed    = lipgloss.Color("196")
	colorOrange = lipgloss.Color("208")
	colorBlue   = lipgloss.Color("33")
	colorWhite  = lipgloss.Color("255")
)

func newForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(huh.ThemeCatppuccin())
}

func AskCreateDir() (bool, error) {
	var answer bool
	err := newForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("プロジェクトディレクトリを作成しますか?").
				Affirmative("はい").
				Negative("いいえ").
				Value(&answer),
		),
	).Run()
	if err != nil {
		return false, err
	}
	return answer, nil
}

func AskProjectName(defaultVal string) (string, error) {
	var answer string
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("プロジェクト名").
				Value(&answer).
				Placeholder(defaultVal).
				Validate(func(s string) error {
					if s == "" {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}
	if answer == "" {
		answer = defaultVal
	}
	return answer, nil
}

func AskDomain(defaultVal string) (string, error) {
	var answer string
	placeholder := "example または example.cybozu.com"
	if defaultVal != "" {
		placeholder = defaultVal
	}
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("kintone ドメイン").
				Description("例: example または example.cybozu.com").
				Value(&answer).
				Placeholder(placeholder).
				Validate(func(s string) error {
					if s == "" && defaultVal == "" {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}
	if answer == "" {
		answer = defaultVal
	}
	return CompleteDomain(answer), nil
}

// CompleteDomain はサブドメインのみの入力を完全なドメインに補完する
func CompleteDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if !strings.Contains(domain, ".") {
		return domain + ".cybozu.com"
	}
	return domain
}

func AskAppID(defaultVal int) (int, error) {
	var answer string
	placeholder := ""
	if defaultVal > 0 {
		placeholder = strconv.Itoa(defaultVal)
	}
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("アプリ ID").
				Value(&answer).
				Placeholder(placeholder).
				Validate(func(s string) error {
					if s == "" && defaultVal == 0 {
						return errRequired
					}
					if s != "" {
						if _, err := strconv.Atoi(s); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return 0, err
	}
	if answer == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(answer)
}

func AskFramework() (Framework, error) {
	return AskFrameworkExcept("")
}

func AskFrameworkExcept(exclude Framework) (Framework, error) {
	cyanStyle := lipgloss.NewStyle().Foreground(colorCyan)
	greenStyle := lipgloss.NewStyle().Foreground(colorGreen)
	orangeStyle := lipgloss.NewStyle().Foreground(colorOrange)
	yellowStyle := lipgloss.NewStyle().Foreground(colorYellow)

	var options []huh.Option[Framework]
	if exclude != FrameworkReact {
		options = append(options, huh.NewOption(cyanStyle.Render("React"), FrameworkReact))
	}
	if exclude != FrameworkVue {
		options = append(options, huh.NewOption(greenStyle.Render("Vue"), FrameworkVue))
	}
	if exclude != FrameworkSvelte {
		options = append(options, huh.NewOption(orangeStyle.Render("Svelte"), FrameworkSvelte))
	}
	if exclude != FrameworkVanilla {
		options = append(options, huh.NewOption(yellowStyle.Render("Vanilla"), FrameworkVanilla))
	}

	var answer Framework
	err := newForm(
		huh.NewGroup(
			huh.NewSelect[Framework]().
				Title("フレームワーク").
				Options(options...).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskLanguage() (Language, error) {
	cyanStyle := lipgloss.NewStyle().Foreground(colorCyan)
	yellowStyle := lipgloss.NewStyle().Foreground(colorYellow)

	var answer Language
	err := newForm(
		huh.NewGroup(
			huh.NewSelect[Language]().
				Title("言語").
				Options(
					huh.NewOption(cyanStyle.Render("TypeScript"), LanguageTypeScript),
					huh.NewOption(yellowStyle.Render("JavaScript"), LanguageJavaScript),
				).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskUsername() (string, error) {
	var answer string
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("kintone ユーザー名").
				Value(&answer).
				Validate(func(s string) error {
					if s == "" {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskPassword() (string, error) {
	var answer string
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("kintone パスワード").
				EchoMode(huh.EchoModePassword).
				Value(&answer).
				Validate(func(s string) error {
					if s == "" {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskPackageManager() (PackageManager, error) {
	redStyle := lipgloss.NewStyle().Foreground(colorRed)
	cyanStyle := lipgloss.NewStyle().Foreground(colorCyan)
	blueStyle := lipgloss.NewStyle().Foreground(colorBlue)
	whiteStyle := lipgloss.NewStyle().Foreground(colorWhite)

	var answer PackageManager
	err := newForm(
		huh.NewGroup(
			huh.NewSelect[PackageManager]().
				Title("パッケージマネージャー").
				Options(
					huh.NewOption(redStyle.Render("npm"), PackageManagerNpm),
					huh.NewOption(cyanStyle.Render("pnpm"), PackageManagerPnpm),
					huh.NewOption(blueStyle.Render("yarn"), PackageManagerYarn),
					huh.NewOption(whiteStyle.Render("bun"), PackageManagerBun),
				).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskTargets(defaultDesktop, defaultMobile bool) (desktop bool, mobile bool, err error) {
	var answers []string
	defaults := []string{}
	if defaultDesktop {
		defaults = append(defaults, "desktop")
	}
	if defaultMobile {
		defaults = append(defaults, "mobile")
	}

	err = newForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("カスタマイズ対象").
				Options(
					huh.NewOption("デスクトップ", "desktop").Selected(defaultDesktop),
					huh.NewOption("モバイル", "mobile").Selected(defaultMobile),
				).
				Value(&answers).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return false, false, err
	}

	for _, a := range answers {
		if a == "desktop" {
			desktop = true
		}
		if a == "mobile" {
			mobile = true
		}
	}

	return desktop, mobile, nil
}

func AskScope(defaultScope Scope) (Scope, error) {
	greenStyle := lipgloss.NewStyle().Foreground(colorGreen)
	yellowStyle := lipgloss.NewStyle().Foreground(colorYellow)
	redStyle := lipgloss.NewStyle().Foreground(colorRed)

	var answer Scope
	err := newForm(
		huh.NewGroup(
			huh.NewSelect[Scope]().
				Title("カスタマイズの適用範囲").
				Options(
					huh.NewOption(greenStyle.Render("すべてのユーザーに適用 (ALL)"), ScopeAll),
					huh.NewOption(yellowStyle.Render("アプリ管理者のみに適用 (ADMIN)"), ScopeAdmin),
					huh.NewOption(redStyle.Render("適用しない (NONE)"), ScopeNone),
				).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func AskOutput(defaultVal string) (string, error) {
	var answer string
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title("出力ファイル名 (拡張子なし)").
				Description("ビルド時に生成されるファイル名 (例: customize → customize.js)").
				Value(&answer).
				Placeholder(defaultVal).
				Validate(func(s string) error {
					if s == "" && defaultVal == "" {
						return errRequired
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}
	if answer == "" {
		answer = defaultVal
	}
	return answer, nil
}
