package version

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

// CurrentVersion は現在のバージョン情報
var (
	CurrentVersion = "dev"
	CurrentCommit  = "none"
	CurrentDate    = "unknown"
)

// チェック結果のキャッシュと有効期限
var (
	lastCheck     time.Time
	cachedResult  *github.RepositoryRelease
	checkInterval = 24 * time.Hour // 1日に1回だけチェック
)

// CheckForUpdates は最新バージョンをチェックし、更新がある場合は通知メッセージを返す
func CheckForUpdates() (string, error) {
	// バージョンが明示的に設定されていない場合（開発中）はチェックしない
	if CurrentVersion == "dev" {
		return "", nil
	}

	// キャッシュが有効なら再チェックしない
	if !lastCheck.IsZero() && time.Since(lastCheck) < checkInterval {
		// キャッシュされた結果がない場合は通知しない
		if cachedResult == nil {
			return "", nil
		}
		return generateUpdateMessage(cachedResult), nil
	}

	// GitHubクライアントを初期化
	client := github.NewClient(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 最新リリースを取得
	release, _, err := client.Repositories.GetLatestRelease(ctx, "hiroyannnn", "ruleforge")
	if err != nil {
		// エラーが発生した場合は静かに失敗（通知なし）
		return "", fmt.Errorf("最新バージョンの確認に失敗: %w", err)
	}

	// 結果とチェック時間をキャッシュ
	lastCheck = time.Now()
	cachedResult = release

	// 最新バージョンと現在のバージョンを比較
	if isNewer(release.GetTagName(), CurrentVersion) {
		return generateUpdateMessage(release), nil
	}

	return "", nil
}

// isNewer は最新バージョンが現在のバージョンより新しいかどうかを判定
func isNewer(latestTag, currentVersion string) bool {
	// バージョン文字列から 'v' プレフィックスを削除
	latestVersion := strings.TrimPrefix(latestTag, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// 単純な文字列比較（セマンティックバージョニングに準拠していると仮定）
	return latestVersion > currentVersion
}

// generateUpdateMessage は更新通知メッセージを生成
func generateUpdateMessage(release *github.RepositoryRelease) string {
	return fmt.Sprintf(`
⚠️ 新しいバージョン %s が利用可能です（現在: %s）
リリースノート: %s
以下のコマンドでアップデートできます:

go install github.com/hiroyannnn/ruleforge/cmd/ruleforge@latest

詳細はこちら: %s`,
		release.GetTagName(),
		CurrentVersion,
		release.GetHTMLURL(),
		release.GetHTMLURL(),
	)
}
