<div align="center">

# kcdev

**kintone カスタマイズ開発を超簡単に**

[![npm version](https://img.shields.io/npm/v/@zygapp/kintone-customize-devtool.svg)](https://www.npmjs.com/package/@zygapp/kintone-customize-devtool)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Built_with-Go-00ADD8.svg)](https://go.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-18+-green.svg)](https://nodejs.org/)

<br />

Vite + HMR で kintone カスタマイズ開発を快適に。<br />
コードを保存すれば、即座に kintone 画面に反映されます。

</div>

---

## Features

| | |
|---|---|
| **Hot Module Replacement** | コード変更が即座に kintone 画面に反映。ページリロード不要。 |
| **モダンフレームワーク対応** | React / Vue / Svelte / Vanilla に対応 |
| **TypeScript サポート** | 型安全な開発環境を標準提供 |
| **シンプルなワークフロー** | `init` → `dev` → `build` → `deploy` の 4 ステップ |
| **クロスプラットフォーム** | macOS / Linux / Windows（Intel & ARM） |

---

## Quick Start

### 1. インストール

```bash
npm install -g @zygapp/kintone-customize-devtool
```

### 2. プロジェクト作成

```bash
kcdev init my-app
cd my-app
```

対話形式で以下を設定します：
- kintone ドメイン（例：`example.cybozu.com`）
- アプリ ID
- フレームワーク（React / Vue / Svelte / Vanilla）
- 言語（TypeScript / JavaScript）
- カスタマイズ対象（デスクトップ / モバイル）
- 認証情報

### 3. 開発開始

```bash
kcdev dev
```

ブラウザが自動で開きます。`https://localhost:3000` の SSL 証明書を許可すると、kintone アプリにリダイレクトされます。

### 4. 本番デプロイ

```bash
kcdev build && kcdev deploy
```

---

## Commands

### `kcdev init [project-name]`

新しいプロジェクトを初期化します。

```bash
kcdev init my-app
```

**オプション:**

| オプション | 説明 |
|-----------|------|
| `-d, --domain` | kintone ドメイン |
| `-a, --app` | アプリ ID |
| `-f, --framework` | フレームワーク（react / vue / svelte / vanilla） |
| `-l, --language` | 言語（typescript / javascript） |
| `-u, --username` | kintone ユーザー名 |
| `-p, --password` | kintone パスワード |
| `--create-dir` | プロジェクトディレクトリを作成 |
| `--no-create-dir` | カレントディレクトリに展開 |

### `kcdev dev`

開発サーバーを起動し、ローダーを kintone に自動デプロイします。

既存のカスタマイズがある場合は確認プロンプトが表示されます。

```bash
kcdev dev
```

**オプション:**

| オプション | 説明 |
|-----------|------|
| `--skip-deploy` | ローダーのデプロイをスキップ（2回目以降の起動時に便利） |
| `--no-browser` | ブラウザを自動で開かない |
| `-f, --force` | 既存カスタマイズの確認をスキップして上書き |

### `kcdev build`

本番用ビルドを生成します。IIFE 形式で `dist/` に出力されます。

```bash
kcdev build
```

**出力ファイル:**
- `dist/customize.js`
- `dist/customize.css`（CSS がある場合）

### `kcdev deploy`

ビルド成果物を kintone にデプロイします。

既存のカスタマイズがある場合は確認プロンプトが表示されます。

```bash
kcdev build && kcdev deploy
```

**オプション:**

| オプション | 説明 |
|-----------|------|
| `-f, --force` | 既存カスタマイズの確認をスキップして上書き |

### `kcdev update`

Vite およびフレームワークプラグインを最新版に更新します。

```bash
kcdev update
```

---

## Project Structure

```
my-app/
├── src/
│   ├── main.tsx          # エントリーポイント
│   ├── App.tsx           # メインコンポーネント
│   └── style.css         # スタイル
├── .kcdev/
│   ├── config.json       # 設定ファイル
│   ├── vite.config.ts    # Vite 設定（自動生成）
│   ├── certs/            # SSL 証明書
│   └── managed/          # ローダー（自動生成）
├── dist/                 # ビルド出力
├── package.json
├── .env                  # 認証情報
└── .gitignore
```

---

## Authentication

認証情報は以下の優先順位で取得されます：

### 1. `.env` ファイル（推奨）

```env
KCDEV_USERNAME=your-username
KCDEV_PASSWORD=your-password
```

### 2. `.kcdev/config.json`

```json
{
  "kintone": {
    "auth": {
      "username": "your-username",
      "password": "your-password"
    }
  }
}
```

> **Note:** `.env` と `.kcdev/config.json` は `.gitignore` に追加されます。認証情報をリポジトリにコミットしないでください。

---

## SSL Certificate

開発サーバーは HTTPS で起動します。初回アクセス時に自己署名証明書の警告が表示されます。

### 証明書を信頼する方法

1. `https://localhost:3000` にアクセス
2. ブラウザの警告画面で「詳細設定」→「安全でないサイトへ進む」を選択
3. または、OS の証明書ストアに `.kcdev/certs/localhost.pem` を登録

---

## How It Works

```
kintone
  ↓ classic script
kintone-dev-loader.js
  ↓ dynamic import
Vite dev server (ESM + HMR)
  ↓
src/main.*
```

kintone は classic script のみ対応ですが、kcdev は開発時に Vite の ESM + HMR を活用できるようローダーを自動生成・デプロイします。

**開発者が意識するのは `src/` 以下のコードだけ。** ローダーや設定ファイルは kcdev が管理します。

### CLI について

kcdev は Go で実装されたネイティブバイナリです。npm 経由でインストールすると、OS・アーキテクチャに応じた実行ファイルが自動選択されます。高速な起動と安定した動作を実現しています。

---

## Requirements

- **Node.js** 18 以上
- **kintone** 環境（cybozu.com）

---

## Supported Platforms

| OS | Architecture |
|----|--------------|
| macOS | Intel (x64) / Apple Silicon (arm64) |
| Linux | x64 / arm64 |
| Windows | x64 / arm64 |

---

## Troubleshooting

### HMR が動作しない

- `https://localhost:3000` の SSL 証明書を許可しているか確認してください
- ブラウザの開発者ツールでコンソールエラーを確認してください

### ローダーのデプロイに失敗する

- `.env` または `.kcdev/config.json` の認証情報を確認してください
- kintone アプリの管理権限があるか確認してください

### Windows で証明書エラーが出る

- PowerShell を管理者権限で実行し、証明書をインポートしてください
- または、ブラウザで `https://localhost:3000` にアクセスして手動で許可してください

---

## License

MIT
