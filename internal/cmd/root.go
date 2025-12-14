package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.1.1"

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
}
