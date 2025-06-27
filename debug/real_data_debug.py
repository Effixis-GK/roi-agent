#!/usr/bin/env python3
"""
ROI Agent Enhanced - Real Data Debug Tools
実データ専用デバッグ・検証ツール
"""

import json
import os
import subprocess
import sys
import time
import socket
from datetime import datetime, timedelta
from pathlib import Path

# Configuration
HOME_DIR = os.path.expanduser("~")
USER_DATA_DIR = os.path.join(HOME_DIR, ".roiagent")
DATA_DIR = os.path.join(USER_DATA_DIR, "data")
LOGS_DIR = os.path.join(USER_DATA_DIR, "logs")

class RealDataDebugTools:
    def __init__(self):
        self.data_dir = DATA_DIR
        self.logs_dir = LOGS_DIR
        
    def verify_real_data_collection(self):
        """実データ収集の検証"""
        print("=== 実データ収集検証 ===")
        
        today = datetime.now().strftime("%Y-%m-%d")
        data_file = os.path.join(self.data_dir, f"combined_{today}.json")
        
        if not os.path.exists(data_file):
            print("❌ 今日の実データファイルが存在しません")
            print(f"   期待されるファイル: {data_file}")
            print("   エージェントが動作していない可能性があります")
            return False
        
        try:
            with open(data_file, 'r') as f:
                data = json.load(f)
            
            print(f"✅ 実データファイル発見: {data_file}")
            
            # データの内容を詳細分析
            apps = data.get('apps', {})
            network = data.get('network', {})
            
            print(f"📱 収集されたアプリ数: {len(apps)}")
            print(f"🌐 収集されたネットワーク接続数: {len(network)}")
            
            if len(apps) > 0:
                print("\n📊 実際のアプリ使用データ:")
                for app_name, app_data in list(apps.items())[:5]:
                    fg_time = app_data.get('foreground_time', 0)
                    focus_time = app_data.get('focus_time', 0)
                    is_active = app_data.get('is_active', False)
                    print(f"   {app_name}: {fg_time}秒 (フォーカス: {focus_time}秒) {'🟢' if is_active else '🔴'}")
            
            if len(network) > 0:
                print("\n🌐 実際のネットワーク接続データ:")
                for conn_key, conn_data in list(network.items())[:5]:
                    domain = conn_data.get('domain', 'Unknown')
                    duration = conn_data.get('duration', 0)
                    app_name = conn_data.get('app_name', 'Unknown')
                    is_active = conn_data.get('is_active', False)
                    print(f"   {domain}: {duration}秒 ({app_name}) {'🟢' if is_active else '🔴'}")
            
            # データの新しさを確認
            file_mtime = os.path.getmtime(data_file)
            last_modified = datetime.fromtimestamp(file_mtime)
            time_diff = datetime.now() - last_modified
            
            print(f"\n⏰ データの最終更新: {last_modified.strftime('%H:%M:%S')}")
            print(f"   更新からの経過時間: {int(time_diff.total_seconds())}秒")
            
            if time_diff.total_seconds() < 30:
                print("✅ データは新鮮です（30秒以内）")
            elif time_diff.total_seconds() < 120:
                print("⚠️  データは少し古いです（2分以内）")
            else:
                print("❌ データが古すぎます - エージェントが停止している可能性")
            
            return True
            
        except Exception as e:
            print(f"❌ データファイル読み込みエラー: {e}")
            return False
    
    def monitor_real_time_data(self):
        """リアルタイムデータ監視"""
        print("=== リアルタイム実データ監視 ===")
        print("15秒間隔でデータの変化を監視します...")
        print("Ctrl+C で停止")
        
        today = datetime.now().strftime("%Y-%m-%d")
        data_file = os.path.join(self.data_dir, f"combined_{today}.json")
        
        prev_data = None
        
        try:
            while True:
                if os.path.exists(data_file):
                    try:
                        with open(data_file, 'r') as f:
                            current_data = json.load(f)
                        
                        current_time = datetime.now().strftime("%H:%M:%S")
                        print(f"\n[{current_time}] データ更新チェック:")
                        
                        apps = current_data.get('apps', {})
                        network = current_data.get('network', {})
                        
                        print(f"  アプリ: {len(apps)}個, ネットワーク: {len(network)}個")
                        
                        # 新しいアプリや接続を検出
                        if prev_data:
                            prev_apps = set(prev_data.get('apps', {}).keys())
                            prev_network = set(prev_data.get('network', {}).keys())
                            
                            current_apps = set(apps.keys())
                            current_network = set(network.keys())
                            
                            new_apps = current_apps - prev_apps
                            new_connections = current_network - prev_network
                            
                            if new_apps:
                                print(f"  🆕 新しいアプリ: {', '.join(new_apps)}")
                            
                            if new_connections:
                                print(f"  🆕 新しい接続: {', '.join(list(new_connections)[:3])}")
                        
                        # アクティブなアプリを表示
                        active_apps = [name for name, data in apps.items() if data.get('is_active', False)]
                        if active_apps:
                            print(f"  🟢 アクティブアプリ: {', '.join(active_apps[:3])}")
                        
                        # フォーカス中のアプリを表示
                        focused_apps = [name for name, data in apps.items() if data.get('is_focused', False)]
                        if focused_apps:
                            print(f"  🎯 フォーカス中: {', '.join(focused_apps)}")
                        
                        prev_data = current_data
                        
                    except Exception as e:
                        print(f"  ❌ データ読み込みエラー: {e}")
                else:
                    print(f"  ⚠️  データファイルが存在しません: {data_file}")
                
                time.sleep(15)
                
        except KeyboardInterrupt:
            print("\n\n監視を停止しました")
    
    def test_current_system_state(self):
        """現在のシステム状態テスト"""
        print("=== 現在のシステム状態テスト ===")
        
        # 現在動作中のアプリケーション
        print("📱 現在動作中のアプリケーション:")
        try:
            result = subprocess.run([
                "osascript", "-e",
                'tell application "System Events" to get name of every application process whose visible is true'
            ], capture_output=True, text=True, timeout=10)
            
            if result.returncode == 0:
                apps = result.stdout.strip().split(', ')
                for i, app in enumerate(apps[:10], 1):
                    print(f"   {i}. {app}")
                if len(apps) > 10:
                    print(f"   ... 他 {len(apps) - 10}個")
            else:
                print("   ❌ アプリケーション取得失敗（アクセシビリティ権限が必要）")
        except Exception as e:
            print(f"   ❌ エラー: {e}")
        
        # フォーカス中のアプリケーション
        print("\n🎯 現在フォーカス中のアプリケーション:")
        try:
            result = subprocess.run([
                "osascript", "-e",
                'tell application "System Events" to get name of first application process whose frontmost is true'
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode == 0:
                focused_app = result.stdout.strip()
                print(f"   {focused_app}")
            else:
                print("   ❌ フォーカスアプリ取得失敗")
        except Exception as e:
            print(f"   ❌ エラー: {e}")
        
        # 現在のネットワーク接続
        print("\n🌐 現在のネットワーク接続:")
        try:
            result = subprocess.run(["netstat", "-an"], capture_output=True, text=True, timeout=10)
            if result.returncode == 0:
                lines = result.stdout.split('\n')
                established = [line for line in lines if 'ESTABLISHED' in line]
                print(f"   確立された接続数: {len(established)}")
                
                # HTTPSポート443の接続
                https_connections = [line for line in established if ':443' in line]
                print(f"   HTTPS接続 (ポート443): {len(https_connections)}")
                
                # サンプル表示
                for line in established[:5]:
                    parts = line.split()
                    if len(parts) >= 5:
                        foreign_addr = parts[4]
                        print(f"     {foreign_addr}")
            else:
                print("   ❌ netstat実行失敗")
        except Exception as e:
            print(f"   ❌ エラー: {e}")
        
        # lsofによる詳細ネットワーク情報
        print("\n🔍 詳細ネットワーク情報 (lsof):")
        try:
            result = subprocess.run(["lsof", "-i", "-n", "-P"], capture_output=True, text=True, timeout=15)
            if result.returncode == 0:
                lines = result.stdout.split('\n')
                tcp_lines = [line for line in lines if 'TCP' in line and ('443' in line or '80' in line)]
                
                print(f"   HTTP/HTTPS関連接続: {len(tcp_lines)}")
                
                for line in tcp_lines[:5]:
                    fields = line.split()
                    if len(fields) >= 9:
                        command = fields[0]
                        node = fields[8]
                        print(f"     {command}: {node}")
            else:
                print("   ⚠️  lsof実行失敗（管理者権限が必要な場合があります）")
        except Exception as e:
            print(f"   ❌ エラー: {e}")
    
    def analyze_collected_data(self):
        """収集済みデータの分析"""
        print("=== 収集済み実データ分析 ===")
        
        today = datetime.now().strftime("%Y-%m-%d")
        data_file = os.path.join(self.data_dir, f"combined_{today}.json")
        
        if not os.path.exists(data_file):
            print("❌ 今日のデータファイルが存在しません")
            return
        
        try:
            with open(data_file, 'r') as f:
                data = json.load(f)
            
            apps = data.get('apps', {})
            network = data.get('network', {})
            app_total = data.get('app_total', {})
            network_total = data.get('network_total', {})
            
            print("📊 アプリケーション使用統計:")
            print(f"   総フォアグラウンド時間: {app_total.get('foreground_time', 0)}秒")
            print(f"   総フォーカス時間: {app_total.get('focus_time', 0)}秒")
            print(f"   監視中アプリ数: {len(apps)}")
            
            # 使用時間トップ5
            if apps:
                sorted_apps = sorted(apps.items(), key=lambda x: x[1].get('foreground_time', 0), reverse=True)
                print("\n   📱 使用時間トップ5:")
                for i, (app_name, app_data) in enumerate(sorted_apps[:5], 1):
                    fg_time = app_data.get('foreground_time', 0)
                    minutes = fg_time // 60
                    seconds = fg_time % 60
                    print(f"     {i}. {app_name}: {minutes}分{seconds}秒")
            
            print(f"\n🌐 ネットワーク接続統計:")
            print(f"   総接続時間: {network_total.get('total_duration', 0)}秒")
            print(f"   送信データ: {network_total.get('total_bytes_sent', 0)} bytes")
            print(f"   受信データ: {network_total.get('total_bytes_received', 0)} bytes")
            print(f"   ユニーク接続数: {network_total.get('unique_connections', 0)}")
            
            # ネットワーク接続トップ5
            if network:
                sorted_network = sorted(network.items(), key=lambda x: x[1].get('duration', 0), reverse=True)
                print("\n   🌐 接続時間トップ5:")
                for i, (conn_key, conn_data) in enumerate(sorted_network[:5], 1):
                    domain = conn_data.get('domain', 'Unknown')
                    duration = conn_data.get('duration', 0)
                    app_name = conn_data.get('app_name', 'Unknown')
                    minutes = duration // 60
                    seconds = duration % 60
                    print(f"     {i}. {domain}: {minutes}分{seconds}秒 ({app_name})")
            
            # データ品質チェック
            print(f"\n🔍 データ品質チェック:")
            
            # アクティブなアプリの確認
            active_apps = sum(1 for app_data in apps.values() if app_data.get('is_active', False))
            print(f"   現在アクティブなアプリ: {active_apps}個")
            
            # アクティブなネットワーク接続の確認
            active_connections = sum(1 for conn_data in network.values() if conn_data.get('is_active', False))
            print(f"   現在アクティブな接続: {active_connections}個")
            
            # データの整合性チェック
            total_fg_time = sum(app_data.get('foreground_time', 0) for app_data in apps.values())
            recorded_total = app_total.get('foreground_time', 0)
            
            if abs(total_fg_time - recorded_total) < 60:  # 1分の誤差は許容
                print("   ✅ アプリデータ整合性: OK")
            else:
                print(f"   ⚠️  アプリデータ整合性: 計算値{total_fg_time}秒 vs 記録値{recorded_total}秒")
            
        except Exception as e:
            print(f"❌ データ分析エラー: {e}")
    
    def check_agent_logs(self):
        """エージェントログの確認"""
        print("=== エージェントログ確認 ===")
        
        if not os.path.exists(self.logs_dir):
            print("❌ ログディレクトリが存在しません")
            return
        
        # 最新のエージェントログファイルを探す
        agent_logs = []
        for filename in os.listdir(self.logs_dir):
            if filename.startswith('agent_') and filename.endswith('.log'):
                filepath = os.path.join(self.logs_dir, filename)
                mtime = os.path.getmtime(filepath)
                agent_logs.append((filepath, mtime))
        
        if not agent_logs:
            print("❌ エージェントログファイルが見つかりません")
            return
        
        # 最新のログファイル
        latest_log = sorted(agent_logs, key=lambda x: x[1], reverse=True)[0][0]
        print(f"📄 最新のエージェントログ: {latest_log}")
        
        try:
            with open(latest_log, 'r') as f:
                lines = f.readlines()
            
            print(f"   ログ行数: {len(lines)}")
            
            # 最新の10行を表示
            print("\n   📋 最新のログ (最後の10行):")
            for line in lines[-10:]:
                line = line.strip()
                if line:
                    timestamp = datetime.now().strftime("%H:%M:%S")
                    print(f"     {line}")
            
            # エラーを探す
            error_lines = [line for line in lines if 'error' in line.lower() or 'failed' in line.lower()]
            if error_lines:
                print(f"\n   ⚠️  エラーメッセージ ({len(error_lines)}件):")
                for error_line in error_lines[-5:]:  # 最新の5件
                    print(f"     {error_line.strip()}")
            else:
                print("\n   ✅ エラーメッセージは見つかりませんでした")
            
            # 成功メッセージを探す
            success_patterns = ['Updated app data', 'Updated network data', 'Saved combined data']
            recent_success = []
            for line in lines[-20:]:  # 最新の20行から
                for pattern in success_patterns:
                    if pattern in line:
                        recent_success.append(line.strip())
            
            if recent_success:
                print(f"\n   ✅ 最近の成功メッセージ:")
                for success_line in recent_success[-3:]:  # 最新の3件
                    print(f"     {success_line}")
        
        except Exception as e:
            print(f"   ❌ ログ読み込みエラー: {e}")
    
    def real_data_full_diagnosis(self):
        """実データの完全診断"""
        print("=== ROI Agent Enhanced 実データ完全診断 ===\n")
        
        # 1. システム状態
        self.test_current_system_state()
        
        print("\n" + "="*50)
        
        # 2. データ収集検証
        self.verify_real_data_collection()
        
        print("\n" + "="*50)
        
        # 3. 収集データ分析
        self.analyze_collected_data()
        
        print("\n" + "="*50)
        
        # 4. ログ確認
        self.check_agent_logs()
        
        print("\n" + "="*50)
        
        # 5. 総合評価
        print("\n=== 総合評価 ===")
        
        issues = []
        successes = []
        
        # データファイル存在チェック
        today = datetime.now().strftime("%Y-%m-%d")
        data_file = os.path.join(self.data_dir, f"combined_{today}.json")
        
        if os.path.exists(data_file):
            successes.append("✅ 実データファイルが存在")
            
            try:
                with open(data_file, 'r') as f:
                    data = json.load(f)
                
                apps = data.get('apps', {})
                network = data.get('network', {})
                
                if len(apps) > 0:
                    successes.append(f"✅ アプリデータ収集中 ({len(apps)}個)")
                else:
                    issues.append("❌ アプリデータが収集されていません")
                
                if len(network) > 0:
                    successes.append(f"✅ ネットワークデータ収集中 ({len(network)}個)")
                else:
                    issues.append("⚠️  ネットワークデータが少ないか未収集")
                
                # データの新しさチェック
                file_mtime = os.path.getmtime(data_file)
                time_diff = time.time() - file_mtime
                
                if time_diff < 30:
                    successes.append("✅ データは最新（30秒以内）")
                elif time_diff < 120:
                    issues.append("⚠️  データがやや古い（2分以内）")
                else:
                    issues.append("❌ データが古すぎる（2分超過）")
                    
            except Exception as e:
                issues.append(f"❌ データファイル読み込みエラー: {e}")
        else:
            issues.append("❌ 実データファイルが存在しません")
        
        # 結果表示
        if successes:
            print("成功項目:")
            for success in successes:
                print(f"  {success}")
        
        if issues:
            print("\n問題・改善点:")
            for issue in issues:
                print(f"  {issue}")
        
        print(f"\n📊 診断結果: {len(successes)}個成功, {len(issues)}個の問題")
        
        # 推奨アクション
        print("\n🎯 推奨アクション:")
        if not os.path.exists(data_file):
            print("  1. エージェントを起動: ./start_real_data_mode.sh")
        elif len(issues) > len(successes):
            print("  1. エージェントを再起動")
            print("  2. アクセシビリティ権限を確認")
            print("  3. ログを詳細確認: tail -f ~/.roiagent/logs/agent_*.log")
        else:
            print("  ✅ 正常に動作しています")
            print("  📊 ダッシュボードで結果を確認: http://localhost:5002")

def main():
    if len(sys.argv) < 2:
        print("ROI Agent Enhanced - Real Data Debug Tools")
        print("")
        print("使用方法:")
        print("  python3 real_data_debug.py [command]")
        print("")
        print("コマンド:")
        print("  verify          - 実データ収集の検証")
        print("  monitor         - リアルタイムデータ監視")
        print("  system          - 現在のシステム状態テスト")
        print("  analyze         - 収集済みデータの分析")
        print("  logs            - エージェントログ確認")
        print("  full            - 完全診断実行")
        print("")
        return
    
    tools = RealDataDebugTools()
    command = sys.argv[1].lower()
    
    if command == "verify":
        tools.verify_real_data_collection()
    elif command == "monitor":
        tools.monitor_real_time_data()
    elif command == "system":
        tools.test_current_system_state()
    elif command == "analyze":
        tools.analyze_collected_data()
    elif command == "logs":
        tools.check_agent_logs()
    elif command == "full":
        tools.real_data_full_diagnosis()
    else:
        print(f"Unknown command: {command}")
        main()

if __name__ == "__main__":
    main()
