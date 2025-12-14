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

	return os.WriteFile(filepath.Join(projectDir, "index.html"), []byte(content), 0644)
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

func generatePackageJSON(projectDir string, answers *prompt.InitAnswers) error {
	deps := getFrameworkDependencies(answers.Framework, answers.Language)

	pkg := map[string]interface{}{
		"name":    answers.ProjectName,
		"version": "0.0.0",
		"private": true,
		"type":    "module",
		"scripts": map[string]string{
			"init": "kcdev init",
			"dev": "kcdev dev",
			"dev:preview": "kcdev dev --preview",
			"build": "kcdev build",
			"deploy": "kcdev deploy",
			"deploy:preview": "kcdev deploy --preview",
		},
		"dependencies":    deps.dependencies,
		"devDependencies": deps.devDependencies,
	}

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projectDir, "package.json"), data, 0644)
}

type dependencies struct {
	dependencies    map[string]string
	devDependencies map[string]string
}

func getFrameworkDependencies(framework prompt.Framework, language prompt.Language) dependencies {
	deps := dependencies{
		dependencies:    make(map[string]string),
		devDependencies: make(map[string]string),
	}

	deps.devDependencies["vite"] = "^5.0.0"

	if language == prompt.LanguageTypeScript {
		deps.devDependencies["typescript"] = "^5.3.0"
	}

	switch framework {
	case prompt.FrameworkReact:
		deps.dependencies["react"] = "^18.2.0"
		deps.dependencies["react-dom"] = "^18.2.0"
		deps.devDependencies["@vitejs/plugin-react"] = "^4.2.0"
		if language == prompt.LanguageTypeScript {
			deps.devDependencies["@types/react"] = "^18.2.0"
			deps.devDependencies["@types/react-dom"] = "^18.2.0"
		}
	case prompt.FrameworkVue:
		deps.dependencies["vue"] = "^3.4.0"
		deps.devDependencies["@vitejs/plugin-vue"] = "^5.0.0"
		if language == prompt.LanguageTypeScript {
			deps.devDependencies["vue-tsc"] = "^1.8.0"
		}
	case prompt.FrameworkSvelte:
		deps.dependencies["svelte"] = "^4.2.0"
		deps.devDependencies["@sveltejs/vite-plugin-svelte"] = "^3.0.0"
		if language == prompt.LanguageTypeScript {
			deps.devDependencies["svelte-check"] = "^3.6.0"
			deps.devDependencies["tslib"] = "^2.6.0"
		}
	}

	return deps
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
