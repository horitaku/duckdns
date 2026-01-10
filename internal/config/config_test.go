package config

import (
	"os"
	"testing"
	"time"
)

// TestLoadFromFile は、YAML設定ファイルからの読み込みをテストします。
func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErr    bool
		wantDomain string
		wantToken  string
	}{
		{
			name: "正常なYAML設定ファイル",
			content: `duckdns:
  domain: "test-domain"
  token: "test-token"
update:
  interval: "5m"
ip_sources:
  - "https://api.ipify.org"
log:
  level: "info"
  format: "json"
`,
			wantErr:    false,
			wantDomain: "test-domain",
			wantToken:  "test-token",
		},
		{
			name: "最小限の設定",
			content: `duckdns:
  domain: "minimal"
  token: "token123"
update:
  interval: "1m"
ip_sources:
  - "https://ifconfig.me"
`,
			wantErr:    false,
			wantDomain: "minimal",
			wantToken:  "token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンポラリーファイルを作成
			tmpFile := t.TempDir() + "/config.yaml"
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("テンポラリーファイルの作成に失敗: %v", err)
			}

			// ファイルから読み込み
			cfg, err := LoadFromFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				if cfg.DuckDNS.Domain != tt.wantDomain {
					t.Errorf("ドメイン名が一致しません。期待: %s, 実際: %s", tt.wantDomain, cfg.DuckDNS.Domain)
				}
				if cfg.DuckDNS.Token != tt.wantToken {
					t.Errorf("トークンが一致しません。期待: %s, 実際: %s", tt.wantToken, cfg.DuckDNS.Token)
				}
			}
		})
	}
}

// TestLoadFromFile_FileNotFound は、ファイルが見つからない場合をテストします。
func TestLoadFromFile_FileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("エラーが返されるべきですが、nilが返されました")
	}
}

// TestLoadFromFile_InvalidYAML は、無効なYAMLをテストします。
func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpFile := t.TempDir() + "/invalid.yaml"
	invalidYAML := "{ invalid yaml content: [unclosed"
	if err := os.WriteFile(tmpFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("テンポラリーファイルの作成に失敗: %v", err)
	}

	_, err := LoadFromFile(tmpFile)
	if err == nil {
		t.Error("無効なYAMLでもエラーが返されるべき")
	}
}

// TestLoadFromEnv は、環境変数からの読み込みをテストします。
func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		setEnv       map[string]string
		wantDomain   string
		wantToken    string
		wantInterval time.Duration
		wantErr      bool
	}{
		{
			name: "すべての環境変数が設定されている",
			setEnv: map[string]string{
				"DUCKDNS_DOMAIN":   "env-domain",
				"DUCKDNS_TOKEN":    "env-token",
				"DUCKDNS_INTERVAL": "10m",
			},
			wantDomain:   "env-domain",
			wantToken:    "env-token",
			wantInterval: 10 * time.Minute,
			wantErr:      false,
		},
		{
			name: "部分的に環境変数が設定されている",
			setEnv: map[string]string{
				"DUCKDNS_DOMAIN": "partial-domain",
			},
			wantDomain:   "partial-domain",
			wantToken:    "",
			wantInterval: 0,
			wantErr:      false,
		},
		{
			name:         "環境変数が設定されていない",
			setEnv:       map[string]string{},
			wantDomain:   "",
			wantToken:    "",
			wantInterval: 0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存の環境変数を保存
			oldDomain := os.Getenv("DUCKDNS_DOMAIN")
			oldToken := os.Getenv("DUCKDNS_TOKEN")
			oldInterval := os.Getenv("DUCKDNS_INTERVAL")
			defer func() {
				os.Setenv("DUCKDNS_DOMAIN", oldDomain)
				os.Setenv("DUCKDNS_TOKEN", oldToken)
				os.Setenv("DUCKDNS_INTERVAL", oldInterval)
			}()

			// テスト環境変数をクリア
			os.Unsetenv("DUCKDNS_DOMAIN")
			os.Unsetenv("DUCKDNS_TOKEN")
			os.Unsetenv("DUCKDNS_INTERVAL")

			// テスト用の環境変数を設定
			for key, value := range tt.setEnv {
				os.Setenv(key, value)
			}

			// 環境変数から読み込み
			cfg, err := LoadFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				if cfg.DuckDNS.Domain != tt.wantDomain {
					t.Errorf("ドメイン名が一致しません。期待: %s, 実際: %s", tt.wantDomain, cfg.DuckDNS.Domain)
				}
				if cfg.DuckDNS.Token != tt.wantToken {
					t.Errorf("トークンが一致しません。期待: %s, 実際: %s", tt.wantToken, cfg.DuckDNS.Token)
				}
				if cfg.Update.Interval != tt.wantInterval {
					t.Errorf("更新間隔が一致しません。期待: %v, 実際: %v", tt.wantInterval, cfg.Update.Interval)
				}
			}
		})
	}
}

