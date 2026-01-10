// Package config は、DuckDNS自動更新プログラムの設定管理を提供します。
// YAML設定ファイルと環境変数からの設定読み込み、バリデーション機能を含みます。
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config は、DuckDNS自動更新プログラムの全体設定を保持する構造体です。
// YAMLファイルまたは環境変数から読み込まれます。
type Config struct {
	// DuckDNS は、DuckDNSサービスへの接続設定を保持します
	DuckDNS DuckDNSConfig `yaml:"duckdns"`

	// Update は、DNS更新の動作設定を保持します
	Update UpdateConfig `yaml:"update"`

	// IPSources は、グローバルIPアドレスを取得するためのURLリストです
	IPSources []string `yaml:"ip_sources"`

	// Log は、ログ出力の設定を保持します
	Log LogConfig `yaml:"log"`
}

// DuckDNSConfig は、DuckDNSサービスへの認証情報を保持する構造体です。
type DuckDNSConfig struct {
	// Domain は、更新するDuckDNSのドメイン名です（例: "your-domain"）
	Domain string `yaml:"domain"`

	// Token は、DuckDNS APIの認証トークンです
	// 環境変数 DUCKDNS_TOKEN からの読み込みを推奨します
	Token string `yaml:"token"`
}

// UpdateConfig は、DNS更新の実行間隔に関する設定を保持する構造体です。
type UpdateConfig struct {
	// Interval は、IPアドレスのチェックと更新を実行する間隔です
	// フォーマット例: "5m", "1h", "30s"
	Interval time.Duration `yaml:"interval"`
}

// LogConfig は、ログ出力の形式とレベルに関する設定を保持する構造体です。
type LogConfig struct {
	// Level は、ログ出力レベルです
	// 有効な値: "debug", "info", "warn", "error"
	Level string `yaml:"level"`

	// Format は、ログ出力形式です
	// 有効な値: "json", "text"
	Format string `yaml:"format"`
}

// ValidationError は、設定のバリデーションエラーを保持する構造体です。
// 複数のエラーメッセージを含むことができます。
type ValidationError struct {
	Errors []string
}

// Error は ValidationError を error インターフェースに実装します。
func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return "unknown validation error"
	}
	return strings.Join(ve.Errors, "\n  - ")
}

