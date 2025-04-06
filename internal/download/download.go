package download

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/hiroyannnn/ruleforge/internal/config"
	"golang.org/x/oauth2"
)

// Execute はダウンロード処理を実行
func Execute(cfg *config.Config) error {
	if cfg.Verbose {
		log.Printf("ベースリポジトリ: %s からファイルをダウンロードします", cfg.BaseRepo)
	}

	// GitHubクライアントの初期化
	client, owner, repo, err := initGitHubClient(cfg)
	if err != nil {
		return fmt.Errorf("GitHubクライアントの初期化に失敗: %w", err)
	}

	// ファイルをダウンロード
	ctx := context.Background()
	for _, filePath := range cfg.Files {
		if cfg.Verbose {
			log.Printf("ファイル '%s' をダウンロード中...", filePath)
		}

		// 汎用パスでまずダウンロードを試みる
		downloadPath := filePath

		// GitHubからファイルコンテンツを取得
		content, _, _, err := client.Repositories.GetContents(
			ctx,
			owner,
			repo,
			downloadPath,
			&github.RepositoryContentGetOptions{},
		)
		if err != nil && cfg.RepoName != "" {
			// 汎用パスでエラーが発生した場合、リポジトリ固有のパスでリトライ
			repoSpecificPath := filepath.Join(cfg.RepoName, filePath)
			if cfg.Verbose {
				log.Printf("汎用パスでファイルが見つかりません。リポジトリ固有のパス '%s' でリトライします", repoSpecificPath)
			}
			content, _, _, err = client.Repositories.GetContents(
				ctx,
				owner,
				repo,
				repoSpecificPath,
				&github.RepositoryContentGetOptions{},
			)
			if err != nil {
				return fmt.Errorf("ファイル '%s' および '%s' の取得に失敗: %w", filePath, repoSpecificPath, err)
			}
			downloadPath = repoSpecificPath
		} else if err != nil {
			return fmt.Errorf("ファイル '%s' の取得に失敗: %w", downloadPath, err)
		}

		// ファイルコンテンツをデコード
		fileContent, err := content.GetContent()
		if err != nil {
			return fmt.Errorf("ファイル '%s' のコンテンツデコードに失敗: %w", filePath, err)
		}

		// ローカルにファイルを書き込む
		localFilePath := filepath.Join(cfg.LocalDir, filePath)

		// ディレクトリが存在しない場合は作成
		if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
			return fmt.Errorf("ディレクトリ '%s' の作成に失敗: %w", filepath.Dir(localFilePath), err)
		}

		// ファイルを書き込む
		if err := os.WriteFile(localFilePath, []byte(fileContent), 0644); err != nil {
			return fmt.Errorf("ファイル '%s' の書き込みに失敗: %w", localFilePath, err)
		}

		log.Printf("ファイル '%s' をダウンロードしました: %s", downloadPath, localFilePath)
	}

	log.Println("すべてのファイルのダウンロードが完了しました")
	return nil
}

// initGitHubClient はGitHubクライアントを初期化し、所有者とリポジトリ名を抽出
func initGitHubClient(cfg *config.Config) (*github.Client, string, string, error) {
	// リポジトリURLからオーナーとリポジトリ名を抽出
	owner, repo, err := parseRepoURL(cfg.BaseRepo)
	if err != nil {
		return nil, "", "", err
	}

	var client *github.Client

	// GitHubトークンが設定されている場合は認証付きクライアントを作成
	if cfg.GitHubToken != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.GitHubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		// 認証なしクライアント（レート制限に注意）
		client = github.NewClient(nil)
	}

	return client, owner, repo, nil
}

// parseRepoURL はGitHubリポジトリURLから所有者とリポジトリ名を抽出
func parseRepoURL(repoURL string) (string, string, error) {
	// URLからgithub.comドメイン以降の部分を抽出
	repoURL = strings.TrimSuffix(repoURL, "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	var parts []string

	// https://github.com/owner/repo 形式
	if strings.HasPrefix(repoURL, "https://github.com/") {
		parts = strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
	} else if strings.HasPrefix(repoURL, "git@github.com:") {
		// git@github.com:owner/repo.git 形式
		parts = strings.Split(strings.TrimPrefix(repoURL, "git@github.com:"), "/")
	} else {
		// owner/repo 形式（短縮形）
		parts = strings.Split(repoURL, "/")
	}

	if len(parts) < 2 {
		return "", "", fmt.Errorf("無効なリポジトリURL形式: %s", repoURL)
	}

	return parts[0], parts[1], nil
}
