# kcdev（kintone customize developer）仕様書 v0.4

## 1. 目的

kintoneカスタマイズ開発を Tailwind v4 のようなコマンド中心のDX で行うためのCLIツール。

- `kcdev init` で雛形と内部設定を生成
- `kcdev dev` で Vite dev server + HMR を使い、kintone画面にリアルタイム反映
- `kcdev build` で本番用 classic script（IIFE）を生成
- `kcdev deploy` で kintone に API 経由デプロイ
- `kcdev config` で対話形式でプロジェクト設定を変更
- `kcdev types` で TypeScript 型定義を生成
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
3. kintoneドメイン（例：`example.cybozu.com`）※自動補完対応
4. アプリID
5. フレームワーク選択：`React` | `Vue` | `Svelte` | `Vanilla`
6. 言語選択：`TypeScript` | `JavaScript`
7. 出力ファイル名（デフォルト：`customize`）
8. カスタマイズ対象：`デスクトップ` | `モバイル`（複数選択可）
9. 適用範囲：`すべてのユーザー (ALL)` | `アプリ管理者のみ (ADMIN)` | `適用しない (NONE)`
10. 認証情報（ユーザー名、パスワード）
11. パッケージマネージャー選択：`npm` | `pnpm` | `yarn` | `bun`

#### CLIオプション

| オプション | 説明 |
|-----------|------|
| `-d, --domain` | kintone ドメイン（自動補完対応） |
| `-a, --app` | アプリ ID |
| `-f, --framework` | フレームワーク（react / vue / svelte / vanilla） |
| `-l, --language` | 言語（typescript / javascript） |
| `-o, --output` | 出力ファイル名（拡張子なし、デフォルト: customize） |
| `-u, --username` | kintone ユーザー名 |
| `-p, --password` | kintone パスワード |
| `-m, --package-manager` | パッケージマネージャー（npm / pnpm / yarn / bun） |
| `--desktop` | デスクトップを対象に含める |
| `--mobile` | モバイルを対象に含める |
| `-s, --scope` | 適用範囲（all / admin / none） |
| `--create-dir` | プロジェクトディレクトリを作成 |
| `--no-create-dir` | カレントディレクトリに展開 |

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
│  ├ eslint.config.js
│  ├ index.html
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

1. 既存のカスタマイズ設定を確認
   - kcdev管理ファイル（`customize.js`, `customize.css`, `kintone-dev-loader.js`）以外がある場合は確認プロンプトを表示
2. ローダー（`.kcdev/managed/kintone-dev-loader.js`）をkintoneにアップロード
3. アプリのJSカスタマイズ設定を更新
4. アプリをデプロイ
5. Vite dev server を起動（`https://localhost:3000`）
6. ブラウザを自動で開く

#### オプション

- `--skip-deploy`: ローダーのデプロイをスキップ（2回目以降の起動時など）
- `--no-browser`: ブラウザを自動で開かない
- `-f, --force`: 既存カスタマイズの確認をスキップして上書き
- `-p, --preview`: プレビュー環境のみにデプロイ（本番反映しない）

#### 起動時の表示

```
⣾ ローダーをデプロイ中...
✓ ローダーをデプロイしました

→ 開発サーバーを起動中...
  https://localhost:3000
  ブラウザで証明書を許可してください
```

### 6.4 kcdev build

#### 目的

本番用ビルド生成

#### 動作

1. バージョン確認プロンプトを表示（パッチ / マイナー / メジャー / カスタム）
2. バージョンが選択された場合、`package.json` を更新
3. `vite build` を実行

#### 内容

- `vite build`
- 出力形式：IIFE
- 出力先：`dist/`
  - `{output}.js`（デフォルト: `customize.js`）
  - `{output}.css`（必要な場合）
- 自動削除：`console.log`, `console.info`, `console.debug`, `console.warn`, `console.trace`, `debugger`
- 残す：`console.error`

#### オプション

| オプション | 説明 |
|-----------|------|
| `--no-minify` | minify を無効化（デバッグ用） |
| `--skip-version` | バージョン確認をスキップ |

#### 起動時の表示

```
現在のバージョン: 1.0.0
? バージョンを更新しますか? はい
? バージョンを選択 パッチ更新 (1.0.1)
✓ バージョンを更新: 1.0.0 → 1.0.1

⣾ ビルド中...
✓ ビルド完了!
出力ファイル:
  dist/customize.js
  dist/customize.css
```

