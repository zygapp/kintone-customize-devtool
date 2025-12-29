package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/generator"
	"github.com/kintone/kcdev/internal/kintone"
	"github.com/spf13/cobra"
)

var skipDeploy bool
var noBrowser bool
var forceDevOverwrite bool
var previewOnlyDev bool

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "開発サーバーを起動",
	Long:  `ローダーをkintoneにデプロイし、Vite dev server を起動します。`,
	RunE:  runDev,
}

func init() {
	devCmd.Flags().BoolVar(&skipDeploy, "skip-deploy", false, "ローダーのデプロイをスキップ")
	devCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "ブラウザを自動で開かない")
	devCmd.Flags().BoolVarP(&forceDevOverwrite, "force", "f", false, "既存カスタマイズを確認せず上書き")
	devCmd.Flags().BoolVarP(&previewOnlyDev, "preview", "p", false, "プレビュー環境のみにデプロイ（本番反映しない）")
}

func runDev(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
	}

	if !generator.CertsExist(projectDir) {
		return fmt.Errorf("証明書が見つかりません。kcdev init を実行してください")
	}

	// 認証情報取得
	username := cfg.Kintone.Auth.Username
	password := cfg.Kintone.Auth.Password

	envCfg, _ := config.LoadEnv(projectDir)
	if envCfg != nil && envCfg.HasAuth() {
		username = envCfg.Username
		password = envCfg.Password
	}

	if username == "" || password == "" {
		return fmt.Errorf("認証情報が見つかりません。.env または .kcdev/config.json に設定してください")
	}

	// デプロイ
	if !skipDeploy {
		if err := deployLoader(projectDir, cfg, username, password, forceDevOverwrite, previewOnlyDev); err != nil {
			return err
		}
	}

	printDevInfo(cfg)

	viteConfig := filepath.Join(projectDir, config.ConfigDir, "vite.config.ts")
	if _, err := os.Stat(filepath.Join(projectDir, "vite.config.ts")); err == nil {
		viteConfig = filepath.Join(projectDir, "vite.config.ts")
	}

	viteCmd := exec.Command("npx", "vite", "--config", viteConfig, "--logLevel", "warn", "--clearScreen", "false")
	viteCmd.Dir = projectDir
	viteCmd.Stdout = os.Stdout
	viteCmd.Stderr = os.Stderr
	viteCmd.Stdin = os.Stdin

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if err := viteCmd.Start(); err != nil {
		return fmt.Errorf("Vite起動エラー: %w", err)
	}

	// ブラウザを自動で開く（localhost:3000でSSL許可後、kintoneにリダイレクト）
	if !noBrowser {
		go func() {
			time.Sleep(2 * time.Second) // Viteの起動を待つ
			openBrowser("https://localhost:3000")
		}()
	}

	go func() {
		<-sigChan
		if viteCmd.Process != nil {
			viteCmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	return viteCmd.Wait()
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func deployLoader(projectDir string, cfg *config.Config, username, password string, force bool, previewOnly bool) error {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if previewOnly {
		fmt.Printf("\n%s ローダーをkintoneプレビュー環境にデプロイ中...\n", cyan("→"))
	} else {
		fmt.Printf("\n%s ローダーをkintoneにデプロイ中...\n", cyan("→"))
	}

	client := kintone.NewClient(cfg.Kintone.Domain, username, password)
	loaderPath := filepath.Join(projectDir, config.ConfigDir, "managed", "kintone-dev-loader.js")

	// 既存カスタマイズの確認
	if !force {
		kcdevFiles := []string{"customize.js", "customize.css", "kintone-dev-loader.js"}
		existing, err := client.GetExistingCustomizations(cfg.Kintone.AppID, kcdevFiles)
		if err != nil {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Printf("  %s 既存カスタマイズの確認をスキップ: %v\n", yellow("⚠"), err)
		} else if existing.HasExisting() {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Printf("\n  %s 既存のカスタマイズが検出されました:\n", yellow("⚠"))

			// 詳細を表示
			if len(existing.Desktop.JS) > 0 {
				fmt.Printf("    デスクトップ JS: %s\n", strings.Join(existing.Desktop.JS, ", "))
			}
			if len(existing.Desktop.CSS) > 0 {
				fmt.Printf("    デスクトップ CSS: %s\n", strings.Join(existing.Desktop.CSS, ", "))
			}
			if len(existing.Mobile.JS) > 0 {
				fmt.Printf("    モバイル JS: %s\n", strings.Join(existing.Mobile.JS, ", "))
			}
			if len(existing.Mobile.CSS) > 0 {
				fmt.Printf("    モバイル CSS: %s\n", strings.Join(existing.Mobile.CSS, ", "))
			}

			fmt.Println()

			var confirm bool
			prompt := &survey.Confirm{
				Message: "これらのカスタマイズは上書きされます。続行しますか?",
				Default: false,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil {
				return fmt.Errorf("キャンセルされました")
			}
			if !confirm {
				fmt.Println("デプロイをキャンセルしました。")
				return fmt.Errorf("デプロイがキャンセルされました")
			}
			fmt.Println()
		}
	}

	var desktopFiles *kintone.CustomizeFiles
	var mobileFiles *kintone.CustomizeFiles

	// デスクトップ用ローダーをアップロード
	if cfg.Targets.Desktop {
		fmt.Printf("  デスクトップ アップロード...")
		fileKey, err := client.UploadFile(loaderPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("ローダーアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", green("✓"))
		desktopFiles = &kintone.CustomizeFiles{JSFileKey: fileKey}
	}

	// モバイル用ローダーをアップロード
	if cfg.Targets.Mobile {
		fmt.Printf("  モバイル アップロード...")
		fileKey, err := client.UploadFile(loaderPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("ローダーアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", green("✓"))
		mobileFiles = &kintone.CustomizeFiles{JSFileKey: fileKey}
	}

	// カスタマイズ設定を更新
	fmt.Printf("  設定...")
	scope := kintone.CustomizeScope(cfg.Scope)
	if scope == "" {
		scope = kintone.ScopeAll
	}
	if err := client.UpdateCustomize(cfg.Kintone.AppID, desktopFiles, mobileFiles, scope); err != nil {
		fmt.Println()
		return fmt.Errorf("カスタマイズ設定エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// アプリをデプロイ（プレビューのみの場合はスキップ）
	if !previewOnly {
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
	} else {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("  %s プレビュー環境のみに適用（本番反映はスキップ）\n", yellow("⚠"))
	}

	return nil
}

func printDevInfo(cfg *config.Config) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// ターゲット表示用の文字列を生成
	var targets []string
	if cfg.Targets.Desktop {
		targets = append(targets, "デスクトップ")
	}
	if cfg.Targets.Mobile {
		targets = append(targets, "モバイル")
	}
	if len(targets) == 0 {
		targets = append(targets, "デスクトップ") // デフォルト
	}

	fmt.Printf("\n%s 開発サーバーを起動中...\n", cyan("→"))
	fmt.Printf("  %s  %s\n", green("➜"), cfg.Dev.Origin)
	fmt.Printf("  %s     %s\n", cyan("エントリー:"), cfg.Dev.Entry)
	fmt.Printf("  %s     %s\n", cyan("ターゲット:"), strings.Join(targets, ", "))

	ok, msg, _ := generator.VerifyLoader(".")
	if ok {
		fmt.Printf("  %s       %s\n\n", green("ローダー:"), msg)
	} else {
		fmt.Printf("  %s       %s\n\n", yellow("ローダー:"), msg)
	}
}
