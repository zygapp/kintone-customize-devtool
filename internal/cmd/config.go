package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/generator"
	"github.com/kintone/kcdev/internal/prompt"
	"github.com/kintone/kcdev/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "プロジェクト設定を変更",
	Long:  `対話形式でプロジェクトの各種設定を変更します。`,
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("設定ファイルが見つかりません。先に kcdev init を実行してください: %w", err)
	}

	for {
		fmt.Print("\033[H\033[2J")
		ui.Title("設定メニュー")
		fmt.Println()

		action, err := askConfigAction()
		if err != nil {
			return err
		}

		switch action {
		case "view":
			showCurrentConfig(cfg)
		case "kintone":
			if err := editKintoneConfig(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "targets":
			if err := editTargets(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "scope":
			if err := editScope(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "output":
			if err := editOutput(cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "entry":
			if err := editEntry(cwd, cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "framework":
			if err := editFramework(cwd, cfg); err != nil {
				return err
			}
			if err := cfg.Save(cwd); err != nil {
				return err
			}
		case "exit":
			fmt.Println("\n設定を終了します。")
			return nil
		}
	}
}

func askConfigAction() (string, error) {
	var answer string
	err := ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("操作を選択してください").
				Options(
					huh.NewOption("現在の設定を表示", "view"),
					huh.NewOption("kintone接続設定（ドメイン、アプリID、認証）", "kintone"),
					huh.NewOption("ターゲット（デスクトップ/モバイル）の設定", "targets"),
					huh.NewOption("適用範囲の設定", "scope"),
					huh.NewOption("出力ファイル名の設定", "output"),
					huh.NewOption("エントリーファイルの設定", "entry"),
					huh.NewOption("フレームワークの変更", "framework"),
					huh.NewOption("終了", "exit"),
				).
				Value(&answer),
		),
	).Run()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func showCurrentConfig(cfg *config.Config) {
	fmt.Println()
	ui.Title("現在の設定")
	fmt.Println()

	successStyle := lipgloss.NewStyle().Foreground(ui.ColorGreen)
	warnStyle := lipgloss.NewStyle().Foreground(ui.ColorYellow)
	errorStyle := lipgloss.NewStyle().Foreground(ui.ColorRed)
	infoStyle := lipgloss.NewStyle().Foreground(ui.ColorCyan)

	// kintone設定
	fmt.Println(infoStyle.Render("kintone:"))
	fmt.Printf("  ドメイン:   %s\n", cfg.Kintone.Domain)
	fmt.Printf("  アプリID:   %d\n", cfg.Kintone.AppID)
	if cfg.Kintone.Auth.Username != "" {
		fmt.Printf("  ユーザー:   %s\n", cfg.Kintone.Auth.Username)
		fmt.Printf("  パスワード: %s\n", "********")
	} else {
		fmt.Printf("  認証:       %s\n", warnStyle.Render("未設定"))
	}

	// ターゲット
	fmt.Println()
	fmt.Println(infoStyle.Render("ターゲット:"))
	if cfg.Targets.Desktop {
		fmt.Printf("  %s デスクトップ\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s デスクトップ\n", errorStyle.Render("✗"))
	}
	if cfg.Targets.Mobile {
		fmt.Printf("  %s モバイル\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s モバイル\n", errorStyle.Render("✗"))
	}

	// 適用範囲
	fmt.Println()
	fmt.Println(infoStyle.Render("適用範囲:"))
	switch cfg.Scope {
	case config.ScopeAll:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", successStyle.Render("✓"))
	case config.ScopeAdmin:
		fmt.Printf("  %s アプリ管理者のみ (ADMIN)\n", warnStyle.Render("✓"))
	case config.ScopeNone:
		fmt.Printf("  %s 適用しない (NONE)\n", errorStyle.Render("✗"))
	default:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", successStyle.Render("✓"))
	}

	// 出力ファイル名
	fmt.Println()
	fmt.Println(infoStyle.Render("出力:"))
	fmt.Printf("  ファイル名: %s.js / %s.css\n", cfg.GetOutputName(), cfg.GetOutputName())

	// Dev設定
	fmt.Println()
	fmt.Println(infoStyle.Render("開発サーバー:"))
	fmt.Printf("  オリジン:   %s\n", cfg.Dev.Origin)
	fmt.Printf("  エントリー: %s\n", cfg.Dev.Entry)

	fmt.Println()
	fmt.Println("Enterキーで戻る...")
	fmt.Scanln()
}

func editKintoneConfig(cfg *config.Config) error {
	fmt.Println()
	ui.Title("kintone接続設定")
	fmt.Println()

	// ドメイン
	domain, err := prompt.AskDomain(cfg.Kintone.Domain)
	if err != nil {
		return err
	}
	cfg.Kintone.Domain = domain

	// アプリID
	appID, err := prompt.AskAppID(cfg.Kintone.AppID)
	if err != nil {
		return err
	}
	cfg.Kintone.AppID = appID

	// 認証情報を更新するか確認
	var updateAuth bool
	err = ui.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("認証情報を更新しますか?").
				Affirmative("はい").
				Negative("いいえ").
				Value(&updateAuth),
		),
	).Run()
	if err != nil {
		return err
	}

	if updateAuth {
		username, err := prompt.AskUsername()
		if err != nil {
			return err
		}
		password, err := prompt.AskPassword()
		if err != nil {
			return err
		}
		cfg.Kintone.Auth.Username = username
		cfg.Kintone.Auth.Password = password
	}

	fmt.Println()
	ui.Success("kintone接続設定を更新しました")
	return nil
}

func editTargets(cfg *config.Config) error {
	fmt.Println()

	desktop, mobile, err := prompt.AskTargets(cfg.Targets.Desktop, cfg.Targets.Mobile)
	if err != nil {
		return err
	}

	cfg.Targets.Desktop = desktop
	cfg.Targets.Mobile = mobile

	fmt.Println()
	ui.Success("ターゲットを更新しました")
	return nil
}

func editScope(cfg *config.Config) error {
	fmt.Println()

	currentScope := prompt.Scope(cfg.Scope)
	if currentScope == "" {
		currentScope = prompt.ScopeAll
	}

	scope, err := prompt.AskScope(currentScope)
	if err != nil {
		return err
	}

	cfg.Scope = string(scope)

	fmt.Println()
	ui.Success("適用範囲を更新しました")
	return nil
}

func editOutput(cfg *config.Config) error {
	fmt.Println()

	output, err := prompt.AskOutput(cfg.GetOutputName())
	if err != nil {
		return err
	}

	cfg.Output = output

	fmt.Println()
	ui.Success(fmt.Sprintf("出力ファイル名を更新しました (%s.js / %s.css)", output, output))
	return nil
}

func editEntry(projectDir string, cfg *config.Config) error {
	fmt.Println()

	// src/ 直下の js, ts, jsx, tsx ファイルを検索
	srcDir := filepath.Join(projectDir, "src")
	var entryFiles []string

	entries, err := os.ReadDir(srcDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
				entryFiles = append(entryFiles, "/src/"+entry.Name())
			}
		}
	}

	if len(entryFiles) == 0 {
		ui.Warn("src/ ディレクトリにエントリーファイルが見つかりません")
		fmt.Println("Enterキーで戻る...")
		fmt.Scanln()
		return nil
	}

	// オプションを作成
	var options []huh.Option[string]
	for _, f := range entryFiles {
		options = append(options, huh.NewOption(f, f))
	}

	var selected string
	err = ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("エントリーファイルを選択").
				Options(options...).
				Value(&selected),
		),
	).Run()
	if err != nil {
		return err
	}

	cfg.Dev.Entry = selected

	fmt.Println()
	ui.Success(fmt.Sprintf("エントリーファイルを更新しました (%s)", selected))
	return nil
}

