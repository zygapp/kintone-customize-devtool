package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "kintone フィールド型定義を生成",
	Long:  `@kintone/dts-gen を使用して、kintone アプリのフィールド型定義を生成します。`,
	RunE:  runTypes,
}

func runTypes(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
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

	return generateTypes(projectDir, cfg, username, password)
}

func generateTypes(projectDir string, cfg *config.Config, username, password string) error {
	fmt.Println()
	ui.Info("型定義を生成中...")

	// 出力ディレクトリを作成
	typesDir := filepath.Join(projectDir, "src", "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	// dts-gen を実行
	dtsGenCmd := exec.Command("npx", "@kintone/dts-gen",
		"--base-url", fmt.Sprintf("https://%s", cfg.Kintone.Domain),
		"--app-id", fmt.Sprintf("%d", cfg.Kintone.AppID),
		"--username", username,
		"--password", password,
		"--output", filepath.Join(typesDir, "kintone.d.ts"),
	)
	dtsGenCmd.Dir = projectDir
	dtsGenCmd.Stdout = os.Stdout
	dtsGenCmd.Stderr = os.Stderr

	if err := dtsGenCmd.Run(); err != nil {
		return fmt.Errorf("dts-gen 実行エラー: %w", err)
	}

	fmt.Println()
	ui.Success("型定義を生成しました: src/types/kintone.d.ts")
	fmt.Println()
	return nil
}
