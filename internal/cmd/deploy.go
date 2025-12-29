package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/kintone"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

var forceOverwrite bool
var previewOnlyDeploy bool
var skipVersionDeploy bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "kintoneにデプロイ",
	Long:  `ビルド成果物をkintoneにアップロードしてデプロイします。`,
	RunE:  runDeploy,
}

func init() {
	deployCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "既存カスタマイズを確認せず上書き")
	deployCmd.Flags().BoolVarP(&previewOnlyDeploy, "preview", "p", false, "プレビュー環境のみにデプロイ（本番反映しない）")
	deployCmd.Flags().BoolVar(&skipVersionDeploy, "skip-version", false, "バージョン確認をスキップ")
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

	successStyle := lipgloss.NewStyle().Foreground(ui.ColorGreen)

	// 設定から出力ファイル名を取得
	outputName := cfg.GetOutputName()

	distDir := filepath.Join(projectDir, "dist")
	jsPath := filepath.Join(distDir, outputName+".js")

	// dist/が存在する場合はビルド確認
	if _, err := os.Stat(distDir); err == nil {
		// dist/が存在する場合、再ビルドするか確認
		var rebuild bool
		err := ui.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("dist/ が存在します。再ビルドしますか?").
					Affirmative("はい").
					Negative("いいえ").
					Value(&rebuild),
			),
		).Run()
		if err != nil {
			return fmt.Errorf("キャンセルされました")
		}
		if rebuild {
			// deploy の --skip-version を build に引き継ぐ
			skipVersion = skipVersionDeploy
			if err := runBuild(nil, nil); err != nil {
				return fmt.Errorf("ビルドエラー: %w", err)
			}
			fmt.Println()
		}
	} else {
		// dist/が存在しない場合は自動でビルド
		ui.Info("dist/ が見つかりません。ビルドを開始...")
		// deploy の --skip-version を build に引き継ぐ
		skipVersion = skipVersionDeploy
		if err := runBuild(nil, nil); err != nil {
			return fmt.Errorf("ビルドエラー: %w", err)
		}
		fmt.Println()
	}

	// ビルド成果物の確認
	if _, err := os.Stat(jsPath); err != nil {
		return fmt.Errorf("ビルド成果物が見つかりません")
	}

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

	fmt.Println()
	if previewOnlyDeploy {
		ui.Info(fmt.Sprintf("プレビュー環境にデプロイ中... (%s, App:%d, %s)", cfg.Kintone.Domain, cfg.Kintone.AppID, strings.Join(targets, "+")))
	} else {
		ui.Info(fmt.Sprintf("デプロイ中... (%s, App:%d, %s)", cfg.Kintone.Domain, cfg.Kintone.AppID, strings.Join(targets, "+")))
	}

	client := kintone.NewClient(cfg.Kintone.Domain, username, password)

	// 既存カスタマイズの確認
	if !forceOverwrite {
		kcdevFiles := []string{outputName + ".js", outputName + ".css", "kintone-dev-loader.js"}
		existing, err := client.GetExistingCustomizations(cfg.Kintone.AppID, kcdevFiles)
		if err != nil {
			ui.Warn(fmt.Sprintf("既存カスタマイズの確認をスキップ: %v", err))
		} else if existing.HasExisting() {
			fmt.Println()
			ui.Warn("既存のカスタマイズが検出されました:")

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
			err := ui.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("これらのカスタマイズは上書きされます。続行しますか?").
						Affirmative("はい").
						Negative("いいえ").
						Value(&confirm),
				),
			).Run()
			if err != nil {
				return fmt.Errorf("キャンセルされました")
			}
			if !confirm {
				fmt.Println("デプロイをキャンセルしました。")
				return nil
			}
			fmt.Println()
		}
	}

	cssPath := filepath.Join(projectDir, "dist", outputName+".css")
	hasCss := false
	if _, err := os.Stat(cssPath); err == nil {
		hasCss = true
	}

	var desktopFiles *kintone.CustomizeFiles
	var mobileFiles *kintone.CustomizeFiles

	// デスクトップ用ファイルをアップロード
	if cfg.Targets.Desktop {
		fmt.Printf("  デスクトップ JS...")
		jsKey, err := client.UploadFile(jsPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("JSファイルアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", successStyle.Render("✓"))

		desktopFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

		if hasCss {
			fmt.Printf("  デスクトップ CSS...")
			cssKey, err := client.UploadFile(cssPath)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("CSSファイルアップロードエラー: %w", err)
			}
			fmt.Printf(" %s\n", successStyle.Render("✓"))
			desktopFiles.CSSFileKey = cssKey
		}
	}

	// モバイル用ファイルをアップロード
	if cfg.Targets.Mobile {
		fmt.Printf("  モバイル JS...")
		jsKey, err := client.UploadFile(jsPath)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("JSファイルアップロードエラー: %w", err)
		}
		fmt.Printf(" %s\n", successStyle.Render("✓"))

		mobileFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

		if hasCss {
			fmt.Printf("  モバイル CSS...")
			cssKey, err := client.UploadFile(cssPath)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("CSSファイルアップロードエラー: %w", err)
			}
			fmt.Printf(" %s\n", successStyle.Render("✓"))
			mobileFiles.CSSFileKey = cssKey
		}
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
	fmt.Printf(" %s\n", successStyle.Render("✓"))

	// アプリをデプロイ（プレビューのみの場合はスキップ）
	if !previewOnlyDeploy {
		fmt.Printf("  デプロイ...")
		if err := client.DeployApp(cfg.Kintone.AppID); err != nil {
			fmt.Println()
			return fmt.Errorf("デプロイ開始エラー: %w", err)
		}

		if err := client.WaitForDeploy(cfg.Kintone.AppID); err != nil {
			fmt.Println()
			return fmt.Errorf("デプロイ待機エラー: %w", err)
		}
		fmt.Printf(" %s\n", successStyle.Render("✓"))

		fmt.Println()
		ui.Success(fmt.Sprintf("完了! https://%s/k/%d/", cfg.Kintone.Domain, cfg.Kintone.AppID))
		fmt.Println()
	} else {
		ui.Warn("プレビュー環境のみに適用（本番反映はスキップ）")
		fmt.Println()
		ui.Success(fmt.Sprintf("プレビュー環境に適用しました! https://%s/k/admin/app/flow?app=%d", cfg.Kintone.Domain, cfg.Kintone.AppID))
		fmt.Println()
	}

	return nil
}
