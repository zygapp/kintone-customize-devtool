package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "依存パッケージを最新版に更新",
	Long:  `Viteおよびフレームワークプラグインを最新版に更新します。`,
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("\n%s 依存パッケージを更新中...\n\n", cyan("→"))

	// package.json を読み込み
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("package.json が見つかりません: %w", err)
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("package.json の解析エラー: %w", err)
	}

	// 更新対象のパッケージを特定
	packagesToUpdate := []string{}

	devDeps, ok := pkg["devDependencies"].(map[string]interface{})
	if !ok {
		devDeps = make(map[string]interface{})
	}

	deps, ok := pkg["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}

	// Vite関連パッケージ
	vitePackages := []string{
		"vite",
		"@vitejs/plugin-react",
		"@vitejs/plugin-vue",
		"@sveltejs/vite-plugin-svelte",
		"typescript",
		"@types/react",
		"@types/react-dom",
		"vue-tsc",
		"svelte-check",
	}

	for _, pkg := range vitePackages {
		if _, exists := devDeps[pkg]; exists {
			packagesToUpdate = append(packagesToUpdate, pkg)
		}
		if _, exists := deps[pkg]; exists {
			packagesToUpdate = append(packagesToUpdate, pkg)
		}
	}

	// フレームワーク本体
	frameworkPackages := []string{
		"react",
		"react-dom",
		"vue",
		"svelte",
	}

	for _, pkg := range frameworkPackages {
		if _, exists := deps[pkg]; exists {
			packagesToUpdate = append(packagesToUpdate, pkg)
		}
	}

	if len(packagesToUpdate) == 0 {
		fmt.Printf("%s 更新対象のパッケージがありません\n\n", yellow("!"))
		return nil
	}

	fmt.Printf("%s 以下のパッケージを更新します:\n", yellow("○"))
	for _, pkg := range packagesToUpdate {
		fmt.Printf("  - %s\n", pkg)
	}
	fmt.Println()

	// パッケージマネージャーを検出
	pm := detectPackageManager(projectDir)

	// 更新コマンドを実行
	var updateArgs []string
	switch pm {
	case "pnpm":
		updateArgs = append([]string{"update", "--latest"}, packagesToUpdate...)
	case "yarn":
		updateArgs = append([]string{"upgrade", "--latest"}, packagesToUpdate...)
	case "bun":
		updateArgs = append([]string{"update", "--latest"}, packagesToUpdate...)
	default: // npm
		updateArgs = append([]string{"update", "--save"}, packagesToUpdate...)
	}

	fmt.Printf("%s %s %s\n", yellow("○"), pm, updateArgs[0])

	updateCmd := exec.Command(pm, updateArgs...)
	updateCmd.Dir = projectDir
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("更新エラー: %w", err)
	}

	fmt.Printf("\n%s パッケージを更新しました!\n\n", green("✓"))
	return nil
}

func detectPackageManager(projectDir string) string {
	// ロックファイルで判定
	if _, err := os.Stat(filepath.Join(projectDir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(projectDir, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(projectDir, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}
