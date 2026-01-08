# duckdns

DuckDNS更新ツール (DuckDNS Update Tool)

## 概要 (Overview)

このツールは、[DuckDNS](https://www.duckdns.org/)のDNSレコードを更新するためのシンプルなコマンドラインツールです。

This is a simple command-line tool for updating [DuckDNS](https://www.duckdns.org/) DNS records.

## インストール (Installation)

```bash
go install github.com/horitaku/duckdns@latest
```

または、リポジトリをクローンしてビルド:

Or clone the repository and build:

```bash
git clone https://github.com/horitaku/duckdns.git
cd duckdns
go build
```

## 使い方 (Usage)

### コマンドラインオプション (Command-line options)

```bash
# ドメインとトークンを指定
./duckdns -domain=mydomain -token=mytoken

# IPアドレスを明示的に指定
./duckdns -domain=mydomain -token=mytoken -ip=1.2.3.4

# 詳細出力を有効化
./duckdns -domain=mydomain -token=mytoken -verbose
```

### 環境変数 (Environment variables)

```bash
export DUCKDNS_DOMAIN=mydomain
export DUCKDNS_TOKEN=mytoken
./duckdns
```

## オプション (Options)

- `-domain`: DuckDNSドメイン (必須) / DuckDNS domain (required)
- `-token`: DuckDNSトークン (必須) / DuckDNS token (required)
- `-ip`: 設定するIPアドレス (オプション、未指定の場合は自動検出) / IP address to set (optional, auto-detected if not specified)
- `-verbose`: 詳細出力を有効化 / Enable verbose output

## DuckDNSについて (About DuckDNS)

DuckDNSは無料のダイナミックDNSサービスです。詳細は[DuckDNS公式サイト](https://www.duckdns.org/)をご覧ください。

DuckDNS is a free dynamic DNS service. For more information, visit the [DuckDNS official site](https://www.duckdns.org/).

## ライセンス (License)

MIT