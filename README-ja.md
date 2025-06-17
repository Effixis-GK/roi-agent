# ROI Agent - 生産性監視ツール

macOS用のアプリケーション使用時間追跡・生産性分析ツールです。バックグラウンドで動作し、どのアプリケーションにどれだけ時間を使っているかをリアルタイムで監視・分析できます。

![ROI Agent Dashboard](https://via.placeholder.com/800x400/007AFF/FFFFFF?text=ROI+Agent+Dashboard)

## 📋 目次

- [機能](#-機能)
- [配布・ダウンロード方法](#-配布ダウンロード方法)
- [インストール方法](#-インストール方法)
- [使用方法](#-使用方法)
- [データの見方](#-データの見方)
- [削除方法](#-削除方法)
- [トラブルシューティング](#-トラブルシューティング)
- [開発者向け情報](#-開発者向け情報)

## 🚀 機能

### 主要機能
- **リアルタイム監視**: 15秒間隔でアプリケーション使用状況を自動追跡
- **3つの測定指標**:
  - **フォアグラウンド時間**: アプリが画面に表示されている時間
  - **フォーカス時間**: アプリのウィンドウがアクティブな時間
  - **バックグラウンド時間**: アプリがバックグラウンドで動作している時間
- **Webダッシュボード**: 使いやすいブラウザベースの分析画面
- **日別データ**: 日ごとの使用パターンを保存・比較
- **使用時間ランキング**: 最も時間を使っているアプリを一目で確認

### 技術的特徴
- **バックグラウンド動作**: 邪魔にならず自動で監視
- **プライバシー重視**: データは全てローカル保存、外部送信なし
- **軽量設計**: CPU・メモリ使用量を最小限に抑制
- **クロスプラットフォーム対応**: Intel・Apple Silicon両対応

## 📦 配布・ダウンロード方法

### 方法1: DMGインストーラー（推奨）

**配布者向け**:
```bash
# DMGファイルを作成
cd /Users/taktakeu/Local/GitHub/roi-agent
chmod +x create_dmg.sh
./create_dmg.sh

# 作成されるファイル: build/ROI-Agent-Installer.dmg
```

**利用者向け**:
1. `ROI-Agent-Installer.dmg` をダウンロード
2. DMGファイルをダブルクリックしてマウント
3. 「ROI Agent」をApplicationsフォルダにドラッグ
4. DMGをアンマウント（取り出し）

### 方法2: ZIPファイル

**配布者向け**:
```bash
# アプリをビルドしてZIP化
./build_app.sh
cd build
zip -r "ROI-Agent-v1.0.zip" "ROI Agent.app"
```

**利用者向け**:
1. ZIPファイルをダウンロード
2. ダブルクリックして解凍
3. 「ROI Agent.app」をApplicationsフォルダに移動

### 方法3: GitHub Release

**配布者向け**:
```bash
# GitHubリリースとして公開
git tag v1.0.0
git push origin v1.0.0
# GitHub Releases ページでDMGファイルを添付
```

**利用者向け**:
1. GitHubのReleasesページにアクセス
2. 最新版のDMGファイルをダウンロード
3. 上記DMG方法でインストール

## 🔧 インストール方法

### 必要な環境
- **macOS**: 10.14 (Mojave) 以降
- **CPU**: Intel または Apple Silicon
- **メモリ**: 50MB以上の空き容量
- **ストレージ**: 100MB以上の空き容量

### インストール手順

1. **アプリケーションの配置**
   ```bash
   # ApplicationsフォルダにコピーまたはFinderで移動
   cp -R "ROI Agent.app" /Applications/
   ```

2. **初回起動**
   ```bash
   # Applicationsフォルダからダブルクリック、または
   open "/Applications/ROI Agent.app"
   ```

3. **アクセシビリティ権限の付与**
   - システム環境設定 → セキュリティとプライバシー → プライバシー
   - 左側の「アクセシビリティ」を選択
   - 鍵マークをクリックしてパスワード入力
   - 「ROI Agent」を追加してチェックを入れる

4. **動作確認**
   - ブラウザで http://localhost:5002 にアクセス
   - ダッシュボードが表示されることを確認

## 📱 使用方法

### 基本的な使い方

#### 起動
- **GUI**: ApplicationsフォルダでROI Agentをダブルクリック
- **ターミナル**: `open "/Applications/ROI Agent.app"`

#### ダッシュボードアクセス
- **自動**: アプリ起動時に自動でブラウザが開く
- **手動**: http://localhost:5002 にアクセス
- **コマンド**: ターミナルで `roi-agent dashboard`

#### バックグラウンド監視
- アプリ起動後、自動でバックグラウンドで監視開始
- メニューバーやDockには表示されません（邪魔になりません）
- 15秒ごとにアプリケーション使用状況を記録

### コマンドライン操作

```bash
# 基本コマンド（/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent の省略形）
alias roi-agent="/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent"

# 監視開始
roi-agent start

# 監視停止
roi-agent stop

# 監視再開
roi-agent restart

# 状況確認
roi-agent status

# ダッシュボードを開く
roi-agent dashboard
```

### 日常的な使用パターン

#### 朝（作業開始時）
1. ROI Agentをダブルクリックして起動
2. アクセシビリティ権限を確認（初回のみ）
3. ダッシュボードで前日のデータを確認
4. バックグラウンドで監視開始

#### 作業中
- アプリは自動でバックグラウンド監視を継続
- 特別な操作は不要
- 15秒ごとにデータが自動保存

#### 休憩時・振り返り時
1. ダッシュボード（http://localhost:5002）にアクセス
2. 現在の使用時間ランキングを確認
3. フォーカス時間と全体時間を比較
4. 生産性を分析

#### 終業時
- アプリは自動で監視を継続（または手動停止）
- データは `~/.roiagent/data/` に自動保存

## 📊 データの見方

### ダッシュボードの構成

#### 上部ステータス
- **緑色**: エージェント正常動作中
- **赤色**: エージェント停止中または権限不足

#### 日別サマリー
- **フォアグラウンド時間**: アプリが画面に表示されていた合計時間
- **フォーカス時間**: 実際にアプリを操作していた時間（生産性の指標）
- **バックグラウンド時間**: アプリがバックグラウンドで動作していた時間
- **アクティブアプリ数**: その日に使用したアプリケーションの総数

#### アプリケーションランキング
- 使用時間の長い順にアプリが表示
- 各アプリの詳細時間（フォアグラウンド・フォーカス・バックグラウンド）
- 現在のアクティブ状態とフォーカス状態

#### コントロール
- **日付選択**: 過去のデータを確認
- **カテゴリ選択**: フォアグラウンド・フォーカス・バックグラウンド時間で並び替え
- **リフレッシュ**: 最新データに更新

### データファイル
```
~/.roiagent/data/usage_2024-12-17.json    # 日別データファイル
~/.roiagent/logs/monitor.log               # 監視ログ
~/.roiagent/logs/web.log                   # Webインターフェースログ
```

### APIアクセス
```bash
# 現在の状況をJSON形式で取得
curl -s http://localhost:5002/api/status

# 今日のデータを取得
curl -s http://localhost:5002/api/data

# 特定日のフォーカス時間データ
curl -s "http://localhost:5002/api/data?date=2024-12-17&category=focus_time"
```

## 🗑 削除方法

### 完全削除（データも含む）
```bash
# 1. 監視停止
/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent stop

# 2. アプリケーション削除
rm -rf "/Applications/ROI Agent.app"

# 3. ユーザーデータ削除
rm -rf ~/.roiagent

# 4. アクセシビリティ権限から除去（手動）
# システム環境設定 → セキュリティとプライバシー → プライバシー → アクセシビリティ
# ROI Agentのチェックを外して削除
```

### アプリのみ削除（データは保持）
```bash
# 監視停止
/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent stop

# アプリのみ削除（データは~/.roiagentに残る）
rm -rf "/Applications/ROI Agent.app"
```

### Finder経由での削除
1. **監視停止**: ターミナルで `roi-agent stop` 実行
2. **アプリ削除**: ApplicationsフォルダでROI Agentをゴミ箱に移動
3. **データ削除**: ホームフォルダの `.roiagent` フォルダを削除（隠しフォルダ）
4. **権限削除**: システム環境設定でアクセシビリティ権限から除去

## 🔧 トラブルシューティング

### よくある問題と解決法

#### 「ROI Agentを開けません」エラー
```bash
# Quarantine属性を除去
xattr -d com.apple.quarantine "/Applications/ROI Agent.app"
```

#### 権限エラー
```bash
# 実行権限を修正
chmod +x "/Applications/ROI Agent.app/Contents/MacOS/roi-agent"
chmod +x "/Applications/ROI Agent.app/Contents/MacOS/monitor"
```

#### データが収集されない
1. システム環境設定でアクセシビリティ権限を確認
2. アプリを再起動
3. 状況確認: `roi-agent status`

#### Webインターフェースにアクセスできない
1. サービス状況確認: `roi-agent status`
2. サービス再起動: `roi-agent restart`
3. ポート確認: `lsof -i :5002`

#### アプリが重い・遅い
```bash
# プロセス確認
ps aux | grep roi-agent
ps aux | grep monitor

# ログ確認
tail -f ~/.roiagent/logs/monitor.log
```

### デバッグコマンド
```bash
# 元のプロジェクトフォルダがある場合
cd /Users/taktakeu/Local/GitHub/roi-agent

# 完全診断
python debug_tools.py full

# テストデータ生成
python debug_tools.py testdata

# システム要件確認
python debug_tools.py system
```

## 👩‍💻 開発者向け情報

### 開発環境セットアップ

#### 必要なツール
- **Go**: 1.21以降
- **Python**: 3.8以降
- **macOS**: 10.14以降（開発環境）
- **Xcode Command Line Tools**: `xcode-select --install`

#### プロジェクト構成
```
roi-agent/
├── agent/                 # Go監視エージェント
│   ├── main.go           # メインプログラム
│   └── go.mod            # Go依存関係
├── web/                  # Python Flask Web UI
│   ├── app.py            # Webアプリケーション
│   ├── requirements.txt  # Python依存関係
│   └── templates/        # HTMLテンプレート
├── build_app.sh          # アプリケーションビルド
├── create_dmg.sh         # DMGインストーラー作成
├── quick_setup.sh        # ワンコマンドセットアップ
└── debug_tools.py        # デバッグツール
```

### ビルド・開発手順

#### 1. 開発環境でのテスト
```bash
# リポジトリクローン
git clone <your-repo>
cd roi-agent

# システム要件確認
python debug_tools.py system

# Goエージェントビルド
cd agent
go build -o monitor main.go

# Python仮想環境セットアップ
cd ../web
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# テストデータ生成
cd ..
python debug_tools.py testdata

# 開発サーバー起動
cd web && python app.py &
cd agent && ./monitor &
```

#### 2. アプリケーションビルド
```bash
# ワンコマンドビルド（推奨）
chmod +x quick_setup.sh
./quick_setup.sh

# 個別ビルド
chmod +x build_app.sh
./build_app.sh
```

**実行結果**:
- `build/ROI Agent.app` が作成される
- ユニバーサルバイナリ（Intel + Apple Silicon対応）
- 自己完結型（Python環境内蔵）
- データは `~/.roiagent/` に保存

#### 3. DMGインストーラー作成
```bash
chmod +x create_dmg.sh
./create_dmg.sh
```

**実行結果**:
- `build/ROI-Agent-Installer.dmg` が作成される
- ドラッグ&ドロップインストール対応
- ユーザー向け説明書付き
- 配布可能なインストーラー

### スクリプト詳細

#### `build_app.sh`
**機能**:
- Goエージェントのクロスコンパイル
- Python仮想環境の作成と依存関係インストール
- macOSアプリバンドル構造の作成
- Info.plistの生成
- 起動スクリプトの作成

**成果物**: 
- 完全に自己完結したmacOSアプリケーション
- Applicationsフォルダにコピー可能
- ダブルクリックで起動可能

#### `create_dmg.sh`
**機能**:
- DMGファイルの作成
- インストール用レイアウトの設定
- README.txtの自動生成
- Applications フォルダへのシンボリックリンク作成

**成果物**:
- プロフェッショナルなDMGインストーラー
- エンドユーザー向け配布パッケージ

#### `quick_setup.sh`
**機能**:
- 全ビルドプロセスの自動実行
- エラーハンドリング付きワークフロー
- オプション付きインストール
- 結果の表示と次のステップガイド

**成果物**:
- アプリとDMGの両方を一度に作成
- 開発からリリースまでのフルワークフロー

### デバッグとテスト

#### デバッグツールの使用
```bash
# 完全診断
python debug_tools.py full

# 個別テスト
python debug_tools.py system      # システム要件
python debug_tools.py permissions # アクセシビリティ権限
python debug_tools.py apps       # アプリ検出機能
python debug_tools.py agent      # エージェント通信
python debug_tools.py data       # データファイル分析
```

#### アプリテスト
```bash
# アプリの動作テスト
open "build/ROI Agent.app"

# コマンドライン機能テスト
"build/ROI Agent.app/Contents/MacOS/roi-agent" status
"build/ROI Agent.app/Contents/MacOS/roi-agent" start
```

### リリースプロセス

#### 1. バージョン管理
```bash
# バージョンタグ作成
git tag v1.0.0
git push origin v1.0.0
```

#### 2. リリースビルド
```bash
# クリーンビルド
rm -rf build/
./quick_setup.sh
```

#### 3. 配布
- GitHub Releasesにタグを作成
- DMGファイルをアップロード
- リリースノートを記載

### アーキテクチャ詳細

#### Go Agent (monitor)
- **言語**: Go 1.21
- **機能**: システム監視、データ収集
- **更新間隔**: 15秒
- **データ保存**: JSON形式
- **権限**: アクセシビリティアクセス

#### Python Web UI (app.py)
- **フレームワーク**: Flask 2.3.3
- **ポート**: 5002
- **機能**: データ可視化、API提供
- **データソース**: JSONファイル読み込み

#### アプリバンドル構造
```
ROI Agent.app/
├── Contents/
│   ├── Info.plist          # アプリメタデータ
│   ├── MacOS/
│   │   ├── roi-agent       # メイン起動スクリプト
│   │   └── monitor         # Goバイナリ
│   └── Resources/
│       ├── web/            # Python Flask アプリ
│       └── config/         # 設定ファイル
```

この構成により、ROI Agentは完全にネイティブなmacOSアプリケーションとして動作し、エンドユーザーにとって使いやすく、開発者にとってメンテナンスしやすいツールになっています。

## 📞 サポート

- **ドキュメント**: このREADME
- **問題報告**: GitHub Issues
- **デバッグ**: 内蔵診断コマンド
- **ログ**: `~/.roiagent/logs/` フォルダ内

---

**ROI Agent** - あなたの時間への投資収益率（ROI）を最大化するためのツールです。
