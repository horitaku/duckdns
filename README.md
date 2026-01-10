# DuckDNS 自動更新プログラム 🦆

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

DuckDNS の DNS レコードを自動的に更新する Go 言語製のプログラムです。定期的にグローバル IP アドレスを取得し、IP アドレスが変更された場合に自動的に DuckDNS を更新します。

## ✨ 機能

- 🌐 **グローバルIP自動取得**: 複数のIPアドレス取得サービスからフェイルオーバーで取得
- 🔄 **自動更新**: 設定した間隔で定期的にIPアドレスをチェック
- 🎯 **変更検知**: IPアドレスが変更された場合のみDuckDNSを更新
- 🔁 **リトライ機能**: 更新失敗時は指数バックオフでリトライ
- 📝 **構造化ログ**: JSON/テキスト形式の詳細なログ出力
- ⚙️ **柔軟な設定**: YAMLファイルまたは環境変数で設定可能
- 🛡️ **グレースフルシャットダウン**: SIGINT/SIGTERM シグナルに対応
- 🐧 **systemd対応**: systemdサービスとして常駐可能

## 📋 前提条件

- Go 1.21 以上（ビルドする場合）
- Linux システム（systemd対応の場合）
- DuckDNS アカウントとトークン

## 🚀 インストール

### 方法1: インストールスクリプトの使用（推奨）

```bash
# リポジトリをクローン
git clone https://github.com/horitaku/duckdns.git
cd duckdns

# 設定ファイルを作成
cp config.yaml.example config.yaml
nano config.yaml  # DuckDNSのドメインとトークンを設定

# インストールスクリプトを実行（要root権限）
sudo ./deploy/install.sh
```

このスクリプトは以下を自動的に実行します：
- プログラムのビルド
- バイナリを `/usr/local/bin/` にコピー
- 設定ファイルを `/etc/duckdns/` にコピー
- systemdサービスの登録と起動

### 方法2: 手動ビルドとインストール

```bash
# リポジトリをクローン
git clone https://github.com/horitaku/duckdns.git
cd duckdns

# ビルド
go build -o duckdns ./cmd/duckdns

# バイナリを配置
sudo cp duckdns /usr/local/bin/

# 設定ファイルを作成
mkdir -p /etc/duckdns
cp config.yaml.example /etc/duckdns/config.yaml
nano /etc/duckdns/config.yaml  # 設定を編集
```

## ⚙️ 設定

### 設定ファイル（config.yaml）

```yaml
# DuckDNS設定
duckdns:
  domain: "your-domain"      # DuckDNSドメイン名（.duckdns.orgは不要）
  token: "your-token"        # DuckDNSトークン

# 更新設定
update:
  interval: "5m"             # チェック間隔（例: 1m, 5m, 1h）

# IP取得ソース（フェイルオーバー対応）
ip_sources:
  - "https://api.ipify.org"
  - "https://ifconfig.me/ip"
  - "https://icanhazip.com"

# ログ設定
log:
  level: "info"              # ログレベル: debug, info, warn, error
  format: "json"             # ログ形式: json, text
```

### 環境変数

環境変数は設定ファイルよりも優先されます：

```bash
# 必須
export DUCKDNS_TOKEN="your-token"
export DUCKDNS_DOMAIN="your-domain"

# オプション
export DUCKDNS_INTERVAL="5m"
export DUCKDNS_LOG_LEVEL="info"
export DUCKDNS_LOG_FORMAT="json"
```

## 📖 使用方法

### 手動実行

```bash
# 設定ファイルを指定して実行
./duckdns -config config.yaml

# 環境変数で実行
export DUCKDNS_TOKEN="your-token"
export DUCKDNS_DOMAIN="your-domain"
./duckdns

# バージョン確認
./duckdns -version
```

### systemdサービスとして実行

```bash
# サービス起動
sudo systemctl start duckdns

# サービス停止
sudo systemctl stop duckdns

# サービス再起動
sudo systemctl restart duckdns

# サービスステータス確認
sudo systemctl status duckdns

# ログ確認
sudo journalctl -u duckdns -f

# 自動起動の有効化
sudo systemctl enable duckdns

# 自動起動の無効化
sudo systemctl disable duckdns
```