// TestLoadFromEnv_InvalidInterval は、無効な間隔フォーマットをテストします。
func TestLoadFromEnv_InvalidInterval(t *testing.T) {
	oldInterval := os.Getenv("DUCKDNS_INTERVAL")
	defer os.Setenv("DUCKDNS_INTERVAL", oldInterval)

	os.Setenv("DUCKDNS_INTERVAL", "invalid-duration")
	_, err := LoadFromEnv()
	if err == nil {
		t.Error("無効な期間フォーマットでもエラーが返されるべき")
	}
}

// TestLoad は、YAMLファイルと環境変数の組み合わせをテストします。
func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		envVars     map[string]string
		wantDomain  string
		wantToken   string
		wantErr     bool
	}{
		{
			name: "環境変数がYAMLを上書き",
			yamlContent: `duckdns:
  domain: "yaml-domain"
  token: "yaml-token"
update:
  interval: "5m"
ip_sources:
  - "https://api.ipify.org"
`,
			envVars: map[string]string{
				"DUCKDNS_DOMAIN": "env-domain",
			},
			wantDomain: "env-domain",
			wantToken:  "yaml-token",
			wantErr:    false,
		},
		{
			name: "YAMLのみが設定されている",
			yamlContent: `duckdns:
  domain: "yaml-only"
  token: "yaml-token-only"
update:
  interval: "3m"
ip_sources:
  - "https://ifconfig.me"
`,
			envVars:    map[string]string{},
			wantDomain: "yaml-only",
			wantToken:  "yaml-token-only",
			wantErr:    false,
		},
		{
			name:        "ファイルパスが空で環境変数が設定されている",
			yamlContent: "",
			envVars: map[string]string{
				"DUCKDNS_DOMAIN": "env-only",
				"DUCKDNS_TOKEN":  "env-token-only",
			},
			wantDomain: "env-only",
			wantToken:  "env-token-only",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存の環境変数を保存
			oldDomain := os.Getenv("DUCKDNS_DOMAIN")
			oldToken := os.Getenv("DUCKDNS_TOKEN")
			oldInterval := os.Getenv("DUCKDNS_INTERVAL")
			defer func() {
				os.Setenv("DUCKDNS_DOMAIN", oldDomain)
				os.Setenv("DUCKDNS_TOKEN", oldToken)
				os.Setenv("DUCKDNS_INTERVAL", oldInterval)
			}()

			// テスト環境変数をクリア
			os.Unsetenv("DUCKDNS_DOMAIN")
			os.Unsetenv("DUCKDNS_TOKEN")
			os.Unsetenv("DUCKDNS_INTERVAL")

			// テスト用の環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			var configPath string
			if tt.yamlContent != "" {
				tmpFile := t.TempDir() + "/config.yaml"
				if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
					t.Fatalf("テンポラリーファイルの作成に失敗: %v", err)
				}
				configPath = tmpFile
			}

			// Load関数を実行
			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				if cfg.DuckDNS.Domain != tt.wantDomain {
					t.Errorf("ドメイン名が一致しません。期待: %s, 実際: %s", tt.wantDomain, cfg.DuckDNS.Domain)
				}
				if cfg.DuckDNS.Token != tt.wantToken {
					t.Errorf("トークンが一致しません。期待: %s, 実際: %s", tt.wantToken, cfg.DuckDNS.Token)
				}
			}
		})
	}
}

// TestValidate_Success は、正常な設定のバリデーションをテストします。
func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		DuckDNS: DuckDNSConfig{
			Domain: "test-domain",
			Token:  "test-token",
		},
		Update: UpdateConfig{
			Interval: 5 * time.Minute,
		},
		IPSources: []string{
			"https://api.ipify.org",
			"https://ifconfig.me",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("バリデーションが失敗しました: %v", err)
	}
}

// TestValidate_MissingDomain は、ドメイン名が設定されていない場合をテストします。
func TestValidate_MissingDomain(t *testing.T) {
	cfg := &Config{
		DuckDNS: DuckDNSConfig{
			Token: "test-token",
		},
		Update: UpdateConfig{
			Interval: 5 * time.Minute,
		},
		IPSources: []string{"https://api.ipify.org"},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("ドメイン名が設定されていない場合、エラーが返されるべき")
	}
}

// TestValidate_MissingToken は、トークンが設定されていない場合をテストします。
func TestValidate_MissingToken(t *testing.T) {
	cfg := &Config{
		DuckDNS: DuckDNSConfig{
			Domain: "test-domain",
		},
		Update: UpdateConfig{
			Interval: 5 * time.Minute,
		},
		IPSources: []string{"https://api.ipify.org"},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("トークンが設定されていない場合、エラーが返されるべき")
	}
}

