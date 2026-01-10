# コントリビューションガイド

DuckDNS 自動更新プログラムへの貢献を検討いただきありがとうございます！

## 📋 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [開発の流れ](#開発の流れ)
- [テストの実行](#テストの実行)
- [コーディング規約](#コーディング規約)
- [リリースプロセス](#リリースプロセス)

## 🔧 開発環境のセットアップ

### 前提条件

- Go 1.21 以上
- Git
- make（オプション）

### セットアップ手順

1. **リポジトリのフォーク**

   GitHubでこのリポジトリをフォークします。

2. **クローン**

   ```bash
   git clone https://github.com/YOUR_USERNAME/duckdns.git
   cd duckdns
   ```

3. **依存関係のインストール**

   ```bash
   go mod download
   go mod verify
   ```

4. **ビルド確認**

   ```bash
   go build -o duckdns ./cmd/duckdns
   ./duckdns -version
   ```

## 🔄 開発の流れ

### 1. ブランチの作成

機能追加やバグ修正用のブランチを作成します：

```bash
# 機能追加
git checkout -b feature/your-feature-name

# バグ修正
git checkout -b fix/issue-number-description
```

### 2. 変更の実施

コードを変更し、適切なテストを追加します。

### 3. テストの実行

```bash
# すべてのテストを実行
go test ./...

# カバレッジ付きでテスト
go test -cover ./...

# 詳細出力
go test -v ./...
```

### 4. コミット

わかりやすいコミットメッセージを書きます：

```bash
git add .
git commit -m "feat: 新機能の説明"
```

#### コミットメッセージのプレフィックス

- `feat:` - 新機能
- `fix:` - バグ修正
- `docs:` - ドキュメントのみの変更
- `style:` - コードの意味に影響しない変更（フォーマットなど）
- `refactor:` - バグ修正でも機能追加でもないコードの変更
- `perf:` - パフォーマンス改善
- `test:` - テストの追加・修正
- `chore:` - ビルドプロセスやツールの変更

### 5. プッシュとPull Request

```bash
git push origin feature/your-feature-name
```

GitHubでPull Requestを作成します。

## 🧪 テストの実行

### ユニットテスト

```bash
# すべてのテストを実行
go test ./...

# 特定のパッケージをテスト
go test ./internal/config/...

# 詳細出力
go test -v ./...

# カバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### レースコンディションの検出

```bash
go test -race ./...
```

### ベンチマーク

```bash
go test -bench=. ./...
```

## 📝 コーディング規約

### Go言語の標準規約に従う

- `gofmt` でコードをフォーマット
- `golangci-lint` でリント
- エラーは適切にハンドリング
- 公開関数にはGoDocコメントを記載

### コード例

```go
// Fetch は指定されたコンテキストでIPアドレスを取得します。
//
// Parameters:
//   - ctx: キャンセルやタイムアウトを制御するコンテキスト
//
// Returns:
//   - string: 取得したIPアドレス（IPv4形式）
//   - error: エラーが発生した場合
func (f *Fetcher) Fetch(ctx context.Context) (string, error) {
    // 実装
}
```

### リント

```bash
# golangci-lint のインストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# リントの実行
golangci-lint run
```

## 🚀 リリースプロセス

### バージョン管理

このプロジェクトは [Semantic Versioning](https://semver.org/) に従います：

- `MAJOR.MINOR.PATCH` (例: 1.2.3)
- `MAJOR`: 互換性のない変更
- `MINOR`: 後方互換性のある機能追加
- `PATCH`: 後方互換性のあるバグ修正

### リリース手順

メンテナー向けのリリース手順：

#### 1. CHANGELOG.md の更新

```markdown
## [1.1.0] - 2026-01-15

### Added
- 新機能の説明

### Fixed
- バグ修正の説明

### Changed
- 変更内容の説明
```

#### 2. バージョンタグの作成

```bash
# 最新の main ブランチに移動
git checkout main
git pull origin main

# バージョンタグを作成
git tag -a v1.1.0 -m "Release v1.1.0"

# タグをプッシュ
git push origin v1.1.0
```

#### 3. 自動リリース

タグをプッシュすると、GitHub Actions が自動的に以下を実行します：

1. テストの実行
2. 複数プラットフォーム向けのビルド
3. GitHub Releases へのアップロード
4. リリースノートの生成

#### 4. リリース確認

[GitHub Releases](https://github.com/horitaku/duckdns/releases) でリリースが正常に作成されたことを確認します。

### ローカルでのリリーステスト

GoReleaser をローカルでテストできます：

```bash
# GoReleaser のインストール
go install github.com/goreleaser/goreleaser@latest

# スナップショットビルド（タグなし）
goreleaser build --snapshot --clean

# 生成された成果物を確認
ls -la dist/
```

## 🐛 バグ報告

バグを見つけた場合は、以下の情報を含めて Issue を作成してください：

- **バグの説明**: 何が起こったか
- **再現手順**: バグを再現する方法
- **期待される動作**: 本来どうあるべきか
- **環境情報**:
  - OS: (例: Ubuntu 22.04)
  - Go バージョン: (例: 1.23)
  - プログラムバージョン: (例: v1.0.0)
- **ログ出力**: エラーメッセージやログ

## 💡 機能提案

新機能の提案は歓迎します！Issue を作成して以下を含めてください：

- **機能の説明**: 何を追加したいか
- **ユースケース**: なぜ必要か
- **実装案**: どのように実装するか（オプション）

## 📞 質問・サポート

質問がある場合は、GitHub Discussions または Issue で気軽に聞いてください。

## 🙏 謝辞

貢献してくださるすべての方に感謝します！

---

ハッピーコーディング！🎉
