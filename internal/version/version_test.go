package version

import (
	"strings"
	"testing"

	"github.com/google/go-github/v60/github"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name          string
		latest        string
		current       string
		expectIsNewer bool
	}{
		{
			name:          "標準的なセマンティックバージョン",
			latest:        "v1.2.0",
			current:       "v1.1.0",
			expectIsNewer: true,
		},
		{
			name:          "パッチバージョンの増加",
			latest:        "v1.1.1",
			current:       "v1.1.0",
			expectIsNewer: true,
		},
		{
			name:          "メジャーバージョンの増加",
			latest:        "v2.0.0",
			current:       "v1.9.9",
			expectIsNewer: true,
		},
		{
			name:          "同じバージョン",
			latest:        "v1.1.0",
			current:       "v1.1.0",
			expectIsNewer: false,
		},
		{
			name:          "古いバージョン",
			latest:        "v1.0.0",
			current:       "v1.1.0",
			expectIsNewer: false,
		},
		{
			name:          "先頭のvが片方にだけある場合",
			latest:        "v1.2.0",
			current:       "1.1.0",
			expectIsNewer: true,
		},
		{
			name:          "両方にvがない場合",
			latest:        "1.2.0",
			current:       "1.1.0",
			expectIsNewer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNewer(tt.latest, tt.current)
			if result != tt.expectIsNewer {
				t.Errorf("isNewer(%s, %s) = %v, want %v",
					tt.latest, tt.current, result, tt.expectIsNewer)
			}
		})
	}
}

func TestGenerateUpdateMessage(t *testing.T) {
	// 固定値でテスト
	tagName := "v1.2.0"
	url := "https://github.com/hiroyannnn/ruleforge/releases/tag/v1.2.0"

	// CurrentVersionを一時的に保存し、テスト後に戻す
	originalVersion := CurrentVersion
	CurrentVersion = "v1.1.0"
	defer func() {
		CurrentVersion = originalVersion
	}()

	release := &github.RepositoryRelease{
		TagName: &tagName,
		HTMLURL: &url,
	}

	message := generateUpdateMessage(release)

	// メッセージに必要な情報が含まれているか確認
	if !contains(message, "v1.2.0") || !contains(message, "v1.1.0") || !contains(message, url) {
		t.Errorf("生成されたメッセージに必要な情報が含まれていません: %s", message)
	}
}

// contains はsの中にsubstringが含まれているかどうかを確認
func contains(s, substring string) bool {
	return strings.Contains(s, substring)
}
