// Package duckdns は、DuckDNS APIへの更新リクエストを行うクライアントを提供します。
// リトライ設定やタイムアウト設定を持ち、将来的な指数バックオフに対応します。
package duckdns

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// defaultBaseURL は DuckDNS の更新APIエンドポイントです。
const defaultBaseURL = "https://www.duckdns.org/update"

// DefaultHTTPTimeout は HTTPクライアントのデフォルトタイムアウトです。
const DefaultHTTPTimeout = 10 * time.Second

// DefaultMaxRetries はリトライのデフォルト最大回数です。
const DefaultMaxRetries = 3

// DefaultBackoff はリトライ時のデフォルトのバックオフ時間です。
var DefaultBackoff = []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

// HTTPDoer は http.Client の Do メソッド互換のインターフェースです。
// テストでモック可能にするため、HTTPクライアントをインターフェース化します。
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetryConfig はリトライの設定を表します。
// 最大リトライ回数とバックオフ時間のリストを持ちます。
type RetryConfig struct {
	MaxRetries int
	Backoff    []time.Duration
}

// Client は DuckDNS API への更新リクエストを実行するためのクライアントです。
// HTTPクライアント、ベースURL、リトライ設定を保持します。
type Client struct {
	httpClient HTTPDoer
	baseURL    string
	retry      RetryConfig
}

// NewClient は既定値で初期化された DuckDNS クライアントを作成します。
// - HTTPタイムアウト: 10秒
// - ベースURL: https://www.duckdns.org/update
// - リトライ: 最大3回、1s/2s/4s のバックオフ
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultHTTPTimeout},
		baseURL:    defaultBaseURL,
		retry: RetryConfig{
			MaxRetries: DefaultMaxRetries,
			Backoff:    append([]time.Duration(nil), DefaultBackoff...),
		},
	}
}

// NewClientWithOptions は指定の HTTP クライアント、ベースURL、リトライ設定で
// DuckDNS クライアントを作成します。引数がゼロ値の場合は適切に既定値を適用します。
func NewClientWithOptions(httpClient HTTPDoer, baseURL string, retry RetryConfig) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	// リトライ設定の既定値補完
	if retry.MaxRetries <= 0 {
		retry.MaxRetries = DefaultMaxRetries
	}
	if len(retry.Backoff) == 0 {
		retry.Backoff = append([]time.Duration(nil), DefaultBackoff...)
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		retry:      retry,
	}
}

// Update は DuckDNS API を呼び出してDNSレコードを更新します。
// domain, token, ip を指定してGETリクエストを送信し、レスポンスボディを返します。
//
// Parameters:
//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
//   - domain: 更新するDuckDNSドメイン名（例: "your-domain"）
//   - token: DuckDNS APIの認証トークン
//   - ip: 更新するIPアドレス（IPv4形式）
//
// Returns:
//   - string: レスポンスボディ（"OK" または "KO"）
//   - error: エラーが発生した場合
func (c *Client) Update(ctx context.Context, domain, token, ip string) (string, error) {
	// クエリパラメータの構築
	params := url.Values{}
	params.Set("domains", domain)
	params.Set("token", token)
	params.Set("ip", ip)

	// URL構築
	reqURL := c.baseURL + "?" + params.Encode()

	slog.Info("DuckDNS更新リクエスト送信",
		"domain", domain,
		"ip", ip,
		"url", c.baseURL,
	)

	// HTTPリクエスト作成
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("リクエスト作成に失敗しました: %w", err)
	}

	// User-Agent設定
	req.Header.Set("User-Agent", "duckdns-updater/1.0")

	// HTTPリクエスト送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("DuckDNS APIリクエスト失敗",
			"domain", domain,
			"error", err,
		)
		return "", fmt.Errorf("HTTPリクエスト実行に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		slog.Error("DuckDNS APIステータスエラー",
			"domain", domain,
			"status_code", resp.StatusCode,
		)
		return "", fmt.Errorf("HTTPステータスエラー: %d", resp.StatusCode)
	}

	// レスポンスボディ読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンス読み込みに失敗しました: %w", err)
	}

	// レスポンス文字列の取得（空白・改行を削除）
	response := strings.TrimSpace(string(body))

	// レスポンス解析："OK" / "KO" の判定
	if response == "OK" {
		slog.Info("DuckDNS更新成功",
			"domain", domain,
			"ip", ip,
			"response", response,
		)
		return response, nil
	}

	// "KO" またはその他の予期しないレスポンス
	slog.Error("DuckDNS更新失敗",
		"domain", domain,
		"ip", ip,
		"response", response,
	)
	return response, fmt.Errorf("DuckDNS更新に失敗しました: レスポンス=%s", response)
}