// TestValidate_InvalidInterval は、無効な更新間隔をテストします。
func TestValidate_InvalidInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantErr  bool
	}{
		{
			name:     "ゼロ間隔",
			interval: 0,
			wantErr:  true,
		},
		{
			name:     "負の間隔",
			interval: -1 * time.Minute,
			wantErr:  true,
		},
		{
			name:     "有効な間隔",
			interval: 5 * time.Minute,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DuckDNS: DuckDNSConfig{
					Domain: "test-domain",
					Token:  "test-token",
				},
				Update: UpdateConfig{
					Interval: tt.interval,
				},
				IPSources: []string{"https://api.ipify.org"},
				Log: LogConfig{
					Level:  "info",
					Format: "json",
				},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}
		})
	}
}

// TestValidate_NoIPSources は、IP取得ソースが設定されていない場合をテストします。
func TestValidate_NoIPSources(t *testing.T) {
	cfg := &Config{
		DuckDNS: DuckDNSConfig{
			Domain: "test-domain",
			Token:  "test-token",
		},
		Update: UpdateConfig{
			Interval: 5 * time.Minute,
		},
		IPSources: []string{},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("IP取得ソースが設定されていない場合、エラーが返されるべき")
	}
}

// TestValidate_InvalidURL は、無効なURLをテストします。
func TestValidate_InvalidURL(t *testing.T) {
	tests := []struct {
		name      string
		ipSources []string
		wantErr   bool
	}{
		{
			name:      "有効なHTTP URL",
			ipSources: []string{"http://api.example.com/ip"},
			wantErr:   false,
		},
		{
			name:      "有効なHTTPS URL",
			ipSources: []string{"https://api.example.com/ip"},
			wantErr:   false,
		},
		{
			name:      "スキームなしのURL",
			ipSources: []string{"api.example.com/ip"},
			wantErr:   true,
		},
		{
			name:      "無効なスキーム",
			ipSources: []string{"ftp://api.example.com/ip"},
			wantErr:   true,
		},
		{
			name:      "ホストなしのURL",
			ipSources: []string{"https://"},
			wantErr:   true,
		},
		{
			name:      "複数ソース（1つが無効）",
			ipSources: []string{"https://api.ipify.org", "invalid-url"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DuckDNS: DuckDNSConfig{
					Domain: "test-domain",
					Token:  "test-token",
				},
				Update: UpdateConfig{
					Interval: 5 * time.Minute,
				},
				IPSources: tt.ipSources,
				Log: LogConfig{
					Level:  "info",
					Format: "json",
				},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}
		})
	}
}

// TestValidate_InvalidLogLevel は、無効なログレベルをテストします。
func TestValidate_InvalidLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "有効: debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "有効: info",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "有効: warn",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "有効: error",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "無効: invalid",
			level:   "invalid",
			wantErr: true,
		},
		{
			name:    "空文字列（許可）",
			level:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DuckDNS: DuckDNSConfig{
					Domain: "test-domain",
					Token:  "test-token",
				},
				Update: UpdateConfig{
					Interval: 5 * time.Minute,
				},
				IPSources: []string{"https://api.ipify.org"},
				Log: LogConfig{
					Level:  tt.level,
					Format: "json",
				},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}
		})
	}
}

// TestValidate_InvalidLogFormat は、無効なログフォーマットをテストします。
func TestValidate_InvalidLogFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "有効: json",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "有効: text",
			format:  "text",
			wantErr: false,
		},
		{
			name:    "無効: invalid",
			format:  "invalid",
			wantErr: true,
		},
		{
			name:    "空文字列（許可）",
			format:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DuckDNS: DuckDNSConfig{
					Domain: "test-domain",
					Token:  "test-token",
				},
				Update: UpdateConfig{
					Interval: 5 * time.Minute,
				},
				IPSources: []string{"https://api.ipify.org"},
				Log: LogConfig{
					Level:  "info",
					Format: tt.format,
				},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("エラーが予期したのと異なります。期待: %v, 実際: %v", tt.wantErr, err)
			}
		})
	}
}

// TestValidationError_Error は、ValidationErrorのError()メソッドをテストします。
func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Errors: []string{
			"エラー1",
			"エラー2",
			"エラー3",
		},
	}

	errStr := ve.Error()
	if errStr == "" {
		t.Error("Error()メソッドが空の文字列を返しました")
	}

	// すべてのエラーメッセージが含まれているか確認
	for _, errMsg := range ve.Errors {
		if !stringContains(errStr, errMsg) {
			t.Errorf("エラーメッセージ '%s' が結果に含まれていません", errMsg)
		}
	}
}

// TestValidationError_Error_Empty は、空のエラーリストをテストします。
func TestValidationError_Error_Empty(t *testing.T) {
	ve := &ValidationError{
		Errors: []string{},
	}

	errStr := ve.Error()
	if errStr == "" {
		t.Error("Error()メソッドが空の文字列を返しました")
	}
}

// stringContains は文字列が含まれているかを確認します
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
