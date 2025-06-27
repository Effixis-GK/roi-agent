# ROI Agent Enhanced - ネットワーク監視機能付き

**macOS向け総合生産性監視アプリケーション**  
アプリケーション使用時間監視 + ネットワーク通信監視を統合したプロフェッショナル版

## 🚀 新機能 - ネットワーク監視

### リアルタイムネットワーク通信監視
- **HTTP/HTTPS通信監視**: ポート80/443 + カスタムポート対応
- **ドメイン別アクセス時間**: どのサイトにどれだけ時間を使っているか
- **帯域幅監視**: 送受信データ量の詳細追跡
- **アプリケーション連携**: どのアプリがどのドメインにアクセスしているか

### 統合ダッシュボード
- **アプリ + ネットワーク**: 統合された生産性分析
- **ドメイン分析機能**: トップアクセスドメインの詳細統計
- **リアルタイム更新**: 15秒間隔での自動データ更新
- **生産性指標**: フォーカス時間とネットワーク使用の相関分析

## 📊 監視項目

### アプリケーション監視
- ✅ フォアグラウンド時間
- ✅ バックグラウンド時間  
- ✅ ウィンドウフォーカス時間
- ✅ アクティブ/非アクティブ状態
- ✅ 使用時間ランキング

### ネットワーク監視 🆕
- ✅ HTTP/HTTPS通信 (ポート80/443)
- ✅ カスタムポート監視 (8080, 3000, 5000等)
- ✅ ドメイン別接続時間
- ✅ 送受信データ量
- ✅ 接続アプリケーション識別
- ✅ プロトコル分析 (TCP/UDP)

## 🛠 技術スタック

- **Goエージェント**: 高性能なシステム監視 (Datadogアーキテクチャ)
- **Python Flask**: 軽量Webダッシュボード
- **HTML/CSS/JS**: モダンなレスポンシブUI
- **JSON データ**: 軽量なデータ保存

## ⚡ クイックスタート

### 1. ワンコマンドセットアップ

```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
chmod +x quick_setup_enhanced.sh
./quick_setup_enhanced.sh
```

自動的に以下を実行します：
- 依存関係チェック
- アプリケーションビルド
- `/Applications`への自動インストール
- 権限設定ガイド
- 初回起動オプション

### 2. 手動セットアップ

```bash
# 1. ネットワーク監視版ビルド
./build_enhanced.sh

# 2. アプリをインストール
cp -R "build/ROI Agent Enhanced.app" /Applications/

# 3. 起動
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start
```

## 🎯 使用方法

### アプリケーション起動

```bash
# 統合起動（推奨）
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start

# 個別コンポーネント起動
cd /Users/taktakeu/Local/GitHub/roi-agent/agent
./monitor &                    # Goエージェント

cd ../web
source venv/bin/activate
python enhanced_app.py &       # ネットワーク対応Web UI
```

### ダッシュボードアクセス

- **メインURL**: http://localhost:5002
- **アプリタブ**: アプリケーション使用時間監視
- **ネットワークタブ**: 通信監視とドメイン統計
- **統合ビュー**: アプリ + ネットワークの総合分析

### コマンドライン操作

```bash
APP_PATH="/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent"

# 基本操作
$APP_PATH start                # 監視開始
$APP_PATH stop                 # 監視停止
$APP_PATH status               # 動作状況確認
$APP_PATH dashboard            # ダッシュボードを開く
$APP_PATH logs                 # ログ表示

# 高精度監視（管理者権限）
sudo $APP_PATH start
```

## 🔧 ネットワーク監視の詳細

### 監視対象ポート

デフォルト監視ポート：
- **80**: HTTP
- **443**: HTTPS  
- **8080**: HTTP代替
- **3000**: 開発サーバー
- **5000**: Flask等
- **8000**: Django等
- **9000**: カスタムアプリ

### 収集データ

**接続情報**:
- ドメイン名
- ポート番号
- プロトコル (HTTP/HTTPS/TCP/UDP)
- 使用アプリケーション

**使用統計**:
- 接続継続時間
- 送信データ量 (bytes)
- 受信データ量 (bytes)
- 初回/最終接続時刻

### データ保存形式

```json
{
  "date": "2025-06-17",
  "network": {
    "google.com:443": {
      "domain": "google.com",
      "port": 443,
      "protocol": "HTTPS",
      "duration": 1200,
      "bytes_sent": 15680,
      "bytes_received": 87432,
      "app_name": "Safari",
      "is_active": true
    }
  },
  "network_total": {
    "total_duration": 4830,
    "total_bytes_sent": 46800,
    "total_bytes_received": 456302,
    "unique_connections": 6
  }
}
```

## 📈 ダッシュボード機能

