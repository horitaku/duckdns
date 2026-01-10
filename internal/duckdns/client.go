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

	slog.Info("DuckDNS APIレスポンス受信",
		"domain", domain,
		"response", response,
	)

	return response, nil
}