// UpdateWithRetry は指数バックオフアルゴリズムでリトライしながら
// DuckDNS API を呼び出してDNSレコードを更新します。
// 最大リトライ回数と各リトライ間のバックオフ時間は Client の retry 設定に従います。
//
// Parameters:
//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
//   - domain: 更新するDuckDNSドメイン名（例: "your-domain"）
//   - token: DuckDNS APIの認証トークン
//   - ip: 更新するIPアドレス（IPv4形式）
//
// Returns:
//   - string: レスポンスボディ（"OK" または "KO"）
//   - error: すべてのリトライが失敗した場合
func (c *Client) UpdateWithRetry(ctx context.Context, domain, token, ip string) (string, error) {
	var lastErr error
	maxAttempts := c.retry.MaxRetries + 1 // 最初の試行 + リトライ回数

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// コンテキストがキャンセルされているか確認
		select {
		case <-ctx.Done():
			slog.Warn("DuckDNS更新がキャンセルされました",
				"domain", domain,
				"attempt", attempt,
				"error", ctx.Err(),
			)
			return "", fmt.Errorf("更新がキャンセルされました: %w", ctx.Err())
		default:
		}

		// 試行開始ログ
		if attempt == 1 {
			slog.Info("DuckDNS更新を開始",
				"domain", domain,
				"ip", ip,
				"max_retries", c.retry.MaxRetries,
			)
		} else {
			slog.Info("DuckDNS更新をリトライ",
				"domain", domain,
				"ip", ip,
				"attempt", attempt,
				"max_attempts", maxAttempts,
			)
		}

		// 更新を試行
		response, err := c.Update(ctx, domain, token, ip)
		if err == nil {
			// 成功
			if attempt > 1 {
				slog.Info("DuckDNS更新がリトライで成功",
					"domain", domain,
					"ip", ip,
					"attempt", attempt,
				)
			}
			return response, nil
		}

		// エラーを記録
		lastErr = err

		// 最後の試行でない場合はバックオフ
		if attempt < maxAttempts {
			// バックオフ時間を取得（範囲外の場合は最後の値を使用）
			backoffIndex := attempt - 1
			if backoffIndex >= len(c.retry.Backoff) {
				backoffIndex = len(c.retry.Backoff) - 1
			}
			backoffDuration := c.retry.Backoff[backoffIndex]

			slog.Warn("DuckDNS更新が失敗、バックオフ後にリトライ",
				"domain", domain,
				"attempt", attempt,
				"backoff", backoffDuration.String(),
				"error", err,
			)

			// バックオフ待機（contextのキャンセルも監視）
			select {
			case <-time.After(backoffDuration):
				// バックオフ完了、次の試行へ
			case <-ctx.Done():
				slog.Warn("バックオフ中にキャンセルされました",
					"domain", domain,
					"error", ctx.Err(),
				)
				return "", fmt.Errorf("バックオフ中にキャンセルされました: %w", ctx.Err())
			}
		}
	}

	// すべての試行が失敗
	slog.Error("DuckDNS更新の全リトライが失敗",
		"domain", domain,
		"ip", ip,
		"attempts", maxAttempts,
		"last_error", lastErr,
	)
	return "", fmt.Errorf("DuckDNS更新に失敗しました（%d回試行）: %w", maxAttempts, lastErr)
}
