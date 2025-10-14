#!/bin/bash

# Cloud Storage IAM設定スクリプト（認証アクセス版）

set -e

echo "========================================"
echo "Cloud Storage IAM設定（認証アクセス）"
echo "========================================"
echo ""

PROJECT_ID="teak-frame-465410-a0"
BUCKET_NAME="gs://roi-agent-releases"

# Cloud Runサービスアカウントを取得
echo "[1/4] Cloud Runサービスアカウントの確認..."

# roi-dashboardのサービスアカウント
ROI_DASHBOARD_SA=$(gcloud run services describe roi-dashboard \
  --region=asia-northeast1 \
  --project=${PROJECT_ID} \
  --format="value(spec.template.spec.serviceAccountName)" 2>/dev/null || echo "")

if [ -z "$ROI_DASHBOARD_SA" ]; then
    echo "⚠️  roi-dashboardのサービスアカウントが見つかりません"
    echo "デフォルトのCompute Engine サービスアカウントを使用します"
    ROI_DASHBOARD_SA="${PROJECT_ID}@appspot.gserviceaccount.com"
fi

echo "roi-dashboard サービスアカウント: ${ROI_DASHBOARD_SA}"
echo ""

# GitHub Actionsサービスアカウント
GITHUB_ACTIONS_SA="github-actions-agent@${PROJECT_ID}.iam.gserviceaccount.com"
echo "GitHub Actions サービスアカウント: ${GITHUB_ACTIONS_SA}"
echo ""

# [2/4] バケットにIAM権限を付与
echo "[2/4] バケットへのIAM権限付与..."

# GitHub Actions: 読み書き権限
echo "GitHub Actionsに書き込み権限を付与..."
gsutil iam ch serviceAccount:${GITHUB_ACTIONS_SA}:roles/storage.objectAdmin ${BUCKET_NAME}
echo "✅ GitHub Actions: objectAdmin"

# Cloud Run (roi-dashboard): 読み取り専用権限
echo "Cloud Runに読み取り権限を付与..."
gsutil iam ch serviceAccount:${ROI_DASHBOARD_SA}:roles/storage.objectViewer ${BUCKET_NAME}
echo "✅ Cloud Run: objectViewer"

echo ""

# [3/4] 現在のIAMポリシー確認
echo "[3/4] 現在のIAMポリシー確認..."
gsutil iam get ${BUCKET_NAME}
echo ""

# [4/4] テスト
echo "[4/4] アクセステスト..."

# GitHub Actionsサービスアカウントでテスト（キーが必要）
if [ -f "github-actions-key.json" ]; then
    echo "GitHub Actionsアカウントでテスト..."
    gcloud auth activate-service-account --key-file=github-actions-key.json
    gsutil ls ${BUCKET_NAME}/ > /dev/null 2>&1 && echo "✅ GitHub Actions: アクセス可能" || echo "❌ GitHub Actions: アクセス不可"
    gcloud auth revoke ${GITHUB_ACTIONS_SA} 2>/dev/null || true
fi

echo ""
echo "========================================"
echo "✅ IAM設定が完了しました！"
echo "========================================"
echo ""
echo "設定内容:"
echo "- GitHub Actions: 読み書き可能 (objectAdmin)"
echo "- Cloud Run (roi-dashboard): 読み取り専用 (objectViewer)"
echo ""
echo "パブリックアクセス: 無効"
echo "セキュリティ: より安全な認証アクセス方式"
echo ""
echo "次のステップ:"
echo "1. GitHub Actionsでエージェントをビルド＆アップロード"
echo "2. Cloud Run (roi-dashboard) からダウンロード機能を使用"
echo ""
