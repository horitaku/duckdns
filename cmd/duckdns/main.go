// DuckDNS 自動更新プログラム
//
// このプログラムは、グローバルIPアドレスを定期的に取得し、
// DuckDNSのDNSレコードを自動的に更新します。
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/horitaku/duckdns/internal/config"
	"github.com/horitaku/duckdns/internal/duckdns"
	"github.com/horitaku/duckdns/internal/ip"
	"github.com/horitaku/duckdns/internal/logger"
	"github.com/horitaku/duckdns/internal/scheduler"
)

// バージョン情報（ビルド時に -ldflags で設定される想定）
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// コマンドライン引数
var (
	configPath  string
	showVersion bool
)

func init() {
	// -config フラグ: 設定ファイルのパスを指定
	flag.StringVar(&configPath, "config", "", "設定ファイルのパス (例: config.yaml)")

	// -version フラグ: バージョン情報を表示
	flag.BoolVar(&showVersion, "version", false, "バージョン情報を表示")

	// ヘルプメッセージのカスタマイズ
	flag.Usage = printUsage
}

// printUsage は、ヘルプメッセージを表示します
func printUsage() {
	fmt.Fprintf(os.Stderr, `DuckDNS 自動更新プログラム

使い方:
  %s [オプション]

オプション:
  -config <path>    設定ファイルのパスを指定 (YAML形式)
                    指定しない場合は環境変数から設定を読み込みます

  -version          バージョン情報を表示して終了

  -h, -help         このヘルプメッセージを表示

環境変数:
  DUCKDNS_DOMAIN    DuckDNS ドメイン名 (必須)
  DUCKDNS_TOKEN     DuckDNS API トークン (必須)
  DUCKDNS_INTERVAL  更新チェック間隔 (例: 5m, 1h) デフォルト: 5m

例:
  # 設定ファイルを使用して起動
  %s -config /etc/duckdns/config.yaml

  # 環境変数を使用して起動
  export DUCKDNS_DOMAIN="your-domain"
  export DUCKDNS_TOKEN="your-token"
  %s

  # バージョン情報を表示
  %s -version

詳細:
  https://github.com/horitaku/duckdns

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

// printVersion は、バージョン情報を表示します
func printVersion() {
	fmt.Printf("DuckDNS 自動更新プログラム\n")
	fmt.Printf("  バージョン: %s\n", version)
	fmt.Printf("  コミット:   %s\n", commit)
	fmt.Printf("  ビルド日時: %s\n", date)
}

// setupSignalHandler は、シグナルハンドリング を設定するますね。
// SIGINT (Ctrl+C) と SIGTERM を受け取って、渡された cancel 関数を呼び出すます。
// グレースフルシャットダウンを実現するますよー。
func setupSignalHandler(cancel context.CancelFunc) {
	// シグナルチャネルを作成するます
	sigChan := make(chan os.Signal, 1)

	// SIGINT (Ctrl+C) と SIGTERM をハンドリング対象に登録するますね
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// goroutineでシグナルを待機するます
	go func() {
		sig := <-sigChan
		slog.Info("シグナルを受け取ったます",
			"signal", sig.String(),
		)

		// context をキャンセルして、すべてのゴルーチンを停止するますよー
		slog.Info("グレースフルシャットダウンを開始するます")
		cancel()
	}()
}

func main() {
	// コマンドライン引数を解析
	flag.Parse()

	// -version フラグが指定された場合: バージョン情報を表示して終了
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	// ========== タスク6.2: ログの初期化 ==========
	// デフォルトのログレベルはinfoにするます
	logLevel := "info"
	logFormat := "text"

	// 環境変数からログレベルを取得するますよー
	if level := os.Getenv("DUCKDNS_LOG_LEVEL"); level != "" {
		logLevel = level
	}

	// 環境変数からログフォーマットを取得するますね
	if format := os.Getenv("DUCKDNS_LOG_FORMAT"); format != "" {
		logFormat = format
	}

	// ログシステムの初期化
	if err := logger.InitLogger(logLevel, logFormat); err != nil {
		fmt.Fprintf(os.Stderr, "ログ初期化に失敗したます: %v\n", err)
		os.Exit(1)
	}

	// ログを使って起動メッセージを出力するますよ
	slog.Info("DuckDNS自動更新プログラムを起動するます",
		"version", version,
		"commit", commit,
		"log_level", logLevel,
		"log_format", logFormat,
		"config_path", configPath,
	)

	// ========== タスク6.3: シグナルハンドリング ==========
	// context.WithCancel を使って、シャットダウン可能なコンテキストを作成するます
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリングを設定するますね
	// SIGINT (Ctrl+C) と SIGTERM を受け取ると、cancel() が呼ばれるます
	setupSignalHandler(cancel)

	slog.Info("シグナルハンドラーが設定されたます")

	// ========== タスク6.4: 各コンポーネントの統合 ==========
	// ここから各コンポーネントを統合するますね！ わくわく! 🎉

	// ===== 設定の読み込み =====
	slog.Info("設定を読み込むます")
	cfg, err := loadConfiguration()
	if err != nil {
		slog.Error("設定の読み込みに失敗したます",
			"error", err,
			"config_path", configPath,
		)
		os.Exit(1)
	}

	slog.Info("設定を読み込みました",
		"domain", cfg.DuckDNS.Domain,
		"interval", cfg.Update.Interval.String(),
		"ip_sources", len(cfg.IPSources),
	)

	// ===== IP Fetcher の初期化 =====
	slog.Info("IP Fetcher を初期化するます")
	fetcher := ip.NewMultipleFetcher(cfg.IPSources)
	slog.Info("IP Fetcher が初期化されたます",
		"sources_count", len(cfg.IPSources),
	)

	// ===== DuckDNS Client の初期化 =====
	slog.Info("DuckDNS クライアントを初期化するます")
	duckDNSClient := duckdns.NewClient()
	slog.Info("DuckDNS クライアントが初期化されたます")

	// ===== Scheduler の初期化と実行 =====
	slog.Info("スケジューラーを初期化するます")
	sch := scheduler.NewScheduler(
		cfg.Update.Interval,
		fetcher,
		duckDNSClient,
		cfg.DuckDNS.Domain,
		cfg.DuckDNS.Token,
	)
	slog.Info("スケジューラーが初期化されたます",
		"interval", cfg.Update.Interval.String(),
	)

	// スケジューラーを実行するます
	// context がキャンセルされるまで実行し続けるますね
	slog.Info("スケジューラーを起動するます")
	sch.Run(ctx)

	// ctx がキャンセルされたら、ここに制御が戻ります
	slog.Info("スケジューラーが停止したます")

	// プログラム終了時のメッセージ
	slog.Info("DuckDNS自動更新プログラムを終了するます")
}

// loadConfiguration は、設定ファイルまたは環境変数から設定を読み込むます。
// 優先度: 環境変数 > 設定ファイル
func loadConfiguration() (*config.Config, error) {
	// Load関数で統一的に設定を読み込む
	// configPath が空文字列の場合は環境変数のみから読み込む
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定の読み込みに失敗: %w", err)
	}

	// バリデーション
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("設定の検証に失敗: %w", err)
	}

	return cfg, nil
}
