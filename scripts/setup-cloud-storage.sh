#!/bin/bash

# Cloud Storage バケット権限設定スクリプト

set -e

echo "========================================"
echo "Cloud Storage Setup for ROI Agent"
echo "========================================"
echo ""

BUCKET_NAME="gs://roi-agent-releases"
PROJECT_ID="teak-frame-465410-a0"

# 1. バケットが存在するか確認
echo "[1/4] バケットの確認..."
if gsutil ls -b "${BUCKET_NAME}" > /dev/null 2>&1; then
    echo "✅ バケットが存在します: ${BUCKET_NAME}"
else
    echo "❌ バケットが存在しません"
    exit 1
fi
echo ""

# 2. バケットのライフサイクル設定（古いバージョンを90日後に削除）
echo "[2/4] ライフサイクルポリシーの設定..."
cat > /tmp/lifecycle.json << EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {
          "type": "Delete"
        },
        "condition": {
          "age": 90,
          "matchesPrefix": ["v"]
        }
      }
    ]
  }
}
EOF

gsutil lifecycle set /tmp/lifecycle.json "${BUCKET_NAME}"
echo "✅ ライフサイクルポリシーを設定しました（90日後に削除）"
echo ""

# 3. バケットのCORS設定（Web UIからアクセス可能にする）
echo "[3/4] CORS設定..."
cat > /tmp/cors.json << EOF
[
  {
    "origin": ["*"],
    "method": ["GET", "HEAD"],
    "responseHeader": ["Content-Type"],
    "maxAgeSeconds": 3600
  }
]
EOF

gsutil cors set /tmp/cors.json "${BUCKET_NAME}"
echo "✅ CORS設定を完了しました"
echo ""

# 4. パブリック読み取り権限の設定（修正版）
echo "[4/4] パブリック読み取り権限の設定..."

# オプションA: バケット全体をパブリックにする（推奨）
echo "オプションA: バケット全体をパブリック読み取り可能にする"
gsutil iam ch allUsers:objectViewer "${BUCKET_NAME}" 2>/dev/null || {
    echo "⚠️  allUsers:objectViewer が失敗しました"
    echo "オプションB: 個別にアクセス制御を設定します"
    
    # 代替方法: uniformBucketLevelAccessを無効化してACLを使用
    gsutil uniformbucketlevelaccess set off "${BUCKET_NAME}"
    gsutil defacl set public-read "${BUCKET_NAME}"
    echo "✅ デフォルトACLを設定しました"
}

echo ""
echo "========================================"
echo "✅ Cloud Storage設定が完了しました！"
echo "========================================"
echo ""
echo "次のステップ:"
echo "1. GitHub Actionsの設定を行う"
echo "2. roi-agentリポジトリにワークフローを追加"
echo ""

# クリーンアップ
rm -f /tmp/lifecycle.json /tmp/cors.json
