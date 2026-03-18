package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/generator"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "プロジェクトと依存パッケージを最新に更新",
	Long:  `Vite設定、依存パッケージを最新のkcdev仕様に更新します。`,
	RunE:  runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := config.Load(projectDir); err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
	}

	infoStyle := lipgloss.NewStyle().Foreground(ui.ColorCyan)

	fmt.Println()
	ui.Info("プロジェクトを更新中...")
	fmt.Println()

	// 1. vite.config.ts の更新チェック
	viteUpdated := false
	viteConfigPath := filepath.Join(projectDir, config.ConfigDir, "vite.config.ts")
	if needsViteConfigUpgrade(viteConfigPath) {
		fmt.Printf("%s vite.config.ts を更新中...", infoStyle.Render("○"))
		framework := detectCurrentFramework(projectDir)
		language := detectCurrentLanguage(projectDir)
		if err := generator.GenerateViteConfig(projectDir, framework, language); err != nil {
			fmt.Printf(" %s\n", ui.ErrorStyle.Render("✗"))
			return fmt.Errorf("vite.config.ts 更新エラー: %w", err)
		}
		fmt.Printf(" %s\n", ui.SuccessStyle.Render("✓"))
		viteUpdated = true
	}

	// 2. パッケージ更新
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("package.json が見つかりません: %w", err)
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("package.json の解析エラー: %w", err)
	}

	packagesToUpdate := collectUpdatePackages(pkg)

	if len(packagesToUpdate) > 0 {
		fmt.Printf("%s 以下のパッケージを更新します:\n", infoStyle.Render("○"))
		for _, p := range packagesToUpdate {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()

		pm := detectPackageManager(projectDir)

		var updateArgs []string
		switch pm {
		case "pnpm":
			updateArgs = append([]string{"update", "--latest"}, packagesToUpdate...)
		case "yarn":
			updateArgs = append([]string{"upgrade", "--latest"}, packagesToUpdate...)
		case "bun":
			updateArgs = append([]string{"update", "--latest"}, packagesToUpdate...)
		default:
			updateArgs = append([]string{"update", "--save"}, packagesToUpdate...)
		}

		fmt.Printf("%s %s %s\n", infoStyle.Render("○"), pm, updateArgs[0])

		updateExec := exec.Command(pm, updateArgs...)
		updateExec.Dir = projectDir
		updateExec.Stdout = os.Stdout
		updateExec.Stderr = os.Stderr

		if err := updateExec.Run(); err != nil {
			return fmt.Errorf("更新エラー: %w", err)
		}
	}

	if !viteUpdated && len(packagesToUpdate) == 0 {
		ui.Success("プロジェクトは最新の状態です")
	} else {
		fmt.Println()
		ui.Success("アップグレード完了!")
	}
	fmt.Println()
	return nil
}

func collectUpdatePackages(pkg map[string]interface{}) []string {
	packagesToUpdate := []string{}

	devDeps, ok := pkg["devDependencies"].(map[string]interface{})
	if !ok {
		devDeps = make(map[string]interface{})
	}

	deps, ok := pkg["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}

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

	for _, p := range vitePackages {
		if _, exists := devDeps[p]; exists {
			packagesToUpdate = append(packagesToUpdate, p)
		}
		if _, exists := deps[p]; exists {
			packagesToUpdate = append(packagesToUpdate, p)
		}
	}

	frameworkPackages := []string{
		"react",
		"react-dom",
		"vue",
		"svelte",
	}

	for _, p := range frameworkPackages {
		if _, exists := deps[p]; exists {
			packagesToUpdate = append(packagesToUpdate, p)
		}
	}

	return packagesToUpdate
}

// needsViteConfigUpgrade は vite.config.ts がアップグレード対象かを判定する
func needsViteConfigUpgrade(viteConfigPath string) bool {
	data, err := os.ReadFile(viteConfigPath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "rollupOptions") || strings.Contains(content, "esbuild")
}

func detectPackageManager(projectDir string) string {
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
