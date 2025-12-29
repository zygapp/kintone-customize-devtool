package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
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

	// スピナーでデプロイ処理
	spinnerTitle := "デプロイ中..."
	if previewOnlyDeploy {
		spinnerTitle = "プレビュー環境にデプロイ中..."
	}

	var deployErr error
	ui.Spinner(spinnerTitle, func() {
		var desktopFiles *kintone.CustomizeFiles
		var mobileFiles *kintone.CustomizeFiles

		// デスクトップ用ファイルをアップロード
		if cfg.Targets.Desktop {
			jsKey, err := client.UploadFile(jsPath)
			if err != nil {
				deployErr = fmt.Errorf("JSファイルアップロードエラー: %w", err)
				return
			}
			desktopFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

			if hasCss {
				cssKey, err := client.UploadFile(cssPath)
				if err != nil {
					deployErr = fmt.Errorf("CSSファイルアップロードエラー: %w", err)
					return
				}
				desktopFiles.CSSFileKey = cssKey
			}
		}

		// モバイル用ファイルをアップロード
		if cfg.Targets.Mobile {
			jsKey, err := client.UploadFile(jsPath)
			if err != nil {
				deployErr = fmt.Errorf("JSファイルアップロードエラー: %w", err)
				return
			}
			mobileFiles = &kintone.CustomizeFiles{JSFileKey: jsKey}

			if hasCss {
				cssKey, err := client.UploadFile(cssPath)
				if err != nil {
					deployErr = fmt.Errorf("CSSファイルアップロードエラー: %w", err)
					return
				}
				mobileFiles.CSSFileKey = cssKey
			}
		}

		// カスタマイズ設定を更新
		scope := kintone.CustomizeScope(cfg.Scope)
		if scope == "" {
			scope = kintone.ScopeAll
		}
		if err := client.UpdateCustomize(cfg.Kintone.AppID, desktopFiles, mobileFiles, scope); err != nil {
			deployErr = fmt.Errorf("カスタマイズ設定エラー: %w", err)
			return
		}

		// アプリをデプロイ（プレビューのみの場合はスキップ）
		if !previewOnlyDeploy {
			if err := client.DeployApp(cfg.Kintone.AppID); err != nil {
				deployErr = fmt.Errorf("デプロイ開始エラー: %w", err)
				return
			}

			if err := client.WaitForDeploy(cfg.Kintone.AppID); err != nil {
				deployErr = fmt.Errorf("デプロイ待機エラー: %w", err)
				return
			}
		}
	})

	if deployErr != nil {
		return deployErr
	}

	if !previewOnlyDeploy {
		ui.Success(fmt.Sprintf("完了! https://%s/k/%d/", cfg.Kintone.Domain, cfg.Kintone.AppID))
	} else {
		ui.Warn("プレビュー環境のみに適用（本番反映はスキップ）")
		ui.Success(fmt.Sprintf("プレビュー環境に適用しました! https://%s/k/admin/app/flow?app=%d", cfg.Kintone.Domain, cfg.Kintone.AppID))
	}
	fmt.Println()

	return nil
}
