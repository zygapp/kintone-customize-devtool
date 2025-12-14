package prompt

import (
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
		Message: "kintone ドメイン (例: example.cybozu.com):",
		Default: defaultVal,
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return answer, nil
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
