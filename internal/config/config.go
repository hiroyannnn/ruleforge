package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション全体の設定
type Config struct {
	// ベースリポジトリのURL
	BaseRepo string `yaml:"base-repo"`

	// 対象ファイルのリスト
	Files []string `yaml:"target-files"`

	// GitHubトークン（環境変数からの読み込みも可）
	GitHubToken string `yaml:"github-token"`

	// コミットメッセージやPRのタイトル/説明
	Message string `yaml:"message"`

	// 詳細なログ出力
	Verbose bool `yaml:"verbose"`

	// ローカルディレクトリパス（カレントディレクトリがデフォルト）
	LocalDir string `yaml:"local-dir"`

	// 作業用ブランチ名（アップロード用）
	BranchName string `yaml:"branch-name"`

	// カレントリポジトリ名（自動検出される）
	RepoName string `yaml:"repo-name"`
}

// Load は設定ファイルとデフォルト値から設定を読み込む
func Load(configFile string) (*Config, error) {
	// デフォルト設定
	cfg := &Config{
		Files:      []string{".cursor/rules.md"},
		LocalDir:   ".",
		BranchName: fmt.Sprintf("update-agent-rules-%d", os.Getpid()),
	}

	// 設定ファイルが存在する場合は読み込む
	if configFile != "" {
		if _, err := os.Stat(configFile); err == nil {
			f, err := os.Open(configFile)
			if err != nil {
				return nil, fmt.Errorf("設定ファイルを開けません: %w", err)
			}
			defer f.Close()

			decoder := yaml.NewDecoder(f)
			if err := decoder.Decode(cfg); err != nil {
				return nil, fmt.Errorf("設定ファイルの解析に失敗: %w", err)
			}
		}
	}

	// 環境変数から GitHub トークンを設定（設定ファイル内で ${GITHUB_TOKEN} の形式で指定されている場合）
	if strings.HasPrefix(cfg.GitHubToken, "${") && strings.HasSuffix(cfg.GitHubToken, "}") {
		envName := strings.TrimSuffix(strings.TrimPrefix(cfg.GitHubToken, "${"), "}")
		cfg.GitHubToken = os.Getenv(envName)
	}

	// GitHub トークンが設定されていない場合は環境変数から直接取得
	if cfg.GitHubToken == "" {
		cfg.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}

	// カレントリポジトリ名の取得を試みる（git remoteから）
	if cfg.RepoName == "" {
		repoName, err := detectRepoName()
		if err == nil {
			cfg.RepoName = repoName
		}
	}

	return cfg, nil
}

// detectRepoName はカレントディレクトリのGitリポジトリからリポジトリ名を検出
func detectRepoName() (string, error) {
	// .git/config ファイルのパスを作成
	gitDir := filepath.Join(".", ".git")
	gitConfig := filepath.Join(gitDir, "config")

	// .git/config ファイルを読み込み
	data, err := os.ReadFile(gitConfig)
	if err != nil {
		return "", fmt.Errorf("Gitリポジトリが見つかりません: %w", err)
	}

	// URL行を探してリポジトリ名を抽出
	configStr := string(data)
	lines := strings.Split(configStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "url = ") {
			url := strings.TrimPrefix(line, "url = ")

			// github.com/username/repo.git 形式からrepo名を抽出
			if strings.Contains(url, "github.com") {
				parts := strings.Split(url, "/")
				if len(parts) > 1 {
					repoName := parts[len(parts)-1]
					repoName = strings.TrimSuffix(repoName, ".git")
					return repoName, nil
				}
			}
		}
	}

	return "", fmt.Errorf("リモートリポジトリ名を検出できませんでした")
}
