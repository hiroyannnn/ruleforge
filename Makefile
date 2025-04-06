.PHONY: build test clean release release-dry-run lint vet format help

# バージョン情報
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ビルド設定
BINARY_NAME = ruleforge
LDFLAGS = -s -w -X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE)
GOFLAGS = -ldflags "$(LDFLAGS)"

# デフォルトのターゲット
.DEFAULT_GOAL := help

# ヘルプコマンド
help:
	@echo "使用可能なコマンド:"
	@echo "  make build        - 実行ファイルをビルド"
	@echo "  make test         - テストを実行"
	@echo "  make lint         - コードの静的解析を実行"
	@echo "  make vet          - go vet を実行"
	@echo "  make format       - コードをフォーマット"
	@echo "  make clean        - 生成したファイルを削除"
	@echo "  make release      - 新しいリリースをビルドして公開"
	@echo "  make release-dry-run - リリースの動作確認（実際には公開しない）"
	@echo ""
	@echo "環境変数:"
	@echo "  VERSION - リリースバージョン (デフォルト: 最新タグまたは'dev')"
	@echo "  COMMIT  - コミットハッシュ (デフォルト: 現在のコミット)"
	@echo "  DATE    - ビルド日時 (デフォルト: 現在時刻)"

# ビルドコマンド
build:
	@echo "🔨 $(BINARY_NAME) をビルドしています..."
	go build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/ruleforge

# テストコマンド
test:
	@echo "🧪 テストを実行しています..."
	go test -v ./...

# 静的解析コマンド
lint:
	@echo "🔍 静的解析を実行しています..."
	golangci-lint run

# go vet コマンド
vet:
	@echo "🔍 go vet を実行しています..."
	go vet ./...

# フォーマットコマンド
format:
	@echo "✨ コードをフォーマットしています..."
	go fmt ./...

# クリーンコマンド
clean:
	@echo "🧹 生成ファイルを削除しています..."
	rm -f $(BINARY_NAME)
	go clean

# リリースのドライラン
release-dry-run:
	@echo "🚀 リリースのドライランを実行しています..."
	goreleaser release --snapshot --clean --skip=publish

# 新しいタグの作成とリリース
release:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "❌ バージョンが指定されていません。VERSION環境変数を設定してください。"; \
		echo "例: make release VERSION=v1.2.3"; \
		exit 1; \
	fi
	@echo "🔖 バージョン $(VERSION) のタグを作成しています..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "🚀 タグが正常にプッシュされました。GitHub Actionsがリリースを自動的に作成します。"
	@echo "   リリースステータスを確認: https://github.com/hiroyannnn/ruleforge/actions"

# インストールコマンド
install:
	@echo "📦 $(BINARY_NAME) をインストールしています..."
	go install $(GOFLAGS) ./cmd/ruleforge
