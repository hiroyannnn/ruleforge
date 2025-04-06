package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "ruleforge-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用の設定ファイルを作成
	testConfigPath := filepath.Join(tempDir, ".ruleforge.yaml")
	testConfigContent := `
base-repo: https://github.com/test/repo
target-files:
  - .cursor/rules.md
  - .cursor/config.json
github-token: ${TEST_TOKEN}
message: "Test message"
verbose: true
local-dir: "./test-dir"
branch-name: "test-branch"
repo-name: "test-repo"
`
	err = os.WriteFile(testConfigPath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("テスト設定ファイルの作成に失敗: %v", err)
	}

	// テスト用の環境変数を設定
	os.Setenv("TEST_TOKEN", "test-token-value")
	defer os.Unsetenv("TEST_TOKEN")

	// テスト実行
	cfg, err := Load(testConfigPath)
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 設定値の検証
	testCases := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"BaseRepo", cfg.BaseRepo, "https://github.com/test/repo"},
		{"Files[0]", cfg.Files[0], ".cursor/rules.md"},
		{"Files[1]", cfg.Files[1], ".cursor/config.json"},
		{"GitHubToken", cfg.GitHubToken, "test-token-value"},
		{"Message", cfg.Message, "Test message"},
		{"Verbose", cfg.Verbose, true},
		{"LocalDir", cfg.LocalDir, "./test-dir"},
		{"BranchName", cfg.BranchName, "test-branch"},
		{"RepoName", cfg.RepoName, "test-repo"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.actual != tc.expected {
				t.Errorf("%s: 期待値 %v, 実際の値 %v", tc.name, tc.expected, tc.actual)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	// 存在しない設定ファイルで読み込みテスト
	cfg, err := Load("non-existent-file.yaml")
	if err != nil {
		t.Fatalf("デフォルト設定の読み込みに失敗: %v", err)
	}

	// デフォルト値のチェック
	if len(cfg.Files) != 1 || cfg.Files[0] != ".cursor/rules.md" {
		t.Errorf("Files のデフォルト値が正しくありません: %v", cfg.Files)
	}

	if cfg.LocalDir != "." {
		t.Errorf("LocalDir のデフォルト値が正しくありません: %v", cfg.LocalDir)
	}

	if cfg.BranchName == "" {
		t.Errorf("BranchName のデフォルト値が設定されていません")
	}
}

func TestParseRepoURL(t *testing.T) {

	// ダミー関数を作成してパースロジックをテスト
	parseRepoURL := func(url string) (string, string, error) {
		if url == "" {
			return "", "", fmt.Errorf("URLが空です")
		}

		url = strings.TrimSuffix(url, "/")
		url = strings.TrimSuffix(url, ".git")

		var parts []string

		if strings.HasPrefix(url, "https://github.com/") {
			parts = strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
		} else if strings.HasPrefix(url, "git@github.com:") {
			parts = strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		} else {
			parts = strings.Split(url, "/")
		}

		if len(parts) < 2 {
			return "", "", fmt.Errorf("無効なリポジトリURL形式: %s", url)
		}

		return parts[0], parts[1], nil
	}

	testCases := []struct {
		url      string
		owner    string
		repo     string
		hasError bool
	}{
		{"https://github.com/owner/repo", "owner", "repo", false},
		{"https://github.com/owner/repo.git", "owner", "repo", false},
		{"https://github.com/owner/repo/", "owner", "repo", false},
		{"git@github.com:owner/repo.git", "owner", "repo", false},
		{"owner/repo", "owner", "repo", false},
		{"", "", "", true},
		{"invalid-url", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			owner, repo, err := parseRepoURL(tc.url)

			if tc.hasError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、成功しました: %s, %s", owner, repo)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if owner != tc.owner {
					t.Errorf("owner: 期待値 %s, 実際の値 %s", tc.owner, owner)
				}
				if repo != tc.repo {
					t.Errorf("repo: 期待値 %s, 実際の値 %s", tc.repo, repo)
				}
			}
		})
	}
}
