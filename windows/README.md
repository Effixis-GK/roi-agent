# ROI Agent - Windows版

Windows用のアプリケーション使用時間追跡・生産性分析ツールです。バックグラウンドで動作し、どのアプリケーションにどれだけ時間を使っているかをリアルタイムで監視・分析できます。

![ROI Agent Windows](https://img.shields.io/badge/Platform-Windows-blue)
![Version](https://img.shields.io/badge/Version-1.0.0-green)
![License](https://img.shields.io/badge/License-Educational-orange)

## 🚀 クイックスタート

### 📋 システム要件

- **OS**: Windows 10/11 (64-bit推奨)
- **Go**: 1.19以降 ([ダウンロード](https://golang.org/dl/))
- **Python**: 3.8以降 ([ダウンロード](https://python.org/downloads/))
- **メモリ**: 100MB以上
- **ストレージ**: 200MB以上

### 📦 インストール・セットアップ

#### 1. 依存関係のインストール

```cmd
# Go言語のインストール確認
go version

# Pythonのインストール確認
python --version
```

#### 2. プロジェクトのビルド

```cmd
# ビルドスクリプトを実行
build.bat
```

#### 3. アプリケーションの起動

```cmd
# 統合起動スクリプト（推奨）
start.bat

# または個別起動
roi-agent-windows.exe    # 監視エージェント
python web_app.py        # Webインターフェース
```

### 🎯 使用方法

1. **ビルド**: `build.bat` を実行してアプリケーションをビルド
2. **起動**: `start.bat` を実行して監視開始
3. **分析**: ブラウザで http://localhost:5002 にアクセス
4. **停止**: Ctrl+C でサービス停止

## ✨ 主な機能

### 📊 監視機能
- **リアルタイム監視**: 15秒間隔でアプリケーション使用状況を追跡
- **3つの指標**: フォアグラウンド・フォーカス・バックグラウンド時間を測定
- **プロセス管理**: Windows APIを使用した正確なプロセス監視
- **システムプロセス除外**: 不要なシステムプロセスを自動除外

### 🎨 Webダッシュボード
- **Windows風デザイン**: Fluent Design Systemに準拠したUI
- **日本語対応**: 完全日本語インターフェース
- **リアルタイム更新**: 30秒間隔での自動データ更新
- **レスポンシブ対応**: デスクトップ・タブレット対応

### 🔒 プライバシー・セキュリティ
- **ローカル保存**: データは全て `%USERPROFILE%\.roiagent` に保存
- **外部送信なし**: インターネット接続不要
- **軽量設計**: CPU・メモリ使用量を最小限に抑制

## 📊 分析できるデータ

### 時間の種類
- **フォアグラウンド時間**: アプリが画面に表示されている時間
- **フォーカス時間**: 実際にアプリを操作している時間（生産性指標）
- **バックグラウンド時間**: アプリがバックグラウンドで動作している時間

### 分析機能
- 日別使用時間ランキング
- アプリケーション別詳細分析
- カテゴリ別時間配分
- 生産性パターンの可視化

## 🎮 操作方法

### GUI操作
```cmd
# アプリケーション起動
start.bat

# ダッシュボードにアクセス
# ブラウザで http://localhost:5002 を開く
```

### コマンドライン操作
```cmd
# 基本操作
roi-agent-windows.exe start      # 監視開始
roi-agent-windows.exe status     # 状況確認
roi-agent-windows.exe stop       # 監視停止

# Webインターフェース
python web_app.py               # Web UI起動
```

## 📱 日常的な使用パターン

### 朝（作業開始）
1. `start.bat` をダブルクリック起動
2. ブラウザで http://localhost:5002 にアクセス
3. 前日のデータを確認
4. バックグラウンド監視が自動開始

### 日中（作業中）
- アプリは自動で監視継続（操作不要）
- 15秒ごとにデータ自動保存
- CPU・メモリ使用量は最小限

### 休憩・振り返り時
1. ダッシュボードで現在の使用時間を確認
2. フォーカス時間を分析して生産性を評価
3. カテゴリ別の時間配分をチェック

## 🗂 ファイル構成

```
windows/
├── main.go                 # Go監視エージェント
├── go.mod                  # Go依存関係
├── web_app.py             # Python Flask Web UI
├── requirements.txt       # Python依存関係
├── config.yaml           # 設定ファイル
├── build.bat             # ビルドスクリプト
├── start.bat             # 起動スクリプト
├── templates/
│   └── index.html        # WebUIテンプレート
└── README.md             # このファイル
```

### 実行時に作成されるファイル
```
%USERPROFILE%\.roiagent/
├── data/
│   ├── usage_2024-01-01.json
│   ├── usage_2024-01-02.json
│   └── ...
├── config/
└── logs/
    └── roi-agent.log
```

## 🔧 設定・カスタマイズ

### config.yaml の主要設定

```yaml
# 監視間隔の変更
agent:
  update_interval: 15  # 秒（デフォルト: 15秒）

# Webポートの変更
web:
  port: 5002  # デフォルト: 5002

# 除外プロセスの追加
windows:
  ignore_processes:
    - "your_process_name"
```

### アプリケーションカテゴリの追加

```yaml
windows:
  categories:
    custom_category:
      - "app1"
      - "app2"
```

## 🛠 トラブルシューティング

### よくある問題

#### ビルドエラー
```cmd
# Go依存関係の更新
go mod tidy
go mod download

# Python依存関係の再インストール
pip install -r requirements.txt --force-reinstall
```

#### 監視が動作しない
1. 管理者権限で実行してみる
2. ウイルス対策ソフトの除外設定を確認
3. Windowsファイアウォールの設定を確認

#### Webページにアクセスできない
```cmd
# ポート使用状況確認
netstat -an | findstr :5002

# プロセス確認
tasklist | findstr python
tasklist | findstr roi-agent
```

#### データが保存されない
1. `%USERPROFILE%\.roiagent` フォルダの権限を確認
2. ディスク容量を確認
3. ログファイルでエラーを確認

### デバッグ方法

```cmd
# 詳細ログ出力
roi-agent-windows.exe start > debug.log 2>&1

# プロセス監視テスト
roi-agent-windows.exe status

# Web UI デバッグモード
set FLASK_DEBUG=1
python web_app.py
```

## 📊 データとプライバシー

### データ保存場所
- **アプリケーションデータ**: `%USERPROFILE%\.roiagent\data\`
- **ログファイル**: `%USERPROFILE%\.roiagent\logs\`
- **設定ファイル**: プロジェクトディレクトリ内

### プライバシー保護
- ✅ **ローカル保存のみ**: データは外部送信されません
- ✅ **オープンソース**: コードは全て検証可能
- ✅ **最小権限**: 必要最小限のシステムアクセス
- ✅ **暗号化不要**: 個人情報は含まれません

## 🛠 技術仕様

### システム要件
- **OS**: Windows 10/11 (x64)
- **CPU**: 1GHz以上
- **メモリ**: 4GB以上（使用量: 約50MB）
- **ストレージ**: 200MB以上

### アーキテクチャ
- **Go Agent**: Windows API使用・システム監視・データ収集
- **Python Flask**: Webインターフェース・データ可視化
- **JSON Storage**: 軽量データ保存形式
- **RESTful API**: プログラマブルデータアクセス

### 使用技術
- **Go 1.19+**: `golang.org/x/sys/windows` パッケージ
- **Python 3.8+**: Flask, Jinja2
- **Windows API**: User32, Kernel32, Psapi
- **HTML5/CSS3/JavaScript**: モダンWebUI

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
- Visual Studio Code: 6時間 (フォーカス: 5.2時間)
- Chrome: 2時間 (フォーカス: 0.8時間)
- Teams: 30分 (フォーカス: 15分)

改善点: ブラウザの使用時間を削減し、コーディングに集中
```

### デザイナー・クリエイター
```
目標: 創作活動時間の可視化
結果:
- Photoshop: 4時間 (フォーカス: 3.5時間)
- Illustrator: 3時間 (フォーカス: 2.8時間)
- 参考サイト閲覧: 1時間 (フォーカス: 0.3時間)

改善点: 参考サイト閲覧時間を制限し、実制作時間を増加
```

### リモートワーカー
```
目標: 在宅勤務の生産性測定
結果:
- 業務アプリ: 7時間 (フォーカス: 5.5時間)
- SNS・エンタメ: 1時間 (フォーカス: 0.8時間)
- Teams: 2時間 (フォーカス: 1.8時間)

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
- ダッシュボードで今日のフォーカス時間トップ5を確認
- 週間傾向の分析
- カテゴリ別時間配分の最適化

## 🔄 定期的なメンテナンス

### データの整理
```cmd
# 古いデータファイルの確認（30日以上前）
forfiles /p "%USERPROFILE%\.roiagent\data" /s /m usage_*.json /d -30

# 古いデータの手動削除（必要に応じて）
forfiles /p "%USERPROFILE%\.roiagent\data" /s /m usage_*.json /d -30 /c "cmd /c del @path"
```

### アプリケーションの更新
```cmd
# 最新版のビルド
git pull origin main
build.bat
```

## 🗑 アンインストール方法

### 完全削除
```cmd
# 1. サービス停止
taskkill /f /im roi-agent-windows.exe
taskkill /f /im python.exe

# 2. データ削除
rmdir /s /q "%USERPROFILE%\.roiagent"

# 3. プロジェクトフォルダ削除
rmdir /s /q "C:\path\to\roi-agent-windows"
```

### アプリのみ削除（データ保持）
```cmd
# 実行ファイルのみ削除
del roi-agent-windows.exe
```

## 🤝 コミュニティ・サポート

### 問題報告
- **GitHub Issues**: バグ報告・機能要望
- **ディスカッション**: 使用方法の質問・アイデア共有

### 貢献方法
- **コードコントリビューション**: プルリクエスト歓迎
- **ドキュメント改善**: 使用方法の改善提案
- **翻訳**: 多言語対応への協力

## 📚 関連リソース

### 生産性関連
- [Deep Work - Cal Newport](https://www.calnewport.com/books/deep-work/)
- [RescueTime](https://www.rescuetime.com/) - 類似ツール
- [Toggl Track](https://toggl.com/track/) - 時間追跡ツール

### 技術ドキュメント
- [Go言語公式](https://golang.org/)
- [Flask公式](https://flask.palletsprojects.com/)
- [Windows API](https://docs.microsoft.com/en-us/windows/win32/api/)

## 📝 ライセンス

このプロジェクトは教育目的で作成されています。個人的な使用および学習目的での利用を推奨します。

## 🙏 謝辞

- **Go言語コミュニティ**: 高性能な監視エージェント開発
- **Flask コミュニティ**: 美しいWebインターフェース構築
- **Microsoft**: Windows APIの提供
- **オープンソースコミュニティ**: 素晴らしいツールとライブラリ

---

**ROI Agent Windows版** を使用して、あなたの時間への投資収益率（ROI）を最大化し、より生産的で充実した日々を送りましょう！

**🚀 今すぐ始める**: `build.bat` を実行してROI Agent Windows版を体験してください。
