package generator

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kintone/kcdev/internal/prompt"
)

//go:embed templates/*
var templates embed.FS

func GenerateProject(projectDir string, answers *prompt.InitAnswers) error {
	templateDir := fmt.Sprintf("%s-%s", answers.Framework, getLanguageShort(answers.Language))

	if err := copyTemplates(projectDir, templateDir); err != nil {
		return err
	}

	if err := generatePackageJSON(projectDir, answers); err != nil {
		return err
	}

	if err := generateIndexHTML(projectDir, answers); err != nil {
		return err
	}

	if err := generateGitignore(projectDir); err != nil {
		return err
	}

	if err := generateReadme(projectDir, answers); err != nil {
		return err
	}

	if err := generateESLintConfig(projectDir, answers.Framework, answers.Language); err != nil {
		return err
	}

	// TypeScript の場合は tsconfig.json と型定義プレースホルダーを生成
	if answers.Language == prompt.LanguageTypeScript {
		if err := generateTSConfig(projectDir, answers.Framework); err != nil {
			return err
		}
		if err := generateTypesPlaceholder(projectDir); err != nil {
			return err
		}
	}

	return nil
}

func generateIndexHTML(projectDir string, answers *prompt.InitAnswers) error {
	appURL := fmt.Sprintf("https://%s/k/%d/", answers.Domain, answers.AppID)
	content := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>kcdev - Redirecting...</title>
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
      margin: 0;
      background: #f5f5f5;
    }
    .container {
      text-align: center;
      padding: 40px;
      background: white;
      border-radius: 8px;
      box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    }
    h1 { color: #333; margin-bottom: 16px; }
    p { color: #666; margin-bottom: 24px; }
    a {
      display: inline-block;
      padding: 12px 24px;
      background: #0066cc;
      color: white;
      text-decoration: none;
      border-radius: 4px;
    }
    a:hover { background: #0052a3; }
  </style>
</head>
<body>
  <div class="container">
    <h1>kcdev Dev Server</h1>
    <p>SSL証明書が許可されました。kintoneアプリに移動します...</p>
    <a href="%s">kintoneアプリを開く</a>
  </div>
  <script>
    setTimeout(function() {
      window.location.href = "%s";
    }, 1500);
  </script>
</body>
</html>
`, appURL, appURL)

	kcdevDir := filepath.Join(projectDir, ".kcdev")
	if err := os.MkdirAll(kcdevDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(kcdevDir, "index.html"), []byte(content), 0644)
}

func getLanguageShort(lang prompt.Language) string {
	if lang == prompt.LanguageTypeScript {
		return "ts"
	}
	return "js"
}

func copyTemplates(projectDir string, templateDir string) error {
	templatePath := filepath.Join("templates", templateDir)

	return fs.WalkDir(templates, templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(projectDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := templates.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(targetPath, content, 0644)
	})
}

type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Private         bool              `json:"private"`
	Type            string            `json:"type"`
	Scripts         packageScripts    `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type packageScripts struct {
	Dev           string `json:"dev"`
	DevPreview    string `json:"dev:preview"`
	Build         string `json:"build"`
	Deploy        string `json:"deploy"`
	DeployPreview string `json:"deploy:preview"`
	Lint          string `json:"lint"`
	Types         string `json:"types,omitempty"`
}

func generatePackageJSON(projectDir string, answers *prompt.InitAnswers) error {
	scripts := packageScripts{
		Dev:           "kcdev dev",
		DevPreview:    "kcdev dev --preview",
		Build:         "kcdev build",
		Deploy:        "kcdev deploy",
		DeployPreview: "kcdev deploy --preview",
		Lint:          "eslint --config .kcdev/eslint.config.js src/",
	}

	// TypeScript の場合は types スクリプトを追加
	if answers.Language == prompt.LanguageTypeScript {
		scripts.Types = "kcdev types"
	}

	// dependencies は空で生成（npm install で最新版を追加）
	pkg := packageJSON{
		Name:            answers.ProjectName,
		Version:         "0.0.0",
		Private:         true,
		Type:            "module",
		Scripts:         scripts,
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
	}

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projectDir, "package.json"), data, 0644)
}

// GetPackageList はフレームワークと言語に応じたパッケージ名リストを返す（バージョンなし）
func GetPackageList(framework prompt.Framework, language prompt.Language) (deps []string, devDeps []string) {
	// 共通 devDependencies
	devDeps = append(devDeps, "vite", "eslint", "@eslint/js", "globals")

	if language == prompt.LanguageTypeScript {
		devDeps = append(devDeps, "typescript", "@kintone/dts-gen", "typescript-eslint")
	}

	switch framework {
	case prompt.FrameworkReact:
		deps = append(deps, "react", "react-dom")
		devDeps = append(devDeps, "@vitejs/plugin-react", "eslint-plugin-react-hooks")
		if language == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "@types/react", "@types/react-dom")
		}
	case prompt.FrameworkVue:
		deps = append(deps, "vue")
		devDeps = append(devDeps, "@vitejs/plugin-vue", "eslint-plugin-vue")
		if language == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "vue-tsc")
		}
	case prompt.FrameworkSvelte:
		deps = append(deps, "svelte")
		devDeps = append(devDeps, "@sveltejs/vite-plugin-svelte", "eslint-plugin-svelte")
		if language == prompt.LanguageTypeScript {
			devDeps = append(devDeps, "svelte-check", "tslib")
		}
	}

	return deps, devDeps
}

