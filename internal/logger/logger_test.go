package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "ログシステムが初期化されたます") {
		t.Errorf("初期化メッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "json", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "ログシステムが初期化されたます") {
		t.Errorf("初期化メッセージがログに含まれていません: %s", output)
	}
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Errorf("JSON形式のログではありません: %s", output)
	}
}

func TestInitLogger_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("debug", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Debug("テストデバッグメッセージ")
	output := buf.String()
	if !strings.Contains(output, "テストデバッグメッセージ") {
		t.Errorf("デバッグメッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Info("テストインフォメッセージ")
	output := buf.String()
	if !strings.Contains(output, "テストインフォメッセージ") {
		t.Errorf("インフォメッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("warn", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Debug("このメッセージは表示されません")
	slog.Warn("テスト警告メッセージ")
	output := buf.String()
	if strings.Contains(output, "このメッセージは表示されません") {
		t.Errorf("デバッグメッセージがログに含まれるべきではありません: %s", output)
	}
	if !strings.Contains(output, "テスト警告メッセージ") {
		t.Errorf("警告メッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("error", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Info("このメッセージは表示されません")
	slog.Error("テストエラーメッセージ")
	output := buf.String()
	if strings.Contains(output, "このメッセージは表示されません") {
		t.Errorf("インフォメッセージがログに含まれるべきではありません: %s", output)
	}
	if !strings.Contains(output, "テストエラーメッセージ") {
		t.Errorf("エラーメッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "invalid-format", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Info("フォールバックテスト")
	output := buf.String()
	if !strings.Contains(output, "フォールバックテスト") {
		t.Errorf("フォールバックメッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_DefaultWriter(t *testing.T) {
	err := InitLogger("info", "text")
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	logger := GetLogger()
	if logger == nil {
		t.Error("デフォルトロガーが取得できません")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		testMsg  string
		logFunc  func(msg string)
		expected bool
	}{
		{
			name:     "debug レベル",
			level:    "debug",
			testMsg:  "デバッグメッセージ",
			logFunc:  func(msg string) { slog.Debug(msg) },
			expected: true,
		},
		{
			name:     "info レベル",
			level:    "info",
			testMsg:  "インフォメッセージ",
			logFunc:  func(msg string) { slog.Info(msg) },
			expected: true,
		},
		{
			name:     "warn レベル",
			level:    "warn",
			testMsg:  "警告メッセージ",
			logFunc:  func(msg string) { slog.Warn(msg) },
			expected: true,
		},
		{
			name:     "error レベル",
			level:    "error",
			testMsg:  "エラーメッセージ",
			logFunc:  func(msg string) { slog.Error(msg) },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := InitLogger(tt.level, "text", &buf)
			if err != nil {
				t.Errorf("エラーが発生しました: %v", err)
			}
			tt.logFunc(tt.testMsg)
			output := buf.String()
			contains := strings.Contains(output, tt.testMsg)
			if contains != tt.expected {
				t.Errorf("期待と異なります。期待: %v, 実際: %v, 出力: %s", tt.expected, contains, output)
			}
		})
	}
}

func TestParseLogLevel_InvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("invalid-level", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Debug("このメッセージは表示されません")
	slog.Info("デフォルトレベルテスト")
	output := buf.String()
	if strings.Contains(output, "このメッセージは表示されません") {
		t.Errorf("デバッグメッセージがログに含まれるべきではありません: %s", output)
	}
	if !strings.Contains(output, "デフォルトレベルテスト") {
		t.Errorf("インフォメッセージがログに含まれていません: %s", output)
	}
}

func TestGetLogger(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	logger := GetLogger()
	if logger == nil {
		t.Error("ロガーが取得できません")
	}
	logger.Info("GetLogger テスト")
	output := buf.String()
	if !strings.Contains(output, "GetLogger テスト") {
		t.Errorf("ログメッセージが含まれていません: %s", output)
	}
}

func TestInitLogger_WarningAlias(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("warning", "text", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Debug("デバッグメッセージ")
	slog.Info("インフォメッセージ")
	slog.Warn("警告メッセージ")
	output := buf.String()
	if strings.Contains(output, "デバッグメッセージ") || strings.Contains(output, "インフォメッセージ") {
		t.Errorf("デバッグ/インフォメッセージがログに含まれるべきではありません: %s", output)
	}
	if !strings.Contains(output, "警告メッセージ") {
		t.Errorf("警告メッセージがログに含まれていません: %s", output)
	}
}

func TestInitLogger_EmptyFormat(t *testing.T) {
	var buf bytes.Buffer
	err := InitLogger("info", "", &buf)
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	slog.Info("空フォーマットテスト")
	output := buf.String()
	if !strings.Contains(output, "空フォーマットテスト") {
		t.Errorf("ログメッセージが含まれていません: %s", output)
	}
}
