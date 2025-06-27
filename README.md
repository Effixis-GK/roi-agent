# ROI Agent Enhanced

**macOS向けネットワーク監視機能付き生産性監視アプリケーション**

## 機能

- 📱 **アプリケーション監視**: 使用時間、フォーカス時間の追跡
- 🌐 **ネットワーク監視**: HTTP/HTTPS通信のリアルタイム追跡
- 📊 **FQDN解決**: IPアドレスから実際のドメイン名を取得
- 🔗 **リダイレクト追跡**: 最終アクセス先を特定
- 📈 **統合ダッシュボード**: Web UIでリアルタイム表示

---

## 開発者向け（コンソール実行・デバッグ）

### 必要な環境
- macOS 10.15+
- Go 1.19+
- Python 3.8+

### クイックスタート
```bash
# 1. リポジトリクローン
git clone <repository-url>
cd roi-agent

# 2. 監視開始
./scripts/start_enhanced_fqdn_monitoring.sh
```

### デバッグコマンド
```bash
# ネットワーク機能テスト
python3 debug/network_fqdn_debug.py full

# 実データ監視
python3 debug/real_data_debug.py monitor

# リアルタイムログ
tail -f ~/.roiagent/logs/enhanced_agent_*.log
```

### 手動実行
```bash
# エージェント起動
cd agent && go run enhanced_network_main.go &

# Web UI起動
cd web && python3 enhanced_app.py &

# ダッシュボード
open http://localhost:5002
```

---

## 配布版（Mac アプリ）

### アプリ作成
```bash
# 1. 完全ビルド＆インストール
./scripts/quick_setup_enhanced.sh

# 2. 手動ビルド（上級者向け）
./scripts/build_enhanced.sh
cp -R "build/ROI Agent Enhanced.app" /Applications/
```

### アプリ使用方法
```bash
# 起動
open "/Applications/ROI Agent Enhanced.app"

# または CLI から
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start

# ダッシュボード
open http://localhost:5002

# 停止
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" stop
```

### 必要な権限設定
1. **アクセシビリティ権限**（必須）
   - システム環境設定 > セキュリティとプライバシー > アクセシビリティ
   - "ROI Agent Enhanced" を追加

2. **ネットワーク監視精度向上**（オプション）
   ```bash
   sudo "/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start
   ```

---

## API エンドポイント

```bash
# 状況確認
curl http://localhost:5002/api/status

# アプリデータ
curl http://localhost:5002/api/data?type=apps

# ネットワークデータ
curl http://localhost:5002/api/data?type=network

# 統合データ
curl http://localhost:5002/api/data?type=both

# ドメイン分析
curl http://localhost:5002/api/network/domains
```

---

## プロジェクト構成

```
roi-agent/
├── agent/                          # Core監視エージェント
│   ├── enhanced_network_main.go    # メインエージェント
│   └── go.mod                      # Go依存関係
├── web/                            # Web UI
│   ├── enhanced_app.py             # Flask Web UI
│   ├── requirements.txt            # Python依存関係
│   └── templates/
│       └── enhanced_index.html     # ダッシュボード
├── config/                         # 設定ファイル
│   └── config.yaml                 # アプリ設定
├── scripts/                        # 🆕 Shell Scripts
│   ├── build_enhanced.sh           # アプリビルド
│   ├── quick_setup_enhanced.sh     # 自動セットアップ
│   ├── start_enhanced_fqdn_monitoring.sh  # 開発モード起動
│   ├── setup_permissions.sh        # 権限設定
│   └── create_dmg.sh              # DMGインストーラー作成
├── debug/                          # 🆕 Debug Tools
│   ├── network_fqdn_debug.py       # ネットワークデバッグ
│   └── real_data_debug.py          # 実データ検証
├── build/                          # ビルド出力
├── data/                           # 実データ保存
├── logs/                           # ログファイル
└── README.md                       # このファイル
```

---

## デバッグツール詳細

### ネットワークデバッグ
```bash
# 完全診断
python3 debug/network_fqdn_debug.py full

# FQDN解決テスト
python3 debug/network_fqdn_debug.py fqdn

# 現在の接続確認
python3 debug/network_fqdn_debug.py connections

# DNS監視テスト
python3 debug/network_fqdn_debug.py dns

# リダイレクト追跡テスト
python3 debug/network_fqdn_debug.py redirects
```

### 実データ検証
```bash
# 実データ収集検証
python3 debug/real_data_debug.py verify

# リアルタイム監視
python3 debug/real_data_debug.py monitor

# システム状態確認
python3 debug/real_data_debug.py system

# データ分析
python3 debug/real_data_debug.py analyze
```

---

## スクリプト詳細

### 開発用スクリプト
- `scripts/start_enhanced_fqdn_monitoring.sh` - 開発モード起動
- `scripts/setup_permissions.sh` - 権限設定ヘルパー

### ビルド用スクリプト
- `scripts/quick_setup_enhanced.sh` - ワンコマンド自動セットアップ
- `scripts/build_enhanced.sh` - アプリビルド
- `scripts/create_dmg.sh` - DMGインストーラー作成

---

## トラブルシューティング

### データが表示されない
```bash
# 権限確認
python3 debug/network_fqdn_debug.py full

# プロセス確認
ps aux | grep enhanced

# ログ確認
tail -f ~/.roiagent/logs/enhanced_*.log
```

### ネットワーク監視が動作しない
```bash
# FQDN解決テスト
python3 debug/network_fqdn_debug.py fqdn

# 管理者権限で実行
sudo ./scripts/start_enhanced_fqdn_monitoring.sh
```

---

**Note**: 実データのみを使用します。テストデータは使用しません。