### アプリケーションタブ
- 使用時間ランキング
- フォアグラウンド/バックグラウンド/フォーカス時間の分類
- アクティブ状態インジケーター
- 日別履歴

### ネットワークタブ 🆕
- ドメイン接続時間ランキング
- 送受信データ量ランキング
- プロトコル別統計
- アプリケーション別ネットワーク使用

### 統合ビュー 🆕
- **生産性指標**: フォーカス時間/総使用時間の割合
- **トップ5ランキング**: アプリとネットワーク
- **データ使用量**: 総転送量表示
- **相関分析**: アプリ使用とネットワーク活動の関係

### ドメイン分析機能 🆕

```javascript
// ドメイン分析ボタンでポップアップ表示
// 例: 使用時間、データ量、接続数の詳細統計
```

## 🔍 デバッグ・トラブルシューティング

### ネットワーク監視専用デバッグツール

```bash
cd /Users/taktakeu/Local/GitHub/roi-agent

# 完全診断（推奨）
python3 network_debug_tools.py full

# 個別テスト
python3 network_debug_tools.py network      # ネットワーク接続テスト
python3 network_debug_tools.py ports        # ポート監視テスト
python3 network_debug_tools.py lsof         # lsof権限テスト
python3 network_debug_tools.py testdata     # テストデータ生成
python3 network_debug_tools.py webui        # Web UI機能テスト
```

### よくある問題と解決方法

#### ❌ ネットワーク データが表示されない

**原因**: lsof権限不足

```bash
# 解決方法1: 管理者権限で実行
sudo "/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start

# 解決方法2: 権限診断
python3 network_debug_tools.py lsof
```

#### ❌ 一部のドメインが検出されない

**原因**: DNS解決やネットワーク設定

```bash
# 診断実行
python3 network_debug_tools.py network

# ネットワーク設定確認
networksetup -getdnsservers Wi-Fi
```

#### ❌ データが古い/更新されない

**原因**: エージェントが停止

```bash
# 状況確認
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" status

# 再起動
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" stop
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" start
```

## ⚙️ 設定とカスタマイズ

### 設定ファイル: `~/.roiagent/config.yaml`

```yaml
monitor:
  interval: 15  # 監視間隔（秒）
  data_retention_days: 30

network:
  monitor_ports: [80, 443, 8080, 3000, 5000, 8000, 9000]
  monitor_protocols: ["HTTP", "HTTPS", "TCP"]
  dns_resolution: true
  packet_capture: false  # 管理者権限必要

web:
  host: "127.0.0.1"
  port: 5002
  auto_refresh: 30  # ダッシュボード自動更新間隔（秒）

security:
  require_accessibility: true
  local_only: true
```

### カスタムポート追加

```yaml
network:
  monitor_ports: [80, 443, 8080, 3000, 5000, 8000, 9000, 4000, 8888]
```

## 🔒 セキュリティとプライバシー

### 必要な権限

1. **アクセシビリティ権限** (必須)
   - システム環境設定 > セキュリティとプライバシー > プライバシー > アクセシビリティ
   - "ROI Agent Enhanced" を追加

2. **ネットワーク権限** (推奨)
   - より詳細なネットワーク監視のため管理者権限での実行を推奨

### データ保護

- **ローカル保存**: すべてのデータは `~/.roiagent/` に保存
- **外部送信なし**: データは一切外部に送信されません
- **暗号化**: 設定により機密データの暗号化可能
- **自動削除**: 設定した日数後に古いデータを自動削除

### プライバシー配慮

- **URL内容**: URLパスやクエリパラメータは記録しません
- **ドメインのみ**: ドメイン名とポート番号のみ記録
- **個人データ**: 個人を特定できる情報は収集しません

## 📦 配布とインストール

### DMGインストーラー作成

```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
chmod +x create_dmg.sh
./create_dmg.sh

# 作成されるファイル: build/ROI-Agent-Enhanced-Installer.dmg
```

### 他のMacへの配布

1. **DMGファイル配布**: 
   - `build/ROI-Agent-Enhanced-Installer.dmg` を配布
   - ダブルクリックでマウント
   - アプリをApplicationsフォルダにドラッグ&ドロップ

2. **ZIP配布**:
   ```bash
   cd build
   zip -r "ROI-Agent-Enhanced-v2.0.zip" "ROI Agent Enhanced.app"
   ```

3. **GitHub Release**: 
   - DMGファイルをGitHub Releasesに添付して配布

## 🚀 高度な使用方法

### API エンドポイント

#### ネットワーク監視専用API

```bash
# ネットワークデータのみ取得
curl "http://localhost:5002/api/data?type=network"

# ドメイン分析データ
curl "http://localhost:5002/api/network/domains"

# ネットワーク接続時間ランキング
curl "http://localhost:5002/api/data?type=network&network_category=duration"

# データ転送量ランキング
curl "http://localhost:5002/api/data?type=network&network_category=bytes_sent"
```