func editFramework(projectDir string, cfg *config.Config) error {
	fmt.Println()
	ui.Title("フレームワークの変更")
	fmt.Println()

	infoStyle := lipgloss.NewStyle().Foreground(ui.ColorCyan)

	// 現在のフレームワークを検出
	currentFramework := detectCurrentFramework(projectDir)
	currentLanguage := detectCurrentLanguage(projectDir)

	fmt.Printf("現在のフレームワーク: %s (%s)\n\n", infoStyle.Render(string(currentFramework)), string(currentLanguage))

	// 新しいフレームワークを選択
	newFramework, err := prompt.AskFramework()
	if err != nil {
		return err
	}

	if newFramework == currentFramework {
		fmt.Println()
		ui.Warn("フレームワークは変更されていません")
		fmt.Println("Enterキーで戻る...")
		fmt.Scanln()
		return nil
	}

	// 確認
	var confirm bool
	err = ui.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("%s から %s に変更しますか?", currentFramework, newFramework)).
				Affirmative("はい").
				Negative("いいえ").
				Value(&confirm),
		),
	).Run()
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	fmt.Println()

	pm := detectPackageManager(projectDir)

	// 1. 旧フレームワークのパッケージをアンインストール
	// パッケージマネージャーごとのコマンドを設定
	var removeCmd, addCmd, addDevFlag string
	switch pm {
	case "yarn", "pnpm", "bun":
		removeCmd = "remove"
		addCmd = "add"
		addDevFlag = "-D"
	default: // npm
		removeCmd = "uninstall"
		addCmd = "install"
		addDevFlag = "-D"
	}

	err = ui.Spinner("旧パッケージを削除中...", func() {
		oldDeps, oldDevDeps := getFrameworkPackageNames(currentFramework, currentLanguage)
		allOldPkgs := append(oldDeps, oldDevDeps...)
		if len(allOldPkgs) > 0 {
			args := append([]string{removeCmd}, allOldPkgs...)
			uninstallCmd := exec.Command(pm, args...)
			uninstallCmd.Dir = projectDir
			uninstallCmd.Stdout = nil
			uninstallCmd.Stderr = nil
			uninstallCmd.Run() // エラーは無視（パッケージがない場合もある）
		}
	})
	if err != nil {
		return err
	}
	ui.Success("旧パッケージを削除しました")

	// 2. 新フレームワークのパッケージをインストール
	fmt.Println()
	newDeps, newDevDeps := getFrameworkPackageNames(newFramework, currentLanguage)

	if len(newDeps) > 0 {
		fmt.Printf("  依存パッケージをインストール中...\n")
		args := append([]string{addCmd}, newDeps...)
		installCmd := exec.Command(pm, args...)
		installCmd.Dir = projectDir
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("依存パッケージインストールエラー: %w", err)
		}
	}

	if len(newDevDeps) > 0 {
		fmt.Printf("  開発パッケージをインストール中...\n")
		args := append([]string{addCmd, addDevFlag}, newDevDeps...)
		installCmd := exec.Command(pm, args...)
		installCmd.Dir = projectDir
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("開発パッケージインストールエラー: %w", err)
		}
	}
	ui.Success("新パッケージをインストールしました")

	// 3. vite.config.ts を再生成
	err = ui.SpinnerWithResult("vite.config.ts を再生成中...", func() error {
		return generator.GenerateViteConfig(projectDir, newFramework, currentLanguage)
	})
	if err != nil {
		return fmt.Errorf("vite.config.ts再生成エラー: %w", err)
	}
	ui.Success("vite.config.ts を再生成しました")

	// 4. eslint.config.js を再生成
	err = ui.SpinnerWithResult("eslint.config.js を再生成中...", func() error {
		return generator.RegenerateESLintConfig(projectDir, newFramework, currentLanguage)
	})
	if err != nil {
		return fmt.Errorf("eslint.config.js再生成エラー: %w", err)
	}
	ui.Success("eslint.config.js を再生成しました")

	// 5. config.json のエントリーパスを更新
	cfg.Dev.Entry = generator.GetEntryPath(newFramework, currentLanguage)

	fmt.Println()
	ui.Success(fmt.Sprintf("フレームワークを %s に変更しました!", newFramework))
	fmt.Println()
	ui.Warn("src/ ディレクトリのコードを手動で書き換えてください")
	fmt.Printf("  エントリーファイル: %s\n\n", cfg.Dev.Entry)
	fmt.Println("Enterキーで戻る...")
	fmt.Scanln()

	return nil
}

