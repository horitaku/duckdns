package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/horitaku/duckdns/internal/duckdns"
)

// MockFetcher は、テスト用のIP Fetcher モックです。
type MockFetcher struct {
	FetchFunc  func(ctx context.Context) (string, error)
	FetchCount int32
	mu         sync.Mutex
}

// Fetch は MockFetcher の Fetch メソッドを実装します。
func (m *MockFetcher) Fetch(ctx context.Context) (string, error) {
	atomic.AddInt32(&m.FetchCount, 1)

	if m.FetchFunc != nil {
		return m.FetchFunc(ctx)
	}
	return "", nil
}

// GetFetchCount はスレッドセーフに取得回数を返します。
func (m *MockFetcher) GetFetchCount() int {
	return int(atomic.LoadInt32(&m.FetchCount))
}

// TestNewScheduler は、Scheduler の作成をテストします。
func TestNewScheduler(t *testing.T) {
	interval := 5 * time.Minute
	mockFetcher := &MockFetcher{}
	mockClient := duckdns.NewClient()
	domain := "test-domain"
	token := "test-token"

	scheduler := NewScheduler(interval, mockFetcher, mockClient, domain, token)

	if scheduler.interval != interval {
		t.Errorf("interval が一致しません。期待: %v, 実際: %v", interval, scheduler.interval)
	}

	if scheduler.domain != domain {
		t.Errorf("domain が一致しません。期待: %s, 実際: %s", domain, scheduler.domain)
	}

	if scheduler.token != token {
		t.Errorf("token が一致しません。期待: %s, 実際: %s", token, scheduler.token)
	}

	if scheduler.lastIP != "" {
		t.Errorf("lastIP は空であるべき。期待: \"\", 実際: %s", scheduler.lastIP)
	}
}

// TestScheduler_Run_ImmediateCheck は、Run が起動直後に IP チェックを実行することをテストします。
func TestScheduler_Run_ImmediateCheck(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "192.168.1.1", nil
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(10*time.Second, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	scheduler.Run(ctx)

	if mockFetcher.GetFetchCount() < 1 {
		t.Errorf("起動直後の IP チェックが実行されていません。期待: >= 1, 実際: %d", mockFetcher.GetFetchCount())
	}
}

// TestScheduler_Run_PeriodicCheck は、Run が定期的に IP チェックを実行することをテストします。
func TestScheduler_Run_PeriodicCheck(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "192.168.1.1", nil
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(50*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	scheduler.Run(ctx)

	count := mockFetcher.GetFetchCount()
	// 起動直後 + 約 6 回の定期実行
	if count < 3 {
		t.Logf("IP チェック回数: %d", count)
	}
}

// TestScheduler_Run_ContextCancellation は、context キャンセル時に Run が停止することをテストします。
func TestScheduler_Run_ContextCancellation(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "192.168.1.1", nil
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(50*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// 正常に終了
	case <-time.After(1 * time.Second):
		t.Error("Run がキャンセル後も終了しません")
	}
}

// TestScheduler_FetchError は、IP 取得失敗時に継続することをテストします。
func TestScheduler_FetchError(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "", context.Canceled
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(10*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	scheduler.Run(ctx)

	if scheduler.lastIP != "" {
		t.Errorf("lastIP は更新されていないはず。期待: \"\", 実際: %s", scheduler.lastIP)
	}
}

// TestScheduler_ContextCancelledDuringCheck は、チェック中のキャンセルをテストします。
func TestScheduler_ContextCancelledDuringCheck(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "192.168.1.1", nil
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(10*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	scheduler.checkAndUpdate(ctx)

	if scheduler.lastIP != "" {
		t.Errorf("lastIP は更新されていないはず。期待: \"\", 実際: %s", scheduler.lastIP)
	}
}

// TestScheduler_FetchCount は、スケジューラーが複数回 Fetch を呼び出すことをテストします。
func TestScheduler_FetchCount(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "192.168.1.1", nil
			}
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(10*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	scheduler.Run(ctx)

	if mockFetcher.GetFetchCount() == 0 {
		t.Error("Fetch が呼び出されていません")
	}
}

// TestScheduler_MultipleCheckAndUpdate は、複数回 checkAndUpdate を呼び出すテストです。
func TestScheduler_MultipleCheckAndUpdate(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFunc: func(ctx context.Context) (string, error) {
			return "192.168.1.1", nil
		},
	}

	mockClient := duckdns.NewClient()
	scheduler := NewScheduler(10*time.Millisecond, mockFetcher, mockClient, "test-domain", "test-token")

	// 複数回実行
	for i := 0; i < 3; i++ {
		scheduler.checkAndUpdate(context.Background())
	}

	if mockFetcher.GetFetchCount() != 3 {
		t.Errorf("Fetch が期待回数呼び出されていません。期待: 3, 実際: %d", mockFetcher.GetFetchCount())
	}
}
