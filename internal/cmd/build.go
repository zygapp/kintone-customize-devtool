package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "本番用ビルドを生成",
	Long:  `Vite build を実行し、IIFE形式のファイルを生成します。`,
	RunE:  runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := config.Load(projectDir); err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\n%s ビルドを開始...\n", cyan("→"))

	viteConfig := filepath.Join(projectDir, config.ConfigDir, "vite.config.ts")
	if _, err := os.Stat(filepath.Join(projectDir, "vite.config.ts")); err == nil {
		viteConfig = filepath.Join(projectDir, "vite.config.ts")
	}

	fmt.Printf("%s バンドル中...\n", yellow("○"))

	viteCmd := exec.Command("npx", "vite", "build", "--config", viteConfig, "--logLevel", "silent")
	viteCmd.Dir = projectDir

	// エラー出力のみキャプチャ
	output, err := viteCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", string(output))
		return fmt.Errorf("ビルドエラー: %w", err)
	}

	fmt.Printf("%s ビルド完了!\n", green("✓"))
	fmt.Printf("出力ファイル:\n")
	fmt.Printf("  dist/customize.js\n")

	if _, err := os.Stat(filepath.Join(projectDir, "dist", "customize.css")); err == nil {
		fmt.Printf("  dist/customize.css\n")
	}

	fmt.Println()
	return nil
}
