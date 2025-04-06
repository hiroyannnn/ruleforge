package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
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
`
	err = os.WriteFile(testConfigPath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("テスト設定ファイルの作成に失敗: %v", err)
	}

	// テスト用の環境変数を設定
	os.Setenv("TEST_TOKEN", "test-token-value")
	defer os.Unsetenv("TEST_TOKEN")

	// グローバル変数を一時的に設定
	origConfigFile := configFile
	origBaseRepo := baseRepo
	origFiles := files
	origMessage := message
	origVerbose := verbose

	// テスト後に元に戻す
	defer func() {
		configFile = origConfigFile
		baseRepo = origBaseRepo
		files = origFiles
		message = origMessage
		verbose = origVerbose
	}()

	// テスト用の値を設定
	configFile = testConfigPath
	baseRepo = ""
	files = []string{".cursor/rules.md"}
	message = ""
	verbose = false

	// テスト実行
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 値の検証
	if cfg.BaseRepo != "https://github.com/test/repo" {
		t.Errorf("BaseRepo: 期待値 %s, 実際の値 %s", "https://github.com/test/repo", cfg.BaseRepo)
	}

	if len(cfg.Files) != 2 || cfg.Files[0] != ".cursor/rules.md" || cfg.Files[1] != ".cursor/config.json" {
		t.Errorf("Files: 期待値 [.cursor/rules.md .cursor/config.json], 実際の値 %v", cfg.Files)
	}

	if cfg.GitHubToken != "test-token-value" {
		t.Errorf("GitHubToken: 期待値 %s, 実際の値 %s", "test-token-value", cfg.GitHubToken)
	}

	if cfg.Message != "Test message" {
		t.Errorf("Message: 期待値 %s, 実際の値 %s", "Test message", cfg.Message)
	}
}

func TestLoadConfigWithOverrides(t *testing.T) {
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
github-token: ${TEST_TOKEN}
message: "File message"
`
	err = os.WriteFile(testConfigPath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("テスト設定ファイルの作成に失敗: %v", err)
	}

	// テスト用の環境変数を設定
	os.Setenv("TEST_TOKEN", "test-token-value")
	defer os.Unsetenv("TEST_TOKEN")

	// グローバル変数を一時的に設定
	origConfigFile := configFile
	origBaseRepo := baseRepo
	origFiles := files
	origMessage := message
	origVerbose := verbose

	// テスト後に元に戻す
	defer func() {
		configFile = origConfigFile
		baseRepo = origBaseRepo
		files = origFiles
		message = origMessage
		verbose = origVerbose
	}()

	// テスト用の値を設定 (コマンドライン引数でオーバーライド)
	configFile = testConfigPath
	baseRepo = "https://github.com/override/repo"
	files = []string{".cursor/rules.md", ".cursor/settings.json"}
	message = "CLI message"
	verbose = true

	// テスト実行
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 値の検証 (コマンドラインのオーバーライドが優先されることを確認)
	if cfg.BaseRepo != "https://github.com/override/repo" {
		t.Errorf("BaseRepo: 期待値 %s, 実際の値 %s", "https://github.com/override/repo", cfg.BaseRepo)
	}

	if len(cfg.Files) != 2 || cfg.Files[0] != ".cursor/rules.md" || cfg.Files[1] != ".cursor/settings.json" {
		t.Errorf("Files: 期待値 [.cursor/rules.md .cursor/settings.json], 実際の値 %v", cfg.Files)
	}

	if cfg.Message != "CLI message" {
		t.Errorf("Message: 期待値 %s, 実際の値 %s", "CLI message", cfg.Message)
	}

	if !cfg.Verbose {
		t.Errorf("Verbose: 期待値 %v, 実際の値 %v", true, cfg.Verbose)
	}
}

func TestLoadConfigError(t *testing.T) {
	// グローバル変数を一時的に設定
	origConfigFile := configFile
	origBaseRepo := baseRepo

	// テスト後に元に戻す
	defer func() {
		configFile = origConfigFile
		baseRepo = origBaseRepo
	}()

	// 設定ファイルなし、ベースリポジトリ指定なしの場合エラーになることを確認
	configFile = "nonexistent-file.yaml"
	baseRepo = ""

	// テスト実行
	_, err := loadConfig()
	if err == nil {
		t.Errorf("エラーが期待されましたが、成功してしまいました")
	}
}
