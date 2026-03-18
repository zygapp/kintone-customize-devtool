package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

// version はビルド時に -ldflags で注入される
var version = "dev"

// upgradeチェックをスキップするコマンド
var skipUpgradeCheck = map[string]bool{
	"init":    true,
	"upgrade": true,
	"help":    true,
	"version": true,
}

var rootCmd = &cobra.Command{
	Use:     "kcdev",
	Short:   "kintone customize developer",
	Long:    `kcdev は kintone カスタマイズ開発を Vite + HMR で行うための CLI ツールです。`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if skipUpgradeCheck[cmd.Name()] {
			return
		}
		checkUpgradeNeeded()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(typesCmd)
}

// checkUpgradeNeeded はプロジェクトのアップグレードが必要か確認し、必要ならメッセージを表示する
func checkUpgradeNeeded() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	// config.json が存在しなければプロジェクト外なのでスキップ
	configPath := filepath.Join(cwd, config.ConfigDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return
	}

	// vite.config.ts のチェック
	viteConfigPath := filepath.Join(cwd, config.ConfigDir, "vite.config.ts")
	if data, err := os.ReadFile(viteConfigPath); err == nil {
		content := string(data)
		if strings.Contains(content, "rollupOptions") || strings.Contains(content, "esbuild") {
			ui.Warn("プロジェクトの設定が古くなっています。 kcdev upgrade を実行してください。")
			fmt.Println()
		}
	}
}