// Validate は設定の妥当性をチェックします。
// 複数のバリデーションエラーがある場合は、すべてを収集して返します。
//
// Returns:
//   - error: バリデーションエラーがある場合
func (c *Config) Validate() error {
	var errors []string

	// 必須項目チェック
	if strings.TrimSpace(c.DuckDNS.Domain) == "" {
		errors = append(errors, "DuckDNSドメイン名が設定されていません (設定項目: duckdns.domain または環境変数: DUCKDNS_DOMAIN)")
	}
	if strings.TrimSpace(c.DuckDNS.Token) == "" {
		errors = append(errors, "DuckDNS APIトークンが設定されていません (設定項目: duckdns.token または環境変数: DUCKDNS_TOKEN)")
	}

	// 更新間隔のチェック
	if c.Update.Interval == 0 {
		errors = append(errors, "更新間隔が設定されていません (設定項目: update.interval または環境変数: DUCKDNS_INTERVAL, 例: \"5m\", \"1h\")")
	} else if c.Update.Interval < 0 {
		errors = append(errors, "更新間隔は正の値である必要があります")
	}

	// IP取得ソースのバリデーション
	if len(c.IPSources) == 0 {
		errors = append(errors, "IP取得ソースが1つも設定されていません (設定項目: ip_sources)")
	} else {
		for i, source := range c.IPSources {
			if strings.TrimSpace(source) == "" {
				errors = append(errors, fmt.Sprintf("IP取得ソース[%d]が空です", i))
				continue
			}

			// URLの妥当性をチェック
			if !isValidURL(source) {
				errors = append(errors, fmt.Sprintf("IP取得ソース[%d] \"%s\" が無効なURLです", i, source))
			}
		}
	}

	// ログレベルのバリデーション
	if c.Log.Level != "" {
		validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if !validLevels[strings.ToLower(c.Log.Level)] {
			errors = append(errors, fmt.Sprintf("無効なログレベル \"%s\" です (有効な値: debug, info, warn, error)", c.Log.Level))
		}
	}

	// ログフォーマットのバリデーション
	if c.Log.Format != "" {
		validFormats := map[string]bool{"json": true, "text": true}
		if !validFormats[strings.ToLower(c.Log.Format)] {
			errors = append(errors, fmt.Sprintf("無効なログフォーマット \"%s\" です (有効な値: json, text)", c.Log.Format))
		}
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}

// isValidURL はURLが有効かどうかを確認します。
func isValidURL(urlStr string) bool {
	// URLをパース
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// スキームがあるか確認（httpまたはhttps）
	if u.Scheme == "" {
		return false
	}

	// HTTPまたはHTTPSスキームのみを許可
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// ホストがあるか確認
	if u.Host == "" {
		return false
	}

	return true
}

// LoadFromFile は、指定されたYAMLファイルから設定を読み込みます。
// ファイルが存在しない場合やYAMLの解析に失敗した場合はエラーを返します。
//
// Parameters:
//   - path: 読み込むYAML設定ファイルのパス
//
// Returns:
//   - *Config: 読み込まれた設定
//   - error: エラーが発生した場合
func LoadFromFile(path string) (*Config, error) {
	// ファイルの存在確認と読み込み
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("設定ファイルが見つかりません: %s", path)
		}
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// YAMLのパース
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("YAML解析に失敗しました: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv は、環境変数から設定を読み込みます。
// 環境変数が設定されていない項目は空のままになります。
//
// 対応する環境変数:
//   - DUCKDNS_DOMAIN: DuckDNSのドメイン名
//   - DUCKDNS_TOKEN: DuckDNS APIトークン
//   - DUCKDNS_INTERVAL: 更新間隔（例: "5m", "1h"）
//
// Returns:
//   - *Config: 環境変数から読み込まれた設定
//   - error: 環境変数のパースに失敗した場合
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}

	// DuckDNS設定の読み込み
	if domain := os.Getenv("DUCKDNS_DOMAIN"); domain != "" {
		cfg.DuckDNS.Domain = domain
	}
	if token := os.Getenv("DUCKDNS_TOKEN"); token != "" {
		cfg.DuckDNS.Token = token
	}

	// 更新間隔の読み込み
	if interval := os.Getenv("DUCKDNS_INTERVAL"); interval != "" {
		duration, err := time.ParseDuration(interval)
		if err != nil {
			return nil, fmt.Errorf("DUCKDNS_INTERVAL の解析に失敗しました: %w", err)
		}
		cfg.Update.Interval = duration
	}

	return cfg, nil
}

// Load は、YAMLファイルと環境変数から設定を読み込みます。
// 環境変数の値は、YAMLファイルの値より優先されます。
//
// Parameters:
//   - path: 読み込むYAML設定ファイルのパス（空文字列の場合は環境変数のみ）
//
// Returns:
//   - *Config: 読み込まれた設定
//   - error: エラーが発生した場合
func Load(path string) (*Config, error) {
	var cfg *Config
	var err error

	// YAMLファイルからの読み込み
	if path != "" {
		cfg, err = LoadFromFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		// ファイルパスが指定されていない場合は空の設定から開始
		cfg = &Config{}
	}

	// 環境変数からの読み込み
	envCfg, err := LoadFromEnv()
	if err != nil {
		return nil, err
	}

	// 環境変数で設定された値をマージ（環境変数が優先）
	if envCfg.DuckDNS.Domain != "" {
		cfg.DuckDNS.Domain = envCfg.DuckDNS.Domain
	}
	if envCfg.DuckDNS.Token != "" {
		cfg.DuckDNS.Token = envCfg.DuckDNS.Token
	}
	if envCfg.Update.Interval != 0 {
		cfg.Update.Interval = envCfg.Update.Interval
	}

	return cfg, nil
}