func detectCurrentFramework(projectDir string) prompt.Framework {
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return prompt.FrameworkVanilla
	}

	content := string(data)
	if strings.Contains(content, `"react"`) {
		return prompt.FrameworkReact
	}
	if strings.Contains(content, `"vue"`) {
		return prompt.FrameworkVue
	}
	if strings.Contains(content, `"svelte"`) {
		return prompt.FrameworkSvelte
	}
	return prompt.FrameworkVanilla
}

func detectCurrentLanguage(projectDir string) prompt.Language {
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return prompt.LanguageJavaScript
	}

	if strings.Contains(string(data), `"typescript"`) {
		return prompt.LanguageTypeScript
	}
	return prompt.LanguageJavaScript
}

// getFrameworkPackageNames はフレームワーク固有のパッケージ名リストを返す（バージョンなし）
func getFrameworkPackageNames(fw prompt.Framework, lang prompt.Language) (deps []string, devDeps []string) {
	switch fw {
	case prompt.FrameworkReact:
		deps = append(deps, "react", "react-dom")
		devDeps = append(devDeps, "@vitejs/plugin-react", "eslint-plugin-react-hooks")
		if lang == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "@types/react", "@types/react-dom")
		}
	case prompt.FrameworkVue:
		deps = append(deps, "vue")
		devDeps = append(devDeps, "@vitejs/plugin-vue", "eslint-plugin-vue")
		if lang == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "vue-tsc")
		}
	case prompt.FrameworkSvelte:
		deps = append(deps, "svelte")
		devDeps = append(devDeps, "@sveltejs/vite-plugin-svelte", "eslint-plugin-svelte")
		if lang == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "svelte-check", "tslib")
		}
	}
	return deps, devDeps
}
