# ROI Agent Enhanced - tcpdump DNS Monitoring

macOS用のアプリケーション使用時間とネットワーク通信（DNS）監視ツール

## 特徴

- **アプリケーション監視**: リアルタイムでアプリの使用時間、フォーカス時間を追跡
- **tcpdump DNS監視**: `sudo tcpdump -i any port 53` を使用したDNS Snooping
- **FQDNとポート表示**: ネットワーク通信先のFQDNとポート番号を表示
- **リアルタイム更新**: 15秒間隔でのデータ更新
- **Web UI**: 直感的なWebインターフェース（ポート5002）

## 要件

- macOS（Accessibility権限が必要）
- Go 1.21以上
- Python 3.x
- sudo権限（tcpdumpのため）

## インストール・起動

### 1. 権限設定

アプリケーション監視のため、Accessibility権限を有効にしてください：
1. システム設定 > プライバシーとセキュリティ > アクセシビリティ
2. ターミナルまたは実行環境を追加

### 2. 起動

```bash
# 実行権限を付与
chmod +x scripts/start_enhanced_fqdn_monitoring.sh

# 監視開始（sudo権限が必要）
./scripts/start_enhanced_fqdn_monitoring.sh
```

### 3. Web UIアクセス

ブラウザで http://localhost:5002 にアクセス

### 4. 停止

```bash
# 監視停止
./scripts/stop_enhanced_monitoring.sh

# または手動で停止
sudo pkill -f "tcpdump.*port 53"
pkill -f main.go
pkill -f enhanced_app.py
```

## 機能詳細

### アプリケーション監視

- **フォアグラウンド時間**: アプリが起動している時間
- **フォーカス時間**: アプリがアクティブ（最前面）にある時間
- **バックグラウンド時間**: バックグラウンドで動作している時間
- **リアルタイム状態**: 現在アクティブ・フォーカス中のアプリを表示

### ネットワーク監視（tcpdump DNS）

- **DNS Snooping**: tcpdumpでDNSクエリ（ポート53）を監視
- **FQDN解決**: DNSクエリからFQDNを抽出
- **ポート推定**: FQDNパターンからHTTP(80)/HTTPS(443)ポートを推定
- **接続状態**: アクティブな接続のみ表示
- **プロトコル分類**: HTTP/HTTPSの自動判別

### Web UI機能

- **3つのタブ**:
  - アプリケーション: アプリ使用時間ランキング
  - ネットワーク: ネットワーク接続ランキング
  - 統合ビュー: 全体サマリーとTop5表示

- **フィルタリング**: アクティブなアプリ・接続のみ表示
- **リアルタイム更新**: 15秒間隔の自動更新
- **日付選択**: 過去データの表示

## ファイル構成

```
roi-agent/
├── agent/
│   ├── main.go              # メインエージェント（tcpdump DNS監視）
│   └── go.mod               # Go モジュール
├── web/
│   ├── enhanced_app.py      # Flask Web UI
│   ├── requirements.txt     # Python依存関係
│   └── templates/
│       └── enhanced_index.html  # Web UIテンプレート
└── scripts/
    ├── start_enhanced_fqdn_monitoring.sh  # 起動スクリプト
    └── stop_enhanced_monitoring.sh        # 停止スクリプト
```

## データ保存

データは `~/.roiagent/data/` に日付別で保存されます：

- `combined_YYYY-MM-DD.json`: 日別の統合データ
- ログファイル: `~/.roiagent/logs/`

## トラブルシューティング

### tcpdumpが起動しない

```bash
# sudo権限を確認
sudo tcpdump --version

# 権限エラーの場合
sudo chmod +s /usr/sbin/tcpdump
```

### Accessibility権限

```bash
# 権限確認
go run main.go check-permissions
```

### DNS監視テスト

```bash
# 30秒間のDNS監視テスト
go run main.go test-dns
```

## 技術仕様

- **DNS監視**: `sudo tcpdump -i any port 53` による暗号化されていないDNS通信の監視
- **更新頻度**: 15秒間隔
- **ポート推定**: FQDNパターンベースでHTTP(80)/HTTPS(443)を自動判別
- **データ形式**: JSON形式での構造化データ
- **Web UI**: Flask + HTML/CSS/JavaScript

## セキュリティ注意事項

- tcpdumpは管理者権限で実行されます
- DNS監視は暗号化されていない通信のみ対象
- ローカルネットワーク内のトラフィックのみ監視
- 収集データはローカルマシンにのみ保存

## 更新履歴

- v1.0: tcpdumpベースのDNS監視実装
- アプリケーション監視とネットワーク監視の統合
- リアルタイムWeb UI対応
- アクティブ接続・アプリのみの表示
