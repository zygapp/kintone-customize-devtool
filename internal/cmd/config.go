package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/prompt"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "プロジェクト設定を変更",
	Long:  `対話形式でプロジェクトの各種設定を変更します。`,
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。先に kcdev init を実行してください: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()

	for {
		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s 設定メニュー\n\n", cyan("⚙"))

		action, err := askConfigAction()
		if err != nil {
			return err
		}

		switch action {
		case "view":
			showCurrentConfig(cfg)
		case "kintone":
			if err := editKintoneConfig(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "targets":
			if err := editTargets(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "scope":
			if err := editScope(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "output":
			if err := editOutput(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "exit":
			fmt.Println("\n設定を終了します。")
			return nil
		}
	}
}

func askConfigAction() (string, error) {
	options := []string{
		"現在の設定を表示",
		"kintone接続設定（ドメイン、アプリID、認証）",
		"ターゲット（デスクトップ/モバイル）の設定",
		"適用範囲の設定",
		"出力ファイル名の設定",
		"終了",
	}

	var answer string
	prompt := &survey.Select{
		Message: "操作を選択してください:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return "view", nil
	case options[1]:
		return "kintone", nil
	case options[2]:
		return "targets", nil
	case options[3]:
		return "scope", nil
	case options[4]:
		return "output", nil
	default:
		return "exit", nil
	}
}

func showCurrentConfig(cfg *config.Config) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\n%s 現在の設定\n\n", cyan("📋"))

	// kintone設定
	fmt.Printf("%s\n", cyan("kintone:"))
	fmt.Printf("  ドメイン: %s\n", cfg.Kintone.Domain)
	fmt.Printf("  アプリID: %d\n", cfg.Kintone.AppID)
	if cfg.Kintone.Auth.Username != "" {
		fmt.Printf("  ユーザー: %s\n", cfg.Kintone.Auth.Username)
		fmt.Printf("  パスワード: %s\n", "********")
	} else {
		fmt.Printf("  認証: %s\n", yellow("未設定"))
	}

	// ターゲット
	fmt.Printf("\n%s\n", cyan("ターゲット:"))
	if cfg.Targets.Desktop {
		fmt.Printf("  %s デスクトップ\n", green("✓"))
	} else {
		fmt.Printf("  ✗ デスクトップ\n")
	}
	if cfg.Targets.Mobile {
		fmt.Printf("  %s モバイル\n", green("✓"))
	} else {
		fmt.Printf("  ✗ モバイル\n")
	}

	// 適用範囲
	fmt.Printf("\n%s\n", cyan("適用範囲:"))
	switch cfg.Scope {
	case config.ScopeAll:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", green("✓"))
	case config.ScopeAdmin:
		fmt.Printf("  %s アプリ管理者のみ (ADMIN)\n", yellow("✓"))
	case config.ScopeNone:
		fmt.Printf("  ✗ 適用しない (NONE)\n")
	default:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", green("✓"))
	}

	// 出力ファイル名
	fmt.Printf("\n%s\n", cyan("出力:"))
	fmt.Printf("  ファイル名: %s.js / %s.css\n", cfg.GetOutputName(), cfg.GetOutputName())

	// Dev設定
	fmt.Printf("\n%s\n", cyan("開発サーバー:"))
	fmt.Printf("  オリジン:     %s\n", cfg.Dev.Origin)
	fmt.Printf("  エントリー:   %s\n", cfg.Dev.Entry)

	fmt.Println()
	fmt.Println("Enterキーで戻る...")
	fmt.Scanln()
}

func editKintoneConfig(cfg *config.Config) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n%s kintone接続設定\n\n", cyan("🔧"))

	// ドメイン
	domain, err := prompt.AskDomain(cfg.Kintone.Domain)
	if err != nil {
		return err
	}
	cfg.Kintone.Domain = domain

	// アプリID
	var appIDStr string
	appIDPrompt := &survey.Input{
		Message: "アプリID:",
		Default: strconv.Itoa(cfg.Kintone.AppID),
	}
	if err := survey.AskOne(appIDPrompt, &appIDStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return fmt.Errorf("アプリIDは数値で入力してください")
	}
	cfg.Kintone.AppID = appID

	// 認証情報を更新するか確認
	var updateAuth bool
	authPrompt := &survey.Confirm{
		Message: "認証情報を更新しますか?",
		Default: false,
	}
	if err := survey.AskOne(authPrompt, &updateAuth); err != nil {
		return err
	}

	if updateAuth {
		username, err := prompt.AskUsername()
		if err != nil {
			return err
		}
		password, err := prompt.AskPassword()
		if err != nil {
			return err
		}
		cfg.Kintone.Auth.Username = username
		cfg.Kintone.Auth.Password = password
	}

	fmt.Printf("\n%s kintone接続設定を更新しました\n", green("✓"))
	return nil
}

func editTargets(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()

	desktop, mobile, err := prompt.AskTargets(cfg.Targets.Desktop, cfg.Targets.Mobile)
	if err != nil {
		return err
	}

	cfg.Targets.Desktop = desktop
	cfg.Targets.Mobile = mobile

	fmt.Printf("\n%s ターゲットを更新しました\n", green("✓"))
	return nil
}

func editScope(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()

	currentScope := prompt.Scope(cfg.Scope)
	if currentScope == "" {
		currentScope = prompt.ScopeAll
	}

	scope, err := prompt.AskScope(currentScope)
	if err != nil {
		return err
	}

	cfg.Scope = string(scope)

	fmt.Printf("\n%s 適用範囲を更新しました\n", green("✓"))
	return nil
}

func editOutput(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()

	output, err := prompt.AskOutput(cfg.GetOutputName())
	if err != nil {
		return err
	}

	cfg.Output = output

	fmt.Printf("\n%s 出力ファイル名を更新しました (%s.js / %s.css)\n", green("✓"), output, output)
	return nil
}
