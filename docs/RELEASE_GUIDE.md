# ROI Agent リリースガイド

このドキュメントでは、ROI Agentのコード変更後の更新・リリース手順を説明します。

## 📋 目次

1. [開発フロー概要](#開発フロー概要)
2. [ローカル開発・テスト](#ローカル開発テスト)
3. [リリース手順](#リリース手順)
4. [手動更新手順](#手動更新手順)
5. [CloudSQLバージョン管理](#cloudsqlバージョン管理)
6. [トラブルシューティング](#トラブルシューティング)

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

### 方法1: 自動リリーススクリプト（推奨）

```bash
cd /path/to/roi-agent

# 1. VERSIONファイルを更新
echo "1.4.1" > VERSION

# 2. リリーススクリプトを実行
./scripts/release.sh
```

スクリプトが自動的に以下を実行:
- ✅ macOS用バイナリビルド（arm64 + amd64）
- ✅ PKGファイル作成
- ✅ GCSにアップロード
- ✅ CloudSQLにバージョン情報登録

### 方法2: 手動リリース

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

## トラブルシューティング

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

- 2024-12-09: 初版作成

