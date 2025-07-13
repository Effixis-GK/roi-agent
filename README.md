# ROI Agent - App & Network Monitor for macOS

macOS用のリアルタイムアプリケーション使用時間とネットワーク通信監視ツール

![ROI Agent](public/icon.png)

## 📋 Features

- **アプリケーション監視**: リアルタイムでアプリの使用時間、フォーカス時間を追跡
- **ネットワーク監視**: DNS監視によるWebサイトアクセス履歴
- **リアルタイムWeb UI**: 直感的なダッシュボード（ポート5002）
- **データ送信**: サーバーへの10分間隔自動送信（オプション）
- **データ保存**: 日別でのJSONデータ保存

## 🚀 Quick Start

### 1. 権限設定
**必須**: macOSのAccessibility権限を有効にしてください
1. システム設定 > プライバシーとセキュリティ > アクセシビリティ
2. ターミナルまたはVS Codeを追加

### 2. 起動（対話式）
```bash
# リポジトリをクローン
git clone <repository-url>
cd roi-agent

# 実行権限を付与
chmod +x scripts/start_enhanced_fqdn_monitoring.sh

# 監視開始（データ送信の有効/無効を選択可能）
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

## 🛠️ Requirements

- macOS（Accessibility権限）
- Go 1.21以上
- Python 3.x
- sudo権限（DNS監視用）

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

Macアプリでは、`.env`ファイルが設定されている場合、データ送信がデフォルトで有効になります。

## 📡 Data Transmission (Optional)

### 基本セットアップ
```bash
# データ送信機能をセットアップ
chmod +x scripts/setup_data_transmission.sh
./scripts/setup_data_transmission.sh
```

### 方法1: 環境変数（推奨）
```bash
# .envファイルを作成
cd data-sender
cp .env.example .env

# .envファイルを編集
# ROI_AGENT_BASE_URL=https://api.yourserver.com/v1/roi-agent
# ROI_AGENT_API_KEY=your-actual-api-key-here
```

### 方法2: コマンドライン設定
```bash
# 現在の設定を確認
./data-sender/data-sender config

# データ送信を有効化
./data-sender/data-sender enable https://api.yourserver.com/v1/roi-agent your-api-key

# テスト送信
./data-sender/data-sender process

# データ送信を無効化
./data-sender/data-sender disable
```

### 方法3: 環境変数で直接設定
```bash
export ROI_AGENT_BASE_URL="https://api.yourserver.com/v1/roi-agent"
export ROI_AGENT_API_KEY="your-actual-api-key"

