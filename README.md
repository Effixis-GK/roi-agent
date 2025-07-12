# ROI Agent - App & Network Monitor for macOS

macOS用のリアルタイムアプリケーション使用時間とネットワーク通信監視ツール

![ROI Agent](public/app-icon.png)

## 📋 Features

- **アプリケーション監視**: リアルタイムでアプリの使用時間、フォーカス時間を追跡
- **ネットワーク監視**: DNS監視によるWebサイトアクセス履歴
- **リアルタイムWeb UI**: 直感的なダッシュボード（ポート5002）
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

# 監視開始（sudo権限が必要）
./scripts/start_enhanced_fqdn_monitoring.sh
```

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
1. アプリアイコンを `public/app-icon.png` に配置
2. 推奨サイズ: 512x512px以上のPNG形式

### アプリビルド
```bash
# アプリビルド（カスタムアイコン付き）
chmod +x scripts/build_mac_app.sh
./scripts/build_mac_app.sh
```

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
├── web/
│   ├── enhanced_app.py      # Flask Web UI
│   ├── requirements.txt
│   └── templates/
│       └── enhanced_index.html
├── scripts/
│   ├── start_enhanced_fqdn_monitoring.sh  # 起動
│   ├── stop_enhanced_monitoring.sh        # 停止
│   └── build_mac_app.sh                   # Macアプリビルド
├── public/
│   └── app-icon.png         # アプリアイコン
└── README.md
```

## 💾 Data Storage

データは `~/.roiagent/` に保存されます：
- **データ**: `~/.roiagent/data/combined_YYYY-MM-DD.json`
- **ログ**: `~/.roiagent/logs/`

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

## 🔒 Security & Privacy

- **ローカル監視のみ**: データは外部に送信されません
- **DNS監視**: 暗号化されていないDNSクエリのみ対象
- **sudo権限**: tcpdumpによるネットワーク監視にのみ使用
- **データ保存**: すべてローカルマシンに保存

## 📝 Tech Stack

- **Backend**: Go (DNS監視エージェント)
- **Frontend**: Python Flask + HTML/CSS/JavaScript
- **Monitoring**: `tcpdump` (DNS) + macOS Accessibility API (アプリ)
- **Update Frequency**: 15秒間隔

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

**ℹ️ Note**: このツールはローカル監視専用です。プライバシーを重視し、すべてのデータはローカルマシンにのみ保存されます。
