# ROI Agent - macOS App & Network Monitor

macOS用のリアルタイムアプリケーション使用時間とネットワーク通信監視ツール

![ROI Agent](public/icon.png)

## 📋 Features

- **アプリケーション監視**: リアルタイムでアプリの使用時間、フォーカス時間を追跡
- **ネットワーク監視**: DNS監視によるWebサイトアクセス履歴
- **リアルタイムWeb UI**: 直感的なダッシュボード（ポート5002）
- **データ送信**: サーバーへの自動送信（オプション、間隔設定可能）
- **データ保存**: 日別でのJSONデータ保存

## 🚀 Quick Start

### 1. 権限設定
**必須**: macOSのAccessibility権限を有効にしてください
1. システム設定 > プライバシーとセキュリティ > アクセシビリティ
2. ターミナルまたはVS Codeを追加

### 2. 起動
```bash
# リポジトリをクローン
git clone <repository-url>
cd roi-agent

# 実行権限を付与
chmod +x scripts/start_enhanced_fqdn_monitoring.sh

# 監視開始
./scripts/start_enhanced_fqdn_monitoring.sh
```

起動時にデータ送信の設定を選択できます：
- **データ送信有効**: サーバーURL/APIキーを入力
- **データ送信無効**: ローカル監視のみ

### 3. ダッシュボードアクセス
ブラウザで **http://localhost:5002** にアクセス

### 4. 停止
```bash
./scripts/stop_enhanced_monitoring.sh
```

## 🔧 Shell Scripts Reference

利用可能なスクリプトとその使用方法：

### `start_enhanced_fqdn_monitoring.sh` - メイン起動スクリプト
```bash
# 実行権限付与
chmod +x scripts/start_enhanced_fqdn_monitoring.sh

# 実行
./scripts/start_enhanced_fqdn_monitoring.sh
```

**機能**:
- 前提条件チェック（Go、Python3、必要ディレクトリ）
- macOS権限チェック（Accessibility、sudo）
- `.env`ファイルからの自動設定読み込み
- データ送信設定の対話的入力（.envファイルがない場合）
- Go Agentと Python Web UIの並行起動
- ブラウザの自動オープン
- ログファイルの追跡

**出力ログ**:
- `~/.roiagent/logs/agent.log` - Go Agent
- `~/.roiagent/logs/webui.log` - Python Web UI

### `stop_enhanced_monitoring.sh` - 停止スクリプト
```bash
./scripts/stop_enhanced_monitoring.sh
```

**機能**:
- すべてのROI Agentプロセスを安全に停止
- tcpdumpプロセスの終了（sudo権限で）
- Go AgentとWeb UIプロセスの終了

### `test.sh` - 統合テストスクリプト
```bash
# 実行権限付与
chmod +x scripts/test.sh

# 全テスト実行
./scripts/test.sh all

# 個別テスト
./scripts/test.sh env              # 環境変数設定テスト
./scripts/test.sh build            # ビルドテスト
./scripts/test.sh data-sender      # データ送信機能テスト
./scripts/test.sh permissions      # macOS権限確認
./scripts/test.sh web              # Web UI動作テスト
./scripts/test.sh status           # 現在の動作状況確認
./scripts/test.sh clean            # クリーンアップ
./scripts/test.sh help             # ヘルプ表示
```

**機能**:
- Go/Python3インストール確認
- 必要ディレクトリの存在確認
- .envファイル設定の確認と検証
- コンパイル/ビルドテスト
- データ送信接続テスト
- macOS権限状態確認
- Flask依存関係チェック
- プロセス動作状況確認
- 古いファイルの自動クリーンアップ

### `build_mac_app.sh` - Macアプリビルドスクリプト
```bash
# 実行権限付与
chmod +x scripts/build_mac_app.sh

# アプリビルド
./scripts/build_mac_app.sh
```

**機能**:
- アプリアイコン準備（`public/icon.png`から）
- Go AgentとData Senderのバイナリビルド
- Python Web UIの統合
- Mac App Bundle（`.app`）の作成
- 実行権限設定とスクリプト統合
- `build/ROI Agent.app`への出力

**前提条件**:
- `public/icon.png`にアプリアイコンを配置（512x512px推奨）

### `setup_data_transmission.sh` - データ送信セットアップ
```bash
# 実行権限付与
chmod +x scripts/setup_data_transmission.sh

# セットアップ実行
./scripts/setup_data_transmission.sh
```

**機能**:
- Data Senderバイナリの自動ビルド
- 対話的な設定入力（サーバーURL、APIキー、送信間隔）
- `.env`ファイルの自動作成
- 設定検証とテスト送信
- データ送信の有効化

### `update_dependencies.sh` - 依存関係更新
```bash
# 実行権限付与
chmod +x scripts/update_dependencies.sh

# 依存関係更新
./scripts/update_dependencies.sh
```

**機能**:
- Go modules依存関係の更新（`go get -u ./...`）
- 不要な依存関係の削除（`go mod tidy`）
- Python依存関係の更新（`pip3 install --upgrade`）
- 全ディレクトリ（agent、data-sender、debug、windows）での一括更新

### `check_unused_go_files.sh` - 未使用Goファイル検出
```bash
# 実行権限付与
chmod +x scripts/check_unused_go_files.sh

# 検出実行
./scripts/check_unused_go_files.sh
```

