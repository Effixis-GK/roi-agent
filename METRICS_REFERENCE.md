# ROI Agent メトリック収集リファレンス

このドキュメントでは、ROI Agentが収集するすべてのメトリックと情報を一覧にしています。

> **参考**: datadog-agentの実装を参考に設計されています。  
> **対象OS**: macOS

---

## 📊 メトリック一覧

### 1. 基本システムメトリック

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `cpu_percent` | float64 | % | CPU使用率 | `top -l 1` |
| `memory_used_mb` | int64 | MB | 使用中メモリ | `vm_stat`, `sysctl` |
| `memory_total_mb` | int64 | MB | 総メモリ | `sysctl hw.memsize` |
| `memory_percent` | float64 | % | メモリ使用率 | 計算値 |
| `process_count` | int | - | プロセス数 | `ps -ax` |
| `system_uptime_sec` | int64 | 秒 | システム稼働時間 | `sysctl kern.boottime` |
| `idle_time_sec` | int64 | 秒 | ユーザーアイドル時間 | `ioreg -c IOHIDSystem` |
| `screen_locked` | bool | - | 画面ロック状態 | `pmset -g assertions` |

---

### 2. Load Average（システム負荷）

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `load_1` | float64 | - | 1分間平均負荷 | `sysctl vm.loadavg` |
| `load_5` | float64 | - | 5分間平均負荷 | `sysctl vm.loadavg` |
| `load_15` | float64 | - | 15分間平均負荷 | `sysctl vm.loadavg` |
| `load_norm_1` | float64 | - | 正規化1分負荷（CPU数で割る） | 計算値 |
| `load_norm_5` | float64 | - | 正規化5分負荷 | 計算値 |
| `load_norm_15` | float64 | - | 正規化15分負荷 | 計算値 |

**活用例**: 
- `load_norm_1 > 1.0` の場合、システムが過負荷状態
- トレンド分析で業務のピーク時間を特定

---

### 3. メモリ・Swap

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `swap_used_mb` | int64 | MB | Swap使用量 | `sysctl vm.swapusage` |
| `swap_total_mb` | int64 | MB | Swap総量 | `sysctl vm.swapusage` |
| `swap_percent` | float64 | % | Swap使用率 | 計算値 |
| `memory_pressure` | string | - | メモリプレッシャー（normal/warning/critical） | `memory_pressure` |

**活用例**:
- `swap_percent > 50%` で物理メモリ不足のアラート
- `memory_pressure = critical` でパフォーマンス問題の予兆

---

### 4. ディスクI/O

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `disk_read_mbps` | float64 | MB/s | ディスク読み取り速度 | `iostat` |
| `disk_write_mbps` | float64 | MB/s | ディスク書き込み速度 | `iostat` |
| `disk_read_ops` | int64 | ops | 読み取り操作数 | `iostat` |
| `disk_write_ops` | int64 | ops | 書き込み操作数 | `iostat` |

---

### 5. ディスク容量

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `disk_used_percent` | float64 | % | ルートディスク使用率 | `df -k /` |
| `disk_free_gb` | float64 | GB | 空き容量 | `df -k /` |
| `disk_total_gb` | float64 | GB | 総容量 | `df -k /` |
| `disk_health` | string | - | 健全性（healthy/warning/critical） | 計算値 |

**しきい値**:
- `healthy`: < 80%
- `warning`: 80-90%
- `critical`: > 90%

---

### 6. ネットワークI/O

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `net_bytes_recv` | uint64 | bytes | 受信バイト数 | `netstat -ib` |
| `net_bytes_sent` | uint64 | bytes | 送信バイト数 | `netstat -ib` |
| `net_packets_recv` | uint64 | packets | 受信パケット数 | `netstat -ib` |
| `net_packets_sent` | uint64 | packets | 送信パケット数 | `netstat -ib` |
| `net_errors_in` | uint64 | - | 入力エラー数 | `netstat -ib` |
| `net_errors_out` | uint64 | - | 出力エラー数 | `netstat -ib` |

---

### 7. TCP接続統計

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `tcp_established` | int | - | ESTABLISHED状態の接続数 | `netstat -an -p tcp` |
| `tcp_time_wait` | int | - | TIME_WAIT状態の接続数 | `netstat -an -p tcp` |
| `tcp_close_wait` | int | - | CLOSE_WAIT状態の接続数 | `netstat -an -p tcp` |

**活用例**:
- `tcp_close_wait` が高い場合、アプリケーションのリソースリーク可能性

---

### 8. WiFi情報（macOS固有）

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `wifi_ssid` | string | - | 接続中のネットワーク名 | `airport -I` |
| `wifi_rssi` | int | dBm | 信号強度（-30〜-90） | `airport -I` |
| `wifi_noise` | int | dBm | ノイズレベル | `airport -I` |
| `wifi_channel` | int | - | 使用チャンネル | `airport -I` |
| `wifi_transmit_rate` | float64 | Mbps | 転送レート | `airport -I` |
| `wifi_phy_mode` | string | - | PHYモード（802.11n/ac/ax） | 推定値 |
| `wifi_signal_quality` | string | - | 信号品質（Excellent/Good/Fair/Poor） | 計算値 |

**信号品質の判定基準**:
| RSSI | 品質 |
|------|------|
| >= -50 | Excellent |
| >= -60 | Good |
| >= -70 | Fair |
| >= -80 | Poor |
| < -80 | Very Poor |

**活用例**:
- リモートワーク環境の品質評価
- WiFi問題による生産性低下の検出

---

