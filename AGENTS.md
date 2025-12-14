# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## プロジェクト概要

kcdev（kintone customize developer）は、Tailwind v4のようなコマンド中心のDXでkintoneカスタマイズ開発を行うためのCLIツールです。kintoneのclassic script環境と、モダンなVite dev server + HMRワークフローを橋渡しします。

## アーキテクチャ

kintoneはclassic scriptのみ対応だが、開発時はESM + HMRを使いたい。この課題を以下の構造で解決：

```
kintone
  ↓ classic script
kintone-dev-loader.js
  ↓ dynamic import
Vite dev server (ESM + HMR)
  ↓
src/main.*
```

ローダー（`.kcdev/managed/kintone-dev-loader.js`）はkintoneとViteを繋ぐ唯一の橋であり、手動で変更してはいけません。

## CLIコマンド

- `kcdev init` - プロジェクト初期化（プロジェクト名、kintoneドメイン、アプリID、フレームワーク、言語を対話形式で入力）
- `kcdev dev` - ローダーをkintoneにデプロイし、Vite dev serverをHTTPS（localhost:3000）で起動
- `kcdev build` - 本番用IIFEバンドルを`dist/`に生成
- `kcdev deploy` - ビルド成果物をkintoneにAPI経由でデプロイ
- `kcdev update` - Viteおよびフレームワークプラグインを最新版に更新

## 配布方式

Go本体 + npmラッパー（全プラットフォームのバイナリを1パッケージに同梱）:
- パッケージ名: `@zygapp/kintone-customize-devtool`
- 対応プラットフォーム: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64, win32-arm64

## 技術スタック

- Go CLI（cobra）
- Vite（dev server、ビルド）
- 自己署名証明書（`.kcdev/certs/`）
- 対応フレームワーク: React, Vue, Svelte, Vanilla（JS/TS）

## 主要ファイル

- `.kcdev/config.json` - プロジェクト設定
- `.kcdev/vite.config.ts` - Vite設定（kcdevが管理、変更禁止）
- `.kcdev/managed/kintone-dev-loader.js` - ローダースクリプト（変更禁止）
- `.kcdev/managed/loader.meta.json` - ローダーメタデータ（変更検知用）
- `.kcdev/certs/` - HTTPS用自己署名証明書

※ Vite設定をカスタマイズする場合は、プロジェクトルートに `vite.config.ts` を作成（そちらが優先される）

## 認証

優先順位: `.env` > `.kcdev/config.json`

認証対話スキップに必要な環境変数:
- `KCDEV_USERNAME`
- `KCDEV_PASSWORD`

## 設計原則

1. ローダーは安定資産 - 自動再生成しない
2. 開発者は`src/main.*`以下だけを意識する
3. devモードではローダーのみkintoneにデプロイ（ソースコードはVite dev serverから配信）
4. deployはビルド成果物のみアップロード

## コミットポリシー

### コミットメッセージ形式

```
<type>: <summary>
```

### type一覧

- `feat` - 新機能
- `fix` - バグ修正
- `docs` - ドキュメントのみの変更
- `refactor` - リファクタリング（機能変更なし）
- `test` - テストの追加・修正
- `chore` - ビルド、CI、依存関係などの雑務

### ルール

- メッセージは日本語で記述
- 1行目は50文字以内を目安
- 動詞で始める（「追加」「修正」「変更」など）
- 1コミット1目的（複数の変更を混ぜない）
- 絵文字を使わない
- クレジット（Co-Authored-Byなど）を記載しない
