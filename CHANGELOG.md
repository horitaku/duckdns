# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-01-11

### 🎉 初回リリース

DuckDNS 自動更新プログラムの最初のリリースです！

### ✨ 追加機能

#### コア機能
- **グローバルIP自動取得**: 複数のIPアドレス取得サービスからフェイルオーバーで取得
  - デフォルトで3つのソース（api.ipify.org、ifconfig.me、icanhazip.com）をサポート
  - タイムアウト設定（デフォルト10秒）
  - IPv4アドレスの自動検証

- **DuckDNS自動更新**: IPアドレス変更時に自動的にDNSレコードを更新
  - IP変更検知機能（変更がない場合は更新をスキップ）
  - 指数バックオフによるリトライ機能（最大3回、1秒/2秒/4秒）
  - context.Context による適切なキャンセル処理

- **定期実行スケジューラー**: 設定した間隔で自動的にチェック・更新を実行
  - 柔軟な間隔設定（例: 5m, 1h, 30s）
  - グレースフルシャットダウン対応（SIGINT/SIGTERM）

#### 設定管理
- **YAMLファイル**: `config.yaml` で詳細な設定が可能
- **環境変数サポート**: 環境変数が設定ファイルより優先
  - `DUCKDNS_TOKEN`: DuckDNSトークン
  - `DUCKDNS_DOMAIN`: DuckDNSドメイン名
  - `DUCKDNS_INTERVAL`: 更新チェック間隔
  - `DUCKDNS_LOG_LEVEL`: ログレベル
  - `DUCKDNS_LOG_FORMAT`: ログ形式
- **バリデーション**: 起動時に設定値を検証し、わかりやすいエラーメッセージを表示

#### ログ機能
- **構造化ログ**: log/slog を使用した構造化ログ出力
- **ログレベル**: debug, info, warn, error の4段階
- **ログ形式**: JSON またはテキスト形式を選択可能
- **ソース情報**: ファイル名・行番号を含む詳細なログ

#### デプロイ
- **systemd対応**: Linuxシステムでサービスとして常駐可能
- **インストールスクリプト**: `deploy/install.sh` で簡単インストール
- **アンインストールスクリプト**: `deploy/uninstall.sh` で完全削除

### 🧪 テスト
- **ユニットテスト**: 全パッケージに対する包括的なテスト
  - config パッケージ: 91.0% カバレッジ
  - duckdns パッケージ: 92.1% カバレッジ
  - ip パッケージ: 94.3% カバレッジ
  - logger パッケージ: 100% カバレッジ
  - scheduler パッケージ: 85.2% カバレッジ
- **統合テスト**: 実際の動作環境でのテスト実施

### 📚 ドキュメント
- **README.md**: 詳細なインストール手順、使用方法、トラブルシューティング
- **GoDocコメント**: すべての公開関数とパッケージにドキュメントコメント
- **設定ファイルサンプル**: `config.yaml.example` で設定例を提供

### 🏗️ アーキテクチャ
- **パッケージ構成**:
  - `cmd/duckdns`: メインプログラム
  - `internal/config`: 設定管理
  - `internal/ip`: IP取得機能
  - `internal/duckdns`: DuckDNS APIクライアント
  - `internal/scheduler`: 定期実行スケジューラー
  - `internal/logger`: ログ管理

### 🔧 技術スタック
- **言語**: Go 1.21+（実際のビルド: Go 1.25.5）
- **標準ライブラリ**: context, net/http, log/slog, os/signal など
- **外部依存**: gopkg.in/yaml.v3（YAML解析のみ）

### 📝 既知の問題

#### 制限事項
1. **IPv6非対応**: 現在はIPv4アドレスのみサポート
   - 将来のバージョンでIPv6対応を検討

2. **環境変数のみでの起動制限**: IP取得ソースは設定ファイルで指定が必要
   - 環境変数 `DUCKDNS_DOMAIN` と `DUCKDNS_TOKEN` だけでは起動不可
   - 最小限の設定ファイル（ip_sourcesのみ）が必要

3. **systemd専用**: 現在はsystemdを持つLinuxシステムのみサポート
   - Docker対応は将来のバージョンで検討
   - Windows/macOSでは手動実行のみ

#### 回避策が必要なケース
- **DuckDNSトークンの保護**: 設定ファイルにトークンを記載する場合は適切な権限設定が必要
  - 推奨: 環境変数 `DUCKDNS_TOKEN` の使用
  - ファイル権限: `chmod 600 /etc/duckdns/config.yaml`

### 🙏 謝辞
このプロジェクトは以下のサービスを利用しています：
- [DuckDNS](https://www.duckdns.org/) - 無料のダイナミックDNSサービス
- [ipify](https://www.ipify.org/) - IP取得API
- [ifconfig.me](https://ifconfig.me/) - IP取得API
- [icanhazip.com](https://icanhazip.com/) - IP取得API

---

[Unreleased]: https://github.com/horitaku/duckdns/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/horitaku/duckdns/releases/tag/v1.0.0