func generateTypesPlaceholder(projectDir string) error {
	typesDir := filepath.Join(projectDir, "src", "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		return err
	}

	content := `// このファイルは kcdev types コマンドで自動生成されます
// kintone アプリのフィールド型定義が含まれます
//
// 生成するには: kcdev types
//
// 注意: このファイルは手動で編集しないでください

declare namespace kintone.types {
  // kcdev types を実行すると、ここにフィールド型が生成されます
}
`
	return os.WriteFile(filepath.Join(typesDir, "kintone.d.ts"), []byte(content), 0644)
}

func generateTSConfig(projectDir string, framework prompt.Framework) error {
	var jsx string
	switch framework {
	case prompt.FrameworkReact:
		jsx = `"jsx": "react-jsx",`
	default:
		jsx = ""
	}

	// jsx の行がある場合は改行を追加
	if jsx != "" {
		jsx = "\n    " + jsx
	}

	content := fmt.Sprintf(`{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,%s
    "esModuleInterop": true,
    "skipLibCheck": true,
    "types": []
  },
  "files": [
    "./node_modules/@kintone/dts-gen/kintone.d.ts",
    "./src/types/kintone.d.ts"
  ],
  "include": [
    "src/**/*"
  ]
}
`, jsx)

	return os.WriteFile(filepath.Join(projectDir, "tsconfig.json"), []byte(content), 0644)
}

func generateGitignore(projectDir string) error {
	content := `# Dependencies
node_modules/

# Build output
dist/

# Environment
.env
.env.local

# kcdev (sensitive)
.kcdev/config.json
.kcdev/certs/

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
`
	return os.WriteFile(filepath.Join(projectDir, ".gitignore"), []byte(content), 0644)
}

func generateReadme(projectDir string, answers *prompt.InitAnswers) error {
	content := fmt.Sprintf(`# %s

kintone カスタマイズプロジェクト

## 開発

`+"```"+`bash
npm install
kcdev dev
`+"```"+`

## ビルド

`+"```"+`bash
kcdev build
`+"```"+`

## デプロイ

`+"```"+`bash
kcdev deploy
`+"```"+`

## 設定

- kintone ドメイン: %s
- アプリ ID: %d
`, answers.ProjectName, answers.Domain, answers.AppID)

	return os.WriteFile(filepath.Join(projectDir, "README.md"), []byte(content), 0644)
}

// RegenerateESLintConfig は既存プロジェクトのESLint設定を再生成する
func RegenerateESLintConfig(projectDir string, framework prompt.Framework, language prompt.Language) error {
	return generateESLintConfig(projectDir, framework, language)
}

func generateESLintConfig(projectDir string, framework prompt.Framework, language prompt.Language) error {
	var content string

	switch framework {
	case prompt.FrameworkReact:
		content = generateESLintReact(language)
	case prompt.FrameworkVue:
		content = generateESLintVue(language)
	case prompt.FrameworkSvelte:
		content = generateESLintSvelte(language)
	default:
		content = generateESLintVanilla(language)
	}

	kcdevDir := filepath.Join(projectDir, ".kcdev")
	if err := os.MkdirAll(kcdevDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(kcdevDir, "eslint.config.js"), []byte(content), 0644)
}

func generateESLintVanilla(language prompt.Language) string {
	if language == prompt.LanguageTypeScript {
		return `import js from "@eslint/js";
import tseslint from "typescript-eslint";
import globals from "globals";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  }
);
`
	}
	return `import js from "@eslint/js";
import globals from "globals";

export default [
  js.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  },
];
`
}

func generateESLintReact(language prompt.Language) string {
	if language == prompt.LanguageTypeScript {
		return `import js from "@eslint/js";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
    },
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  }
);
`
	}
	return `import js from "@eslint/js";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";

export default [
  js.configs.recommended,
  {
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
    },
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  },
];
`
}

func generateESLintVue(language prompt.Language) string {
	if language == prompt.LanguageTypeScript {
		return `import js from "@eslint/js";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import globals from "globals";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  }
);
`
	}
	return `import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import globals from "globals";

export default [
  js.configs.recommended,
  ...pluginVue.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  },
];
`
}

func generateESLintSvelte(language prompt.Language) string {
	if language == prompt.LanguageTypeScript {
		return `import js from "@eslint/js";
import tseslint from "typescript-eslint";
import svelte from "eslint-plugin-svelte";
import globals from "globals";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  }
);
`
	}
	return `import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import globals from "globals";

export default [
  js.configs.recommended,
  ...svelte.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        kintone: "readonly",
      },
    },
  },
  {
    ignores: ["dist/", ".kcdev/"],
  },
];
`
}

func GetEntryPath(framework prompt.Framework, language prompt.Language) string {
	ext := getEntryExtension(framework, language)
	return fmt.Sprintf("/src/main.%s", ext)
}

func getEntryExtension(framework prompt.Framework, language prompt.Language) string {
	switch framework {
	case prompt.FrameworkReact:
		if language == prompt.LanguageTypeScript {
			return "tsx"
		}
		return "jsx"
	case prompt.FrameworkVue, prompt.FrameworkSvelte:
		if language == prompt.LanguageTypeScript {
			return "ts"
		}
		return "js"
	default:
		if language == prompt.LanguageTypeScript {
			return "ts"
		}
		return "js"
	}
}

func GetMainFileName(framework prompt.Framework, language prompt.Language) string {
	return fmt.Sprintf("main.%s", getEntryExtension(framework, language))
}

func replaceTemplateVars(content string, answers *prompt.InitAnswers) string {
	content = strings.ReplaceAll(content, "{{PROJECT_NAME}}", answers.ProjectName)
	return content
}
