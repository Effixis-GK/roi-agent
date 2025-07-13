# Debug Tools for ROI Agent Data Transmission

このフォルダには、ROI Agentのデータ送信機能をテストするためのデバッグツールが含まれています。

## ファイル構成

- `test_data_transmission.go` - データ送信のテストツール（メインプログラム）
- `go.mod` - Go モジュール設定
- `run_debug.sh` - 自動化されたデバッグスクリプト
- `README.md` - このファイル（使用方法説明）

## Debug Scripts Usage Guide

### Script 1: `run_debug.sh` - 自動デバッグスクリプト

**目的**: 環境設定から実行まで全自動でデバッグテストを実行

**使用方法**:
```bash
# Project root から実行
cd debug
chmod +x run_debug.sh
./run_debug.sh
```

**このスクリプトが行うこと**:
1. `.env`ファイルの存在確認（`data-sender/.env`）
2. 環境変数の表示（APIキーは隠蔽）
3. Go依存関係の自動ダウンロード
4. デバッグツールのビルド
5. テストの自動実行

**出力例**:
```
🔧 ROI Agent Data Transmission Debug
====================================
Project root: /Users/user/roi-agent
✅ Found .env file at /Users/user/roi-agent/data-sender/.env

📄 Current .env configuration:
==============================
ROI_AGENT_BASE_URL=https://test-bjdnhp7xna-an.a.run.app/api/v1/device
ROI_AGENT_API_KEY=***hidden***

📦 Checking Go dependencies...
🔨 Building debug tool...
✅ Build successful!

🚀 Running debug test...
=======================
```

### Script 2: `test_data_transmission.go` - メインデバッグツール

**目的**: APIの詳細テストと2つのヘッダー方式の比較

**直接実行方法**:
```bash
cd debug
go run test_data_transmission.go
```

**ビルドして実行**:
```bash
cd debug
go build -o debug-tool test_data_transmission.go
./debug-tool
```

**このプログラムが行うこと**:
1. `.env`ファイルから環境変数を読み込み
2. テスト用のペイロードを生成
3. **Test 1**: `Authorization: Bearer`ヘッダーでテスト（古い方式）
4. **Test 2**: `X-API-Key`ヘッダーでテスト（正しい方式）
5. 詳細なリクエスト/レスポンス情報を表示

## 詳細な使用手順

### Step 1: 環境設定の確認

`.env`ファイルが`data-sender/.env`に正しく設定されていることを確認：

```bash
# .envファイルの存在確認
ls -la ../data-sender/.env

# 内容確認（APIキーが設定されているか）
cat ../data-sender/.env
```

期待される内容:
```
ROI_AGENT_BASE_URL=https://test-bjdnhp7xna-an.a.run.app/api/v1/device
ROI_AGENT_API_KEY=VvxFHdH4KKoux6n7
```

### Step 2: クイックテスト実行

**推奨方法（初回）**:
```bash
# Project root から
cd debug
./run_debug.sh
```

**手動実行（詳細制御したい場合）**:
```bash
cd debug
go mod tidy
go run test_data_transmission.go
```

### Step 3: 結果の解釈

**成功例**:
```
🧪 Test 1: Current Implementation (Authorization Bearer)
❌ Current implementation failed: server returned status 401

🧪 Test 2: Corrected Implementation (X-API-Key)
✅ Request successful!
```

**失敗例**:
```
❌ Configuration error: missing required environment variables
```
→ `.env`ファイルを確認してください

### Step 4: トラブルシューティング用コマンド

**環境変数が読み込まれない場合**:
```bash
# 直接環境変数を設定
export ROI_AGENT_BASE_URL="https://test-bjdnhp7xna-an.a.run.app/api/v1/device"
export ROI_AGENT_API_KEY="VvxFHdH4KKoux6n7"
go run test_data_transmission.go
```

**Go依存関係の問題**:
```bash
# 依存関係のクリーンインストール
rm go.sum
go clean -modcache
go mod download
go mod tidy
```

**詳細ログが必要な場合**:
```bash
# より詳細な出力
go run test_data_transmission.go 2>&1 | tee debug_output.log
```

## Debug Scripts の使い分け

| 用途 | 使用するScript | コマンド |
|------|---------------|----------|
| 初回テスト・自動化 | `run_debug.sh` | `./run_debug.sh` |
| 詳細な制御・カスタマイズ | `test_data_transmission.go` | `go run test_data_transmission.go` |
| 本番前の最終確認 | `run_debug.sh` | `./run_debug.sh` |
| エラー調査 | `test_data_transmission.go` | `go run test_data_transmission.go` |
| CI/CD環境 | `test_data_transmission.go` | `go test` (将来対応) |

## テスト内容

デバッグツールは2つのテストを実行します：

### Test 1: 現在の実装（Authorization Bearer）
- `Authorization: Bearer {API_KEY}` ヘッダーを使用
- 現在のdata-senderの実装

### Test 2: 修正された実装（X-API-Key）
- `X-API-Key: {API_KEY}` ヘッダーを使用
- curlの例に合わせた正しい実装

## 出力例

```
🔧 ROI Agent Data Transmission Debug Tool
=========================================

📡 Server URL: https://test-bjdnhp7xna-an.a.run.app/api/v1/device
🔑 API Key: VvxFHdH4...

🧪 Test 1: Current Implementation (Authorization Bearer)
======================================================
=== Request Payload ===
{
  "device_id": "MacBook-Pro-1752306890",
  "timestamp": "2025-07-12T12:00:00Z",
  "interval_minutes": 10,
  "apps": [...],
  "networks": [...],
  "metadata": {...}
}

=== Using Authorization Bearer Header (Current) ===
POST https://test-bjdnhp7xna-an.a.run.app/api/v1/device
Headers:
  Content-Type: application/json
  Authorization: Bearer VvxFHdH4...
  User-Agent: ROI-Agent-Debug/1.0.0

=== Response Status: 401 Unauthorized ===
❌ Current implementation failed: server returned status 401

🧪 Test 2: Corrected Implementation (X-API-Key)
==============================================
=== Using X-API-Key Header (Correct) ===
POST https://test-bjdnhp7xna-an.a.run.app/api/v1/device
Headers:
  Content-Type: application/json
  X-API-Key: VvxFHdH4...
  User-Agent: ROI-Agent-Debug/1.0.0

=== Response Status: 200 OK ===
✅ Request successful!
```

## トラブルシューティング

### 環境変数が読み込まれない場合

```bash
# 環境変数を直接設定
export ROI_AGENT_BASE_URL="https://test-bjdnhp7xna-an.a.run.app/api/v1/device"
export ROI_AGENT_API_KEY="VvxFHdH4KKoux6n7"

# デバッグツール実行
./debug-tool
```

### Goの依存関係エラー

```bash
# 依存関係の再インストール
go mod download
go mod tidy
go build -o debug-tool test_data_transmission.go
```

## APIヘッダーの違い

| 実装 | ヘッダー | 状態 |
|------|----------|------|
| 現在のdata-sender | `Authorization: Bearer {API_KEY}` | ❌ 不正 |
| 修正版 | `X-API-Key: {API_KEY}` | ✅ 正常 |

正しい実装ではcurlの例に合わせて`X-API-Key`ヘッダーを使用します。
