package ip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestValidateIPv4 は、IPv4アドレスの検証をテストします。
func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "有効なIPv4: 192.168.1.1",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "有効なIPv4: 0.0.0.0",
			ip:      "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "有効なIPv4: 255.255.255.255",
			ip:      "255.255.255.255",
			wantErr: false,
		},
		{
			name:    "有効なIPv4: 8.8.8.8",
			ip:      "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "空文字列",
			ip:      "",
			wantErr: true,
		},
		{
			name:    "無効: 256.256.256.256",
			ip:      "256.256.256.256",
			wantErr: true,
		},
		{
			name:    "無効: 192.168.1",
			ip:      "192.168.1",
			wantErr: true,
		},
		{
			name:    "無効: 192.168.1.1.1",
			ip:      "192.168.1.1.1",
			wantErr: true,
		},
		{
			name:    "無効: not-an-ip",
			ip:      "not-an-ip",
			wantErr: true,
		},
		{
			name:    "無効: IPv6",
			ip:      "::1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPv4(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}
		})
	}
}

// TestNewHTTPFetcher は、HTTPFetcherの作成をテストします。
func TestNewHTTPFetcher(t *testing.T) {
	url := "https://api.ipify.org"
	fetcher := NewHTTPFetcher(url)

	if fetcher.URL != url {
		t.Errorf("URL が一致しません。期待: %s, 実際: %s", url, fetcher.URL)
	}

	if fetcher.client.Timeout != DefaultHTTPTimeout {
		t.Errorf("タイムアウトが一致しません。期待: %v, 実際: %v", DefaultHTTPTimeout, fetcher.client.Timeout)
	}
}

// TestNewHTTPFetcherWithTimeout は、カスタムタイムアウト設定でのHTTPFetcher作成をテストします。
func TestNewHTTPFetcherWithTimeout(t *testing.T) {
	url := "https://api.ipify.org"
	timeout := 5 * time.Second
	fetcher := NewHTTPFetcherWithTimeout(url, timeout)

	if fetcher.URL != url {
		t.Errorf("URL が一致しません。期待: %s, 実際: %s", url, fetcher.URL)
	}

	if fetcher.client.Timeout != timeout {
		t.Errorf("タイムアウトが一致しません。期待: %v, 実際: %v", timeout, fetcher.client.Timeout)
	}
}

// TestHTTPFetcher_Fetch_Success は、IP取得成功をテストします。
func TestHTTPFetcher_Fetch_Success(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL)
	ip, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if ip != "192.168.1.1" {
		t.Errorf("IPが一致しません。期待: 192.168.1.1, 実際: %s", ip)
	}
}

// TestHTTPFetcher_Fetch_WithWhitespace は、ホワイトスペース付きのIP取得をテストします。
func TestHTTPFetcher_Fetch_WithWhitespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("  \n192.168.1.1\t\n  "))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL)
	ip, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if ip != "192.168.1.1" {
		t.Errorf("IPが一致しません。期待: 192.168.1.1, 実際: %s", ip)
	}
}

// TestHTTPFetcher_Fetch_InvalidIP は、無効なIPを返すサーバーをテストします。
func TestHTTPFetcher_Fetch_InvalidIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-an-ip"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL)
	_, err := fetcher.Fetch(context.Background())

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestHTTPFetcher_Fetch_EmptyResponse は、空のレスポンスをテストします。
func TestHTTPFetcher_Fetch_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL)
	_, err := fetcher.Fetch(context.Background())

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestHTTPFetcher_Fetch_StatusNotOK は、ステータスコード エラーをテストします。
func TestHTTPFetcher_Fetch_StatusNotOK(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			fetcher := NewHTTPFetcher(server.URL)
			_, err := fetcher.Fetch(context.Background())

			if err == nil {
				t.Error("エラーが返されるべきですが、nilが返されました")
			}
		})
	}
}

// TestHTTPFetcher_Fetch_Timeout は、リクエストタイムアウトをテストします。
func TestHTTPFetcher_Fetch_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// タイムアウトより長い時間待機
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server.Close()

	// 100ミリ秒のタイムアウトを設定
	fetcher := NewHTTPFetcherWithTimeout(server.URL, 100*time.Millisecond)
	_, err := fetcher.Fetch(context.Background())

	if err == nil {
		t.Error("タイムアウトエラーが返されるべき")
	}
}

