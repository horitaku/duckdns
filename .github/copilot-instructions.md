# DuckDNS 自動更新プログラム - Copilot 指示書

## プロジェクト概要

このプロジェクトは、Go言語で実装されたDuckDNS自動更新ツールです。
定期的にグローバルIPアドレスを取得し、DuckDNSのDNSレコードを自動的に更新します。

## 主要機能

1. **グローバルIP取得**: 外部サービスから現在のグローバルIPアドレスを取得
2. **DuckDNS更新**: DuckDNS APIを使用してDNSレコードを更新
3. **定期実行**: 設定された間隔で自動的にチェック・更新を実行
4. **ログ記録**: 更新履歴とエラーを適切にログ出力
5. **設定管理**: 環境変数またはYAML設定ファイルで設定を管理

## 技術スタック

- **言語**: Go 1.21+
- **主要ライブラリ**:
  - `net/http`: HTTPクライアント
  - `time`: スケジューリング
  - `log/slog`: 構造化ログ
  - `gopkg.in/yaml.v3`: 設定ファイル解析

## コーディング規約

### Go言語スタイル

- Go標準の命名規則に従う
- `gofmt`でフォーマット
- `golangci-lint`でリント
- エラーハンドリングを適切に実装
- context.Contextを使用した適切なキャンセル処理

### コメント

- 公開関数・型には必ずGoDocコメントを記述
- 複雑なロジックには日本語で説明コメントを追加
- TODOやFIXMEは明確に記載

### エラーハンドリング

```go
// 良い例
if err != nil {
    return fmt.Errorf("failed to fetch IP: %w", err)
}

// エラーをラップして詳細な情報を提供
```

## プロジェクト構成

```
duckdns/
├── cmd/
│   └── duckdns/
│       └── main.go          # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go        # 設定管理
│   ├── ip/
│   │   └── fetcher.go       # IP取得ロジック
│   ├── duckdns/
│   │   └── client.go        # DuckDNS APIクライアント
│   └── scheduler/
│       └── scheduler.go     # 定期実行ロジック
├── config.yaml              # 設定ファイル例
├── go.mod
├── go.sum
├── README.md
└── .github/
    └── copilot-instructions.md

```

## 設定ファイルフォーマット

```yaml
duckdns:
  domain: "your-domain"      # DuckDNSドメイン名
  token: "your-token"        # DuckDNSトークン
  
update:
  interval: "5m"             # 更新チェック間隔
  
ip_sources:
  - "https://api.ipify.org"
  - "https://ifconfig.me/ip"
  - "https://icanhazip.com"
  
log:
  level: "info"              # debug, info, warn, error
  format: "json"             # json, text
```

## セキュリティ要件

1. **トークン管理**
   - DuckDNSトークンは環境変数優先
   - 設定ファイルに含める場合は`.gitignore`に追加
   - 例: `DUCKDNS_TOKEN`環境変数

2. **HTTPSの使用**
   - すべての外部通信はHTTPS
   - 証明書検証を無効化しない

3. **タイムアウト設定**
   - HTTPリクエストには適切なタイムアウトを設定
   - デフォルト: 10秒

## エラー処理戦略

1. IP取得失敗時は複数のソースを順次試行
2. DuckDNS更新失敗時は指数バックオフでリトライ
3. 致命的エラーはログに記録してプログラム継続
4. 一時的エラーは警告レベルでログ出力

## ログ出力

```go
// 構造化ログの使用例
slog.Info("IP updated successfully",
    "domain", domain,
    "ip", newIP,
    "old_ip", oldIP,
)

slog.Error("failed to update DuckDNS",
    "error", err,
    "domain", domain,
)
```

## テスト要件

1. **ユニットテスト**
   - 各パッケージに`*_test.go`ファイルを作成
   - テストカバレッジ目標: 80%以上

2. **モック**
   - HTTPクライアントはインターフェース化してモック可能に
   - `httptest`パッケージを活用

3. **統合テスト**
   - 環境変数を使ったテスト設定
   - CI/CDでの自動実行

## ビルド・実行

```bash
# ビルド
go build -o duckdns ./cmd/duckdns

# 実行（環境変数使用）
export DUCKDNS_TOKEN="your-token"
export DUCKDNS_DOMAIN="your-domain"
./duckdns

# 実行（設定ファイル使用）
./duckdns -config config.yaml
```

## Docker対応

- Dockerfileを提供
- マルチステージビルドで最小イメージサイズ
- scratch または alpine ベース

## 参考リンク

- [DuckDNS API仕様](https://www.duckdns.org/spec.jsp)
- [Go言語仕様](https://go.dev/ref/spec)
- [Effective Go](https://go.dev/doc/effective_go)

## 開発の進め方

1. プロジェクト構造の作成
2. 設定管理の実装
3. IP取得機能の実装
4. DuckDNS更新機能の実装
5. スケジューラーの実装
6. ログ・エラーハンドリングの統合
7. テストの作成
8. Docker化
9. ドキュメント整備
