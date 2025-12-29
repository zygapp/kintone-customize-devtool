package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

var errVersionRequired = errors.New("入力必須です")

var (
	noMinify    bool
	skipVersion bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "本番用ビルドを生成",
	Long:  `Vite build を実行し、IIFE形式のファイルを生成します。`,
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().BoolVar(&noMinify, "no-minify", false, "minifyを無効化（デバッグ用）")
	buildCmd.Flags().BoolVar(&skipVersion, "skip-version", false, "バージョン確認をスキップ")
}

func runBuild(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := config.Load(projectDir); err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。kcdev init を実行してください: %w", err)
	}

	// バージョン確認（スキップフラグがない場合）
	if !skipVersion {
		pkg, err := loadPackageJSON(projectDir)
		if err != nil {
			return fmt.Errorf("package.json の読み込みに失敗しました: %w", err)
		}

		currentVersion := fmt.Sprintf("%v", pkg["version"])
		newVersion, err := askVersionUpdate(currentVersion)
		if err != nil {
			return err
		}

		// バージョンが変更された場合は保存
		if newVersion != currentVersion {
			pkg["version"] = newVersion
			if err := savePackageJSON(projectDir, pkg); err != nil {
				return fmt.Errorf("package.json の保存に失敗しました: %w", err)
			}
			ui.Success(fmt.Sprintf("バージョンを更新: %s → %s", currentVersion, newVersion))
		}
	}

	viteConfig := filepath.Join(projectDir, config.ConfigDir, "vite.config.ts")
	if _, err := os.Stat(filepath.Join(projectDir, "vite.config.ts")); err == nil {
		viteConfig = filepath.Join(projectDir, "vite.config.ts")
	}

	viteArgs := []string{"vite", "build", "--config", viteConfig, "--logLevel", "silent"}
	if noMinify {
		viteArgs = append(viteArgs, "--minify", "false")
	}

	var buildErr error
	var buildOutput []byte
	ui.Spinner("ビルド中...", func() {
		viteCmd := exec.Command("npx", viteArgs...)
		viteCmd.Dir = projectDir
		buildOutput, buildErr = viteCmd.CombinedOutput()
	})

	if buildErr != nil {
		fmt.Printf("%s\n", string(buildOutput))
		return fmt.Errorf("ビルドエラー: %w", buildErr)
	}

	// 設定から出力ファイル名を取得
	cfg, _ := config.Load(projectDir)
	outputName := "customize"
	if cfg != nil {
		outputName = cfg.GetOutputName()
	}

	ui.Success("ビルド完了!")
	fmt.Println("出力ファイル:")
	fmt.Printf("  dist/%s.js\n", outputName)

	if _, err := os.Stat(filepath.Join(projectDir, "dist", outputName+".css")); err == nil {
		fmt.Printf("  dist/%s.css\n", outputName)
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

func askVersionUpdate(currentVersion string) (string, error) {
	fmt.Printf("現在のバージョン: %s\n", currentVersion)

	// まずバージョンを更新するか確認
	var updateVersion bool
	err := ui.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("バージョンを更新しますか?").
				Affirmative("はい").
				Negative("いいえ").
				Value(&updateVersion),
		),
	).Run()
	if err != nil {
		return "", err
	}

	if !updateVersion {
		return currentVersion, nil
	}

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

	var answer string
	err = ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("バージョンを選択").
				Options(
					huh.NewOption(fmt.Sprintf("パッチ更新 (%s)", patchVersion), "patch"),
					huh.NewOption(fmt.Sprintf("マイナー更新 (%s)", minorVersion), "minor"),
					huh.NewOption(fmt.Sprintf("メジャー更新 (%s)", majorVersion), "major"),
					huh.NewOption("カスタム入力", "custom"),
				).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}

	switch answer {
	case "patch":
		return patchVersion, nil
	case "minor":
		return minorVersion, nil
	case "major":
		return majorVersion, nil
	default:
		// カスタム入力
		var customVersion string
		err := ui.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("バージョンを入力").
					Value(&customVersion).
					Placeholder(patchVersion).
					Validate(func(s string) error {
						if s == "" {
							return errVersionRequired
						}
						return nil
					}),
			),
		).Run()
		if err != nil {
			return "", err
		}
		if customVersion == "" {
			customVersion = patchVersion
		}
		return customVersion, nil
	}
}
