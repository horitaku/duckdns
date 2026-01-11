package duckdns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockHTTPDoer はテスト用のモック HTTP クライアントです。
type MockHTTPDoer struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do は MockHTTPDoer の Do メソッドを実装します。
func (m *MockHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

// TestNewClient は、NewClient が既定値で正しく初期化されることをテストします。
func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.baseURL != defaultBaseURL {
		t.Errorf("ベースURLが一致しません。期待: %s, 実際: %s", defaultBaseURL, client.baseURL)
	}

	if client.retry.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetriesが一致しません。期待: %d, 実際: %d", DefaultMaxRetries, client.retry.MaxRetries)
	}

	if len(client.retry.Backoff) != len(DefaultBackoff) {
		t.Errorf("Backoffの長さが一致しません。期待: %d, 実際: %d", len(DefaultBackoff), len(client.retry.Backoff))
	}
}

// TestNewClientWithOptions は、NewClientWithOptions でオプション指定できることをテストします。
func TestNewClientWithOptions(t *testing.T) {
	mockClient := &MockHTTPDoer{}
	customURL := "https://custom.example.com/api"
	customRetry := RetryConfig{
		MaxRetries: 5,
		Backoff:    []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
	}

	client := NewClientWithOptions(mockClient, customURL, customRetry)

	if client.baseURL != customURL {
		t.Errorf("ベースURLが一致しません。期待: %s, 実際: %s", customURL, client.baseURL)
	}

	if client.retry.MaxRetries != customRetry.MaxRetries {
		t.Errorf("MaxRetriesが一致しません。期待: %d, 実際: %d", customRetry.MaxRetries, client.retry.MaxRetries)
	}
}

// TestNewClientWithOptions_ZeroValues は、ゼロ値の場合に既定値が適用されることをテストします。
func TestNewClientWithOptions_ZeroValues(t *testing.T) {
	client := NewClientWithOptions(nil, "", RetryConfig{})

	if client.baseURL != defaultBaseURL {
		t.Errorf("ベースURLがデフォルトに設定されるべき。期待: %s, 実際: %s", defaultBaseURL, client.baseURL)
	}

	if client.retry.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetriesがデフォルトに設定されるべき。期待: %d, 実際: %d", DefaultMaxRetries, client.retry.MaxRetries)
	}
}

// TestClient_Update_Success は、更新成功をテストします。
func TestClient_Update_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("domains") != "test-domain" {
			t.Errorf("domains パラメータが一致しません。期待: test-domain, 実際: %s", query.Get("domains"))
		}
		if query.Get("token") != "test-token" {
			t.Errorf("token パラメータが一致しません。期待: test-token, 実際: %s", query.Get("token"))
		}
		if query.Get("ip") != "192.168.1.1" {
			t.Errorf("ip パラメータが一致しません。期待: 192.168.1.1, 実際: %s", query.Get("ip"))
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	response, err := client.Update(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}
}

// TestClient_Update_Failure は、更新失敗をテストします。
func TestClient_Update_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("KO")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	_, err := client.Update(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestClient_Update_StatusError は、ステータスコードエラーをテストします。
func TestClient_Update_StatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	_, err := client.Update(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestClient_Update_ContextCancelled は、キャンセルされたコンテキストをテストします。
func TestClient_Update_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	_, err := client.Update(ctx, "test-domain", "test-token", "192.168.1.1")

	if err == nil {
		t.Error("キャンセルエラーが返されるべき")
	}
}

// TestClient_Update_WithWhitespace は、ホワイトスペース付きのレスポンスをテストします。
func TestClient_Update_WithWhitespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("  \nOK\t\n  ")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	response, err := client.Update(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}
}

// TestClient_UpdateWithRetry_Success は、最初の試行で成功をテストします。
func TestClient_UpdateWithRetry_Success(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{10 * time.Millisecond},
	})

	response, err := client.UpdateWithRetry(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}

	if attemptCount != 1 {
		t.Errorf("試行回数が一致しません。期待: 1, 実際: %d", attemptCount)
	}
}

// TestClient_UpdateWithRetry_SuccessAfterRetry は、リトライで成功をテストします。
func TestClient_UpdateWithRetry_SuccessAfterRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				t.Errorf("w.Write failed: %v", err)
			}
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	})

	response, err := client.UpdateWithRetry(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}

	if attemptCount != 2 {
		t.Errorf("試行回数が一致しません。期待: 2, 実際: %d", attemptCount)
	}
}

// TestClient_UpdateWithRetry_AllFail は、すべてのリトライが失敗をテストします。
func TestClient_UpdateWithRetry_AllFail(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 2,
		Backoff:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond},
	})

	_, err := client.UpdateWithRetry(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}

	if attemptCount != 3 {
		t.Errorf("試行回数が一致しません。期待: 3, 実際: %d", attemptCount)
	}
}

// TestClient_UpdateWithRetry_ContextCancelled は、リトライ中のキャンセルをテストします。
func TestClient_UpdateWithRetry_ContextCancelled(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 5,
		Backoff:    []time.Duration{500 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond},
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := client.UpdateWithRetry(ctx, "test-domain", "test-token", "192.168.1.1")

	if err == nil {
		t.Error("キャンセルエラーが返されるべき")
	}

	// 最初の試行で失敗してバックオフ開始、そこでキャンセルされるので attemptCount は 1
	if attemptCount != 1 {
		t.Errorf("試行回数が一致しません。期待: 1, 実際: %d", attemptCount)
	}
}

// TestClient_UpdateWithRetry_BackoffWait は、バックオフ待機をテストします。
func TestClient_UpdateWithRetry_BackoffWait(t *testing.T) {
	attemptCount := 0
	startTime := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				t.Errorf("w.Write failed: %v", err)
			}
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{100 * time.Millisecond},
	})

	response, err := client.UpdateWithRetry(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}

	elapsed := time.Since(startTime)
	if elapsed < 100*time.Millisecond {
		t.Errorf("バックオフ時間が不足しています。期待: >= 100ms, 実際: %v", elapsed)
	}
}

// TestClient_Update_UserAgent は、User-Agent ヘッダーが正しく設定されているかテストします。
func TestClient_Update_UserAgent(t *testing.T) {
	var receivedAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{})
	_, err := client.Update(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	expectedAgent := "duckdns-updater/1.0"
	if receivedAgent != expectedAgent {
		t.Errorf("User-Agentが一致しません。期待: %s, 実際: %s", expectedAgent, receivedAgent)
	}
}

// TestClient_UpdateWithRetry_RecoveryMessage は、リトライ成功時のメッセージをテストします。
func TestClient_UpdateWithRetry_RecoveryMessage(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				t.Errorf("w.Write failed: %v", err)
			}
		}
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{}, server.URL, RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond},
	})

	response, err := client.UpdateWithRetry(context.Background(), "test-domain", "test-token", "192.168.1.1")

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if response != "OK" {
		t.Errorf("レスポンスが一致しません。期待: OK, 実際: %s", response)
	}

	if attemptCount != 3 {
		t.Errorf("試行回数が一致しません。期待: 3, 実際: %d", attemptCount)
	}
}

// TestMockHTTPDoer は、MockHTTPDoer のテストです。
func TestMockHTTPDoer(t *testing.T) {
	mockClient := &MockHTTPDoer{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := mockClient.Do(req)

	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが一致しません。期待: %d, 実際: %d", http.StatusOK, resp.StatusCode)
	}
}
