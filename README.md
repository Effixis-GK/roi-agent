# ROI Agent - 生産性監視ツール

macOS用のアプリケーション使用時間追跡・生産性分析ツールです。バックグラウンドで動作し、どのアプリケーションにどれだけ時間を使っているかをリアルタイムで監視・分析できます。

> **English documentation**: See [README-EN.md](README-EN.md) for English version.

![ROI Agent](https://img.shields.io/badge/Platform-macOS-blue)
![Version](https://img.shields.io/badge/Version-1.0.0-green)
![License](https://img.shields.io/badge/License-Educational-orange)

## 🚀 クイックスタート

### 📦 ダウンロード・インストール

#### 方法1: ワンコマンドセットアップ（開発者向け）
```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
chmod +x quick_setup.sh
./quick_setup.sh
```

#### 方法2: DMGインストーラー（エンドユーザー向け）
1. [Releases](../../releases) から最新のDMGファイルをダウンロード
2. DMGをダブルクリックしてマウント
3. 「ROI Agent」をApplicationsフォルダにドラッグ
4. アプリを起動してアクセシビリティ権限を許可

### 🎯 使用方法

1. **起動**: ApplicationsフォルダでROI Agentをダブルクリック
2. **権限許可**: アクセシビリティ権限を付与（初回のみ）
3. **監視開始**: 自動でバックグラウンド監視が開始
4. **分析**: http://localhost:5002 でダッシュボードを確認

## ✨ 主な機能

- **📊 リアルタイム監視**: 15秒間隔でアプリケーション使用状況を追跡
- **🎯 3つの指標**: フォアグラウンド・フォーカス・バックグラウンド時間を測定
- **📱 Webダッシュボード**: 美しく使いやすいブラウザインターフェース
- **🔒 プライバシー重視**: データは全てローカル保存、外部送信なし
- **⚡ 軽量設計**: CPU・メモリ使用量を最小限に抑制
- **🍎 ネイティブアプリ**: macOSアプリとして完全統合

## 📊 分析できるデータ

### 時間の種類
- **フォアグラウンド時間**: アプリが画面に表示されている時間
- **フォーカス時間**: 実際にアプリを操作している時間（生産性指標）
- **バックグラウンド時間**: アプリがバックグラウンドで動作している時間

### 分析機能
- 日別使用時間ランキング
- アプリケーション別詳細分析
- 生産性パターンの可視化
- 時間配分の最適化提案

## 🎮 操作方法

### GUI操作
```bash
# アプリを起動
open "/Applications/ROI Agent.app"

# ダッシュボードにアクセス
# ブラウザで http://localhost:5002 を開く
```

### コマンドライン操作
```bash
# エイリアス設定（推奨）
alias roi-agent="/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent"

# 基本操作
roi-agent start      # 監視開始
roi-agent stop       # 監視停止
roi-agent status     # 状況確認
roi-agent dashboard  # ダッシュボードを開く
```

## 📱 日常的な使用パターン

### 朝（作業開始）
1. ROI Agentをダブルクリック起動
2. 前日のデータをダッシュボードで確認
3. バックグラウンド監視が自動開始

### 日中（作業中）
- アプリは自動で監視継続（操作不要）
- 15秒ごとにデータ自動保存
- CPU・メモリ使用量は最小限

### 休憩・振り返り時
1. http://localhost:5002 にアクセス
2. 現在の使用時間ランキングを確認
3. フォーカス時間を分析して生産性を評価

## 🗑 アンインストール方法

### 完全削除
```bash
# 1. 監視停止
/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent stop

# 2. アプリ削除
rm -rf "/Applications/ROI Agent.app"

# 3. データ削除
rm -rf ~/.roiagent

# 4. システム環境設定でアクセシビリティ権限から除去
```

### アプリのみ削除（データ保持）
```bash
# アプリのみ削除（データは残る）
rm -rf "/Applications/ROI Agent.app"
```

## 📦 配布・開発者向け情報

### アプリケーションビルド
```bash
# DMGインストーラー作成
./create_dmg.sh

# 成果物: build/ROI-Agent-Installer.dmg
```

### 開発環境セットアップ
```bash
# 依存関係確認
python debug_tools.py system

# 開発サーバー起動
./start_web.sh &
cd agent && ./monitor &
```

### プロジェクト構成
```
roi-agent/
├── agent/                 # Go監視エージェント
├── web/                   # Python Flask Web UI
├── build_app.sh          # アプリビルドスクリプト
├── create_dmg.sh         # DMG作成スクリプト
├── quick_setup.sh        # ワンコマンドセットアップ
└── debug_tools.py        # デバッグツール
```

### ビルドプロセス詳細

#### `build_app.sh` の動作
1. **Goエージェントビルド**: クロスコンパイルでユニバーサルバイナリ作成
2. **Pythonアプリ準備**: 仮想環境作成と依存関係インストール
3. **アプリバンドル作成**: macOS標準のアプリ構造を構築
4. **統合**: 全コンポーネントを単一アプリに統合

#### `create_dmg.sh` の動作
1. **DMGファイル作成**: 配布用ディスクイメージ生成
2. **インストーラーUI**: ドラッグ&ドロップインストール画面
3. **README追加**: インストール手順書を同梱
4. **圧縮最適化**: ファイルサイズを最小化

## 🔧 トラブルシューティング

### よくある問題

#### 「開けません」エラー
```bash
xattr -d com.apple.quarantine "/Applications/ROI Agent.app"
```

#### データが収集されない
1. システム環境設定 → セキュリティとプライバシー → アクセシビリティ
2. ROI Agentにチェックを入れる
3. アプリを再起動

#### Webページにアクセスできない
```bash
roi-agent restart
lsof -i :5002  # ポート使用状況確認
```

### デバッグコマンド
```bash
# 完全診断
python debug_tools.py full

# テストデータ生成
python debug_tools.py testdata

# システム要件確認
python debug_tools.py system
```

## 📊 データとプライバシー

### データ保存場所
- **アプリケーションデータ**: `~/.roiagent/data/`
- **ログファイル**: `~/.roiagent/logs/`
- **設定ファイル**: アプリバンドル内

### プライバシー保護
- ✅ **ローカル保存のみ**: データは外部送信されません
- ✅ **オープンソース**: コードは全て検証可能
- ✅ **権限最小化**: 必要最小限のシステムアクセス
- ✅ **暗号化不要**: 個人情報は含まれません

## 🛠 技術仕様

### システム要件
- **OS**: macOS 10.14 (Mojave) 以降
- **CPU**: Intel または Apple Silicon
- **メモリ**: 50MB以上
- **ストレージ**: 100MB以上

### アーキテクチャ
- **Go Agent**: システム監視・データ収集
- **Python Flask**: Webインターフェース・データ可視化
- **JSON Storage**: 軽量データ保存形式
- **RESTful API**: プログラマブルデータアクセス

### 性能
- **CPU使用率**: < 1%
- **メモリ使用量**: 約30-50MB
- **データ更新**: 15秒間隔
- **応答速度**: リアルタイム

## 📈 使用例・ユースケース

### 開発者・エンジニア
```
目標: コーディング時間とツール切り替えを最適化
結果: 
- VS Code: 6時間 (フォーカス: 5.2時間)
- ブラウザ: 2時間 (フォーカス: 0.8時間)
- Slack: 30分 (フォーカス: 15分)

改善点: ブラウザの使用時間を削減し、コーディングに集中
```

### デザイナー・クリエイター
```
目標: 創作活動時間の可視化
結果:
- Figma: 4時間 (フォーカス: 3.5時間)
- Photoshop: 3時間 (フォーカス: 2.8時間)
- 参考サイト閲覧: 1時間 (フォーカス: 0.3時間)

改善点: 参考サイト閲覧時間を制限し、実制作時間を増加
```

### リモートワーカー
```
目標: 在宅勤務の生産性測定
結果:
- 業務アプリ: 7時間 (フォーカス: 5.5時間)
- SNS・エンタメ: 1時間 (フォーカス: 0.8時間)
- 会議ツール: 2時間 (フォーカス: 1.8時間)

改善点: 業務集中時間を増やし、気になるアプリの使用を管理
```

## 🎯 生産性向上のヒント

### 1. フォーカス時間の最大化
- **目標設定**: フォーカス時間÷フォアグラウンド時間の比率を80%以上に
- **アプリ切り替え削減**: 短時間での頻繁な切り替えを避ける
- **通知管理**: 集中作業中は通知をオフに

### 2. 時間泥棒アプリの特定
- **SNS・エンタメアプリ**: 意図しない長時間使用を発見
- **バックグラウンド時間**: 不要なアプリの自動終了
- **会議効率**: 会議時間と実作業時間のバランス

### 3. 日次・週次の振り返り
```bash
# 今日のフォーカス時間トップ5を確認
curl -s "http://localhost:5002/api/data?category=focus_time" | jq '.ranking[:5]'

# 今週の傾向分析（手動集計）
ls ~/.roiagent/data/usage_*.json | tail -7
```

## 🔄 定期的なメンテナンス

### データの整理
```bash
# 古いデータファイルの確認（30日以上前）
find ~/.roiagent/data/ -name "usage_*.json" -mtime +30

# 古いデータの手動削除（必要に応じて）
find ~/.roiagent/data/ -name "usage_*.json" -mtime +30 -delete
```

### アプリの更新
```bash
# 最新版の確認（開発者向け）
cd /Users/taktakeu/Local/GitHub/roi-agent
git pull origin main
./quick_setup.sh
```

## 🤝 コミュニティ・サポート

### 問題報告
- **GitHub Issues**: バグ報告・機能要望
- **ディスカッション**: 使用方法の質問・アイデア共有

### 貢献方法
- **コードコントリビューション**: プルリクエスト歓迎
- **ドキュメント改善**: 使用方法の改善提案
- **翻訳**: 多言語対応への協力

### コミュニティガイドライン
- 建設的なフィードバック
- プライバシーの尊重
- オープンソース精神の重視

## 📚 関連リソース

### 生産性関連
- [Deep Work - Cal Newport](https://www.calnewport.com/books/deep-work/)
- [RescueTime](https://www.rescuetime.com/) - 類似ツール
- [Toggl Track](https://toggl.com/track/) - 時間追跡ツール

### 技術ドキュメント
- [Go言語公式](https://golang.org/)
- [Flask公式](https://flask.palletsprojects.com/)
- [macOS アプリ開発](https://developer.apple.com/documentation/)

## 📝 ライセンス

このプロジェクトは教育目的で作成されています。個人的な使用および学習目的での利用を推奨します。

## 🙏 謝辞

- **Go言語コミュニティ**: 高性能な監視エージェント開発
- **Flask コミュニティ**: 美しいWebインターフェース構築
- **macOS開発者**: アクセシビリティAPIの提供
- **オープンソースコミュニティ**: 素晴らしいツールとライブラリ

---

**ROI Agent** を使用して、あなたの時間への投資収益率（ROI）を最大化し、より生産的で充実した日々を送りましょう！

**🚀 今すぐ始める**: `./quick_setup.sh` を実行してROI Agentを体験してください。
