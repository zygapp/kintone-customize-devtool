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
	"github.com/kintone/kcdev/internal/generator"
	"github.com/kintone/kcdev/internal/prompt"
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

	cyan := color.New(color.FgCyan).SprintFunc()

	for {
		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s 設定メニュー\n\n", cyan("⚙"))

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
	options := []string{
		"現在の設定を表示",
		"kintone接続設定（ドメイン、アプリID、認証）",
		"ターゲット（デスクトップ/モバイル）の設定",
		"適用範囲の設定",
		"出力ファイル名の設定",
		"エントリーファイルの設定",
		"フレームワークの変更",
		"終了",
	}

	var answer string
	prompt := &survey.Select{
		Message: "操作を選択してください:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}

	switch answer {
	case options[0]:
		return "view", nil
	case options[1]:
		return "kintone", nil
	case options[2]:
		return "targets", nil
	case options[3]:
		return "scope", nil
	case options[4]:
		return "output", nil
	case options[5]:
		return "entry", nil
	case options[6]:
		return "framework", nil
	default:
		return "exit", nil
	}
}

func showCurrentConfig(cfg *config.Config) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\n%s 現在の設定\n\n", cyan("📋"))

	// kintone設定
	fmt.Printf("%s\n", cyan("kintone:"))
	fmt.Printf("  ドメイン: %s\n", cfg.Kintone.Domain)
	fmt.Printf("  アプリID: %d\n", cfg.Kintone.AppID)
	if cfg.Kintone.Auth.Username != "" {
		fmt.Printf("  ユーザー: %s\n", cfg.Kintone.Auth.Username)
		fmt.Printf("  パスワード: %s\n", "********")
	} else {
		fmt.Printf("  認証: %s\n", yellow("未設定"))
	}

	// ターゲット
	fmt.Printf("\n%s\n", cyan("ターゲット:"))
	if cfg.Targets.Desktop {
		fmt.Printf("  %s デスクトップ\n", green("✓"))
	} else {
		fmt.Printf("  ✗ デスクトップ\n")
	}
	if cfg.Targets.Mobile {
		fmt.Printf("  %s モバイル\n", green("✓"))
	} else {
		fmt.Printf("  ✗ モバイル\n")
	}

	// 適用範囲
	fmt.Printf("\n%s\n", cyan("適用範囲:"))
	switch cfg.Scope {
	case config.ScopeAll:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", green("✓"))
	case config.ScopeAdmin:
		fmt.Printf("  %s アプリ管理者のみ (ADMIN)\n", yellow("✓"))
	case config.ScopeNone:
		fmt.Printf("  ✗ 適用しない (NONE)\n")
	default:
		fmt.Printf("  %s すべてのユーザー (ALL)\n", green("✓"))
	}

	// 出力ファイル名
	fmt.Printf("\n%s\n", cyan("出力:"))
	fmt.Printf("  ファイル名: %s.js / %s.css\n", cfg.GetOutputName(), cfg.GetOutputName())

	// Dev設定
	fmt.Printf("\n%s\n", cyan("開発サーバー:"))
	fmt.Printf("  オリジン:     %s\n", cfg.Dev.Origin)
	fmt.Printf("  エントリー:   %s\n", cfg.Dev.Entry)

	fmt.Println()
	fmt.Println("Enterキーで戻る...")
	fmt.Scanln()
}

func editKintoneConfig(cfg *config.Config) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n%s kintone接続設定\n\n", cyan("🔧"))

	// ドメイン
	domain, err := prompt.AskDomain(cfg.Kintone.Domain)
	if err != nil {
		return err
	}
	cfg.Kintone.Domain = domain

	// アプリID
	var appIDStr string
	appIDPrompt := &survey.Input{
		Message: "アプリID:",
		Default: strconv.Itoa(cfg.Kintone.AppID),
	}
	if err := survey.AskOne(appIDPrompt, &appIDStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return fmt.Errorf("アプリIDは数値で入力してください")
	}
	cfg.Kintone.AppID = appID

	// 認証情報を更新するか確認
	var updateAuth bool
	authPrompt := &survey.Confirm{
		Message: "認証情報を更新しますか?",
		Default: false,
	}
	if err := survey.AskOne(authPrompt, &updateAuth); err != nil {
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

	fmt.Printf("\n%s kintone接続設定を更新しました\n", green("✓"))
	return nil
}

