package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/spf13/cobra"
)

var (
	noMinify bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "本番用ビルドを生成",
	Long:  `Vite build を実行し、IIFE形式のファイルを生成します。`,
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().BoolVar(&noMinify, "no-minify", false, "minifyを無効化（デバッグ用）")
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

	// package.json からバージョンを読み込み
	pkg, err := loadPackageJSON(projectDir)
	if err != nil {
		return fmt.Errorf("package.json の読み込みに失敗しました: %w", err)
	}

	currentVersion := fmt.Sprintf("%v", pkg["version"])
	newVersion, err := askVersion(currentVersion)
	if err != nil {
		return err
	}

	// バージョンが変更された場合は保存
	if newVersion != currentVersion {
		pkg["version"] = newVersion
		if err := savePackageJSON(projectDir, pkg); err != nil {
			return fmt.Errorf("package.json の保存に失敗しました: %w", err)
		}
		fmt.Printf("%s バージョンを更新: %s → %s\n", green("✓"), currentVersion, newVersion)
	}

	fmt.Printf("\n%s ビルドを開始...\n", cyan("→"))

	viteConfig := filepath.Join(projectDir, config.ConfigDir, "vite.config.ts")
	if _, err := os.Stat(filepath.Join(projectDir, "vite.config.ts")); err == nil {
		viteConfig = filepath.Join(projectDir, "vite.config.ts")
	}

	fmt.Printf("%s バンドル中...\n", yellow("○"))

	viteArgs := []string{"vite", "build", "--config", viteConfig, "--logLevel", "silent"}
	if noMinify {
		viteArgs = append(viteArgs, "--minify", "false")
	}

	viteCmd := exec.Command("npx", viteArgs...)
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

func loadPackageJSON(projectDir string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
	if err != nil {
		return nil, err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

func savePackageJSON(projectDir string, pkg map[string]interface{}) error {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projectDir, "package.json"), data, 0644)
}

func askVersion(currentVersion string) (string, error) {
	cyan := color.New(color.FgCyan).SprintFunc()

	// バージョンをパース
	parts := strings.Split(currentVersion, ".")
	major, minor, patch := 0, 0, 0
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}

	// バージョン選択肢を作成
	patchVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	minorVersion := fmt.Sprintf("%d.%d.%d", major, minor+1, 0)
	majorVersion := fmt.Sprintf("%d.%d.%d", major+1, 0, 0)

	options := []string{
		fmt.Sprintf("現在のまま (%s)", currentVersion),
		fmt.Sprintf("パッチ更新 (%s)", patchVersion),
		fmt.Sprintf("マイナー更新 (%s)", minorVersion),
		fmt.Sprintf("メジャー更新 (%s)", majorVersion),
		"カスタム入力",
	}

	fmt.Printf("現在のバージョン: %s\n\n", cyan(currentVersion))

	var answer string
	prompt := &survey.Select{
		Message: "バージョンを選択:",
		Options: options,
		Default: options[0],
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return currentVersion, nil
	case options[1]:
		return patchVersion, nil
	case options[2]:
		return minorVersion, nil
	case options[3]:
		return majorVersion, nil
	default:
		// カスタム入力
		var customVersion string
		inputPrompt := &survey.Input{
			Message: "バージョンを入力:",
			Default: currentVersion,
		}
		if err := survey.AskOne(inputPrompt, &customVersion, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		return customVersion, nil
	}
}
