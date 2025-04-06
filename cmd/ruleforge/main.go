package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hiroyannnn/ruleforge/internal/config"
	"github.com/hiroyannnn/ruleforge/internal/download"
	"github.com/hiroyannnn/ruleforge/internal/upload"
	"github.com/hiroyannnn/ruleforge/internal/version"
	"github.com/spf13/cobra"
)

var (
	// バージョン情報（ビルド時に設定）
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var (
	configFile string
	baseRepo   string
	files      []string
	message    string
	verbose    bool
)

func init() {
	// バージョン情報をバージョンパッケージに設定
	version.CurrentVersion = buildVersion
	version.CurrentCommit = buildCommit
	version.CurrentDate = buildDate
}

func main() {
	// ルートコマンド
	rootCmd := &cobra.Command{
		Use:     "ruleforge",
		Short:   "AIエージェントのルール管理ツール",
		Version: fmt.Sprintf("%s (commit: %s, built at: %s)", version.CurrentVersion, version.CurrentCommit, version.CurrentDate),
	}

	// フラグ定義
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".ruleforge.yaml", "設定ファイルのパス")
	rootCmd.PersistentFlags().StringVarP(&baseRepo, "base-repo", "b", "", "ベースリポジトリのURL")
	rootCmd.PersistentFlags().StringSliceVarP(&files, "files", "f", []string{".cursor/rules.md"}, "対象ファイルのリスト")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "詳細なログ出力")

	// downloadコマンド
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "ベースリポジトリからエージェントルールをダウンロード",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			return download.Execute(cfg)
		},
	}

	// uploadコマンド
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "カレントリポジトリのエージェントルールをベースリポジトリにPRとして送信",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			return upload.Execute(cfg)
		},
	}
	uploadCmd.Flags().StringVarP(&message, "message", "m", "", "PRのメッセージ")
	if err := uploadCmd.MarkFlagRequired("message"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}

	// コマンド追加
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(uploadCmd)

	// バージョンチェックを実行
	go func() {
		updateMsg, err := version.CheckForUpdates()
		if err != nil {
			// エラーは無視（ログに残さない）
			return
		}
		if updateMsg != "" {
			// 更新通知を表示
			fmt.Fprintln(os.Stderr, updateMsg)
		}
	}()

	// コマンド実行
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}

// 設定を読み込む
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, fmt.Errorf("設定の読み込みに失敗: %w", err)
	}

	// コマンドライン引数で上書き
	if baseRepo != "" {
		cfg.BaseRepo = baseRepo
	}

	if len(files) > 0 && !(len(files) == 1 && files[0] == ".cursor/rules.md") {
		cfg.Files = files
	}

	if message != "" {
		cfg.Message = message
	}

	cfg.Verbose = verbose

	// 必須項目の検証
	if cfg.BaseRepo == "" {
		return nil, fmt.Errorf("ベースリポジトリURLが指定されていません。--base-repo フラグまたは設定ファイルで指定してください")
	}

	return cfg, nil
}