func editTargets(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()

	desktop, mobile, err := prompt.AskTargets(cfg.Targets.Desktop, cfg.Targets.Mobile)
	if err != nil {
		return err
	}

	cfg.Targets.Desktop = desktop
	cfg.Targets.Mobile = mobile

	fmt.Printf("\n%s ターゲットを更新しました\n", green("✓"))
	return nil
}

func editScope(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

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

	fmt.Printf("\n%s 適用範囲を更新しました\n", green("✓"))
	return nil
}

func editOutput(cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()

	output, err := prompt.AskOutput(cfg.GetOutputName())
	if err != nil {
		return err
	}

	cfg.Output = output

	fmt.Printf("\n%s 出力ファイル名を更新しました (%s.js / %s.css)\n", green("✓"), output, output)
	return nil
}

func editEntry(projectDir string, cfg *config.Config) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

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
		fmt.Printf("  %s src/ ディレクトリにエントリーファイルが見つかりません\n", yellow("⚠"))
		fmt.Println("Enterキーで戻る...")
		fmt.Scanln()
		return nil
	}

	// 現在の設定を先頭に表示
	currentEntry := cfg.Dev.Entry
	defaultIndex := 0
	for i, f := range entryFiles {
		if f == currentEntry {
			defaultIndex = i
			break
		}
	}

	var selected string
	selectPrompt := &survey.Select{
		Message: "エントリーファイルを選択:",
		Options: entryFiles,
		Default: entryFiles[defaultIndex],
	}
	if err := survey.AskOne(selectPrompt, &selected); err != nil {
		return err
	}

	cfg.Dev.Entry = selected

	fmt.Printf("\n%s エントリーファイルを更新しました (%s)\n", green("✓"), selected)
	return nil
}

