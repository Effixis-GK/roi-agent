#!/bin/bash

# Cloud Storage診断スクリプト

echo "========================================"
echo "Cloud Storage診断"
echo "========================================"
echo ""

PROJECT_ID="teak-frame-465410-a0"
BUCKET_NAME="gs://roi-agent-releases"

# 1. プロジェクトの組織ポリシー確認
echo "[1/5] 組織ポリシーの確認..."
echo "Uniform Bucket-Level Access ポリシー:"
gcloud resource-manager org-policies describe \
  constraints/storage.uniformBucketLevelAccess \
  --project=${PROJECT_ID} 2>/dev/null || echo "  組織ポリシーは設定されていません（デフォルト: 強制）"
echo ""

# 2. バケットの現在の設定確認
echo "[2/5] バケット設定の確認..."
gsutil uniformbucketlevelaccess get ${BUCKET_NAME}
echo ""

# 3. バケットのIAMポリシー確認
echo "[3/5] バケットのIAMポリシー確認..."
gsutil iam get ${BUCKET_NAME}
echo ""

# 4. プロジェクトのIAMポリシー確認（Storage関連のみ）
echo "[4/5] プロジェクトのStorage関連IAMロール..."
gcloud projects get-iam-policy ${PROJECT_ID} \
  --flatten="bindings[].members" \
  --filter="bindings.role:roles/storage.*" \
  --format="table(bindings.role, bindings.members)"
echo ""

# 5. 推奨される設定方法
echo "[5/5] 推奨される設定方法"
echo "========================================"
echo ""
echo "組織ポリシーでUniform Bucket-Level Accessが強制されているため、"
echo "ACLではなくIAMポリシーを使用する必要があります。"
echo ""
echo "解決策: IAMポリシーでallUsersに読み取り権限を付与"
echo ""
echo "実行コマンド:"
echo "  gsutil iam ch allUsers:objectViewer ${BUCKET_NAME}"
echo ""
echo "もしこれが失敗する場合は、組織管理者に以下を依頼してください："
echo "  1. 組織ポリシー 'constraints/storage.uniformBucketLevelAccess' の確認"
echo "  2. 必要に応じてポリシーの例外設定"
echo ""
echo "代替案: Cloud Run からの認証アクセス"
echo "  - バケットをパブリックにせず、Cloud Runのサービスアカウントに権限付与"
echo "  - より安全な方法"
echo ""
