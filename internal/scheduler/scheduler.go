// Package scheduler は、DuckDNSのDNSレコードを定期的に更新するスケジューラーを提供します。
// グローバルIPアドレスの変更を監視し、変更があった場合にDuckDNSを自動更新します。
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/horitaku/duckdns/internal/duckdns"
	"github.com/horitaku/duckdns/internal/ip"
)

// Scheduler は、定期的にIPアドレスをチェックし、DuckDNSを更新する構造体です。
// IP変更を検知した場合のみ更新を実行することで、不要なAPI呼び出しを削減します。
type Scheduler struct {
	// interval は更新チェックの実行間隔です
	interval time.Duration

	// ipFetcher はグローバルIPアドレスを取得するためのインターフェースです
	ipFetcher ip.Fetcher

	// duckDNSClient はDuckDNS APIへの更新リクエストを行うクライアントです
	duckDNSClient *duckdns.Client

	// domain はDuckDNSに登録されているドメイン名です
	domain string

	// token はDuckDNS APIのアクセストークンです
	token string

	// lastIP は前回取得したIPアドレスを保持します（変更検知に使用）
	lastIP string
}

// NewScheduler は、指定された設定で新しいSchedulerを作成します。
//
// Parameters:
//   - interval: 更新チェックの実行間隔
//   - ipFetcher: グローバルIPアドレスを取得するFetcherインターフェース
//   - duckDNSClient: DuckDNS APIクライアント
//   - domain: DuckDNSドメイン名
//   - token: DuckDNS APIトークン
//
// Returns:
//   - *Scheduler: 初期化されたSchedulerインスタンス
func NewScheduler(
	interval time.Duration,
	ipFetcher ip.Fetcher,
	duckDNSClient *duckdns.Client,
	domain string,
	token string,
) *Scheduler {
	slog.Info("Scheduler を初期化します",
		"interval", interval,
		"domain", domain,
	)

	return &Scheduler{
		interval:      interval,
		ipFetcher:     ipFetcher,
		duckDNSClient: duckDNSClient,
		domain:        domain,
		token:         token,
		lastIP:        "", // 初回は必ず更新を実行
	}
}

// Run は、スケジューラーを起動して定期的にIPアドレスをチェックし、
// 必要に応じてDuckDNSを更新します。
// context がキャンセルされるまで実行を継続します。
//
// Parameters:
//   - ctx: 実行を制御するコンテキスト（キャンセルで停止）
//
// この関数はブロッキングします。バックグラウンドで実行する場合は
// goroutine で起動してください。
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	scheduler.Run(ctx)
func (s *Scheduler) Run(ctx context.Context) {
	slog.Info("スケジューラーを開始します",
		"interval", s.interval,
		"domain", s.domain,
	)

	// 初回実行: 起動直後に一度チェックを実行
	s.checkAndUpdate(ctx)

	// Ticker を作成して定期実行を設定
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop() // 終了時にTickerを停止してリソースを解放

	// select 文で定期実行とコンテキストキャンセルを監視
	for {
		select {
		case <-ticker.C:
			// Ticker が発火: 定期チェックを実行
			s.checkAndUpdate(ctx)

		case <-ctx.Done():
			// コンテキストがキャンセルされた: 終了処理
			slog.Info("スケジューラーを停止します",
				"reason", ctx.Err(),
			)
			return
		}
	}
}

// checkAndUpdate は、現在のIPアドレスを取得し、
// 前回と異なる場合にDuckDNSを更新します（内部用ヘルパー関数）
func (s *Scheduler) checkAndUpdate(ctx context.Context) {
	// TODO: タスク5.3で実装予定
	// この関数は次のタスクで実装します
	slog.Debug("checkAndUpdate を実行します（未実装）")
}
