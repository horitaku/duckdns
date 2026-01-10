// Package ip は、グローバルIPアドレスの取得機能を提供します。
// 複数の外部ソースからIPアドレスを取得し、フェイルオーバーに対応しています。
package ip

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DefaultHTTPTimeout は、HTTPリクエストのデフォルトタイムアウト設定です。
const DefaultHTTPTimeout = 10 * time.Second

// Fetcher は、グローバルIPアドレスを取得するためのインターフェースです。
// 異なるIPソースの実装をサポートするために設計されています。
type Fetcher interface {
	// Fetch は、コンテキストを受け取ってIPアドレスを取得します。
	// 取得に成功した場合、IPv4アドレスを文字列として返します。
	// エラーが発生した場合は、エラーを返します。
	//
	// Parameters:
	//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
	//
	// Returns:
	//   - string: 取得したIPアドレス（IPv4形式）
	//   - error: エラーが発生した場合
	Fetch(ctx context.Context) (string, error)
}

// HTTPFetcher は、HTTPリクエストを使ってIPアドレスを取得する構造体です。
// 指定されたURLからレスポンスボディをIPアドレスとして解析します。
type HTTPFetcher struct {
	// URL は、IPアドレスを取得するエンドポイントです
	URL string

	// client は、タイムアウト設定付きのHTTPクライアントです
	client *http.Client
}

// NewHTTPFetcher は、タイムアウト設定付きのHTTPFetcherを作成します。
// デフォルトのタイムアウト (10秒) が適用されます。
//
// Parameters:
//   - url: IPアドレスを取得するエンドポイントのURL
//
// Returns:
//   - *HTTPFetcher: 作成されたHTTPFetcher
func NewHTTPFetcher(url string) *HTTPFetcher {
	return &HTTPFetcher{
		URL: url,
		client: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
	}
}

// NewHTTPFetcherWithTimeout は、カスタムタイムアウト設定でHTTPFetcherを作成します。
//
// Parameters:
//   - url: IPアドレスを取得するエンドポイントのURL
//   - timeout: HTTPリクエストのタイムアウト
//
// Returns:
//   - *HTTPFetcher: 作成されたHTTPFetcher
func NewHTTPFetcherWithTimeout(url string, timeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		URL: url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Fetch は、HTTPリクエストを使ってIPアドレスを取得します。
// コンテキストがキャンセルされた場合は、リクエストもキャンセルされます。
//
// Parameters:
//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
//
// Returns:
//   - string: 取得したIPアドレス
//   - error: エラーが発生した場合
func (f *HTTPFetcher) Fetch(ctx context.Context) (string, error) {
	// リクエスト作成
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return "", fmt.Errorf("リクエスト作成に失敗しました: %w", err)
	}

	// User-Agent設定
	req.Header.Set("User-Agent", "duckdns-updater/1.0")

	// リクエスト実行
	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTPリクエスト実行に失敗しました (%s): %w", f.URL, err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTPステータスエラー: %d (URL: %s)", resp.StatusCode, f.URL)
	}

	// レスポンスボディを読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンス読み込みに失敗しました: %w", err)
	}

	// IPアドレス抽出（空白やタブ、改行を削除）
	ip := strings.TrimSpace(string(body))

	if ip == "" {
		return "", fmt.Errorf("レスポンスが空です (URL: %s)", f.URL)
	}

	// IPv4アドレスのバリデーション
	if err := ValidateIPv4(ip); err != nil {
		return "", fmt.Errorf("無効なIPアドレス: %s (URL: %s, エラー: %w)", ip, f.URL, err)
	}

	return ip, nil
}

