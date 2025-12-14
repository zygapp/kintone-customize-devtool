# kcdev（kintone customize developer）仕様書 v0.1

## 1. 目的

kintoneカスタマイズ開発を Tailwind v4 のようなコマンド中心のDX で行うためのCLIツール。

- `kcdev init` で雛形と内部設定を生成
- `kcdev dev` で Vite dev server + HMR を使い、kintone画面にリアルタイム反映
- `kcdev build` で本番用 classic script（IIFE）を生成
- `kcdev deploy` で kintone に API 経由デプロイ
- `kcdev update` で依存パッケージを最新版に更新

## 2. 基本思想（重要）

- kintoneは classic script のみ対応
- Vite dev は ESM

よって dev 時は以下の構造を取る：

```
kintone
  ↓ classic script
kintone-dev-loader.js
  ↓ dynamic import
Vite dev server (ESM + HMR)
  ↓
src/main.*
```

- エンジニアは `src/main.*` 以下だけを意識すればよい
- loader は触らない・安定資産

## 3. 非目的（v0.1ではやらない）

- チーム共有用トンネル（ngrok / Cloudflare Tunnel）
- OSの証明書ストアへの信頼登録自動化
- プラグインzip生成
- APIトークン / Basic認証対応

## 4. 対応環境

- OS：macOS / Linux / Windows（x64, ARM64）
- Node.js：18 以上（Vite実行用）
- CLI配布：npm公開（Go本体 + Nodeラッパー）

## 5. CLI 配布方式

### 採用方式

Go本体 + npmラッパー（全プラットフォームのバイナリを1パッケージに同梱）

### npmパッケージ構成

- `@zygapp/kintone-customize-devtool`（全プラットフォーム対応）

対応プラットフォーム：
- darwin-x64（macOS Intel）
- darwin-arm64（macOS Apple Silicon）
- linux-x64
- linux-arm64
- win32-x64
- win32-arm64

`npm install` 時に全バイナリがインストールされ、postinstall でOS/CPUに合ったバイナリが選択される。

## 6. コマンド仕様

### 6.1 kcdev init

#### 目的

- プロジェクト初期化
- Vite設定
- 自己署名証明書生成
- kintone-dev-loader と meta の生成

#### 対話フロー

1. ディレクトリ作成の確認
2. プロジェクト名
3. kintoneドメイン（例：`example.cybozu.com`）
4. アプリID
5. フレームワーク選択：`React` | `Vue` | `Svelte` | `Vanilla`
6. 言語選択：`TypeScript` | `JavaScript`
7. 認証情報（ユーザー名、パスワード）
8. パッケージマネージャー選択：`npm` | `pnpm` | `yarn` | `bun`

#### 対話スキップ条件

既存ファイルに応じて対話をスキップし、値を自動取得する：

| 既存ファイル | スキップする対話 | 取得方法 |
|-------------|----------------|---------|
| `package.json` | プロジェクト名、パッケージマネージャー | 既存プロジェクトとして扱う |
| `.kcdev/managed/loader.meta.json` | プロジェクト名、ドメイン、アプリID、フレームワーク、言語 | メタデータから取得 |
| `.kcdev/config.json` | kintoneドメイン、アプリID | 設定値から取得 |
| `.env`（認証情報あり） | ユーザー名、パスワード | 環境変数から取得 |

#### 認証情報の取得優先順位

1. `.env` の `KCDEV_USERNAME` / `KCDEV_PASSWORD`
2. `.kcdev/config.json` の `auth`
3. 対話で入力（password はマスク入力）

#### 生成物構成

```
./{projectName}/
├ src/
│  ├ main.(js|ts|tsx|vue|svelte)
│  ├ App.*
│  └ style.css
├ .kcdev/
│  ├ config.json
│  ├ vite.config.ts
│  ├ certs/
│  │  ├ localhost-key.pem
│  │  └ localhost.pem
│  └ managed/
│     ├ kintone-dev-loader.js
│     └ loader.meta.json
├ package.json
├ .gitignore
└ README.md
```

#### Vite設定の扱い

- `.kcdev/vite.config.ts` はkcdevが管理（フレームワーク選択に応じたプラグインを自動挿入）
- ユーザーはこのファイルを編集しない
- カスタマイズが必要な場合：プロジェクトルートに `vite.config.ts` を作成すると、そちらが優先される

### 6.2 証明書生成仕様

- `openssl` を使用
- SAN を必ず含める：
  - `DNS: localhost`
  - `IP: 127.0.0.1`
  - `IP: ::1`
- 生成先：`.kcdev/certs/`
- OSの信頼登録はユーザー手動

### 6.3 kcdev dev

#### 目的

- ローダーをkintoneに自動デプロイ
- Vite dev server 起動（https）
- HMR 有効

#### 動作

1. ローダー（`.kcdev/managed/kintone-dev-loader.js`）をkintoneにアップロード
2. アプリのJSカスタマイズ設定を更新
3. アプリをデプロイ
4. Vite dev server を起動（`https://localhost:3000`）
5. ブラウザを自動で開く

#### オプション

- `--skip-deploy`: ローダーのデプロイをスキップ（2回目以降の起動時など）
- `--no-browser`: ブラウザを自動で開かない

#### 起動時の表示

