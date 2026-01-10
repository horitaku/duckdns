#!/bin/bash

# DuckDNS 自動更新プログラム - アンインストールスクリプト
# このスクリプトは、DuckDNS自動更新プログラムをシステムから削除するます。

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
trap 'log_error "アンインストール中にエラーが発生しました"; exit 1' ERR

# ========== パーミッション確認 ==========
if [[ $EUID -ne 0 ]]; then
    log_error "このスクリプトは root 権限で実行する必要があるます。"
    log_error "以下のコマンドで実行してください："
    log_error "  sudo $0"
    exit 1
fi

# ========== 変数定義 ==========
BINARY_NAME="duckdns"
INSTALL_BIN="/usr/local/bin/${BINARY_NAME}"
INSTALL_CONFIG="/etc/duckdns"
SERVICE_FILE="/etc/systemd/system/duckdns.service"
WORK_DIR="/var/lib/duckdns"

log_info "DuckDNS 自動更新プログラムをアンインストールするます..."
echo ""

# ========== 確認メッセージ ==========
log_warning "以下のファイルが削除されます:"
echo "  - バイナリ: ${INSTALL_BIN}"
echo "  - サービスファイル: ${SERVICE_FILE}"
echo ""
log_warning "以下は保持されます（手動削除が必要）:"
echo "  - 設定ファイル: ${INSTALL_CONFIG}/"
echo "  - ワーキングディレクトリ: ${WORK_DIR}/"
echo "  - ログファイル: /var/log/duckdns/"
echo ""

# ========== ユーザー確認 ==========
read -p "本当にアンインストールしますか？ (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "アンインストールをキャンセルしました。"
    exit 0
fi

# ========== 1. サービスを停止 ==========
log_info "ステップ 1: サービスを停止するます..."

if systemctl is-active --quiet duckdns.service 2>/dev/null; then
    systemctl stop duckdns.service
    log_success "サービスを停止しました"
else
    log_warning "サービスは起動していません"
fi

# ========== 2. サービスを無効化 ==========
log_info "ステップ 2: サービスを無効化するます..."

if systemctl is-enabled --quiet duckdns.service 2>/dev/null; then
    systemctl disable duckdns.service
    log_success "サービスを無効化しました"
else
    log_warning "サービスは無効化されています"
fi

# ========== 3. サービスファイルを削除 ==========
log_info "ステップ 3: サービスファイルを削除するます..."

if [[ -f "${SERVICE_FILE}" ]]; then
    rm -f "${SERVICE_FILE}"
    log_success "サービスファイルを削除しました: ${SERVICE_FILE}"
else
    log_warning "サービスファイルが見つかりません"
fi

# ========== 4. systemd デーモンをリロード ==========
log_info "ステップ 4: systemd デーモンをリロードするます..."
systemctl daemon-reload

log_success "systemd デーモンをリロードしました"

# ========== 5. バイナリを削除 ==========
log_info "ステップ 5: バイナリを削除するます..."

if [[ -f "${INSTALL_BIN}" ]]; then
    rm -f "${INSTALL_BIN}"
    log_success "バイナリを削除しました: ${INSTALL_BIN}"
else
    log_warning "バイナリが見つかりません"
fi

# ========== 6. 設定ファイルの確認 ==========
log_info "ステップ 6: 設定ファイルと関連ディレクトリの確認..."

if [[ -d "${INSTALL_CONFIG}" ]]; then
    log_warning "設定ファイルが残っています:"
    log_warning "  ディレクトリ: ${INSTALL_CONFIG}/"
    log_warning "  手動で削除する場合は以下のコマンドを実行してください:"
    log_warning "    sudo rm -rf ${INSTALL_CONFIG}"
fi

if [[ -d "${WORK_DIR}" ]]; then
    log_warning "ワーキングディレクトリが残っています:"
    log_warning "  ディレクトリ: ${WORK_DIR}/"
    log_warning "  必要に応じて手動で削除してください"
fi

if [[ -d "/var/log/duckdns" ]]; then
    log_warning "ログディレクトリが残っています:"
    log_warning "  ディレクトリ: /var/log/duckdns/"
    log_warning "  必要に応じて手動で削除してください"
fi

# ========== 完了メッセージ ==========
echo ""
log_success "=================================================================================="
log_success "DuckDNS 自動更新プログラムのアンインストールが完了しました！"
log_success "=================================================================================="
echo ""

echo "🗑️ 削除されたファイル:"
echo "  ✓ バイナリ: ${INSTALL_BIN}"
echo "  ✓ サービスファイル: ${SERVICE_FILE}"
echo ""

echo "⚠️ 手動削除が必要なファイル（オプション）:"
echo "  - 設定ファイル: ${INSTALL_CONFIG}/"
echo "    削除コマンド: sudo rm -rf ${INSTALL_CONFIG}"
echo ""
echo "  - ワーキングディレクトリ: ${WORK_DIR}/"
echo "    削除コマンド: sudo rm -rf ${WORK_DIR}"
echo ""
echo "  - ログディレクトリ: /var/log/duckdns/"
echo "    削除コマンド: sudo rm -rf /var/log/duckdns"
echo ""

log_success "アンインストール完了！"
