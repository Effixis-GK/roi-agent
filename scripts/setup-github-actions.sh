#!/bin/bash

# GitHub Actions用サービスアカウント設定スクリプト

set -e

echo "========================================"
echo "GitHub Actions Service Account Setup"
echo "========================================"
echo ""

PROJECT_ID="teak-frame-465410-a0"
SA_NAME="github-actions-agent"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
KEY_FILE="github-actions-key.json"

# 1. サービスアカウント作成
echo "[1/4] サービスアカウントの作成..."
if gcloud iam service-accounts describe ${SA_EMAIL} --project=${PROJECT_ID} > /dev/null 2>&1; then
    echo "✅ サービスアカウントは既に存在します: ${SA_EMAIL}"
else
    gcloud iam service-accounts create ${SA_NAME} \
      --display-name="GitHub Actions for ROI Agent" \
      --project=${PROJECT_ID}
    echo "✅ サービスアカウントを作成しました: ${SA_EMAIL}"
fi
echo ""

# 2. Cloud Storage権限付与
echo "[2/4] Cloud Storage権限の付与..."
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin" \
  --condition=None > /dev/null 2>&1

echo "✅ Cloud Storage権限を付与しました"
echo ""

# 3. サービスアカウントキー作成
echo "[3/4] サービスアカウントキーの作成..."
if [ -f "${KEY_FILE}" ]; then
    echo "⚠️  既存のキーファイルが見つかりました: ${KEY_FILE}"
    read -p "上書きしますか？ (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "キー作成をスキップしました"
        KEY_FILE_EXISTS=true
    else
        rm "${KEY_FILE}"
        KEY_FILE_EXISTS=false
    fi
else
    KEY_FILE_EXISTS=false
fi

if [ "$KEY_FILE_EXISTS" != "true" ]; then
    gcloud iam service-accounts keys create ${KEY_FILE} \
      --iam-account=${SA_EMAIL} \
      --project=${PROJECT_ID}
    echo "✅ サービスアカウントキーを作成しました: ${KEY_FILE}"
fi
echo ""

# 4. GitHub Secrets設定手順を表示
echo "[4/4] GitHub Secrets設定手順"
echo "========================================"
echo ""
echo "以下の手順でGitHub Secretsに登録してください："
echo ""
echo "1. GitHubリポジトリにアクセス"
echo "   https://github.com/Effixis-GK/roi-agent"
echo ""
echo "2. Settings → Secrets and variables → Actions"
echo ""
echo "3. 'New repository secret' をクリック"
echo ""
echo "4. 以下の内容で登録："
echo "   Name: GCP_SA_KEY"
echo "   Value: ${KEY_FILE} の内容をコピー&ペースト"
echo ""
echo "キーファイルの内容を表示:"
echo "----------------------------------------"
cat ${KEY_FILE}
echo "----------------------------------------"
echo ""
echo "⚠️  重要: このキーファイルは機密情報です！"
echo "   - GitHub Secretsに登録後、ローカルファイルを削除してください"
echo "   - Gitにコミットしないでください"
echo ""
echo "削除コマンド:"
echo "  rm ${KEY_FILE}"
echo ""
