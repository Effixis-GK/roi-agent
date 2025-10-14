#!/bin/bash

# 初回リリーステストスクリプト

set -e

echo "========================================"
echo "ROI Agent - First Release Test"
echo "========================================"
echo ""

# 現在のブランチ確認
CURRENT_BRANCH=$(git branch --show-current)
echo "現在のブランチ: ${CURRENT_BRANCH}"
echo ""

# Gitステータス確認
if [[ -n $(git status -s) ]]; then
    echo "⚠️  コミットされていない変更があります："
    git status -s
    echo ""
    read -p "続行しますか？ (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "中止しました"
        exit 1
    fi
fi

# GitHub Actionsワークフローがコミットされているか確認
if [ ! -f ".github/workflows/build-and-release.yml" ]; then
    echo "❌ GitHub Actionsワークフローが見つかりません"
    echo "   まず以下をコミット＆プッシュしてください："
    echo "   git add .github/workflows/build-and-release.yml"
    echo "   git commit -m 'Add GitHub Actions workflow for agent release'"
    echo "   git push origin main"
    exit 1
fi

# バージョン番号入力
echo "リリースバージョンを入力してください（例: v1.0.0）:"
read -p "Version: " VERSION

if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ バージョン形式が正しくありません（例: v1.0.0）"
    exit 1
fi

echo ""
echo "リリース内容:"
echo "- バージョン: ${VERSION}"
echo "- ブランチ: ${CURRENT_BRANCH}"
echo ""
read -p "このバージョンでリリースしますか？ (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "中止しました"
    exit 1
fi

# タグ作成
echo ""
echo "タグを作成中..."
git tag -a ${VERSION} -m "Release ${VERSION}"
echo "✅ タグを作成しました: ${VERSION}"

# タグをプッシュ
echo ""
echo "タグをプッシュ中..."
git push origin ${VERSION}
echo "✅ タグをプッシュしました"

echo ""
echo "========================================"
echo "✅ リリースプロセスを開始しました！"
echo "========================================"
echo ""
echo "GitHub Actionsでビルドが開始されます："
echo "https://github.com/Effixis-GK/roi-agent/actions"
echo ""
echo "完了まで数分かかります。"
echo ""
echo "確認方法:"
echo "1. GitHubのActionsタブでビルド状況を確認"
echo "2. 完了後、以下でファイルを確認:"
echo "   gsutil ls -lh gs://roi-agent-releases/latest/"
echo "   gsutil ls -lh gs://roi-agent-releases/${VERSION}/"
echo ""