### 9. バッテリー情報

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `battery_level` | int | % | バッテリー残量（0-100） | `pmset -g batt` |
| `battery_charging` | bool | - | 充電中かどうか | `pmset -g batt` |
| `battery_time_remaining` | int | 分 | 残り使用可能時間 | `pmset -g batt` |

---

### 10. ファイルハンドル

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `open_file_handles` | int64 | - | 開いているファイル数 | `sysctl kern.num_files` |
| `max_file_handles` | int64 | - | 最大ファイル数 | `sysctl kern.maxfiles` |

**活用例**:
- システムリソースの枯渇検出
- アプリケーションのリソースリーク検出

---

### 11. ディスプレイ情報

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `total_displays` | int | - | 接続ディスプレイ総数 | `system_profiler SPDisplaysDataType` |
| `external_displays` | int | - | 外部ディスプレイ数 | `system_profiler SPDisplaysDataType` |

**活用例**:
- マルチモニター環境の把握
- 作業環境の充実度評価

---

### 12. Bluetooth情報

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `bluetooth_devices` | int | - | 接続中Bluetoothデバイス数 | `system_profiler SPBluetoothDataType` |

**活用例**:
- 周辺機器の利用状況把握

---

### 13. ユーザーアクティビティ

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `meeting_active` | bool | - | ミーティング中かどうか | アプリ検出（Zoom, Teams等） |
| `camera_in_use` | bool | - | カメラ使用中 | ミーティング中を推定 |
| `microphone_in_use` | bool | - | マイク使用中 | ミーティング中を推定 |
| `browser_tabs` | int | - | ブラウザタブ総数 | AppleScript |
| `focus_score` | float64 | 0-100 | 集中度スコア | 計算値 |

**ミーティング検出対象アプリ**:
- Zoom
- Microsoft Teams
- Slack
- Google Meet
- FaceTime
- Webex
- Discord
- Skype

**ブラウザタブ取得対象**:
- Safari
- Google Chrome

---

### 14. アプリケーション使用情報

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `active_app` | string | - | アクティブなアプリ名 | AppleScript |
| `focused_app` | string | - | フォーカス中のアプリ名 | AppleScript |
| `focus_time_seconds` | int | 秒 | フォーカス時間 | 計算値 |
| `foreground_time_seconds` | int | 秒 | フォアグラウンド時間 | 計算値 |

---

### 15. ネットワーク接続情報

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `fqdn` | string | - | 接続先ドメイン | DNS監視（tcpdump） |
| `port` | int | - | ポート番号 | DNS監視 |
| `access_count` | int | - | アクセス回数 | 計算値 |
| `protocol` | string | - | プロトコル（HTTP/HTTPS） | 推定値 |

---

### 16. プロセスメトリック

| メトリック名 | 型 | 単位 | 説明 | 取得方法 |
|-------------|---|------|------|---------|
| `pid` | int | - | プロセスID | `ps -ax` |
| `name` | string | - | プロセス名 | `ps -ax` |
| `cpu_percent` | float64 | % | プロセスCPU使用率 | `ps -ax` |
| `memory_mb` | int64 | MB | プロセスメモリ使用量 | `ps -ax` |

---

## 🗂 メタデータ

各送信ペイロードには以下のメタデータが含まれます：

| フィールド | 説明 |
|-----------|------|
| `device_id` | デバイス固有ID |
| `hostname` | ホスト名 |
| `os_version` | OS種別（macOS） |
| `agent_version` | エージェントバージョン |
| `employee_name` | 従業員名 |
| `employee_email` | 従業員メール |
| `timestamp` | 送信タイムスタンプ |
| `interval_minutes` | 収集間隔（分） |

---

## 📁 ファイル構成

```
roi-agent/
├── agent/
│   └── main.go              # メイン収集ロジック
├── collectors/              # メトリック収集モジュール
│   ├── load_collector.go    # Load Average
│   ├── wifi_collector_darwin.go    # WiFi情報
│   ├── network_io_collector.go     # ネットワークI/O
│   ├── memory_collector.go         # メモリ・Swap
│   ├── disk_collector.go           # ディスク容量
│   ├── system_resources_collector.go # ファイルハンドル、ディスプレイ、Bluetooth
│   ├── user_activity_collector.go  # ユーザーアクティビティ
│   └── platform_collector_darwin.go # プラットフォーム情報
├── data-sender/
│   ├── types.go             # データ型定義
│   ├── processor.go         # データ処理・送信
│   └── ...
└── METRICS_REFERENCE.md     # このドキュメント
```

---

## 🎯 ROI分析への活用マッピング

| メトリックカテゴリ | ROI活用シナリオ |
|------------------|----------------|
| **Load Average** | オーバーワークの定量化、リソース需要予測 |
| **WiFi品質** | リモートワーク環境評価、接続問題による生産性低下検出 |
| **Swap/メモリプレッシャー** | ハードウェアアップグレード提案の根拠 |
| **ディスク容量** | ストレージ不足アラート、計画的な拡張 |
| **ミーティング検出** | 会議時間の可視化、会議過多の検出 |
| **ブラウザタブ数** | マルチタスキング傾向分析 |
| **フォーカススコア** | 生産性スコアリング、チーム比較 |
| **外部ディスプレイ** | 作業環境の充実度評価、設備投資効果測定 |
| **アプリ使用時間** | 業務アプリケーション利用率、ツール採用率 |

---

## 📝 更新履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2024-12-09 | 1.0.0 | 初版作成 - datadog-agent参考のメトリック追加 |

---

## 🔗 参考リンク

- [Datadog Agent GitHub](https://github.com/DataDog/datadog-agent)
- [macOS sysctl リファレンス](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man3/sysctl.3.html)
- [Apple Core WLAN Framework](https://developer.apple.com/documentation/corewlan)

