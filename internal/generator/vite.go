package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/prompt"
)

func GenerateViteConfig(projectDir string, framework prompt.Framework, language prompt.Language) error {
	kcdevDir := filepath.Join(projectDir, config.ConfigDir)
	if err := os.MkdirAll(kcdevDir, 0755); err != nil {
		return err
	}

	content := generateViteConfigContent(framework, language)
	return os.WriteFile(filepath.Join(kcdevDir, "vite.config.ts"), []byte(content), 0644)
}

func generateViteConfigContent(framework prompt.Framework, language prompt.Language) string {
	imports := getViteImports(framework)
	plugins := getVitePlugins(framework)
	entry := GetEntryPath(framework, language)

	return fmt.Sprintf(`import { defineConfig, build } from 'vite'
%s
import fs from 'fs'
import path from 'path'

const certDir = path.resolve(__dirname, 'certs')
const srcEntry = path.resolve(__dirname, '..%s')

let cachedBundle: string | null = null

const kcdevPlugin = {
  name: 'kcdev',
  configureServer(server) {
    // CORS/PNA ヘッダー
    server.middlewares.use((req, res, next) => {
      res.setHeader('Access-Control-Allow-Origin', '*')
      res.setHeader('Access-Control-Allow-Methods', 'GET, OPTIONS')
      res.setHeader('Access-Control-Allow-Headers', '*')
      res.setHeader('Access-Control-Allow-Private-Network', 'true')

      if (req.method === 'OPTIONS') {
        res.statusCode = 204
        res.end()
        return
      }
      next()
    })

    // /customize.js - Vite でリアルタイムバンドル
    server.middlewares.use(async (req, res, next) => {
      if (!req.url?.startsWith('/customize.js')) {
        return next()
      }

      try {
        const result = await build({
          configFile: false,
          logLevel: 'silent',
          plugins: [%s],
          define: {
            'process.env.NODE_ENV': JSON.stringify('development'),
          },
          build: {
            write: false,
            lib: {
              entry: srcEntry,
              name: 'customize',
              formats: ['iife'],
              fileName: () => 'customize.js',
            },
            rollupOptions: {
              output: {
                assetFileNames: 'customize.[ext]',
              },
            },
          },
        })

        const output = Array.isArray(result) ? result[0] : result
        const jsChunk = output.output.find((o: any) => o.fileName === 'customize.js')
        const cssChunk = output.output.find((o: any) => o.fileName?.endsWith('.css'))

        if (jsChunk && 'code' in jsChunk) {
          let code = jsChunk.code

          // CSS をインライン化
          if (cssChunk && 'source' in cssChunk) {
            const cssCode = ` + "`" + `(function(){var s=document.createElement('style');s.textContent=${JSON.stringify(cssChunk.source)};document.head.appendChild(s);})();` + "`" + `
            code = cssCode + code
          }

          cachedBundle = code
          res.setHeader('Content-Type', 'application/javascript')
          res.end(code)
        } else {
          throw new Error('Build output not found')
        }
      } catch (err) {
        console.error('Build error:', err)
        res.statusCode = 500
        res.end(` + "`" + `console.error(${JSON.stringify(String(err))})` + "`" + `)
      }
    })
  },
  // ソース変更時にフルリロード
  handleHotUpdate({ server }) {
    cachedBundle = null
    server.ws.send({ type: 'full-reload' })
    return []
  },
}

export default defineConfig({
  plugins: [%skcdevPlugin],
  server: {
    https: {
      key: fs.readFileSync(path.join(certDir, 'localhost-key.pem')),
      cert: fs.readFileSync(path.join(certDir, 'localhost.pem')),
    },
    port: 3000,
    strictPort: true,
    origin: 'https://localhost:3000',
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
  },
  build: {
    lib: {
      entry: srcEntry,
      name: 'customize',
      formats: ['iife'],
      fileName: () => 'customize.js',
    },
    outDir: path.resolve(__dirname, '../dist'),
    emptyOutDir: true,
    rollupOptions: {
      output: {
        assetFileNames: 'customize.[ext]',
      },
    },
  },
  esbuild: {
    drop: ['debugger'],
    pure: ['console.log', 'console.info', 'console.debug', 'console.warn', 'console.trace'],
  },
})
`, imports, entry, getBuildPlugins(framework), plugins)
}

func getViteImports(framework prompt.Framework) string {
	switch framework {
	case prompt.FrameworkReact:
		return "import react from '@vitejs/plugin-react'"
	case prompt.FrameworkVue:
		return "import vue from '@vitejs/plugin-vue'"
	case prompt.FrameworkSvelte:
		return "import { svelte } from '@sveltejs/vite-plugin-svelte'"
	default:
		return ""
	}
}

func getVitePlugins(framework prompt.Framework) string {
	switch framework {
	case prompt.FrameworkReact:
		return "react(), "
	case prompt.FrameworkVue:
		return "vue(), "
	case prompt.FrameworkSvelte:
		return "svelte(), "
	default:
		return ""
	}
}

func getBuildPlugins(framework prompt.Framework) string {
	switch framework {
	case prompt.FrameworkReact:
		return "react()"
	case prompt.FrameworkVue:
		return "vue()"
	case prompt.FrameworkSvelte:
		return "svelte()"
	default:
		return ""
	}
}
