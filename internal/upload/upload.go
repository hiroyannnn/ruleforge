package upload

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

// Execute はアップロード処理を実行
func Execute(cfg *config.Config) error {
	if cfg.GitHubToken == "" {
		return fmt.Errorf("GitHub APIトークンが設定されていません。環境変数 GITHUB_TOKEN を設定するか、設定ファイルで指定してください")
	}

	if cfg.Message == "" {
		return fmt.Errorf("コミットメッセージが指定されていません。--message フラグまたは設定ファイルで指定してください")
	}

	if cfg.Verbose {
		log.Printf("ファイルをベースリポジトリ %s にアップロードします", cfg.BaseRepo)
	}

	// GitHubクライアントの初期化
	client, owner, repo, err := initGitHubClient(cfg)
	if err != nil {
		return fmt.Errorf("GitHubクライアントの初期化に失敗: %w", err)
	}

	ctx := context.Background()

	// ベースブランチ（通常は main または master）を取得
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("リポジトリ情報の取得に失敗: %w", err)
	}
	baseBranch := repository.GetDefaultBranch()

	// ベースブランチのリファレンスを取得
	baseRef, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+baseBranch)
	if err != nil {
		return fmt.Errorf("ベースブランチのリファレンス取得に失敗: %w", err)
	}

	// 新しいブランチを作成
	branchName := cfg.BranchName
	if cfg.RepoName != "" {
		// リポジトリ名をプレフィックスとして追加
		branchName = fmt.Sprintf("%s-%s", cfg.RepoName, branchName)
	}

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: baseRef.Object,
	}

	_, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		if !strings.Contains(err.Error(), "Reference already exists") {
			return fmt.Errorf("ブランチの作成に失敗: %w", err)
		}
		log.Printf("ブランチ '%s' は既に存在します。既存のブランチに追加します", branchName)
	} else {
		log.Printf("ブランチ '%s' を作成しました", branchName)
	}

	// 各ファイルをアップロード
	for _, filePath := range cfg.Files {
		// ローカルファイルパス
		localFilePath := filepath.Join(cfg.LocalDir, filePath)

		// ファイルが存在するか確認
		if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
			log.Printf("警告: ファイル '%s' が見つかりません。スキップします", localFilePath)
			continue
		}

		// ファイルコンテンツを読み込む
		content, err := os.ReadFile(localFilePath)
		if err != nil {
			return fmt.Errorf("ファイル '%s' の読み込みに失敗: %w", localFilePath, err)
		}

		// 既存ファイルの情報を取得（SHA取得のため）
		var existingSHA string
		fileContent, _, _, err := client.Repositories.GetContents(
			ctx,
			owner,
			repo,
			filePath,
			&github.RepositoryContentGetOptions{Ref: branchName},
		)

		if err == nil && fileContent != nil {
			existingSHA = fileContent.GetSHA()
		}

		// ファイルのアップロード先パス
		targetPath := filePath
		if cfg.RepoName != "" {
			// リポジトリ固有のディレクトリにファイルを配置
			dir := filepath.Dir(filePath)
			base := filepath.Base(filePath)
			targetPath = filepath.Join(dir, cfg.RepoName, base)
		}

		// ファイルをアップロード（更新または作成）
		opts := &github.RepositoryContentFileOptions{
			Message: github.String(cfg.Message),
			Content: content,
			Branch:  github.String(branchName),
		}

		if existingSHA != "" {
			opts.SHA = github.String(existingSHA)
		}

		if cfg.Verbose {
			log.Printf("ファイル '%s' をパス '%s' にアップロード中...", localFilePath, targetPath)
		}

		_, _, err = client.Repositories.CreateFile(ctx, owner, repo, targetPath, opts)
		if err != nil {
			return fmt.Errorf("ファイル '%s' のアップロードに失敗: %w", targetPath, err)
		}

		log.Printf("ファイル '%s' をアップロードしました: %s", localFilePath, targetPath)
	}

	// プルリクエストを作成
	title := cfg.Message
	if cfg.RepoName != "" {
		title = fmt.Sprintf("[%s] %s", cfg.RepoName, title)
	}

	body := fmt.Sprintf("このPRは %s から自動生成されました。\n\nAIエージェントルールの更新を含みます。", cfg.RepoName)

	pr := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branchName),
		Base:                github.String(baseBranch),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pullRequest, _, err := client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		// PR作成エラーチェック - 既に同じブランチでPRが存在する可能性がある
		if strings.Contains(err.Error(), "pull request already exists") {
			log.Printf("警告: このブランチからのPRは既に存在します")

			// 既存PRを探す
			prs, _, listErr := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
				Head:  branchName,
				Base:  baseBranch,
				State: "open",
			})

			if listErr == nil && len(prs) > 0 {
				log.Printf("既存のPR #%d にコンテンツが追加されました: %s", prs[0].GetNumber(), prs[0].GetHTMLURL())
				return nil
			}

			return fmt.Errorf("PR作成に失敗し、既存のPRも特定できません: %w", err)
		}
		return fmt.Errorf("プルリクエストの作成に失敗: %w", err)
	}

	log.Printf("プルリクエスト #%d を作成しました: %s", pullRequest.GetNumber(), pullRequest.GetHTMLURL())
	return nil
}

// initGitHubClient はGitHubクライアントを初期化し、所有者とリポジトリ名を抽出
func initGitHubClient(cfg *config.Config) (*github.Client, string, string, error) {
	// リポジトリURLからオーナーとリポジトリ名を抽出
	owner, repo, err := parseRepoURL(cfg.BaseRepo)
	if err != nil {
		return nil, "", "", err
	}

	// GitHub API認証
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

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
