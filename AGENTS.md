# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## 作業前の確認事項

質問や要望を受けた際は、すぐに回答・実装せず、以下の手順でプロジェクトを理解してから対応すること：

1. **ドキュメントを確認** - CLAUDE.md、AGENTS.md、SPECIFIC.md を読む
2. **関連コードを確認** - 要望に関連するソースコード（`internal/cmd/`配下など）を探して読む
3. **参考リポジトリがあれば確認** - 類似プロジェクトの実装を参照する
4. **要望の意図を正しく理解** - 上記を踏まえて、何を求められているか把握する

これを怠ると、的外れな実装や無駄な作業が発生する。

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

- `kcdev init` - プロジェクト初期化（プロジェクト名、kintoneドメイン、アプリID、フレームワーク、言語、カスタマイズ対象、適用範囲を対話形式で入力）
- `kcdev dev` - ローダーをkintoneにデプロイし、Vite dev serverをHTTPS（localhost:3000）で起動
  - 既存カスタマイズがある場合は確認プロンプトを表示
  - `-f, --force`: 確認をスキップして上書き
  - `-p, --preview`: プレビュー環境のみにデプロイ（本番反映しない）
- `kcdev build` - 本番用IIFEバンドルを`dist/`に生成（console.error以外のconsole.*とdebuggerは自動削除）
- `kcdev deploy` - ビルド成果物をkintoneにAPI経由でデプロイ
  - 既存カスタマイズがある場合は確認プロンプトを表示
  - `-f, --force`: 確認をスキップして上書き
  - `-p, --preview`: プレビュー環境のみにデプロイ（本番反映しない）
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

## バージョン管理

バージョンは **Makefile の `VERSION` 変数で一元管理**される。

```makefile
VERSION := 0.2.1
```

この値が以下に自動反映される:
- **Go CLI** - ビルド時に `-ldflags` で注入
- **npm パッケージ** - `make build-all` 時に `package.json` を自動更新

### リリース手順

1. `Makefile` の `VERSION` を更新
2. ビルド: `make build-all`（package.json も自動更新される）
3. コミット: `chore: バージョンを v0.x.x に更新`
4. タグ追加: `git tag v0.x.x && git push origin v0.x.x`
5. npm公開: `make npm-publish-token TOKEN=xxx`
6. GitHubリリース作成: 下記「GitHubリリース作成」セクション参照

### GitHubリリース作成

`gh release create` コマンドでリリースを作成する。リリースノートには各変更のコミットIDを含める（GitHubが自動でリンクに変換する）。

```bash
gh release create v0.x.x --title "v0.x.x" --notes "$(cat <<'EOF'
## 変更内容

### 新機能
- 機能の説明 (a1b2c3d)

### 改善
- 改善の説明 (e4f5g6h)

### バグ修正
- 修正の説明 (i7j8k9l)
EOF
)"
```

#### リリースノートの書き方

1. **変更をカテゴリ分け**: 新機能、改善、バグ修正、その他
2. **各変更にコミットIDを追記**: `(abc1234)` 形式
3. **コミット一覧の取得方法**:
   ```bash
   # 前回リリースからの変更一覧を取得
   git log v0.x.x..HEAD --oneline
   ```

#### リリースノート例

```markdown
## 変更内容

### 新機能
- init コマンドに --scope オプションを追加 (a1b2c3d)
- ドメイン自動補完機能を追加 (e4f5g6h)

### 改善
- CLI出力を日本語に統一 (i7j8k9l)

### バグ修正
- package.json のキー順序を修正 (m0n1o2p)
```

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
