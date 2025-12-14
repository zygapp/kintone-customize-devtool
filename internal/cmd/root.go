package cmd

import (
	"github.com/spf13/cobra"
)

// version はビルド時に -ldflags で注入される
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "kcdev",
	Short:   "kintone customize developer",
	Long:    `kcdev は kintone カスタマイズ開発を Vite + HMR で行うための CLI ツールです。`,
	Version: version,
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
