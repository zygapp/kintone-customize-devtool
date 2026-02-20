.PHONY: build clean install test build-all npm-prepare npm-package npm-publish npm-publish-token npm-dry-run version-sync

# .env から環境変数を読み込み
-include .env
export

# ========================================
# バージョン（ここを変更するだけで全体に反映）
# ========================================
VERSION := 0.4.1

BINARY_NAME := kcdev
BUILD_DIR := build
NPM_DIR := npm/@zygapp/kintone-customize-devtool
LDFLAGS := -s -w -X github.com/kintone/kcdev/internal/cmd.version=$(VERSION)

build:
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kcdev

install:
	go install -ldflags="$(LDFLAGS)" ./cmd/kcdev

clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist
	rm -rf $(NPM_DIR)/bin/darwin-*
	rm -rf $(NPM_DIR)/bin/linux-*
	rm -rf $(NPM_DIR)/bin/win32-*
	rm -f $(NPM_DIR)/bin/.binary-path

test:
	go test ./...

# package.json のバージョンを同期
version-sync:
	@echo "Syncing version to $(VERSION)..."
	@sed -i 's/"version": "[^"]*"/"version": "$(VERSION)"/' $(NPM_DIR)/package.json
	@echo "Done!"

# 全プラットフォーム向けビルド
build-all: clean version-sync
	@mkdir -p $(BUILD_DIR)
	@echo "Building darwin-x64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-x64 ./cmd/kcdev
	@echo "Building darwin-arm64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/kcdev
	@echo "Building linux-x64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-x64 ./cmd/kcdev
	@echo "Building linux-arm64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/kcdev
	@echo "Building win32-x64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-win32-x64.exe ./cmd/kcdev
	@echo "Building win32-arm64..."
	@GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-win32-arm64.exe ./cmd/kcdev
	@echo "Build complete!"

# npmパッケージにバイナリをコピー
npm-prepare: build-all
	@echo "Copying binaries to npm package..."
	@mkdir -p $(NPM_DIR)/bin/darwin-x64
	@mkdir -p $(NPM_DIR)/bin/darwin-arm64
	@mkdir -p $(NPM_DIR)/bin/linux-x64
	@mkdir -p $(NPM_DIR)/bin/linux-arm64
	@mkdir -p $(NPM_DIR)/bin/win32-x64
	@mkdir -p $(NPM_DIR)/bin/win32-arm64
	@cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-x64 $(NPM_DIR)/bin/darwin-x64/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(NPM_DIR)/bin/darwin-arm64/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-linux-x64 $(NPM_DIR)/bin/linux-x64/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(NPM_DIR)/bin/linux-arm64/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-win32-x64.exe $(NPM_DIR)/bin/win32-x64/$(BINARY_NAME).exe
	@cp $(BUILD_DIR)/$(BINARY_NAME)-win32-arm64.exe $(NPM_DIR)/bin/win32-arm64/$(BINARY_NAME).exe
	@chmod +x $(NPM_DIR)/bin/*/$(BINARY_NAME) 2>/dev/null || true
	@echo "Done!"

# npm公開用パッケージ作成
npm-package: npm-prepare
	@echo "npm package ready in $(NPM_DIR)/"
	@du -sh $(NPM_DIR)/bin/*/

# npm公開テスト (dry-run)
npm-dry-run: npm-package
	cd $(NPM_DIR) && npm publish --access public --dry-run

# npm公開 (TOTP認証アプリの6桁コード: make npm-publish OTP=123456)
npm-publish: npm-package
ifndef OTP
	$(error OTP is required. Usage: make npm-publish OTP=<TOTP認証アプリの6桁コード>)
endif
	cd $(NPM_DIR) && npm publish --access public --otp=$(OTP)
	@echo "Published successfully!"

# npm公開 (トークン使用: make npm-publish-token)
# 環境変数 NPM_TOKEN または引数 TOKEN=npm_xxx を使用
TOKEN ?= $(NPM_TOKEN)
npm-publish-token: npm-package
ifeq ($(TOKEN),)
	$(error NPM_TOKEN or TOKEN is required. Set NPM_TOKEN env var or use: make npm-publish-token TOKEN=npm_xxx)
endif
	cd $(NPM_DIR) && npm publish --access public --registry=https://registry.npmjs.org/ --//registry.npmjs.org/:_authToken=$(TOKEN)
	@echo "Published successfully!"