**機能**:
- 各ディレクトリのGo modulesとファイル構成の確認
- main関数やexport関数の有無チェック
- 未使用ファイルの検出と推奨アクション表示
- プロジェクト全体のGoファイル構成レポート

## 📡 Data Transmission

### 設定方法

```bash
# .envファイルを作成
cd data-sender
cat > .env << EOF
ROI_AGENT_BASE_URL=https://api.yourserver.com/v1/device
ROI_AGENT_API_KEY=your-actual-api-key
ROI_AGENT_INTERVAL_MINUTES=10
EOF
```

### 送信されるデータ形式

**エンドポイント**: `POST {BASE_URL}`

**ヘッダー**:
```
Content-Type: application/json
X-API-Key: {API_KEY}
User-Agent: ROI-Agent/1.0.0
```

**ペイロード例**:
```json
{
  "device_id": "MacBook-Pro-1752306890",
  "timestamp": "2025-07-19T00:25:00Z",
  "interval_minutes": 10,
  "apps": [
    {
      "active_app": "Cursor",
      "focused_app": "Cursor",
      "focus_time_seconds": 180,
      "timestamp": "2025-07-19T00:25:00Z"
    }
  ],
  "networks": [
    {
      "fqdn": "www.yahoo.co.jp",
      "port": 443,
      "access_count": 3,
      "protocol": "HTTPS",
      "timestamp": "2025-07-19T00:25:00Z"
    }
  ],
  "metadata": {
    "os_version": "macOS",
    "agent_version": "1.0.0",
    "total_apps": 18,
    "total_domains": 3
  }
}
```

## 📱 Mac App Creation

### アイコン準備
1. アプリアイコンを `public/icon.png` に配置
2. 推奨サイズ: 512x512px以上のPNG形式

### アプリビルド
```bash
# アプリビルド（カスタムアイコン付き）
chmod +x scripts/build_mac_app.sh
./scripts/build_mac_app.sh
```

ビルドされたアプリは `build/ROI Agent.app` に作成されます。

**Mac App特徴**:
- `.env`ファイルが設定されていればデータ送信が自動有効化
- アプリケーションフォルダにドラッグ&ドロップで簡単インストール
- システム起動時の自動実行設定可能

## 🛠️ Requirements

- macOS（Accessibility権限）
- Go 1.21以上
- Python 3.x
- sudo権限（DNS監視用）

## 📁 File Structure

```
roi-agent/
├── agent/
│   ├── main.go              # メインエージェント
│   └── go.mod
├── data-sender/
│   ├── main.go              # データ送信機能
│   ├── config.go            # 設定管理
│   ├── processor.go         # データ処理
│   ├── sender.go            # HTTP送信
│   ├── logger.go            # ログ機能
│   ├── types.go             # データ型定義
│   ├── utils.go             # ユーティリティ
│   ├── .env                 # 環境変数設定
│   └── go.mod
├── web/
│   ├── enhanced_app.py      # Flask Web UI
│   ├── requirements.txt
│   └── templates/
│       └── enhanced_index.html
├── scripts/
│   ├── start_enhanced_fqdn_monitoring.sh  # 起動スクリプト
│   ├── stop_enhanced_monitoring.sh        # 停止スクリプト
│   ├── test.sh                           # 統合テストスクリプト
│   ├── build_mac_app.sh                  # Macアプリビルド
│   ├── setup_data_transmission.sh        # データ送信セットアップ
│   ├── update_dependencies.sh            # 依存関係更新
│   └── check_unused_go_files.sh          # 未使用ファイル検出
├── public/
│   └── icon.png             # アプリアイコン
├── build/
│   └── ROI Agent.app        # ビルド済みMacアプリ
├── debug/                   # デバッグツール
└── windows/                 # Windows版
```

## 💾 Data Storage

データは `~/.roiagent/` に保存されます：
- **データ**: `~/.roiagent/data/combined_YYYY-MM-DD.json`
- **ログ**: `~/.roiagent/logs/`
- **送信データ**: `~/.roiagent/transmission/`
- **送信ログ**: `~/.roiagent/transmission_logs.json`

**ファイル清理**: 7日以上古いファイルは自動清理されます。

## 📊 Dashboard Features

### アプリケーション監視
- **フォアグラウンド時間**: アプリが起動している時間
- **フォーカス時間**: アプリがアクティブ（最前面）な時間
- **リアルタイム状態**: 現在のアクティブ・フォーカスアプリ

### ネットワーク監視
- **DNS Snooping**: ユーザーがアクセスしたWebサイトのみ表示
- **FQDN + ポート**: `www.example.com:443` 形式
- **プロトコル**: HTTP/HTTPS自動判別
- **アクティブ接続**: 現在接続中のサイトのみ

### Web UI
- **3つのタブ**: アプリケーション / ネットワーク / 統合ビュー
- **リアルタイム更新**: 15秒間隔の自動更新
- **日付選択**: 過去データの表示

## 📝 Tech Stack

- **Backend**: Go (DNS監視エージェント)
- **Frontend**: Python Flask + HTML/CSS/JavaScript
- **Monitoring**: `tcpdump` (DNS) + macOS Accessibility API (アプリ)
- **Data Transmission**: Go + HTTP Client
- **Update Frequency**: 15秒間隔（監視）/ 設定可能間隔（送信）


## 📄 License

MIT License - see [LICENSE](LICENSE) file for details
