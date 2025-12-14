package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/generator"
	"github.com/kintone/kcdev/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	flagDomain    string
	flagAppID     int
	flagFramework string
	flagLanguage  string
	flagUsername  string
	flagPassword  string
	flagCreateDir bool
	flagNoCreateDir bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "新しいプロジェクトを初期化",
	Long:  `kintone カスタマイズ開発用の新しいプロジェクトを作成します。`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVarP(&flagDomain, "domain", "d", "", "kintone ドメイン")
	initCmd.Flags().IntVarP(&flagAppID, "app", "a", 0, "アプリ ID")
	initCmd.Flags().StringVarP(&flagFramework, "framework", "f", "", "フレームワーク (react|vue|svelte|vanilla)")
	initCmd.Flags().StringVarP(&flagLanguage, "language", "l", "", "言語 (typescript|javascript)")
	initCmd.Flags().StringVarP(&flagUsername, "username", "u", "", "kintone ユーザー名")
	initCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "kintone パスワード")
	initCmd.Flags().BoolVar(&flagCreateDir, "create-dir", false, "プロジェクトディレクトリを作成")
	initCmd.Flags().BoolVar(&flagNoCreateDir, "no-create-dir", false, "カレントディレクトリに展開")
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	}

	answers, err := collectAnswers(cwd, projectName)
	if err != nil {
		return err
	}

	var projectDir string
	if answers.CreateDir {
		projectDir = filepath.Join(cwd, answers.ProjectName)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return err
		}
	} else {
		projectDir = cwd
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// 既存プロジェクトかどうか判定
	isExisting := false
	if _, err := os.Stat(filepath.Join(projectDir, "package.json")); err == nil {
		isExisting = true
	}

	if isExisting {
		fmt.Printf("\n%s 既存プロジェクトを再初期化中...\n", cyan("→"))
	} else {
		fmt.Printf("\n%s プロジェクトを作成中...\n", cyan("→"))
		fmt.Printf("  テンプレート...")
		if err := generator.GenerateProject(projectDir, answers); err != nil {
			fmt.Println()
			return fmt.Errorf("プロジェクト生成エラー: %w", err)
		}
		fmt.Printf(" %s\n", green("✓"))
	}

	fmt.Printf("  Vite設定...")
	if err := generator.GenerateViteConfig(projectDir, answers.Framework, answers.Language); err != nil {
		fmt.Println()
		return fmt.Errorf("Vite設定生成エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	fmt.Printf("  ローダー...")
	if err := generator.GenerateLoader(projectDir, answers); err != nil {
		fmt.Println()
		return fmt.Errorf("ローダー生成エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	fmt.Printf("  証明書...")
	if err := generator.GenerateCerts(projectDir); err != nil {
		fmt.Println()
		return fmt.Errorf("証明書生成エラー: %w", err)
	}
	fmt.Printf(" %s\n", green("✓"))

	cfg := &config.Config{
		Kintone: config.KintoneConfig{
			Domain: answers.Domain,
			AppID:  answers.AppID,
			Auth: config.AuthConfig{
				Username: answers.Username,
				Password: answers.Password,
			},
		},
		Dev: config.DevConfig{
			Origin: "https://localhost:3000",
			Entry:  generator.GetEntryPath(answers.Framework, answers.Language),
		},
		Targets: config.TargetsConfig{
			Desktop: answers.TargetDesktop,
			Mobile:  answers.TargetMobile,
		},
		Scope: string(answers.Scope),
	}
	if err := cfg.Save(projectDir); err != nil {
		return fmt.Errorf("設定保存エラー: %w", err)
	}

	// 新規プロジェクトの場合、パッケージをインストール
	if !isExisting && answers.PackageManager != "" {
		fmt.Printf("\n%s パッケージをインストール中... (%s)\n", cyan("→"), answers.PackageManager)

		installCmd := exec.Command(string(answers.PackageManager), "install")
		installCmd.Dir = projectDir
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("インストールエラー: %w", err)
		}

		// TypeScript の場合、型定義を生成
		if answers.Language == prompt.LanguageTypeScript {
			if err := generateTypes(projectDir, cfg, answers.Username, answers.Password); err != nil {
				// 型定義生成の失敗は警告のみ（プロジェクト作成は成功として扱う）
				yellow := color.New(color.FgYellow).SprintFunc()
				fmt.Printf("\n%s 型定義の生成をスキップしました: %v\n", yellow("⚠"), err)
				fmt.Printf("  後で %s を実行して型定義を生成できます\n", cyan("kcdev types"))
			}
		}
	}

	printSuccess(projectDir, answers, isExisting)
	return nil
}

func collectAnswers(projectDir string, projectName string) (*prompt.InitAnswers, error) {
	answers := &prompt.InitAnswers{}

	// 既存プロジェクトかどうか判定
	isExisting := false
	if _, err := os.Stat(filepath.Join(projectDir, "package.json")); err == nil {
		isExisting = true
	}

	// loader.meta.json から既存の設定を読み込み
	meta, _ := generator.LoadLoaderMeta(projectDir)

	// ディレクトリ作成（既存プロジェクトでは不要）
	if isExisting {
		answers.CreateDir = false
	} else if flagCreateDir {
		answers.CreateDir = true
	} else if flagNoCreateDir {
		answers.CreateDir = false
	} else {
		createDir, err := prompt.AskCreateDir()
		if err != nil {
			return nil, err
		}
		answers.CreateDir = createDir
	}

	// プロジェクト名
	if projectName != "" {
		answers.ProjectName = projectName
	} else if meta != nil && meta.Project.Name != "" {
		answers.ProjectName = meta.Project.Name
	} else {
		defaultName := filepath.Base(projectDir)
		if answers.CreateDir {
			defaultName = "my-kintone-app"
		}
		name, err := prompt.AskProjectName(defaultName)
		if err != nil {
			return nil, err
		}
		answers.ProjectName = name
	}

	// ドメイン・アプリID
	if flagDomain != "" && flagAppID > 0 {
		answers.Domain = flagDomain
		answers.AppID = flagAppID
	} else if meta != nil && meta.Kintone.Domain != "" && meta.Kintone.AppID > 0 {
		answers.Domain = meta.Kintone.Domain
		answers.AppID = meta.Kintone.AppID
	} else if cfg, err := config.Load(projectDir); err == nil {
		answers.Domain = cfg.Kintone.Domain
		answers.AppID = cfg.Kintone.AppID
	} else {
		if flagDomain != "" {
			answers.Domain = flagDomain
		} else {
			domain, err := prompt.AskDomain("")
			if err != nil {
				return nil, err
			}
			answers.Domain = domain
		}

		if flagAppID > 0 {
			answers.AppID = flagAppID
		} else {
			appID, err := prompt.AskAppID(0)
			if err != nil {
				return nil, err
			}
			answers.AppID = appID
		}
	}

	// フレームワーク・言語
	if flagFramework != "" && flagLanguage != "" {
		answers.Framework = prompt.Framework(flagFramework)
		answers.Language = prompt.Language(flagLanguage)
	} else if meta != nil && meta.Project.Framework != "" {
		answers.Framework = prompt.Framework(meta.Project.Framework)
		answers.Language = prompt.Language(meta.Project.Language)
	} else if fw, lang := detectFromPackageJSON(projectDir); fw != "" {
		answers.Framework = fw
		answers.Language = lang
	} else {
		if flagFramework != "" {
			answers.Framework = prompt.Framework(flagFramework)
		} else {
			framework, err := prompt.AskFramework()
			if err != nil {
				return nil, err
			}
			answers.Framework = framework
		}

		if flagLanguage != "" {
			answers.Language = prompt.Language(flagLanguage)
		} else {
			language, err := prompt.AskLanguage()
			if err != nil {
				return nil, err
			}
			answers.Language = language
		}
	}

	// 認証情報
	if flagUsername != "" && flagPassword != "" {
		answers.Username = flagUsername
		answers.Password = flagPassword
	} else {
		envCfg, _ := config.LoadEnv(projectDir)
		if envCfg != nil && envCfg.HasAuth() {
			answers.Username = envCfg.Username
			answers.Password = envCfg.Password
		} else {
			if flagUsername != "" {
				answers.Username = flagUsername
			} else {
				username, err := prompt.AskUsername()
				if err != nil {
					return nil, err
				}
				answers.Username = username
			}

			if flagPassword != "" {
				answers.Password = flagPassword
			} else {
				password, err := prompt.AskPassword()
				if err != nil {
					return nil, err
				}
				answers.Password = password
			}
		}
	}

	// パッケージマネージャー（新規プロジェクトのみ）
	if !isExisting {
		pm, err := prompt.AskPackageManager()
		if err != nil {
			return nil, err
		}
		answers.PackageManager = pm
	}

	// カスタマイズ対象（desktop/mobile）
	defaultDesktop := true
	defaultMobile := false
	defaultScope := prompt.ScopeAll
	if cfg, err := config.Load(projectDir); err == nil {
		defaultDesktop = cfg.Targets.Desktop
		defaultMobile = cfg.Targets.Mobile
		// 既存設定がない場合のデフォルト
		if !defaultDesktop && !defaultMobile {
			defaultDesktop = true
		}
		if cfg.Scope != "" {
			defaultScope = prompt.Scope(cfg.Scope)
		}
	}
	desktop, mobile, err := prompt.AskTargets(defaultDesktop, defaultMobile)
	if err != nil {
		return nil, err
	}
	answers.TargetDesktop = desktop
	answers.TargetMobile = mobile

	// カスタマイズの適用範囲
	scope, err := prompt.AskScope(defaultScope)
	if err != nil {
		return nil, err
	}
	answers.Scope = scope

	return answers, nil
}

func detectProjectName(projectDir string) string {
	pkgPath := filepath.Join(projectDir, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return ""
	}
	return ""
}

func detectFromPackageJSON(projectDir string) (prompt.Framework, prompt.Language) {
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", ""
	}

	content := string(data)

	var framework prompt.Framework
	if strings.Contains(content, `"react"`) {
		framework = prompt.FrameworkReact
	} else if strings.Contains(content, `"vue"`) {
		framework = prompt.FrameworkVue
	} else if strings.Contains(content, `"svelte"`) {
		framework = prompt.FrameworkSvelte
	} else {
		return "", ""
	}

	var language prompt.Language
	if strings.Contains(content, `"typescript"`) {
		language = prompt.LanguageTypeScript
	} else if _, err := os.Stat(filepath.Join(projectDir, "src/main.ts")); err == nil {
		language = prompt.LanguageTypeScript
	} else if _, err := os.Stat(filepath.Join(projectDir, "src/main.tsx")); err == nil {
		language = prompt.LanguageTypeScript
	} else {
		language = prompt.LanguageJavaScript
	}

	return framework, language
}

func printSuccess(projectDir string, answers *prompt.InitAnswers, isExisting bool) {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if isExisting {
		fmt.Printf("%s プロジェクトを再初期化しました!\n\n", green("✓"))
		fmt.Printf("次のステップ:\n")
		fmt.Printf("  %s\n", cyan("kcdev dev"))
	} else {
		fmt.Printf("%s プロジェクトが作成されました!\n\n", green("✓"))
		fmt.Printf("次のステップ:\n")
		if answers.CreateDir {
			fmt.Printf("  %s %s\n", cyan("cd"), answers.ProjectName)
		}
		fmt.Printf("  %s\n", cyan("kcdev dev"))
	}
	fmt.Println()
}
