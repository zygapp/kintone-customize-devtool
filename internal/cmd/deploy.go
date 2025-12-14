package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/kintone"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "kintoneにデプロイ",
	Long:  `ビルド成果物をkintoneにアップロードしてデプロイします。`,
	RunE:  runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
	}

	envCfg, _ := config.LoadEnv(projectDir)

	username := cfg.Kintone.Auth.Username
	password := cfg.Kintone.Auth.Password

	if envCfg != nil && envCfg.HasAuth() {
		username = envCfg.Username
		password = envCfg.Password
	}

	if username == "" || password == "" {
		return fmt.Errorf("認証情報が見つかりません。.env または .kcdev/config.json に設定してください")
	}

	jsPath := filepath.Join(projectDir, "dist", "customize.js")
	if _, err := os.Stat(jsPath); err != nil {
		return fmt.Errorf("ビルド成果物が見つかりません。先に kcdev build を実行してください")
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// ターゲット表示用の文字列を生成
	var targets []string
	if cfg.Targets.Desktop {
		targets = append(targets, "Desktop")
	}
	if cfg.Targets.Mobile {
		targets = append(targets, "Mobile")
	}
	if len(targets) == 0 {
		targets = append(targets, "Desktop") // デフォルト
	}

	fmt.Printf("\n%s デプロイ中... (%s, App:%d, %s)\n", cyan("→"), cfg.Kintone.Domain, cfg.Kintone.AppID, strings.Join(targets, "+"))

	client := kintone.NewClient(cfg.Kintone.Domain, username, password)

	cssPath := filepath.Join(projectDir, "dist", "customize.css")
	hasCss := false
	if _, err := os.Stat(cssPath); err == nil {
		hasCss = true
	}

	var desktopFiles *kintone.CustomizeFiles
	var mobileFiles *kintone.CustomizeFiles

	// デスクトップ用ファイルをアップロード
	if cfg.Targets.Desktop {
		fmt.Printf("  Desktop JS...")
		jsKey, err := client.UploadFile(jsPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("JSファイルアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", green("✓"))

		desktopFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

		if hasCss {
			fmt.Printf("  Desktop CSS...")
			cssKey, err := client.UploadFile(cssPath)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("CSSファイルアップロードエラー: %w", err)
			}
			fmt.Printf(" %s\n", green("✓"))
			desktopFiles.CSSFileKey = cssKey
		}
	}

	// モバイル用ファイルをアップロード
	if cfg.Targets.Mobile {
		fmt.Printf("  Mobile JS...")
		jsKey, err := client.UploadFile(jsPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("JSファイルアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", green("✓"))

		mobileFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

		if hasCss {
			fmt.Printf("  Mobile CSS...")
			cssKey, err := client.UploadFile(cssPath)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("CSSファイルアップロードエラー: %w", err)
			}
			fmt.Printf(" %s\n", green("✓"))
			mobileFiles.CSSFileKey = cssKey
		}
	}

	// カスタマイズ設定を更新
	fmt.Printf("  設定...")
	if err := client.UpdateCustomize(cfg.Kintone.AppID, desktopFiles, mobileFiles); err != nil {
		fmt.Println()
		return fmt.Errorf("カスタマイズ設定エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// アプリをデプロイ
	fmt.Printf("  デプロイ...")
	if err := client.DeployApp(cfg.Kintone.AppID); err != nil {
		fmt.Println()
		return fmt.Errorf("デプロイ開始エラー: %w", err)
	}

	if err := client.WaitForDeploy(cfg.Kintone.AppID); err != nil {
		fmt.Println()
		return fmt.Errorf("デプロイ待機エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	fmt.Printf("\n%s 完了! https://%s/k/%d/\n\n", green("✓"), cfg.Kintone.Domain, cfg.Kintone.AppID)

	return nil
}
