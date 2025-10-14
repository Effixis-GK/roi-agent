#!/bin/bash

# Workload Identity Federation セットアップスクリプト（修正版）

set -e

echo "========================================"
echo "Workload Identity Federation Setup"
echo "========================================"
echo ""

PROJECT_ID="teak-frame-465410-a0"
PROJECT_NUMBER="607617540267"
SA_EMAIL="github-actions-agent@${PROJECT_ID}.iam.gserviceaccount.com"
POOL_NAME="github-pool"
PROVIDER_NAME="github-provider"
REPO_OWNER="Effixis-GK"
REPO_NAME="roi-agent"

echo "プロジェクト: ${PROJECT_ID}"
echo "サービスアカウント: ${SA_EMAIL}"
echo "GitHubリポジトリ: ${REPO_OWNER}/${REPO_NAME}"
echo ""

# 1. Workload Identity Pool確認（既に作成済み）
echo "[1/5] Workload Identity Poolの確認..."
if gcloud iam workload-identity-pools describe ${POOL_NAME} \
   --location=global \
   --project=${PROJECT_ID} > /dev/null 2>&1; then
    echo "✅ Workload Identity Poolは既に存在します"
else
    echo "❌ Workload Identity Poolが見つかりません"
    echo "再作成します..."
    gcloud iam workload-identity-pools create ${POOL_NAME} \
      --location=global \
      --display-name="GitHub Actions Pool" \
      --project=${PROJECT_ID}
    echo "✅ Workload Identity Poolを作成しました"
fi
echo ""

# 2. Workload Identity Provider作成（attribute-mapping修正）
echo "[2/5] Workload Identity Providerの作成..."

# 既存のProviderを削除（存在する場合）
if gcloud iam workload-identity-pools providers describe ${PROVIDER_NAME} \
   --workload-identity-pool=${POOL_NAME} \
   --location=global \
   --project=${PROJECT_ID} > /dev/null 2>&1; then
    echo "既存のProviderを削除します..."
    gcloud iam workload-identity-pools providers delete ${PROVIDER_NAME} \
      --workload-identity-pool=${POOL_NAME} \
      --location=global \
      --project=${PROJECT_ID} \
      --quiet
fi

# 新しいProviderを作成（attribute-mappingを修正）
gcloud iam workload-identity-pools providers create-oidc ${PROVIDER_NAME} \
  --workload-identity-pool=${POOL_NAME} \
  --location=global \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner=='${REPO_OWNER}'" \
  --project=${PROJECT_ID}

echo "✅ Workload Identity Providerを作成しました"
echo ""

# 3. サービスアカウントにIAMバインディング追加
echo "[3/5] サービスアカウントへのIAMバインディング..."
WORKLOAD_IDENTITY_USER="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_NAME}/attribute.repository/${REPO_OWNER}/${REPO_NAME}"

gcloud iam service-accounts add-iam-policy-binding ${SA_EMAIL} \
  --role="roles/iam.workloadIdentityUser" \
  --member="${WORKLOAD_IDENTITY_USER}" \
  --project=${PROJECT_ID} \
  --condition=None

echo "✅ IAMバインディングを追加しました"
echo ""

# 4. Workload Identity Provider名を取得
echo "[4/5] Workload Identity Provider情報..."
WORKLOAD_IDENTITY_PROVIDER="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_NAME}/providers/${PROVIDER_NAME}"

echo "Workload Identity Provider:"
echo "${WORKLOAD_IDENTITY_PROVIDER}"
echo ""

# 5. GitHub Secrets設定手順
echo "[5/5] GitHub Secrets設定手順"
echo "========================================"
echo ""
echo "以下の情報をGitHub Secretsに登録してください："
echo ""
echo "1. GitHubリポジトリにアクセス"
echo "   https://github.com/${REPO_OWNER}/${REPO_NAME}/settings/secrets/actions"
echo ""
echo "2. 'New repository secret' をクリックして以下を追加:"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Secret 1:"
echo "  Name: GCP_PROJECT_ID"
echo "  Value: ${PROJECT_ID}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Secret 2:"
echo "  Name: GCP_SA_EMAIL"
echo "  Value: ${SA_EMAIL}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Secret 3:"
echo "  Name: GCP_WORKLOAD_IDENTITY_PROVIDER"
echo "  Value:"
echo "  ${WORKLOAD_IDENTITY_PROVIDER}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "========================================"
echo "✅ セットアップ完了！"
echo "========================================"
echo ""
echo "次のステップ:"
echo "1. GitHub Secretsに上記3つの値を登録"
echo "2. Storage IAM設定: ./scripts/setup-storage-iam.sh"
echo "3. Gitにコミット＆プッシュ"
echo "4. タグを作成してリリース: ./scripts/release-agent.sh"
echo ""

# 設定を保存
cat > workload-identity-config.txt << EOF
# Workload Identity Federation設定情報
# GitHub Secretsに登録してください

GCP_PROJECT_ID=${PROJECT_ID}
GCP_SA_EMAIL=${SA_EMAIL}
GCP_WORKLOAD_IDENTITY_PROVIDER=${WORKLOAD_IDENTITY_PROVIDER}

# GitHub Secrets登録URL
https://github.com/${REPO_OWNER}/${REPO_NAME}/settings/secrets/actions
EOF

echo "設定情報を workload-identity-config.txt に保存しました"
echo ""