## 🔧 トラブルシューティング

### よくある問題と解決方法

#### 1. "validation error: domain is required" エラー

**原因**: DuckDNSドメインが設定されていません。

**解決方法**:
```bash
# 設定ファイルでドメインを設定
nano /etc/duckdns/config.yaml

# または環境変数で設定
export DUCKDNS_DOMAIN="your-domain"
```

#### 2. "validation error: token is required" エラー

**原因**: DuckDNSトークンが設定されていません。

**解決方法**:
```bash
# 設定ファイルでトークンを設定
nano /etc/duckdns/config.yaml

# または環境変数で設定（推奨）
export DUCKDNS_TOKEN="your-token"
```

#### 3. "DuckDNS update failed: KO" エラー

**原因**: DuckDNS APIが更新を拒否しました。

**解決方法**:
- ドメイン名が正しいか確認（`.duckdns.org` は不要）
- トークンが正しいか確認
- DuckDNS の管理画面でドメインが有効か確認

#### 4. サービスが起動しない

**原因**: 設定ファイルの権限または内容に問題があります。

**解決方法**:
```bash
# ログを確認
sudo journalctl -u duckdns -n 50

# 設定ファイルの権限確認
ls -la /etc/duckdns/config.yaml

# 設定ファイルの内容確認
sudo /usr/local/bin/duckdns -config /etc/duckdns/config.yaml
```

#### 5. IP取得に失敗する

**原因**: すべてのIP取得ソースにアクセスできません。

**解決方法**:
- ネットワーク接続を確認
- ファイアウォール設定を確認
- プロキシ設定が必要な場合は環境変数を設定

### ログの確認方法

#### systemdサービスのログ

```bash
# リアルタイムでログを表示
sudo journalctl -u duckdns -f

# 最新50行を表示
sudo journalctl -u duckdns -n 50

# 特定の期間のログを表示
sudo journalctl -u duckdns --since "2026-01-11 00:00:00"

# エラーレベルのログのみ表示
sudo journalctl -u duckdns -p err
```

#### 手動実行時のログ

```bash
# デバッグログを有効にして実行
export DUCKDNS_LOG_LEVEL="debug"
./duckdns -config config.yaml

# ログをファイルに出力
./duckdns -config config.yaml 2>&1 | tee duckdns.log
```

## 🗑️ アンインストール

```bash
# アンインストールスクリプトを実行
sudo ./deploy/uninstall.sh
```

手動でアンインストールする場合：

```bash
# サービス停止と無効化
sudo systemctl stop duckdns
sudo systemctl disable duckdns

# ファイル削除
sudo rm /etc/systemd/system/duckdns.service
sudo rm /usr/local/bin/duckdns
sudo rm -rf /etc/duckdns

# systemd設定再読み込み
sudo systemctl daemon-reload
```

## 🧪 開発

### テストの実行

```bash
# すべてのテストを実行
go test ./...

# カバレッジ付きでテスト
go test -cover ./...

# 詳細出力
go test -v ./...
```

### ビルド

```bash
# 開発用ビルド
go build -o duckdns ./cmd/duckdns

# リリース用ビルド（最適化）
go build -ldflags="-s -w" -o duckdns ./cmd/duckdns
```

## 📄 ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🔗 参考リンク

- [DuckDNS 公式サイト](https://www.duckdns.org/)
- [DuckDNS API仕様](https://www.duckdns.org/spec.jsp)
- [Go言語公式サイト](https://go.dev/)

## 🙏 謝辞

このプロジェクトは以下のサービスを利用しています：
- [DuckDNS](https://www.duckdns.org/) - 無料のダイナミックDNSサービス
- [ipify](https://www.ipify.org/) - IP取得API
- [ifconfig.me](https://ifconfig.me/) - IP取得API
- [icanhazip.com](https://icanhazip.com/) - IP取得API