# GitHub Actions セットアップガイド

このドキュメントでは、GitHub Actionsを使用した自動リリースのセットアップ方法を説明します。

## 📋 必要な準備

### 1. GitHub Secrets の設定

リポジトリの Settings → Secrets and variables → Actions で以下を設定：

| Secret名 | 説明 | 取得方法 |
|---------|------|---------|
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Workload Identity プロバイダー | GCP IAM設定 |
| `GCP_SERVICE_ACCOUNT` | サービスアカウント | `github-actions@PROJECT.iam.gserviceaccount.com` |
| `CLOUDSQL_PASSWORD` | CloudSQLパスワード | CloudSQL管理画面 |
| `GITHUB_TOKEN` | GitHubトークン | 自動提供（設定不要） |

### 2. GCP Workload Identity の設定

```bash
# プロジェクト設定
PROJECT_ID="teak-frame-465410-a0"
REPO="Effixis-GK/roi-agent"

# Workload Identity Pool作成
gcloud iam workload-identity-pools create "github-actions" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --display-name="GitHub Actions Pool"

# Workload Identity Provider作成
gcloud iam workload-identity-pools providers create-oidc "github-actions-provider" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --display-name="GitHub Actions Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# サービスアカウント作成
gcloud iam service-accounts create github-actions \
  --project="${PROJECT_ID}" \
  --display-name="GitHub Actions"

# 権限付与
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# Workload Identity バインディング
gcloud iam service-accounts add-iam-policy-binding \
  "github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions/attribute.repository/${REPO}"
```

### 3. PostgreSQL クライアントのインストール（ローカル）

```bash
# macOS
brew install postgresql@15

# 環境変数設定
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
```

---

## 🚀 使い方

### 方法1: タグをプッシュして自動リリース

```bash
# バージョンタグを作成
git tag v1.4.1
git push origin v1.4.1
```

→ GitHub Actionsが自動的に実行されます

### 方法2: GitHub UIから手動実行

1. GitHubリポジトリの **Actions** タブを開く
2. **Release ROI Agent** ワークフローを選択
3. **Run workflow** をクリック
4. バージョン番号を入力（例: `1.4.1`）
5. **Run workflow** を実行

---

## 📦 ワークフローの流れ

```
1. コードチェックアウト
2. Go環境セットアップ
3. バージョン決定
4. macOSバイナリビルド（arm64 + amd64）
5. PKGファイル作成
6. GCP認証
7. GCSにアップロード
8. PostgreSQLクライアントインストール
9. Cloud SQL Proxy経由でCloudSQL更新
10. GitHub Releaseの作成
```

---

## 🔧 トラブルシューティング

### Workload Identity エラー

```
Error: google-github-actions/auth failed with: retry function failed after 3 attempts
```

**解決方法**:
1. Workload Identity Poolが正しく設定されているか確認
2. サービスアカウントの権限を確認
3. リポジトリ名が正確か確認（`Effixis-GK/roi-agent`）

### CloudSQL接続エラー

```
Error: could not connect to server
```

**解決方法**:
1. Cloud SQL Proxyが正しく起動しているか確認
2. `CLOUDSQL_PASSWORD` Secretが正しく設定されているか確認
3. CloudSQLインスタンスが起動しているか確認

### GCS アップロードエラー

```
Error: AccessDeniedException: 403
```

**解決方法**:
1. サービスアカウントに `roles/storage.objectAdmin` 権限があるか確認
2. GCSバケットが存在するか確認

---

## 📝 ローカルスクリプトとの違い

| 項目 | ローカルスクリプト | GitHub Actions |
|-----|------------------|----------------|
| 実行環境 | 開発者のMac | GitHub Runners |
| 認証 | gcloud CLI | Workload Identity |
| CloudSQL接続 | Cloud SQL Proxy | Cloud SQL Proxy |
| トリガー | 手動実行 | タグプッシュ/手動 |
| ログ | ターミナル | GitHub UI |

---

## 🔐 セキュリティ

- **Secretsは暗号化されて保存**されます
- **Workload Identity**により、サービスアカウントキーの管理が不要
- **最小権限の原則**に従い、必要な権限のみ付与

---

## 📚 関連ドキュメント

- [GitHub Actions - Workload Identity](https://github.com/google-github-actions/auth)
- [Cloud SQL Proxy](https://cloud.google.com/sql/docs/postgres/connect-auth-proxy)
- [GCS 権限管理](https://cloud.google.com/storage/docs/access-control/iam)

---

## 更新履歴

- 2024-12-09: 初版作成

