#!/bin/bash

# DuckDNS 自動更新プログラム - インストールスクリプト
# このスクリプトは、DuckDNS自動更新プログラムをシステムにインストールし、
# systemd サービスとして有効化・起動するます。

set -e  # エラー時に即座に終了するます

# ========== カラー出力の定義 ==========
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # Color reset

# ========== ログ関数 ==========
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# ========== エラーハンドリング ==========
trap 'log_error "インストール中にエラーが発生しました"; exit 1' ERR

# ========== パーミッション確認 ==========
if [[ $EUID -ne 0 ]]; then
    log_error "このスクリプトは root 権限で実行する必要があるます。"
    log_error "以下のコマンドで実行してください："
    log_error "  sudo $0"
    exit 1
fi

# ========== 変数定義 ==========
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_NAME="duckdns"
BINARY_PATH="${PROJECT_ROOT}/${BINARY_NAME}"
INSTALL_BIN="/usr/local/bin/${BINARY_NAME}"
INSTALL_CONFIG="/etc/duckdns"
CONFIG_FILE="${INSTALL_CONFIG}/config.yaml"
SERVICE_FILE="/etc/systemd/system/duckdns.service"
WORK_DIR="/var/lib/duckdns"
LOG_DIR="/var/log/duckdns"

log_info "DuckDNS 自動更新プログラムをインストールするます..."
log_info "プロジェクトディレクトリ: ${PROJECT_ROOT}"

# ========== 1. ビルド処理 ==========
log_info "ステップ 1: バイナリをビルドするます..."
cd "${PROJECT_ROOT}"

if [[ ! -f "go.mod" ]]; then
    log_error "go.mod が見つかりません。プロジェクトディレクトリを確認してください。"
    exit 1
fi

log_info "ビルド中... (これには少し時間がかかるかもしれません)"
go build -o "${BINARY_NAME}" ./cmd/duckdns

if [[ ! -f "${BINARY_PATH}" ]]; then
    log_error "ビルドに失敗しました。"
    exit 1
fi

log_success "ビルド完了: ${BINARY_PATH}"

# ========== 2. バイナリのインストール ==========
log_info "ステップ 2: バイナリを ${INSTALL_BIN} にインストールするます..."
mkdir -p "$(dirname "${INSTALL_BIN}")"
cp "${BINARY_PATH}" "${INSTALL_BIN}"
chmod +x "${INSTALL_BIN}"

# バイナリが実行可能か確認
if ! "${INSTALL_BIN}" -version &>/dev/null; then
    log_warning "バイナリの実行テストに失敗しました。"
fi

log_success "バイナリをインストール完了: ${INSTALL_BIN}"

# ========== 3. 設定ファイルのディレクトリ作成 ==========
log_info "ステップ 3: 設定ファイルのディレクトリを作成するます..."
mkdir -p "${INSTALL_CONFIG}"
mkdir -p "${WORK_DIR}"
mkdir -p "${LOG_DIR}"
chmod 755 "${INSTALL_CONFIG}"
chmod 755 "${WORK_DIR}"
chmod 755 "${LOG_DIR}"

log_success "ディレクトリを作成完了"

# ========== 4. 設定ファイルのインストール ==========
log_info "ステップ 4: 設定ファイルをインストールするます..."

if [[ ! -f "config.yaml.example" ]]; then
    log_warning "config.yaml.example が見つかりません。スキップするます。"
else
    if [[ ! -f "${CONFIG_FILE}" ]]; then
        cp "config.yaml.example" "${CONFIG_FILE}"
        chmod 600 "${CONFIG_FILE}"
        log_success "設定ファイルを作成完了: ${CONFIG_FILE}"
        log_warning "設定ファイルを編集してください:"
        log_warning "  $EDITOR ${CONFIG_FILE}"
    else
        log_warning "設定ファイル ${CONFIG_FILE} は既に存在するます。スキップするます。"
        log_warning "更新が必要な場合は、手動で編集してください。"
    fi
fi

# ========== 5. systemd サービスファイルのインストール ==========
log_info "ステップ 5: systemd サービスファイルをインストールするます..."

if [[ ! -f "${PROJECT_ROOT}/deploy/duckdns.service" ]]; then
    log_error "duckdns.service が見つかりません。"
    exit 1
fi

cp "${PROJECT_ROOT}/deploy/duckdns.service" "${SERVICE_FILE}"
chmod 644 "${SERVICE_FILE}"

log_success "サービスファイルをインストール完了: ${SERVICE_FILE}"

# ========== 6. systemd デーモンをリロード ==========
log_info "ステップ 6: systemd デーモンをリロードするます..."
systemctl daemon-reload

log_success "systemd デーモンをリロード完了"

# ========== 7. サービスを有効化 ==========
log_info "ステップ 7: サービスを有効化するます..."
systemctl enable duckdns.service

log_success "サービスを有効化完了"

# ========== 8. サービスを開始 ==========
log_info "ステップ 8: サービスを開始するます..."
systemctl start duckdns.service

# サービスの状態確認
if systemctl is-active --quiet duckdns.service; then
    log_success "サービスが起動しました！"
else
    log_warning "サービスの起動に失敗しました。ログを確認してください："
    log_warning "  journalctl -u duckdns.service -n 20"
    exit 1
fi

# ========== 完了メッセージ ==========
echo ""
log_success "=================================================================================="
log_success "DuckDNS 自動更新プログラムのインストールが完了しました！ わくわく! ✨"
log_success "=================================================================================="
echo ""

echo "📋 インストール情報:"
echo "  バイナリ: ${INSTALL_BIN}"
echo "  設定ファイル: ${CONFIG_FILE}"
echo "  サービスファイル: ${SERVICE_FILE}"
echo "  ワーキングディレクトリ: ${WORK_DIR}"
echo "  ログディレクトリ: ${LOG_DIR}"
echo ""

echo "🚀 サービス管理コマンド:"
echo "  状態確認: systemctl status duckdns.service"
echo "  ログ確認: journalctl -u duckdns.service -f"
echo "  再起動: systemctl restart duckdns.service"
echo "  停止: systemctl stop duckdns.service"
echo ""

echo "⚙️ 設定方法:"
if [[ ! -f "${CONFIG_FILE}" ]]; then
    echo "  設定ファイルを作成し、DuckDNS のトークンと ドメイン名を設定してください:"
    echo "  $EDITOR ${CONFIG_FILE}"
else
    echo "  設定ファイルが既に存在するます。必要に応じて編集してください:"
    echo "  $EDITOR ${CONFIG_FILE}"
    echo "  編集後、サービスを再起動してください:"
    echo "  systemctl restart duckdns.service"
fi
echo ""

echo "📖 詳細は以下のコマンドでログを確認できます:"
echo "  journalctl -u duckdns.service -n 50"
echo ""

log_success "インストール完了！"
