package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HistoryEntry represents a Chrome history entry
type HistoryEntry struct {
	ID          int
	URL         string
	Title       string
	VisitCount  int
	LastVisited time.Time
}

// ChromeHistoryReader reads Chrome browser history
type ChromeHistoryReader struct {
	historyPath string
}

// NewChromeHistoryReader creates a new Chrome history reader
func NewChromeHistoryReader() *ChromeHistoryReader {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	// MacOS Chrome history path
	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "History")
	
	return &ChromeHistoryReader{
		historyPath: historyPath,
	}
}

// GetHistory retrieves Chrome browsing history
func (chr *ChromeHistoryReader) GetHistory(limit int) ([]HistoryEntry, error) {
	// Chrome locks the database when running, so we need to copy it first
	tempPath := "/tmp/chrome_history_copy.db"
	if err := chr.copyHistoryFile(tempPath); err != nil {
		return nil, fmt.Errorf("failed to copy history file: %v", err)
	}
	defer os.Remove(tempPath)

	// Open the copied database
	db, err := sql.Open("sqlite3", tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %v", err)
	}
	defer db.Close()

	// Query history
	query := `
		SELECT id, url, title, visit_count, last_visit_time
		FROM urls 
		ORDER BY last_visit_time DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %v", err)
	}
	defer rows.Close()

	var history []HistoryEntry
	for rows.Next() {
		var entry HistoryEntry
		var lastVisitTime int64

		err := rows.Scan(&entry.ID, &entry.URL, &entry.Title, &entry.VisitCount, &lastVisitTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert Chrome timestamp to Go time
		// Chrome uses WebKit timestamp: microseconds since 1601-01-01 00:00:00 UTC
		// But we need to be careful about the conversion
		if lastVisitTime > 0 {
			// Chrome's WebKit timestamp is microseconds since January 1, 1601 (Windows FILETIME)
			// Convert to Unix timestamp (seconds since January 1, 1970)
			const webkitToUnixOffset = 11644473600000000 // microseconds between 1601 and 1970
			unixMicroseconds := lastVisitTime - webkitToUnixOffset
			if unixMicroseconds > 0 {
				entry.LastVisited = time.Unix(unixMicroseconds/1000000, (unixMicroseconds%1000000)*1000)
			} else {
				entry.LastVisited = time.Time{} // Invalid timestamp
			}
		} else {
			entry.LastVisited = time.Time{} // Zero time for invalid timestamps
		}

		history = append(history, entry)
	}

	return history, nil
}

// copyHistoryFile copies the Chrome history file to a temporary location
func (chr *ChromeHistoryReader) copyHistoryFile(destPath string) error {
	// Check if source file exists
	if _, err := os.Stat(chr.historyPath); os.IsNotExist(err) {
		return fmt.Errorf("Chrome history file not found at: %s", chr.historyPath)
	}

	// Read source file
	sourceData, err := os.ReadFile(chr.historyPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	// Write to destination
	err = os.WriteFile(destPath, sourceData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file: %v", err)
	}

	return nil
}

// GetTopDomains retrieves the most visited domains
func (chr *ChromeHistoryReader) GetTopDomains(limit int) ([]DomainStats, error) {
	tempPath := "/tmp/chrome_history_copy.db"
	if err := chr.copyHistoryFile(tempPath); err != nil {
		return nil, fmt.Errorf("failed to copy history file: %v", err)
	}
	defer os.Remove(tempPath)

	db, err := sql.Open("sqlite3", tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %v", err)
	}
	defer db.Close()

	query := `
		SELECT 
			substr(url, 1, instr(substr(url, 8), '/') + 6) as domain,
			SUM(visit_count) as total_visits,
			COUNT(*) as url_count
		FROM urls 
		WHERE url LIKE 'http%'
		GROUP BY domain
		ORDER BY total_visits DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains: %v", err)
	}
	defer rows.Close()

	var domains []DomainStats
	for rows.Next() {
		var domain DomainStats
		err := rows.Scan(&domain.Domain, &domain.TotalVisits, &domain.URLCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// DomainStats represents statistics for a domain
type DomainStats struct {
	Domain      string
	TotalVisits int
	URLCount    int
}

// PrintHistory prints the browsing history in a formatted way
func PrintHistory(history []HistoryEntry) {
	fmt.Println("=== Chrome ブラウジング履歴 ===")
	fmt.Printf("%-5s %-50s %-8s %-20s %s\n", "ID", "URL", "訪問回数", "最終訪問日時", "タイトル")
	fmt.Println(strings.Repeat("-", 120))

	for _, entry := range history {
		title := entry.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		
		url := entry.URL
		if len(url) > 47 {
			url = url[:44] + "..."
		}

		fmt.Printf("%-5d %-50s %-8d %-20s %s\n", 
			entry.ID, 
			url, 
			entry.VisitCount,
			entry.LastVisited.Format("2006/01/02 15:04:05"),
			title)
	}
}

// PrintTopDomains prints the top visited domains
func PrintTopDomains(domains []DomainStats) {
	fmt.Println("\n=== 最もアクセス頻度の高いドメイン ===")
	fmt.Printf("%-40s %-12s %-10s\n", "ドメイン", "総訪問回数", "URL数")
	fmt.Println(strings.Repeat("-", 70))

	for _, domain := range domains {
		fmt.Printf("%-40s %-12d %-10d\n", 
			domain.Domain, 
			domain.TotalVisits, 
			domain.URLCount)
	}
}

func main() {
	fmt.Println("=== Chrome Browser Access Tool ===")
	fmt.Println("プログラムを開始します...")
	
	reader := NewChromeHistoryReader()
	
	fmt.Printf("履歴ファイルパス: %s\n", reader.historyPath)

	// Get recent browsing history
	fmt.Println("\nChromeブラウジング履歴を読み込み中...")
	history, err := reader.GetHistory(20)
	if err != nil {
		log.Printf("履歴の読み込みエラー: %v\n", err)
		fmt.Println("エラーの詳細:")
		fmt.Println("- Chromeが起動中の場合は一度終了してから再試行してください")
		fmt.Println("- または、Chromeの履歴ファイルへのアクセス権限を確認してください")
		return
	}

	if len(history) == 0 {
		fmt.Println("履歴が見つかりませんでした。")
		fmt.Println("- Chromeを使用してウェブサイトを訪問してから再試行してください")
		return
	}

	fmt.Printf("履歴の読み込み成功: %d件\n", len(history))
	PrintHistory(history)

	// Get top visited domains
	fmt.Println("\n最もアクセス頻度の高いドメインを取得中...")
	domains, err := reader.GetTopDomains(10)
	if err != nil {
		log.Printf("ドメイン情報の読み込みエラー: %v\n", err)
		return
	}

	PrintTopDomains(domains)

	fmt.Printf("\n✅ 処理完了: 合計 %d件の履歴エントリを取得しました\n", len(history))
}