// ValidateIPv4 は、IPv4アドレスが有効かどうかを確認します。
// 正規表現チェックと net.ParseIP による検証を行います。
//
// Parameters:
//   - ip: 検証するIPアドレス文字列
//
// Returns:
//   - error: 無効なIPアドレスの場合
func ValidateIPv4(ip string) error {
	// 空文字列チェック
	if ip == "" {
		return fmt.Errorf("IPアドレスが空です")
	}

	// IPv4フォーマットの正規表現パターン
	// 0.0.0.0 から 255.255.255.255 までを許可
	ipv4Pattern := regexp.MustCompile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`)

	// 正規表現チェック
	if !ipv4Pattern.MatchString(ip) {
		return fmt.Errorf("IPv4形式に一致していません")
	}

	// net.ParseIP による厳密な検証
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("net.ParseIP による検証に失敗しました")
	}

	// IPv4かどうかを確認
	if parsedIP.To4() == nil {
		return fmt.Errorf("IPv6ではなくIPv4である必要があります")
	}

	return nil
}

// MultipleFetcher は、複数のIPソースからIPアドレスを順次試行して取得する構造体です。
// フェイルオーバー機能を提供し、最初に成功したソースのIPアドレスを返します。
type MultipleFetcher struct {
	// URLs は、試行するIPアドレス取得エンドポイントのURLリストです
	URLs []string

	// timeout は、各HTTPリクエストのタイムアウト設定です
	timeout time.Duration
}

// NewMultipleFetcher は、複数のURLから順次IPアドレスを取得する
// MultipleFetcherを作成します。デフォルトタイムアウト (10秒) が適用されます。
//
// Parameters:
//   - urls: 試行するIPアドレス取得エンドポイントのURLリスト
//
// Returns:
//   - *MultipleFetcher: 作成されたMultipleFetcher
func NewMultipleFetcher(urls []string) *MultipleFetcher {
	return NewMultipleFetcherWithTimeout(urls, DefaultHTTPTimeout)
}

// NewMultipleFetcherWithTimeout は、カスタムタイムアウト設定で
// MultipleFetcherを作成します。
//
// Parameters:
//   - urls: 試行するIPアドレス取得エンドポイントのURLリスト
//   - timeout: 各HTTPリクエストのタイムアウト
//
// Returns:
//   - *MultipleFetcher: 作成されたMultipleFetcher
func NewMultipleFetcherWithTimeout(urls []string, timeout time.Duration) *MultipleFetcher {
	return &MultipleFetcher{
		URLs:    urls,
		timeout: timeout,
	}
}

// Fetch は、複数のIPソースから順次試行してIPアドレスを取得します。
// 最初に成功したソースのIPアドレスを返します。
// すべての試行に失敗した場合は、詳細なエラーメッセージを返します。
//
// Parameters:
//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
//
// Returns:
//   - string: 取得したIPアドレス
//   - error: すべてのソースから取得できなかった場合
func (mf *MultipleFetcher) Fetch(ctx context.Context) (string, error) {
	if len(mf.URLs) == 0 {
		return "", fmt.Errorf("IP取得ソースが設定されていません")
	}

	// 各試行のエラーを記録
	var errors []string

	// 各URLを順次試行
	for i, url := range mf.URLs {
		// 空のURLをスキップ
		if strings.TrimSpace(url) == "" {
			errors = append(errors, fmt.Sprintf("[%d] URLが空です", i))
			slog.Warn("IPソースURLが空のためスキップ",
				"index", i,
				"url", url,
			)
			continue
		}

		// 試行開始ログ
		slog.Info("IP取得を試行",
			"index", i,
			"url", url,
			"timeout", mf.timeout.String(),
		)

		// HTTPFetcherで取得を試行
		fetcher := NewHTTPFetcherWithTimeout(url, mf.timeout)
		ip, err := fetcher.Fetch(ctx)

		// 成功時はIPを返す
		if err == nil {
			slog.Info("IP取得に成功",
				"index", i,
				"url", url,
				"ip", ip,
			)
			return ip, nil
		}

		// 失敗をログに記録
		errors = append(errors, fmt.Sprintf("[%d] %s: %v", i, url, err))
		slog.Warn("IP取得に失敗",
			"index", i,
			"url", url,
			"error", err,
		)
	}

	// すべての試行が失敗した場合
	errorMessage := "すべてのIP取得ソースから取得に失敗しました:\n  - " + strings.Join(errors, "\n  - ")
	slog.Error("IP取得ソースの全試行が失敗",
		"errors", errors,
	)
	return "", fmt.Errorf("%s", errorMessage)
}