// TestHTTPFetcher_Fetch_ContextCancelled は、キャンセルされたコンテキストをテストします。
func TestHTTPFetcher_Fetch_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	fetcher := NewHTTPFetcher(server.URL)
	_, err := fetcher.Fetch(ctx)

	if err == nil {
		t.Error("キャンセルエラーが返されるべき")
	}
}

// TestNewMultipleFetcher は、MultipleFetcherの作成をテストします。
func TestNewMultipleFetcher(t *testing.T) {
	urls := []string{
		"https://api.ipify.org",
		"https://ifconfig.me",
	}
	fetcher := NewMultipleFetcher(urls)

	if len(fetcher.URLs) != len(urls) {
		t.Errorf("URLの数が一致しません。期待: %d, 実際: %d", len(urls), len(fetcher.URLs))
	}

	if fetcher.timeout != DefaultHTTPTimeout {
		t.Errorf("タイムアウトが一致しません。期待: %v, 実際: %v", DefaultHTTPTimeout, fetcher.timeout)
	}
}

// TestNewMultipleFetcherWithTimeout は、カスタムタイムアウトでのMultipleFetcher作成をテストします。
func TestNewMultipleFetcherWithTimeout(t *testing.T) {
	urls := []string{"https://api.ipify.org"}
	timeout := 5 * time.Second
	fetcher := NewMultipleFetcherWithTimeout(urls, timeout)

	if fetcher.timeout != timeout {
		t.Errorf("タイムアウトが一致しません。期待: %v, 実際: %v", timeout, fetcher.timeout)
	}
}

// TestMultipleFetcher_Fetch_FirstSuccess は、最初のソースで成功をテストします。
func TestMultipleFetcher_Fetch_FirstSuccess(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	fetcher := NewMultipleFetcher([]string{server1.URL, server2.URL})
	ip, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if ip != "192.168.1.1" {
		t.Errorf("IPが一致しません。期待: 192.168.1.1, 実際: %s", ip)
	}
}

// TestMultipleFetcher_Fetch_SecondSuccess は、2番目のソースで成功をテストします。
func TestMultipleFetcher_Fetch_SecondSuccess(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("10.0.0.1"))
	}))
	defer server2.Close()

	fetcher := NewMultipleFetcher([]string{server1.URL, server2.URL})
	ip, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if ip != "10.0.0.1" {
		t.Errorf("IPが一致しません。期待: 10.0.0.1, 実際: %s", ip)
	}
}

// TestMultipleFetcher_Fetch_AllFail は、すべてのソースの失敗をテストします。
func TestMultipleFetcher_Fetch_AllFail(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server2.Close()

	fetcher := NewMultipleFetcher([]string{server1.URL, server2.URL})
	_, err := fetcher.Fetch(context.Background())

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestMultipleFetcher_Fetch_NoURLs は、URLが設定されていないをテストします。
func TestMultipleFetcher_Fetch_NoURLs(t *testing.T) {
	fetcher := NewMultipleFetcher([]string{})
	_, err := fetcher.Fetch(context.Background())

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestMultipleFetcher_Fetch_EmptyURL は、空のURLをスキップをテストします。
func TestMultipleFetcher_Fetch_EmptyURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("172.16.0.1"))
	}))
	defer server.Close()

	// 最初のURLは空、2番目が有効
	fetcher := NewMultipleFetcher([]string{"", server.URL})
	ip, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if ip != "172.16.0.1" {
		t.Errorf("IPが一致しません。期待: 172.16.0.1, 実際: %s", ip)
	}
}

// TestMultipleFetcher_Fetch_ContextTimeout は、コンテキストタイムアウトをテストします。
func TestMultipleFetcher_Fetch_ContextTimeout(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("10.0.0.1"))
	}))
	defer server2.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	fetcher := NewMultipleFetcher([]string{server1.URL, server2.URL})
	_, err := fetcher.Fetch(ctx)

	if err == nil {
		t.Error("タイムアウトエラーが返されるべき")
	}
}

// TestHTTPFetcher_UserAgent は、User-Agent ヘッダーが正しく設定されているかテストします。
func TestHTTPFetcher_UserAgent(t *testing.T) {
	var receivedAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.1"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL)
	_, err := fetcher.Fetch(context.Background())

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	expectedAgent := "duckdns-updater/1.0"
	if receivedAgent != expectedAgent {
		t.Errorf("User-Agentが一致しません。期待: %s, 実際: %s", expectedAgent, receivedAgent)
	}
}