#### 統合データAPI

```bash
# アプリ + ネットワーク統合データ
curl "http://localhost:5002/api/data?type=both"

# 特定日の統合データ
curl "http://localhost:5002/api/data?type=both&date=2025-06-17"
```

### 自動化スクリプト例

#### ネットワーク使用量レポート

```bash
#!/bin/bash
# network_report.sh - 日次ネットワーク使用量レポート

DATE=$(date +%Y-%m-%d)
API_URL="http://localhost:5002/api/network/domains"

echo "=== ネットワーク使用量レポート $DATE ==="
curl -s "$API_URL" | jq -r '
  .[:5][] | 
  "Domain: " + .domain + 
  " | Time: " + .formatted_duration + 
  " | Data: " + .formatted_bytes_sent + "/" + .formatted_bytes_received
'
```

#### 生産性アラート

```bash
#!/bin/bash
# productivity_alert.sh - フォーカス時間アラート

FOCUS_DATA=$(curl -s "http://localhost:5002/api/data?app_category=focus_time" | jq '.app_total.focus_time')
THRESHOLD=7200  # 2時間

if [ "$FOCUS_DATA" -lt "$THRESHOLD" ]; then
    osascript -e 'display notification "フォーカス時間が不足しています" with title "ROI Agent Alert"'
fi
```

## 🔄 アップデート

### 自動アップデート確認

```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
git pull origin main
./quick_setup_enhanced.sh
```

### バージョン確認

```bash
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" --version
```

## 🗑️ アンインストール

### 完全削除

```bash
#!/bin/bash
# uninstall.sh - ROI Agent Enhanced 完全削除

echo "ROI Agent Enhanced を完全削除します..."

# 1. 監視停止
"/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" stop 2>/dev/null

# 2. アプリケーション削除
rm -rf "/Applications/ROI Agent Enhanced.app"

# 3. データ削除
rm -rf ~/.roiagent

# 4. 自動起動設定削除
rm -f ~/Library/LaunchAgents/com.roiagent.enhanced.plist

echo "✅ 削除完了"
echo "手動で以下も確認してください:"
echo "- システム環境設定 > セキュリティとプライバシー > アクセシビリティ権限"
echo "- システム環境設定 > ユーザとグループ > ログイン項目"
```

### データのみ保持

```bash
# アプリのみ削除（データは保持）
rm -rf "/Applications/ROI Agent Enhanced.app"
```

## 🆘 サポートとコミュニティ

### バグレポート

問題が発生した場合：

1. **診断実行**:
   ```bash
   cd /Users/taktakeu/Local/GitHub/roi-agent
   python3 network_debug_tools.py full > debug_report.txt
   ```

2. **ログ収集**:
   ```bash
   "/Applications/ROI Agent Enhanced.app/Contents/MacOS/roi-agent" logs > app_logs.txt
   ```

3. **GitHub Issue**: 上記ファイルと共にissueを作成

### 機能リクエスト

- GitHub Discussionsで機能提案
- 詳細な使用ケースと共に投稿

## 📊 ベンチマークとパフォーマンス

### システムリソース使用量

- **CPU使用率**: 通常時 < 1%
- **メモリ使用量**: 約50-100MB
- **ディスク容量**: データ保存で日次約1-5MB
- **ネットワーク**: 監視のみ、通信は発生しません

### 対応スケール

- **アプリケーション**: 100+個のアプリの同時監視
- **ネットワーク接続**: 1000+個の同時接続追跡
- **データ保持**: デフォルト30日（設定で変更可能）

## 🎯 今後のロードマップ

### Version 2.1 予定機能

- **プロキシ経由監視**: 企業ネットワーク対応
- **SSL証明書情報**: 接続先の証明書詳細
- **地理位置情報**: IPアドレスの地理的位置
- **エクスポート機能**: CSV/PDF形式でのレポート出力

### Version 2.2 予定機能

- **Windows対応**: クロスプラットフォーム展開
- **チーム機能**: 複数ユーザーでのデータ共有
- **アラート機能**: カスタム閾値でのアラート設定
- **API webhook**: 外部システムとの連携

---

## 🎉 まとめ

**ROI Agent Enhanced** は、macOS向けの最も包括的な生産性監視ツールです。

- ✅ **アプリケーション監視**: 従来の使用時間追跡
- ✅ **ネットワーク監視**: 革新的な通信追跡機能  
- ✅ **統合ダッシュボード**: 統一された分析ビュー
- ✅ **プライバシー重視**: ローカル処理のみ
- ✅ **カスタマイズ可能**: 豊富な設定オプション

生産性の向上と時間の最適化を、今すぐ始めましょう！

```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
./quick_setup_enhanced.sh
```

**Happy Productivity! 🚀📊**