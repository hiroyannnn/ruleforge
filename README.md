# RuleForge

AIエージェントのルールを管理するためのCLIツール。ベースリポジトリとローカルリポジトリ間でAIエージェントのルールファイル（`.cursor/rules.md`など）を同期します。

## 概要

このツールは以下の機能を提供します：

1. **ダウンロード**: ベースリポジトリからCursor Rulesファイルをカレントディレクトリにコピー
2. **アップロード**: カレントディレクトリのCursor RulesファイルをベースリポジトリにPRとして送信

## インストール

```bash
go install github.com/yourusername/ruleforge@latest
```

または、リポジトリをクローンして手動でビルド：

```bash
git clone https://github.com/yourusername/ruleforge.git
cd ruleforge
go build
```

## 使い方

### 基本コマンド

```bash
# ヘルプを表示
ruleforge --help

# ベースリポジトリからルールをダウンロード
ruleforge download --base-repo https://github.com/organization/base-rules-repo

# カレントディレクトリのルールをベースリポジトリにアップロードしてPRを作成
ruleforge upload --base-repo https://github.com/organization/base-rules-repo --message "Update rules for my-project"
```

### 設定ファイル

`.ruleforge.yaml`という設定ファイルを作成することで、コマンドライン引数を省略できます：

```yaml
base-repo: https://github.com/organization/base-rules-repo
target-files:
  - .cursor/rules.md
  - .cursor/config.json
github-token: ${GITHUB_TOKEN} # 環境変数から読み込み
```

## アーキテクチャ

```
cmd/             # エントリーポイントとCLIコマンド定義
  ruleforge/
    main.go
internal/        # 内部パッケージ
  config/        # 設定ファイル関連
  download/      # ダウンロード機能
  upload/        # アップロード機能
  github/        # GitHub API操作
  file/          # ファイル操作ユーティリティ
  logger/        # ロギング
pkg/             # 公開APIパッケージ（必要な場合）
```

## 開発

### 必要条件

- Go 1.20以上
- GitHub Personal Access Token（アップロード機能で使用）

### テスト

```bash
go test ./...
```

### ビルド

```bash
go build -o ruleforge ./cmd/ruleforge
```

## ライセンス

MIT

## 貢献

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチをプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成