```
→ ローダーをkintoneにデプロイ中...
  アップロード... ✓
  設定... ✓
  デプロイ... ✓

→ Dev server を起動中...
  ➜  https://localhost:3000
  Entry:  /src/main.tsx
  Loader:  OK（再登録不要）
```

### 6.4 kcdev build

#### 目的

本番用ビルド生成

#### 内容

- `vite build`
- 出力形式：IIFE
- 出力先：`dist/`
  - `customize.js`
  - `customize.css`（必要な場合）

#### 起動時の表示

```
→ ビルドを開始...
○ バンドル中...
✓ ビルド完了!
出力ファイル:
  dist/customize.js
  dist/customize.css
```

### 6.5 kcdev deploy

#### 目的

kintone に API 経由で反映

#### 手順

1. `POST /k/v1/file.json`（JS/CSSアップロード）
2. `PUT /k/v1/preview/app/customize.json`
3. `POST /k/v1/preview/app/deploy.json`
4. `GET /k/v1/preview/app/deploy.json`（完了待ち）

#### 認証

- `X-Cybozu-Authorization: base64(username:password)`
- `.env` → `.kcdev/config.json` の順で取得

#### 起動時の表示

```
→ デプロイ中... (example.cybozu.com, App:123)
  JS... ✓
  CSS... ✓
  設定... ✓
  デプロイ... ✓

✓ 完了! https://example.cybozu.com/k/123/
```

### 6.6 kcdev update

#### 目的

Viteおよびフレームワークプラグインを最新版に更新

#### 更新対象

- vite
- @vitejs/plugin-react
- @vitejs/plugin-vue
- @sveltejs/vite-plugin-svelte
- typescript
- @types/react, @types/react-dom
- react, react-dom
- vue, vue-tsc
- svelte, svelte-check

#### 動作

1. package.json から更新対象パッケージを特定
2. ロックファイルからパッケージマネージャーを検出
3. 各パッケージマネージャーの update コマンドを実行

## 7. kintone-dev-loader.js 仕様

### 役割

kintone（classic）と Vite（ESM）をつなぐ唯一の橋

### 内容（例）

```javascript
// kcdev-loader
// schemaVersion: 1
// generatedAt: 2025-12-13T09:00:00+09:00
// origin: https://localhost:3000
// entry: /src/main.tsx

(function() {
  var origin = "https://localhost:3000";
  var entry = "/src/main.tsx";

  var xhr = new XMLHttpRequest();
  xhr.open("GET", origin + "/@vite/client", false);
  xhr.send();

  var script = document.createElement("script");
  script.type = "module";
  script.textContent = xhr.responseText;
  document.head.appendChild(script);

  import(origin + entry);
})();
```

### ルール

- `.kcdev/managed/` に配置
- kcdev は勝手に上書きしない
- 再生成は将来の明示コマンドでのみ行う

## 8. loader.meta.json 仕様

### 目的

- loader の生成条件と状態を記録
- 再登録が必要かどうかを判定

### スキーマ

```json
{
  "schemaVersion": 1,
  "kcdevVersion": "0.1.0",
  "generatedAt": "2025-12-13T09:00:00+09:00",
  "dev": {
    "origin": "https://localhost:3000",
    "entry": "/src/main.tsx"
  },
  "project": {
    "name": "sample",
    "framework": "react",
    "language": "typescript"
  },
  "kintone": {
    "domain": "example.cybozu.com",
    "appId": 123
  },
  "files": {
    "loaderPath": ".kcdev/managed/kintone-dev-loader.js",
    "loaderSha256": "hexstring...",
    "certKeyPath": ".kcdev/certs/localhost-key.pem",
    "certCertPath": ".kcdev/certs/localhost.pem"
  }
}
```

### 判定ルール

- loader の sha256 が不一致 → 再登録警告
- entry / origin が meta と不一致 → 再登録警告
- 自動再生成はしない

## 9. .kcdev/config.json 仕様

```json
{
  "kintone": {
    "domain": "example.cybozu.com",
    "appId": 123,
    "auth": {
      "username": "xxx",
      "password": "yyy"
    }
  },
  "dev": {
    "origin": "https://localhost:3000",
    "entry": "/src/main.tsx"
  }
}
```

### 優先順位

1. `.env`
2. `.kcdev/config.json`

## 10. Git管理ルール

`.gitignore` に必須：

```
.env
.kcdev/config.json
.kcdev/certs/
node_modules/
dist/
```

### 追跡対象とする理由

チーム開発を考慮し、以下は追跡対象とする：

- `.kcdev/vite.config.ts` - フレームワーク設定を共有
- `.kcdev/managed/` - ローダーとメタデータを共有

## 11. 実装技術（Go本体）

| 用途 | パッケージ |
|------|-----------|
| CLI | github.com/spf13/cobra |
| 対話 | github.com/AlecAivazis/survey/v2 |
| 色付き出力 | github.com/fatih/color |
| .env | github.com/joho/godotenv |
| HTTP | net/http |
| JSON | encoding/json |
| プロセス | os/exec |
| hash | crypto/sha256 |

## 12. 最重要設計原則（再掲）

1. loader は触らせない
2. main 以下だけ考えさせる
3. dev はローダーのみ自動デプロイ（ソースコードはdev serverから配信）
4. deploy は build 成果物だけ
