#!/bin/bash

# ROI Agent 統合テストスクリプト
# すべてのテスト機能を統合した包括的なテストツール

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# カラー出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ログ関数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# ヘルプ表示
show_help() {
    echo "ROI Agent 統合テストスクリプト"
    echo ""
    echo "使用方法:"
    echo "  $0 [コマンド] [オプション]"
    echo ""
    echo "コマンド:"
    echo "  all              - 全テストを実行"
    echo "  env              - 環境変数とファイル設定のテスト"
    echo "  build            - ビルドテスト"
    echo "  data-sender      - データ送信機能のテスト"
    echo "  permissions      - macOS権限の確認"
    echo "  web              - Web UIの動作テスト"
    echo "  status           - 現在の動作状況確認"
    echo "  clean            - 古いファイルのクリーンアップ"
    echo "  help             - このヘルプを表示"
    echo ""
    echo "例:"
    echo "  $0 all           # 全テスト実行"
    echo "  $0 data-sender   # データ送信テストのみ"
    echo "  $0 clean         # クリーンアップのみ"
}

# 前提条件チェック
check_prerequisites() {
    log_info "前提条件をチェック中..."
    
    # Goインストール確認
    if ! command -v go &> /dev/null; then
        log_error "Go がインストールされていません"
        return 1
    fi
    log_success "Go $(go version | cut -d' ' -f3) が見つかりました"
    
    # Python3インストール確認
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 がインストールされていません"
        return 1
    fi
    log_success "Python3 $(python3 --version | cut -d' ' -f2) が見つかりました"
    
    # 必要なディレクトリの確認
    for dir in "agent" "data-sender" "web"; do
        if [ ! -d "$PROJECT_ROOT/$dir" ]; then
            log_error "必要なディレクトリが見つかりません: $dir"
            return 1
        fi
    done
    log_success "必要なディレクトリが全て存在します"
    
    return 0
}

# 環境変数テスト
test_environment() {
    log_info "環境変数設定をテスト中..."
    
    ENV_FILE="$PROJECT_ROOT/data-sender/.env"
    
    if [ -f "$ENV_FILE" ]; then
        log_success ".env ファイルが見つかりました: $ENV_FILE"
        
        echo "📄 .env ファイル内容:"
        cat "$ENV_FILE" | sed 's/API_KEY=.*/API_KEY=***隠匿***/'
        echo ""
        
        # .envファイルから読み込み
        source "$ENV_FILE"
        
        echo "🔍 読み込まれた環境変数:"
        echo "  ROI_AGENT_BASE_URL: ${ROI_AGENT_BASE_URL:-未設定}"
        echo "  ROI_AGENT_API_KEY: ${ROI_AGENT_API_KEY:0:8}***"
        echo "  ROI_AGENT_INTERVAL_MINUTES: ${ROI_AGENT_INTERVAL_MINUTES:-未設定}"
        
        if [ -n "$ROI_AGENT_BASE_URL" ] && [ -n "$ROI_AGENT_API_KEY" ]; then
            log_success "データ送信設定が正常に設定されています"
        else
            log_warning "データ送信設定が不完全です"
        fi
    else
        log_warning ".env ファイルが見つかりません: $ENV_FILE"
        log_info "データ送信は無効になります"
    fi
}

# ビルドテスト
test_build() {
    log_info "ビルドテストを実行中..."
    
    # Data Senderのビルド
    log_info "Data Sender をビルド中..."
    cd "$PROJECT_ROOT/data-sender"
    
    if go mod tidy && go build -o test-data-sender . ; then
        log_success "Data Sender のビルドが成功しました"
        rm -f test-data-sender
    else
        log_error "Data Sender のビルドに失敗しました"
        return 1
    fi
    
    # Agentのビルド（コンパイル確認のみ）
    log_info "Agent をビルド中..."
    cd "$PROJECT_ROOT/agent"
    
    if go mod tidy && go build -o test-agent . ; then
        log_success "Agent のビルドが成功しました"
        rm -f test-agent
    else
        log_error "Agent のビルドに失敗しました"
        return 1
    fi
    
    return 0
}

# データ送信テスト
test_data_sender() {
    log_info "データ送信機能をテスト中..."
    
    cd "$PROJECT_ROOT/data-sender"
    
    # ビルド
    if ! go build -o test-data-sender . ; then
        log_error "Data Sender のビルドに失敗しました"
        return 1
    fi
    
    # 設定確認
    log_info "設定を確認中..."
    ./test-data-sender status
    echo ""
    
    # 接続テスト
    log_info "サーバー接続をテスト中..."
    if ./test-data-sender test; then
        log_success "データ送信テストが成功しました"
    else
        log_warning "データ送信テストに失敗しました（設定未完了の可能性）"
    fi
    
    # クリーンアップ
    rm -f test-data-sender
    return 0
}