# 監視開始（自動的に有効化）
./scripts/start_enhanced_fqdn_monitoring.sh
```

### 送信されるデータ（10分間隔）

**エンドポイント**: `POST {BASE_URL}/data`

**ヘッダー**:
```
Content-Type: application/json
Authorization: Bearer {API_KEY}
User-Agent: ROI-Agent/1.0.0
```

**ペイロード例**:
```json
{
  "device_id": "MacBook-Pro-1752306890",
  "timestamp": "2025-07-12T07:00:00Z",
  "interval_minutes": 10,
  "apps": [
    {
      "active_app": "Google Chrome",
      "focused_app": "Google Chrome",
      "focus_time_seconds": 180,
      "timestamp": "2025-07-12T07:00:00Z"
    }
  ],
  "networks": [
    {
      "fqdn": "www.yahoo.co.jp",
      "port": 443,
      "access_count": 3,
      "protocol": "HTTPS",
      "timestamp": "2025-07-12T07:00:00Z"
    },
    {
      "fqdn": "chatgpt.com",
      "port": 443,
      "access_count": 1,
      "protocol": "HTTPS",
      "timestamp": "2025-07-12T07:00:00Z"
    }
  ],
  "metadata": {
    "os_version": "macOS",
    "agent_version": "1.0.0",
    "total_apps": 15,
    "total_domains": 8
  }
}
```

**送信されるデータ詳細**:

**アプリケーション**:
- `active_app`: 現在アクティブなアプリ名
- `focused_app`: 現在フォーカス中のアプリ名  
- `focus_time_seconds`: フォーカス時間（秒）

**ネットワーク**:
- `fqdn`: アクセス先FQDN（例: www.example.com）
- `port`: ポート番号（例: 443）
- `access_count`: 10分間のアクセス回数
- `protocol`: プロトコル（HTTP/HTTPS）

**メタデータ**:
- `device_id`: デバイス固有識別子
- `os_version`: OS版本
- `agent_version`: エージェント版本
- `total_apps`: アプリ総数
- `total_domains`: ドメイン総数

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

## 📁 File Structure

```
roi-agent/
├── agent/
│   ├── main.go              # メインエージェント
│   └── go.mod
├── data-sender/
│   ├── main.go              # データ送信機能
│   ├── go.mod
│   ├── .env.example         # 環境変数テンプレート
│   └── GO_DEPENDENCIES_GUIDE.md
├── web/
│   ├── enhanced_app.py      # Flask Web UI
│   ├── requirements.txt
│   └── templates/
│       └── enhanced_index.html
├── scripts/
│   ├── start_enhanced_fqdn_monitoring.sh  # 起動（対話式）
│   ├── stop_enhanced_monitoring.sh        # 停止
│   ├── build_mac_app.sh                   # Macアプリビルド
│   ├── setup_data_transmission.sh         # データ送信セットアップ
│   └── update_dependencies.sh             # Go依存関係更新
├── public/
│   └── icon.png             # アプリアイコン
└── README.md
```

## 💾 Data Storage

データは `~/.roiagent/` に保存されます：
- **データ**: `~/.roiagent/data/combined_YYYY-MM-DD.json`
- **ログ**: `~/.roiagent/logs/`
- **送信データ**: `~/.roiagent/transmission/`
- **設定**: `~/.roiagent/transmission_config.json`

## 🔧 Troubleshooting

### DNS監視が動作しない
```bash
# sudo権限を確認
sudo tcpdump --version

# DNS監視テスト（30秒）
cd agent
go run main.go test-dns
```

### Accessibility権限エラー
```bash
# 権限確認
go run main.go check-permissions
```

### Web UIでデータが表示されない
```bash
# データファイル確認
ls -la ~/.roiagent/data/

# API直接テスト
curl -s http://localhost:5002/api/data | jq '.'
curl -s http://localhost:5002/api/status | jq '.'
```

### データ送信のトラブルシューティング
```bash
# 設定確認
./data-sender/data-sender config

# 送信テスト
./data-sender/data-sender process

# 送信ログ確認
ls -la ~/.roiagent/transmission/
```

## 🔒 Security & Privacy

- **ローカル監視のみ**: データ送信は完全にオプション
- **DNS監視**: 暗号化されていないDNSクエリのみ対象
- **sudo権限**: tcpdumpによるネットワーク監視にのみ使用
- **データ保存**: すべてローカルマシンに保存

### データ送信セキュリティ（オプション）
- **明示的な有効化**: 起動時またはコマンドで明示的に有効化
- **HTTPS暗号化**: すべてのデータ送信はHTTPSで暗号化
- **APIキー認証**: サーバー認証にはAPIキーが必要
- **ローカルログ**: 送信データはローカルにも保存
- **設定管理**: 環境変数または設定ファイルで管理

## 📝 Tech Stack

- **Backend**: Go (DNS監視エージェント)
- **Frontend**: Python Flask + HTML/CSS/JavaScript
- **Monitoring**: `tcpdump` (DNS) + macOS Accessibility API (アプリ)
- **Data Transmission**: Go + HTTP Client
- **Update Frequency**: 15秒間隔（監視）/ 10分間隔（送信）

## 🔄 Go Dependencies

```bash
# 依存関係追加
go get github.com/joho/godotenv

# 依存関係更新
go get -u ./...

# 不要な依存関係削除
go mod tidy

# 依存関係一覧
go list -m all
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

**ℹ️ Note**: このツールはローカル監視専用です。データ送信機能は完全にオプションで、明示的に有効化しない限りデータは外部に送信されません。プライバシーを重視し、すべてのデータはローカルマシンにも保存されます。