func editFramework(projectDir string, cfg *config.Config) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\n%s フレームワークの変更\n\n", cyan("🔧"))

	// 現在のフレームワークを検出
	currentFramework := detectCurrentFramework(projectDir)
	currentLanguage := detectCurrentLanguage(projectDir)

	fmt.Printf("現在のフレームワーク: %s (%s)\n\n", cyan(string(currentFramework)), string(currentLanguage))

	// 新しいフレームワークを選択
	newFramework, err := prompt.AskFramework()
	if err != nil {
		return err
	}

	if newFramework == currentFramework {
		fmt.Printf("\n%s フレームワークは変更されていません\n", yellow("⚠"))
		fmt.Println("Enterキーで戻る...")
		fmt.Scanln()
		return nil
	}

	// 確認
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("%s から %s に変更しますか?", currentFramework, newFramework),
		Default: true,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	fmt.Println()

	// 1. package.json を更新
	fmt.Printf("  package.json を更新...")
	if err := updatePackageJSONFramework(projectDir, currentFramework, newFramework, currentLanguage); err != nil {
		fmt.Println()
		return fmt.Errorf("package.json更新エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// 2. vite.config.ts を再生成
	fmt.Printf("  vite.config.ts を再生成...")
	if err := generator.GenerateViteConfig(projectDir, newFramework, currentLanguage); err != nil {
		fmt.Println()
		return fmt.Errorf("vite.config.ts再生成エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// 3. eslint.config.js を再生成
	fmt.Printf("  eslint.config.js を再生成...")
	if err := generator.RegenerateESLintConfig(projectDir, newFramework, currentLanguage); err != nil {
		fmt.Println()
		return fmt.Errorf("eslint.config.js再生成エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// 4. node_modules を削除
	fmt.Printf("  node_modules を削除...")
	nodeModulesPath := filepath.Join(projectDir, "node_modules")
	if err := os.RemoveAll(nodeModulesPath); err != nil {
		fmt.Println()
		return fmt.Errorf("node_modules削除エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	// 5. パッケージマネージャーを検出してインストール
	pm := detectPackageManager(projectDir)
	fmt.Printf("\n%s パッケージを再インストール中... (%s)\n", cyan("→"), pm)

	installCmd := exec.Command(pm, "install")
	installCmd.Dir = projectDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("インストールエラー: %w", err)
	}

	// 6. config.json のエントリーパスを更新
	cfg.Dev.Entry = generator.GetEntryPath(newFramework, currentLanguage)

	fmt.Printf("\n%s フレームワークを %s に変更しました!\n\n", green("✓"), newFramework)
	fmt.Printf("%s src/ ディレクトリのコードを手動で書き換えてください\n", yellow("⚠"))
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

func updatePackageJSONFramework(projectDir string, oldFw, newFw prompt.Framework, lang prompt.Language) error {
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	deps, _ := pkg["dependencies"].(map[string]interface{})
	if deps == nil {
		deps = make(map[string]interface{})
		pkg["dependencies"] = deps
	}

	devDeps, _ := pkg["devDependencies"].(map[string]interface{})
	if devDeps == nil {
		devDeps = make(map[string]interface{})
		pkg["devDependencies"] = devDeps
	}

	// 旧フレームワークのパッケージを削除
	removeFrameworkPackages(deps, devDeps, oldFw)

	// 新フレームワークのパッケージを追加
	addFrameworkPackages(deps, devDeps, newFw, lang)

	// JSON を書き出し
	output, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkgPath, output, 0644)
}

func removeFrameworkPackages(deps, devDeps map[string]interface{}, fw prompt.Framework) {
	switch fw {
	case prompt.FrameworkReact:
		delete(deps, "react")
		delete(deps, "react-dom")
		delete(devDeps, "@vitejs/plugin-react")
		delete(devDeps, "eslint-plugin-react-hooks")
		delete(devDeps, "@types/react")
		delete(devDeps, "@types/react-dom")
	case prompt.FrameworkVue:
		delete(deps, "vue")
		delete(devDeps, "@vitejs/plugin-vue")
		delete(devDeps, "eslint-plugin-vue")
		delete(devDeps, "vue-tsc")
	case prompt.FrameworkSvelte:
		delete(deps, "svelte")
		delete(devDeps, "@sveltejs/vite-plugin-svelte")
		delete(devDeps, "eslint-plugin-svelte")
		delete(devDeps, "svelte-check")
		delete(devDeps, "tslib")
	}
}

func addFrameworkPackages(deps, devDeps map[string]interface{}, fw prompt.Framework, lang prompt.Language) {
	switch fw {
	case prompt.FrameworkReact:
		deps["react"] = "^18.2.0"
		deps["react-dom"] = "^18.2.0"
		devDeps["@vitejs/plugin-react"] = "^4.2.0"
		devDeps["eslint-plugin-react-hooks"] = "^5.0.0"
		if lang == prompt.LanguageTypeScript {
			devDeps["@types/react"] = "^18.2.0"
			devDeps["@types/react-dom"] = "^18.2.0"
		}
	case prompt.FrameworkVue:
		deps["vue"] = "^3.4.0"
		devDeps["@vitejs/plugin-vue"] = "^5.0.0"
		devDeps["eslint-plugin-vue"] = "^9.0.0"
		if lang == prompt.LanguageTypeScript {
			devDeps["vue-tsc"] = "^1.8.0"
		}
	case prompt.FrameworkSvelte:
		deps["svelte"] = "^4.2.0"
		devDeps["@sveltejs/vite-plugin-svelte"] = "^3.0.0"
		devDeps["eslint-plugin-svelte"] = "^2.0.0"
		if lang == prompt.LanguageTypeScript {
			devDeps["svelte-check"] = "^3.6.0"
			devDeps["tslib"] = "^2.6.0"
		}
	}
}