# macOS権限チェック
test_permissions() {
    log_info "macOS権限をチェック中..."
    
    # Accessibility権限チェック
    cd "$PROJECT_ROOT/agent"
    if go run main.go check-permissions 2>/dev/null | grep -q "OK"; then
        log_success "Accessibility権限が許可されています"
    else
        log_warning "Accessibility権限が必要です"
        log_info "システム設定 > プライバシーとセキュリティ > アクセシビリティ でターミナルを許可してください"
    fi
    
    # sudo権限チェック
    if sudo -n tcpdump --version &> /dev/null; then
        log_success "sudo権限が利用可能です"
    else
        log_warning "DNS監視にはsudo権限が必要です"
    fi
}

# Web UIテスト
test_web_ui() {
    log_info "Web UI をテスト中..."
    
    cd "$PROJECT_ROOT/web"
    
    # Python依存関係チェック
    if python3 -c "import flask" 2>/dev/null; then
        log_success "Flask がインストールされています"
    else
        log_warning "Flask がインストールされていません"
        log_info "pip3 install -r requirements.txt を実行してください"
        return 1
    fi
    
    # Web UI起動テスト（短時間）
    log_info "Web UI の起動テスト中..."
    timeout 5 python3 enhanced_app.py &> /dev/null &
    WEBUI_PID=$!
    sleep 2
    
    if kill -0 $WEBUI_PID 2>/dev/null; then
        log_success "Web UI が正常に起動しました"
        kill $WEBUI_PID 2>/dev/null || true
    else
        log_warning "Web UI の起動に問題があります"
    fi
}

# 動作状況確認
check_status() {
    log_info "現在の動作状況を確認中..."
    
    # プロセス確認
    if pgrep -f "main.go" > /dev/null; then
        log_success "Agent プロセスが動作中です"
    else
        log_info "Agent プロセスは停止中です"
    fi
    
    if pgrep -f "enhanced_app.py" > /dev/null; then
        log_success "Web UI プロセスが動作中です"
    else
        log_info "Web UI プロセスは停止中です"
    fi
    
    if pgrep -f "tcpdump.*port 53" > /dev/null; then
        log_success "DNS監視プロセスが動作中です"
    else
        log_info "DNS監視プロセスは停止中です"
    fi
    
    # ログファイル確認
    LOG_DIR="$HOME/.roiagent/logs"
    if [ -d "$LOG_DIR" ]; then
        log_success "ログディレクトリが存在します: $LOG_DIR"
        
        for logfile in "agent.log" "webui.log"; do
            if [ -f "$LOG_DIR/$logfile" ]; then
                size=$(du -h "$LOG_DIR/$logfile" | cut -f1)
                log_info "  $logfile: $size"
            fi
        done
    else
        log_info "ログディレクトリが作成されていません"
    fi
    
    # データディレクトリ確認
    DATA_DIR="$HOME/.roiagent/data"
    if [ -d "$DATA_DIR" ]; then
        file_count=$(ls -1 "$DATA_DIR" | wc -l)
        log_success "データディレクトリ: $file_count ファイル"
    else
        log_info "データディレクトリが作成されていません"
    fi
}

# クリーンアップ
cleanup_files() {
    log_info "不要なファイルをクリーンアップ中..."
    
    # 古いログファイル削除（7日以上前）
    find "$HOME/.roiagent" -name "*.log" -type f -mtime +7 -delete 2>/dev/null || true
    find "$HOME/.roiagent" -name "combined_*.json" -type f -mtime +7 -delete 2>/dev/null || true
    find "$HOME/.roiagent" -name "transmission_*.json" -type f -mtime +7 -delete 2>/dev/null || true
    
    # 一時ファイル削除
    find "$PROJECT_ROOT" -name "test-*" -type f -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.tmp" -type f -delete 2>/dev/null || true
    
    log_success "クリーンアップが完了しました"
}

# 全テスト実行
run_all_tests() {
    echo "========================================"
    echo "    ROI Agent 統合テスト開始"
    echo "========================================"
    echo ""
    
    if ! check_prerequisites; then
        log_error "前提条件チェックに失敗しました"
        exit 1
    fi
    echo ""
    
    test_environment
    echo ""
    
    test_build
    echo ""
    
    test_permissions
    echo ""
    
    test_data_sender  
    echo ""
    
    test_web_ui
    echo ""
    
    check_status
    echo ""
    
    echo "========================================"
    echo "    ROI Agent 統合テスト完了"
    echo "========================================"
}

# メイン処理
main() {
    case "${1:-help}" in
        "all")
            run_all_tests
            ;;
        "env"|"environment")
            test_environment
            ;;
        "build")
            check_prerequisites && test_build
            ;;
        "data-sender")
            test_data_sender
            ;;
        "permissions")
            test_permissions
            ;;
        "web")
            test_web_ui
            ;;
        "status")
            check_status
            ;;
        "clean")
            cleanup_files
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            log_error "不明なコマンド: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# スクリプト実行
main "$@"
