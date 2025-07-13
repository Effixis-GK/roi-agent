#!/usr/bin/env python3
"""
ROI Agent Enhanced - Network FQDN Debug Tools
実際のパケットキャプチャとFQDN解決のデバッグツール
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

class NetworkFQDNDebugTools:
    def __init__(self):
        self.data_dir = DATA_DIR
        self.logs_dir = LOGS_DIR
        
    def test_fqdn_resolution(self):
        """FQDN解決テスト"""
        print("=== FQDN解決テスト ===")
        
        # よく知られたIPアドレスでテスト
        test_ips = [
            ("140.82.112.4", "GitHub"),
            ("172.217.14.196", "Google"),
            ("13.107.42.14", "Microsoft"),
            ("151.101.1.140", "Reddit"),
            ("104.16.123.96", "Cloudflare"),
        ]
        
        successful_resolutions = 0
        
        for ip, expected_service in test_ips:
            try:
                start_time = time.time()
                hostnames = socket.gethostbyaddr(ip)
                resolution_time = time.time() - start_time
                
                primary_hostname = hostnames[0]
                print(f"✅ {ip} -> {primary_hostname} ({expected_service}) ({resolution_time:.3f}s)")
                successful_resolutions += 1
                
                # 追加のホスト名がある場合は表示
                if len(hostnames[1]) > 1:
                    print(f"   追加: {', '.join(hostnames[1][:3])}")
                    
            except socket.herror as e:
                print(f"❌ {ip} -> 解決失敗 ({expected_service}): {e}")
            except Exception as e:
                print(f"❌ {ip} -> エラー ({expected_service}): {e}")
        
        print(f"\n📊 解決成功率: {successful_resolutions}/{len(test_ips)} ({successful_resolutions/len(test_ips)*100:.1f}%)")
        
        return successful_resolutions > 0
    
    def test_current_connections(self):
        """現在のネットワーク接続をテスト"""
        print("=== 現在のネットワーク接続テスト ===")
        
        # lsofを使って実際の接続を確認
        try:
            result = subprocess.run(["lsof", "-i", "-n", "-P"], capture_output=True, text=True, timeout=15)
            if result.returncode == 0:
                lines = result.stdout.split('\n')
                
                # HTTP/HTTPS接続を抽出
                http_connections = []
                for line in lines:
                    if 'TCP' in line and ('->127.0.0.1' not in line) and ('ESTABLISHED' in line or 'SYN_SENT' in line):
                        fields = line.split()
                        if len(fields) >= 9:
                            command = fields[0]
                            node = fields[8]
                            
                            # ポート443または80を含む接続を探す
                            if ':443' in node or ':80' in node or ':8080' in node:
                                http_connections.append({
                                    'app': command,
                                    'connection': node,
                                    'full_line': line.strip()
                                })
                
                print(f"📡 HTTP/HTTPS接続数: {len(http_connections)}")
                
                if http_connections:
                    print("\n🔍 詳細接続情報:")
                    for i, conn in enumerate(http_connections[:10], 1):  # 最初の10個
                        print(f"  {i}. {conn['app']}: {conn['connection']}")
                        
                        # IPアドレスを抽出してFQDN解決を試行
                        connection_parts = conn['connection'].split('->')
                        if len(connection_parts) == 2:
                            remote_part = connection_parts[1]
                            if ':' in remote_part:
                                ip_part = remote_part.split(':')[0]
                                try:
                                    # IPアドレスかどうか確認
                                    socket.inet_aton(ip_part)  # IPv4チェック
                                    
                                    # FQDN解決を試行
                                    try:
                                        hostname = socket.gethostbyaddr(ip_part)[0]
                                        print(f"     -> FQDN: {hostname}")
                                    except:
                                        print(f"     -> FQDN解決失敗: {ip_part}")
                                except:
                                    # IPv4でない場合はそのまま表示
                                    print(f"     -> ホスト: {ip_part}")
                else:
                    print("⚠️  HTTP/HTTPS接続が検出されませんでした")
                    print("   ブラウザでWebサイトにアクセスしてから再実行してください")
                
                return len(http_connections) > 0
            else:
                print("❌ lsofコマンド実行失敗")
                print("   管理者権限が必要な場合があります: sudo python3 network_fqdn_debug.py test-connections")
                return False
                
        except subprocess.TimeoutExpired:
            print("❌ lsofコマンドタイムアウト")
            return False
        except Exception as e:
            print(f"❌ 接続テストエラー: {e}")
            return False
    
    def test_dns_monitoring(self):
        """DNS監視機能のテスト"""
        print("=== DNS監視機能テスト ===")
        
        # システムログからDNSクエリを監視
        try:
            print("最近のDNSクエリを確認中...")
            result = subprocess.run([
                "log", "show", 
                "--predicate", "subsystem == 'com.apple.network.dnsproxy'", 
                "--style", "syslog", 
                "--last", "2m"
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                lines = result.stdout.split('\n')
                dns_queries = []
                
                for line in lines:
                    if 'query' in line.lower() or 'resolve' in line.lower():
                        # ドメイン名を抽出
                        words = line.split()
                        for word in words:
                            if ('.' in word and 
                                (word.endswith('.com') or word.endswith('.org') or 
                                 word.endswith('.net') or word.endswith('.io') or
                                 word.endswith('.co.jp'))):
                                domain = word.strip('.,()[]')
                                if len(domain) > 3:
                                    dns_queries.append(domain)
                
                unique_domains = list(set(dns_queries))
                print(f"✅ 最近のDNSクエリ: {len(unique_domains)}個のユニークドメイン")
                
                if unique_domains:
                    print("\n🔍 検出されたドメイン:")
                    for domain in unique_domains[:10]:
                        print(f"   {domain}")
                else:
                    print("⚠️  DNSクエリが検出されませんでした")
                    print("   ブラウザでWebサイトにアクセスしてから再実行してください")
                
                return len(unique_domains) > 0
            else:
                print("❌ DNSログ取得失敗")
                print("   権限不足またはシステム設定の問題の可能性があります")
                return False
                
        except subprocess.TimeoutExpired:
            print("❌ DNSログ取得タイムアウト")
            return False
        except Exception as e:
            print(f"❌ DNS監視テストエラー: {e}")
            return False
    
    def test_redirect_following(self):
        """HTTPリダイレクト追跡のテスト"""
        print("=== HTTPリダイレクト追跡テスト ===")
        
        # よく知られたリダイレクトするURLをテスト
        test_urls = [
            "http://github.com",
            "http://google.com", 
            "http://facebook.com",
            "http://twitter.com",
        ]
        
        successful_redirects = 0
        
        for url in test_urls:
            try:
                print(f"\n🔗 テスト中: {url}")
                
                import urllib.request
                import urllib.parse
                
                # カスタムHTTPハンドラーでリダイレクトを追跡
                class RedirectHandler(urllib.request.HTTPRedirectHandler):
                    def __init__(self):
                        self.redirects = []
                    
                    def http_error_302(self, req, fp, code, msg, headers):
                        location = headers.get('Location')
                        if location:
                            self.redirects.append(location)
                            print(f"   リダイレクト: {location}")
                        return super().http_error_302(req, fp, code, msg, headers)
                    
                    def http_error_301(self, req, fp, code, msg, headers):
                        location = headers.get('Location')
                        if location:
                            self.redirects.append(location)
                            print(f"   リダイレクト: {location}")
                        return super().http_error_301(req, fp, code, msg, headers)
                
                redirect_handler = RedirectHandler()
                opener = urllib.request.build_opener(redirect_handler)
                
                start_time = time.time()
                response = opener.open(url, timeout=10)
                response_time = time.time() - start_time
                
                final_url = response.url
                original_domain = urllib.parse.urlparse(url).netloc
                final_domain = urllib.parse.urlparse(final_url).netloc
                
                print(f"   最終URL: {final_url}")
                print(f"   応答時間: {response_time:.3f}秒")
                print(f"   ステータス: {response.status}")
                
                if original_domain != final_domain:
                    print(f"   ✅ リダイレクト検出: {original_domain} -> {final_domain}")
                    successful_redirects += 1
                else:
                    print(f"   📝 リダイレクトなし")
                    successful_redirects += 1
                
            except Exception as e:
                print(f"   ❌ エラー: {e}")
        
        print(f"\n📊 テスト成功率: {successful_redirects}/{len(test_urls)}")
        return successful_redirects > 0
    
    def verify_enhanced_agent_data(self):
        """拡張エージェントのデータ検証"""
        print("=== 拡張エージェントデータ検証 ===")
        
        today = datetime.now().strftime("%Y-%m-%d")
        data_file = os.path.join(self.data_dir, f"combined_{today}.json")
        
        if not os.path.exists(data_file):
            print("❌ 今日のデータファイルが存在しません")
            print("   拡張エージェントが動作していない可能性があります")
            return False
        
        try:
            with open(data_file, 'r') as f:
                data = json.load(f)
            
            print(f"✅ データファイル読み込み成功: {data_file}")
            
            # 拡張機能の確認
            network = data.get('network', {})
            dns_queries = data.get('dns_queries', [])
            http_transactions = data.get('http_transactions', [])
            
            print(f"📊 データ統計:")
            print(f"   ネットワーク接続: {len(network)}個")
            print(f"   DNSクエリ: {len(dns_queries)}個") 
            print(f"   HTTPトランザクション: {len(http_transactions)}個")
            
            # FQDN解決の確認
            fqdn_connections = 0
            ip_connections = 0
            
            print(f"\n🔍 ネットワーク接続詳細:")
            for key, conn in list(network.items())[:10]:
                domain = conn.get('domain', 'Unknown')
                remote_ip = conn.get('remote_ip', 'Unknown')
                app_name = conn.get('app_name', 'Unknown')
                original_domain = conn.get('original_domain', domain)
                
                print(f"   {domain}")
                print(f"     IP: {remote_ip}")
                print(f"     アプリ: {app_name}")
                
                if original_domain != domain:
                    print(f"     リダイレクト: {original_domain} -> {domain}")
                
                # FQDNか判定
                if '.' in domain and not domain.replace('.', '').isdigit():
                    fqdn_connections += 1
                else:
                    ip_connections += 1
            
            print(f"\n📈 FQDN解決統計:")
            print(f"   FQDN接続: {fqdn_connections}個")
            print(f"   IP接続: {ip_connections}個")
            
            if fqdn_connections > 0:
                print(f"   ✅ FQDN解決が動作しています")
                resolution_rate = fqdn_connections / (fqdn_connections + ip_connections) * 100
                print(f"   解決率: {resolution_rate:.1f}%")
            else:
                print(f"   ⚠️  FQDN解決が機能していません")
            
            # DNSクエリの確認
            if len(dns_queries) > 0:
                print(f"\n🔍 最近のDNSクエリ:")
                for query in dns_queries[-5:]:  # 最新の5個
                    domain = query.get('domain', 'Unknown')
                    timestamp = query.get('timestamp', 'Unknown')
                    print(f"   {domain} ({timestamp})")
            
            return len(network) > 0
            
        except Exception as e:
            print(f"❌ データ検証エラー: {e}")
            return False
    
    def test_enhanced_agent_binary(self):
        """拡張エージェントバイナリのテスト"""
        print("=== 拡張エージェントバイナリテスト ===")
        
        agent_path = "/Users/taktakeu/Local/GitHub/roi-agent/agent/enhanced_network_main.go"
        
        if not os.path.exists(agent_path):
            print(f"❌ 拡張エージェントソース未発見: {agent_path}")
            return False
        
        print(f"✅ 拡張エージェントソース発見: {agent_path}")
        
        # Go環境の確認
        try:
            result = subprocess.run(["go", "version"], capture_output=True, text=True)
            if result.returncode == 0:
                go_version = result.stdout.strip()
                print(f"✅ Go環境: {go_version}")
            else:
                print("❌ Go環境が見つかりません")
                return False
        except FileNotFoundError:
            print("❌ Goコマンドが見つかりません")
            return False
        
        # ビルドテスト
        try:
            print("🔨 ビルドテスト中...")
            agent_dir = os.path.dirname(agent_path)
            
            # go.modの確認
            go_mod_path = os.path.join(agent_dir, "go.mod")
            if not os.path.exists(go_mod_path):
                print("   go.modを作成中...")
                subprocess.run(["go", "mod", "init", "roi-agent-enhanced"], 
                             cwd=agent_dir, check=True)
            
            # ビルド実行
            result = subprocess.run([
                "go", "build", "-o", "test_enhanced_monitor", "enhanced_network_main.go"
            ], cwd=agent_dir, capture_output=True, text=True)
            
            if result.returncode == 0:
                print("✅ ビルド成功")
                
                # バイナリテスト
                test_binary = os.path.join(agent_dir, "test_enhanced_monitor")
                if os.path.exists(test_binary):
                    print("🧪 機能テスト中...")
                    
                    # FQDN解決テスト
                    result = subprocess.run([test_binary, "test-fqdn"], 
                                          capture_output=True, text=True, timeout=30)
                    if result.returncode == 0:
                        print("✅ FQDN解決テスト成功")
                        print(result.stdout)
                    else:
                        print(f"⚠️  FQDN解決テスト問題: {result.stderr}")
                    
                    # 接続テスト
                    result = subprocess.run([test_binary, "test-connections"], 
                                          capture_output=True, text=True, timeout=30)
                    if result.returncode == 0:
                        print("✅ 接続テスト成功")
                        print(result.stdout)
                    else:
                        print(f"⚠️  接続テスト問題: {result.stderr}")
                    
                    # クリーンアップ
                    os.remove(test_binary)
                    
                return True
            else:
                print(f"❌ ビルド失敗:")
                print(result.stderr)
                return False
                
        except Exception as e:
            print(f"❌ ビルドテストエラー: {e}")
            return False
    
    def comprehensive_network_diagnosis(self):
        """包括的ネットワーク診断"""
        print("=== ROI Agent Enhanced - 包括的ネットワーク診断 ===\n")
        
        results = {}
        
        # 1. FQDN解決テスト
        print("1. FQDN解決テスト")
        results['fqdn_resolution'] = self.test_fqdn_resolution()
        
        print("\n" + "="*50 + "\n")
        
        # 2. 現在の接続テスト
        print("2. 現在のネットワーク接続テスト")
        results['current_connections'] = self.test_current_connections()
        
        print("\n" + "="*50 + "\n")
        
        # 3. DNS監視テスト
        print("3. DNS監視機能テスト")
        results['dns_monitoring'] = self.test_dns_monitoring()
        
        print("\n" + "="*50 + "\n")
        
        # 4. リダイレクト追跡テスト
        print("4. HTTPリダイレクト追跡テスト")
        results['redirect_following'] = self.test_redirect_following()
        
        print("\n" + "="*50 + "\n")
        
        # 5. 拡張エージェントバイナリテスト
        print("5. 拡張エージェントバイナリテスト")
        results['enhanced_agent'] = self.test_enhanced_agent_binary()
        
        print("\n" + "="*50 + "\n")
        
        # 6. 拡張エージェントデータ検証
        print("6. 拡張エージェントデータ検証")
        results['enhanced_data'] = self.verify_enhanced_agent_data()
        
        print("\n" + "="*50 + "\n")
        
        # 総合評価
        print("=== 総合評価 ===")
        
        passed_tests = sum(1 for result in results.values() if result)
        total_tests = len(results)
        
        print(f"テスト結果: {passed_tests}/{total_tests} 成功")
        
        if passed_tests == total_tests:
            print("🎉 すべてのテストが成功しました！")
            print("   拡張ネットワーク監視機能は正常に動作しています")
        elif passed_tests >= total_tests * 0.8:
            print("✅ ほとんどのテストが成功しました")
            print("   軽微な問題はありますが、基本機能は動作しています")
        elif passed_tests >= total_tests * 0.5:
            print("⚠️  一部のテストが失敗しました")
            print("   設定や権限の確認が必要です")
        else:
            print("❌ 多くのテストが失敗しました")
            print("   システム設定や権限の大幅な見直しが必要です")
        
        print(f"\n📋 詳細結果:")
        for test_name, result in results.items():
            status = "✅" if result else "❌"
            print(f"   {status} {test_name}")
        
        # 推奨アクション
        print(f"\n🎯 推奨アクション:")
        
        if not results.get('enhanced_agent', False):
            print("   1. 拡張エージェントをビルド: cd agent && go build enhanced_network_main.go")
        
        if not results.get('current_connections', False):
            print("   2. lsof権限確認: sudo権限でテスト実行")
            
        if not results.get('dns_monitoring', False):
            print("   3. DNS監視権限確認: システムログアクセス権限")
        
        if results.get('enhanced_agent', False) and not results.get('enhanced_data', False):
            print("   4. 拡張エージェント起動: ./enhanced_monitor &")
        
        if passed_tests >= total_tests * 0.8:
            print("   ✅ 準備完了: 実際のネットワーク監視を開始できます")

def main():
    if len(sys.argv) < 2:
        print("ROI Agent Enhanced - DNS Snooping Debug Tools")
        print("")
        print("使用方法:")
        print("  python3 network_fqdn_debug.py [command]")
        print("")
        print("コマンド:")
        print("  fqdn              - FQDN解決テスト")
        print("  dns               - DNS Snooping機能テスト (要sudo)")
        print("  connections       - DNS Snooping接続テスト (要sudo)")
        print("  redirects         - HTTPリダイレクト追跡テスト")
        print("  agent             - 拡張エージェントバイナリテスト")
        print("  verify            - 拡張エージェントデータ検証")
        print("  full              - 包括的DNS Snooping診断 (要sudo)")
        print("")
        print("例:")
        print("  sudo python3 network_fqdn_debug.py full")
        print("  sudo python3 network_fqdn_debug.py dns")
        print("  python3 network_fqdn_debug.py fqdn")
        print("")
        print("注意: DNS Snooping機能にはsudo権限が必要です")
        print("")
        return
    
    tools = NetworkFQDNDebugTools()
    command = sys.argv[1].lower()
    
    if command == "fqdn":
        tools.test_fqdn_resolution()
    elif command == "connections":
        tools.test_current_connections()
    elif command == "dns":
        tools.test_dns_snooping()
    elif command == "redirects":
        tools.test_redirect_following()
    elif command == "agent":
        tools.test_enhanced_agent_binary()
    elif command == "verify":
        tools.verify_enhanced_agent_data()
    elif command == "full":
        tools.comprehensive_network_diagnosis()
    else:
        print(f"Unknown command: {command}")
        main()

if __name__ == "__main__":
    main()
