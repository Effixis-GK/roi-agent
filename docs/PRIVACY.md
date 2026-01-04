# ROI Agent プライバシーポリシー

このドキュメントでは、ROI Agentのプライバシーに関する設計方針を記載しています。

---

## 📋 目次

1. [収集しない情報](#収集しない情報)
2. [macOSプライバシー保護への対応](#macosプライバシー保護への対応)
3. [データの取り扱い](#データの取り扱い)

---

## 🚫 収集しない情報

以下の情報は意図的に収集していません：

### WiFi SSID / BSSID

| 項目 | 状態 | 理由 |
|------|------|------|
| WiFi SSID（ネットワーク名） | ❌ 収集しない | 位置情報として利用可能なため |
| WiFi BSSID（アクセスポイントMAC） | ❌ 収集しない | 位置情報として利用可能なため |

**背景:**
- macOS Sonoma (14) / Sequoia (15) 以降、AppleはSSID/BSSIDを位置情報として扱うようになりました
- システムコマンドで取得しようとすると `<redacted>` として返されます
- Location Services APIを使用すれば取得可能ですが、以下の理由により採用しません：
  1. **ユーザー許可が必要**: バックグラウンドサービス（LaunchDaemon）ではUIを表示できない
  2. **プライバシーへの配慮**: SSIDから自宅・勤務先などの場所が特定可能
  3. **業務ROI分析に不要**: WiFi品質（RSSI/信号強度）のみで十分

**代替情報:**
WiFi接続品質の評価には以下の情報を使用します：
- `wifi_rssi`: 信号強度（dBm）
- `wifi_noise`: ノイズレベル（dBm）
- `wifi_channel`: 使用チャンネル
- `wifi_transmit_rate`: 転送レート（Mbps）
- `wifi_signal_quality`: 信号品質（Excellent/Good/Fair/Poor）

---

## 🍎 macOSプライバシー保護への対応

### 対応済みのmacOSプライバシー機能

| macOS機能 | roi-agentの対応 |
|-----------|----------------|
| SSID/BSSIDの秘匿 | ✅ 収集せず空文字列を記録 |
| Location Services | ✅ 使用しない |
| Full Disk Access | ⚠️ DNSログ取得に必要（tcpdump） |
| Accessibility | ⚠️ アプリケーション情報取得に必要 |

### `<redacted>` の検出と処理

```go
// macOSが返す<redacted>を検出したら空文字列に置換
if value == "<redacted>" {
    metrics.SSID = ""
    metrics.BSSID = ""
}
```

---

## 📦 データの取り扱い

### ローカルストレージ

| パス | 内容 | 保持期間 |
|------|------|---------|
| `~/.roiagent/data/` | 収集したメトリクス | 送信後リセット |
| `~/.roiagent/transmission/` | 送信済みデータのコピー | 直近の送信のみ |
| `/var/lib/roiagent/transmission/` | 共有ディレクトリ（検証用） | 直近の送信のみ |

### 送信データ

- **HTTPS暗号化**: すべての通信はHTTPS経由
- **API Key認証**: 組織ごとの固有キーで認証
- **最小限のデータ**: 業務ROI分析に必要な情報のみ送信

---

## 📝 更新履歴

| 日付 | 変更内容 |
|------|---------|
| 2025-01-04 | 初版作成 - WiFi SSID/BSSID非収集方針を明記 |

---

## 🔗 関連ドキュメント

- [METRICS_REFERENCE.md](./METRICS_REFERENCE.md) - 収集メトリクス一覧
- [RELEASE_GUIDE.md](./RELEASE_GUIDE.md) - リリースガイド

