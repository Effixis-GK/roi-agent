# ROI Agent リリースガイド

このドキュメントでは、ROI Agentのコード変更後の更新・リリース手順を説明します。

## 📋 目次

1. [開発フロー概要](#開発フロー概要)
2. [ローカル開発・テスト](#ローカル開発テスト)
3. [リリース手順](#リリース手順)
4. [手動更新手順](#手動更新手順)
5. [CloudSQLバージョン管理](#cloudsqlバージョン管理)
6. [リモート設定（ダッシュボード連携）](#リモート設定ダッシュボード連携)
7. [トラブルシューティング](#トラブルシューティング)

---

## 開発フロー概要

```
コード変更 → ローカルビルド → テスト → バージョン更新 → GCSアップロード → CloudSQL更新 → リリース
```

### 関連コンポーネント

| コンポーネント | 役割 |
|--------------|------|
| `roi-agent/agent/` | メトリクス収集エージェント |
| `roi-agent/data-sender/` | データ送信・自動更新 |
| `data-collection_gcp` | APIサーバー（Cloud Run） |
| `roi-database` | DBスキーマ・マイグレーション |
| GCS `roi-agent-releases` | PKGファイルホスティング |
| CloudSQL `agent_releases` | バージョン情報管理 |

---

## ローカル開発・テスト

### 1. コード変更後のビルド

```bash
cd /path/to/roi-agent

# agentのビルド
cd agent && go build -o /tmp/roi-agent-test . && cd ..

# data-senderのビルド
cd data-sender && go build -o /tmp/data-sender-test . && cd ..
```

### 2. ローカルテスト

```bash
# エージェントをテスト実行
/tmp/roi-agent-test

# data-senderでアップデートテスト
sudo /tmp/data-sender-test update
```

### 3. 実行中のエージェントを一時的に置き換え

```bash
# エージェントを停止
sudo launchctl unload /Library/LaunchDaemons/com.roiagent.*.plist

# バイナリを置き換え
sudo cp /tmp/roi-agent-test "/Applications/ROI Agent/bin/roi-agent"
sudo cp /tmp/data-sender-test "/Applications/ROI Agent/bin/data-sender"

# 再起動
sudo launchctl load /Library/LaunchDaemons/com.roiagent.*.plist
```

---

## リリース手順

### 方法1: Gitタグでリリース（推奨）🚀

**タグをpushするだけで自動的にリリースが完了します！**

```bash
cd /path/to/roi-agent

# 1. VERSIONファイルを更新
echo "1.4.7" > VERSION

# 2. 変更をコミット
git add VERSION
git commit -m "release: v1.4.7"

# 3. タグを作成してpush
git tag -a v1.4.7 -m "v1.4.7: リリースノート"
git push origin feature#27
git push origin v1.4.7
```

#### 自動で実行される処理

タグをpushすると、GitHub Actionsが自動的に以下を実行します：

| ステップ | 内容 |
|---------|------|
| 1. ビルド | macOS用バイナリビルド（arm64 + amd64） |
| 2. PKG作成 | インストーラーパッケージ作成 |
| 3. GCSアップロード | `gs://roi-agent-releases/` にPKGをアップロード |
| 4. CloudSQL更新 | `agent_releases`テーブルにバージョン情報を登録 |
| 5. 自動更新配信 | 既存Agentが自動的に新バージョンを検出・更新 |

#### 既存Agentの自動アップグレード

リリース後、既存のAgentは以下のタイミングで自動更新されます：

1. **リモート設定のポーリング時**（10分間隔）に新バージョンを検出
2. `update_mode: "auto"` が設定されている場合、自動でPKGをダウンロード
3. バックグラウンドでインストールを実行
4. Agentが自動的に再起動

```
[Agent] → [API Server] "最新バージョンは？"
[API Server] → [Agent] "v1.4.7です（update_mode: auto）"
[Agent] → [GCS] PKGをダウンロード
[Agent] インストール実行 → 再起動
```

### 方法2: ローカルリリーススクリプト

CI/CDを使わずにローカルからリリースする場合：

```bash
cd /path/to/roi-agent

# 1. VERSIONファイルを更新
echo "1.4.7" > VERSION

# 2. リリーススクリプトを実行
./scripts/release.sh
```

スクリプトが自動的に以下を実行:
- ✅ macOS用バイナリビルド（arm64 + amd64）
- ✅ PKGファイル作成
- ✅ GCSにアップロード
- ✅ CloudSQLにバージョン情報登録

### 方法3: 手動リリース

#### Step 1: バージョン更新

```bash
# VERSION ファイルを更新
echo "1.4.1" > VERSION
```

#### Step 2: ビルド

```bash
# arm64 (Apple Silicon)
cd agent && GOOS=darwin GOARCH=arm64 go build -o ../build/macos-arm64/roi-agent .
cd ../data-sender && GOOS=darwin GOARCH=arm64 go build -o ../build/macos-arm64/data-sender .

# amd64 (Intel)
cd ../agent && GOOS=darwin GOARCH=amd64 go build -o ../build/macos-amd64/roi-agent .
cd ../data-sender && GOOS=darwin GOARCH=amd64 go build -o ../build/macos-amd64/data-sender .
```

#### Step 3: PKG作成

```bash
# ペイロード準備
mkdir -p build/payload/Applications/ROI\ Agent/{bin,Resources}
cp build/macos-arm64/* "build/payload/Applications/ROI Agent/bin/"
cp VERSION "build/payload/Applications/ROI Agent/Resources/"

# PKGビルド
pkgbuild --root build/payload \
         --identifier "com.roiagent.monitor" \
         --version "1.4.1" \
         --install-location "/" \
         build/ROI-Agent-macOS-arm64.pkg
```

#### Step 4: GCSにアップロード

```bash
gsutil cp build/ROI-Agent-macOS-arm64.pkg \
    gs://roi-agent-releases/latest/macos/ROI-Agent-macOS-arm64.pkg
```

#### Step 5: CloudSQL更新

```bash
gcloud sql connect roi-production --user=admin --project=teak-frame-465410-a0 << 'EOF'
INSERT INTO agent_releases (
    version, macos_arm64_url, is_latest, release_notes, published_at
) VALUES (
    '1.4.1',
    'https://storage.googleapis.com/roi-agent-releases/latest/macos/ROI-Agent-macOS-arm64.pkg',
    TRUE,
    'Release 1.4.1',
    CURRENT_TIMESTAMP
) ON CONFLICT (version) DO UPDATE SET 
    is_latest = TRUE,
    macos_arm64_url = EXCLUDED.macos_arm64_url,
    published_at = CURRENT_TIMESTAMP;

UPDATE agent_releases SET is_latest = false WHERE version != '1.4.1';
EOF
```

#### Step 6: コミット & Push

```bash
git add -A
git commit -m "release: v1.4.1"
git push origin feature#27
git tag v1.4.1
git push origin v1.4.1
```

---

## 手動更新手順

ユーザーのマシンで手動でエージェントを更新する方法：

### 自動更新コマンド

```bash
sudo /Applications/ROI\ Agent/bin/data-sender update
```

### 手動バイナリ置き換え

```bash
# 1. エージェント停止
sudo launchctl unload /Library/LaunchDaemons/com.roiagent.sender.plist
sudo launchctl unload /Library/LaunchDaemons/com.roiagent.monitor.plist

# 2. バイナリ置き換え
sudo cp /path/to/new/roi-agent "/Applications/ROI Agent/bin/"
sudo cp /path/to/new/data-sender "/Applications/ROI Agent/bin/"
sudo cp /path/to/VERSION "/Applications/ROI Agent/Resources/"

# 3. 再起動
sudo launchctl load /Library/LaunchDaemons/com.roiagent.monitor.plist
sudo launchctl load /Library/LaunchDaemons/com.roiagent.sender.plist
```

---

## CloudSQLバージョン管理

### テーブル構造

```sql
CREATE TABLE agent_releases (
    id UUID PRIMARY KEY,
    version VARCHAR(20) UNIQUE NOT NULL,
    macos_arm64_url TEXT,
    macos_amd64_url TEXT,
    windows_amd64_url TEXT,
    linux_amd64_url TEXT,
    macos_arm64_checksum VARCHAR(64),
    macos_amd64_checksum VARCHAR(64),
    windows_amd64_checksum VARCHAR(64),
    linux_amd64_checksum VARCHAR(64),
    release_notes TEXT,
    is_latest BOOLEAN DEFAULT FALSE,
    is_required BOOLEAN DEFAULT FALSE,
    min_version VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ
);
```

### バージョン確認

```sql
SELECT version, is_latest, published_at 
FROM agent_releases 
ORDER BY created_at DESC 
LIMIT 5;
```

### 強制アップデートの設定

```sql
UPDATE agent_releases 
SET is_required = TRUE, min_version = '1.4.0' 
WHERE version = '1.4.1';
```

---

## リモート設定（ダッシュボード連携）

ROI Dashboardで設定した値（データ送信間隔など）はリモート設定としてAgentに適用されます。

### 設定の流れ

```
ROI Dashboard → API Server → Agent (fetch-config) → キャッシュファイル → 起動時に適用
```

### キャッシュファイルの場所

| ファイル | 説明 |
|---------|------|
| `/var/lib/roiagent/remote_config.json` | リモート設定キャッシュ（推奨） |
| `~/.roiagent/remote_config.json` | フォールバック（ユーザー権限用） |

### 設定の確認

```bash
# キャッシュされた設定を確認
cat /var/lib/roiagent/remote_config.json

# 手動でリモート設定を取得
sudo /Applications/ROI\ Agent/bin/data-sender fetch-config

# 現在のリモート設定を表示
sudo /Applications/ROI\ Agent/bin/data-sender show-config
```

### 設定の優先順位

1. **リモート設定（キャッシュ）** ← 最優先（ダッシュボードで設定した値）
2. 環境変数（`.env`ファイル）
3. デフォルト値

### v1.4.7以降の動作

- **起動時にキャッシュが無い場合**: 自動的にサーバーからリモート設定をフェッチ
- **起動時にキャッシュがある場合**: キャッシュから設定を読み込み（即座に反映）
- **定期的なポーリング**: 10分間隔でリモート設定の変更を確認

### 主なリモート設定項目

| 項目 | 説明 | デフォルト |
|-----|------|----------|
| `interval_minutes` | データ送信間隔（分） | 10 |
| `enabled` | データ送信有効/無効 | true |
| `collect_apps` | アプリ使用状況の収集 | true |
| `collect_network` | ネットワーク接続の収集 | true |
| `sample_rate_seconds` | サンプリング間隔（秒） | 15 |

---

## トラブルシューティング

### ダッシュボードの設定が反映されない

1. **リモート設定を手動フェッチ**
   ```bash
   sudo /Applications/ROI\ Agent/bin/data-sender fetch-config
   ```

2. **Agentを再起動**
   ```bash
   sudo launchctl kickstart -k system/com.roiagent.daemon
   ```

3. **ログを確認**
   ```bash
   tail -30 /var/log/roiagent/roiagent.log | grep -i interval
   ```
   
   正常な場合、以下のようなログが表示されます：
   ```
   Using cached remote config interval: 5 minutes
   ```

### 自動更新が失敗する

1. **403 Forbidden エラー**
   - GCSバケットにファイルがアップロードされているか確認
   - バケットのIAM権限を確認

2. **Unsupported platform/architecture エラー**
   - `data-collection_gcp/handlers/agent_download.go`で対応プラットフォームを確認

3. **署名検証エラー**
   - 開発環境では署名なしPKGが許可されています
   - 本番環境ではDeveloper ID署名が必要

### ログの確認

```bash
# エージェントログ
tail -f /var/log/roiagent/agent.log

# data-senderログ
tail -f /var/log/roiagent/sender.log
```

### CloudSQL接続

```bash
gcloud sql connect roi-production --user=admin --project=teak-frame-465410-a0
```

---

## 関連ファイル

| ファイル | 説明 |
|---------|------|
| `VERSION` | 現在のバージョン番号 |
| `scripts/release.sh` | 自動リリーススクリプト |
| `agent/main.go` | エージェント本体 |
| `data-sender/auto_updater.go` | 自動更新ロジック |
| `METRICS_REFERENCE.md` | 収集メトリクス一覧 |

---

## 更新履歴

- 2026-01-04: v1.4.7 リモート設定（ダッシュボード連携）セクション追加
- 2024-12-09: 初版作成