### 6.5 kcdev deploy

#### 目的

kintone に API 経由で反映

#### 動作

1. `dist/` が存在しない場合は自動で `kcdev build` を実行
2. `dist/` が存在する場合は再ビルドするか確認プロンプトを表示
3. 既存のカスタマイズ設定を確認
   - kcdev管理ファイル（`{output}.js`, `{output}.css`, `kintone-dev-loader.js`）以外がある場合は確認プロンプトを表示
4. `POST /k/v1/file.json`（JS/CSSアップロード）
5. `PUT /k/v1/preview/app/customize.json`
6. `POST /k/v1/preview/app/deploy.json`
7. `GET /k/v1/preview/app/deploy.json`（完了待ち）

#### オプション

| オプション | 説明 |
|-----------|------|
| `-f, --force` | 既存カスタマイズの確認をスキップして上書き |
| `-p, --preview` | プレビュー環境のみにデプロイ（本番反映しない） |
| `--skip-version` | バージョン確認をスキップ |

#### 認証

- `X-Cybozu-Authorization: base64(username:password)`
- `.env` → `.kcdev/config.json` の順で取得

#### 起動時の表示

```
? dist/ が存在します。再ビルドしますか? はい
[ビルド処理...]

⣾ デプロイ中...
✓ 完了! https://example.cybozu.com/k/123/
```

### 6.6 kcdev types

#### 目的

TypeScript プロジェクトで kintone アプリのフィールド型定義を生成

#### 動作

1. `@kintone/dts-gen` を使用して型定義を生成
2. 出力先：`src/types/kintone.d.ts`
3. 認証情報は `.env` → `.kcdev/config.json` の順で取得

#### 起動時の表示

```
→ 型定義を生成中...
[dts-gen output]
✓ 型定義を生成しました: src/types/kintone.d.ts
```

#### 補足

- TypeScript プロジェクトでは、`kcdev init` 実行時に自動的に型定義が生成される
- フィールドを追加・変更した場合は、このコマンドで再生成

### 6.7 kcdev update

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

### 6.8 kcdev config

#### 目的

対話形式でプロジェクト設定を変更

#### 設定可能な項目

- `kintone`: kintone 接続設定（ドメイン、アプリ ID、認証情報）
- `targets`: カスタマイズ対象（デスクトップ / モバイル）
- `scope`: 適用範囲（ALL / ADMIN / NONE）
- `output`: 出力ファイル名
- `entry`: エントリーファイル
- `framework`: フレームワーク変更

#### フレームワーク変更時の動作

1. 古いフレームワークの依存パッケージをアンインストール
2. 新しいフレームワークの依存パッケージをインストール
3. `.kcdev/vite.config.ts` を再生成
4. `.kcdev/eslint.config.js` を再生成

#### 起動時の表示

```
? 変更する項目を選択
  > kintone接続設定
    ターゲット（デスクトップ/モバイル）
    適用範囲
    出力ファイル名
    エントリーファイル
    フレームワーク
```

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
  },
  "targets": {
    "desktop": true,
    "mobile": false
  },
  "output": "customize",
  "scope": "ALL"
}
```

### フィールド説明

| フィールド | 説明 |
|-----------|------|
| `kintone.domain` | kintone ドメイン |
| `kintone.appId` | アプリ ID |
| `kintone.auth` | 認証情報（`.env` 推奨） |
| `dev.origin` | 開発サーバーの URL |
| `dev.entry` | エントリーファイルのパス |
| `targets.desktop` | デスクトップを対象にするか |
| `targets.mobile` | モバイルを対象にするか |
| `output` | 出力ファイル名（拡張子なし） |
| `scope` | 適用範囲（ALL / ADMIN / NONE） |

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
- `.kcdev/eslint.config.js` - ESLint設定を共有
- `.kcdev/index.html` - 開発用HTMLを共有
- `.kcdev/managed/` - ローダーとメタデータを共有

## 11. 実装技術（Go本体）

| 用途 | パッケージ |
|------|-----------|
| CLI | github.com/spf13/cobra |
| 対話・フォーム | github.com/charmbracelet/huh |
| スピナー | github.com/charmbracelet/huh/spinner |
| スタイル | github.com/charmbracelet/lipgloss |
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
