package prompt

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

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

func AskCreateDir() (bool, error) {
	var answer bool
	prompt := &survey.Confirm{
		Message: "プロジェクトディレクトリを作成しますか?",
		Default: true,
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return false, err
	}
	return answer, nil
}

func AskProjectName(defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: "プロジェクト名:",
		Default: defaultVal,
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return answer, nil
}

func AskDomain(defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: "kintone ドメイン (例: example または example.cybozu.com):",
		Default: defaultVal,
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
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
	var answer int
	prompt := &survey.Input{
		Message: "アプリ ID:",
	}
	if defaultVal > 0 {
		prompt.Default = string(rune(defaultVal))
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return 0, err
	}
	return answer, nil
}

func AskFramework() (Framework, error) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	hiRed := color.New(color.FgHiRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	options := []string{
		cyan("React"),
		green("Vue"),
		hiRed("Svelte"),
		yellow("Vanilla"),
	}

	var answer string
	prompt := &survey.Select{
		Message: "フレームワーク:",
		Options: options,
		Default: options[0],
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return FrameworkReact, nil
	case options[1]:
		return FrameworkVue, nil
	case options[2]:
		return FrameworkSvelte, nil
	case options[3]:
		return FrameworkVanilla, nil
	}
	return FrameworkReact, nil
}

func AskLanguage() (Language, error) {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	options := []string{
		cyan("TypeScript"),
		yellow("JavaScript"),
	}

	var answer string
	prompt := &survey.Select{
		Message: "言語:",
		Options: options,
		Default: options[0],
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return LanguageTypeScript, nil
	case options[1]:
		return LanguageJavaScript, nil
	}
	return LanguageTypeScript, nil
}

func AskUsername() (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: "kintone ユーザー名:",
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return answer, nil
}

func AskPassword() (string, error) {
	var answer string
	prompt := &survey.Password{
		Message: "kintone パスワード:",
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return answer, nil
}

func AskPackageManager() (PackageManager, error) {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	options := []string{
		red("npm"),
		cyan("pnpm"),
		blue("yarn"),
		white("bun"),
	}

	var answer string
	prompt := &survey.Select{
		Message: "パッケージマネージャー:",
		Options: options,
		Default: options[0],
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return PackageManagerNpm, nil
	case options[1]:
		return PackageManagerPnpm, nil
	case options[2]:
		return PackageManagerYarn, nil
	case options[3]:
		return PackageManagerBun, nil
	}
	return PackageManagerNpm, nil
}

func AskTargets(defaultDesktop, defaultMobile bool) (desktop bool, mobile bool, err error) {
	options := []string{
		"デスクトップ",
		"モバイル",
	}

	defaults := []string{}
	if defaultDesktop {
		defaults = append(defaults, options[0])
	}
	if defaultMobile {
		defaults = append(defaults, options[1])
	}

	var answers []string
	prompt := &survey.MultiSelect{
		Message: "カスタマイズ対象:",
		Options: options,
		Default: defaults,
	}
	if err := survey.AskOne(prompt, &answers, survey.WithValidator(survey.MinItems(1))); err != nil {
		return false, false, err
	}

	for _, a := range answers {
		if a == options[0] {
			desktop = true
		}
		if a == options[1] {
			mobile = true
		}
	}

	return desktop, mobile, nil
}

func AskScope(defaultScope Scope) (Scope, error) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	options := []string{
		green("すべてのユーザーに適用 (ALL)"),
		yellow("アプリ管理者のみに適用 (ADMIN)"),
		red("適用しない (NONE)"),
	}

	defaultIndex := 0
	switch defaultScope {
	case ScopeAdmin:
		defaultIndex = 1
	case ScopeNone:
		defaultIndex = 2
	}

	var answer string
	prompt := &survey.Select{
		Message: "カスタマイズの適用範囲:",
		Options: options,
		Default: options[defaultIndex],
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return ScopeAll, nil
	case options[1]:
		return ScopeAdmin, nil
	case options[2]:
		return ScopeNone, nil
	}
	return ScopeAll, nil
}

func AskOutput(defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: "出力ファイル名 (拡張子なし):",
		Default: defaultVal,
		Help:    "ビルド時に生成されるファイル名を指定します (例: customize → customize.js, customize.css)",
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return answer, nil
}
