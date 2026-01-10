// Package logger は、アプリケーションの構造化ログ管理を提供します。
// log/slog を使用して JSON またはテキスト形式でのログ出力に対応します。
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// InitLogger は、指定されたログレベルとフォーマットでロガーを初期化します。
//
// パラメータ:
//   - level: ログレベル ("debug", "info", "warn", "error")
//   - format: ログフォーマット ("json" または "text")
//   - writer: ログ出力先 (デフォルト: os.Stderr)
//
// 戻り値:
//   - エラーが発生した場合は error を返します
func InitLogger(level, format string, writer ...io.Writer) error {
	// 出力先を決定
	var output io.Writer = os.Stderr
	if len(writer) > 0 && writer[0] != nil {
		output = writer[0]
	}

	// ログレベルを解析
	logLevel := parseLogLevel(level)

	// ログハンドラーを作成
	var handler slog.Handler

	switch strings.ToLower(format) {
	case "json":
		// JSON形式でのログ出力
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	case "text", "":
		// テキスト形式でのログ出力（デフォルト）
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	default:
		// 不正なフォーマット
		slog.Error("不正なログフォーマットが指定されました", "format", format)
		// テキスト形式にフォールバック
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	}

	// デフォルトロガーを設定
	slog.SetDefault(slog.New(handler))

	// 初期化完了をログ出力
	slog.Info("ログシステムが初期化されました",
		"level", level,
		"format", format,
	)

	return nil
}

// parseLogLevel は、文字列のログレベルを slog.Level に変換します。
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// デフォルトは Info レベル
		return slog.LevelInfo
	}
}

// GetLogger は、デフォルトロガーを取得します。
func GetLogger() *slog.Logger {
	return slog.Default()
}
