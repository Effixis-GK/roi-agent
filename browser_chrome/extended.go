package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BookmarkEntry represents a Chrome bookmark
type BookmarkEntry struct {
	ID       int
	Title    string
	URL      string
	DateAdded time.Time
}

// Download represents a Chrome download entry
type DownloadEntry struct {
	ID           int
	TargetPath   string
	URL          string
	StartTime    time.Time
	TotalBytes   int64
	ReceivedBytes int64
}

// SearchResult represents a search query from history
type SearchResult struct {
	Query     string
	Count     int
	LastUsed  time.Time
}

// ExtendedChromeReader provides additional Chrome data reading capabilities
type ExtendedChromeReader struct {
	*ChromeHistoryReader
	profilePath string
}

// NewExtendedChromeReader creates a new extended Chrome reader
func NewExtendedChromeReader() *ExtendedChromeReader {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user home directory: " + err.Error())
	}

	profilePath := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default")
	
	return &ExtendedChromeReader{
		ChromeHistoryReader: NewChromeHistoryReader(),
		profilePath:         profilePath,
	}
}

// GetDownloads retrieves Chrome download history
func (ecr *ExtendedChromeReader) GetDownloads(limit int) ([]DownloadEntry, error) {
	historyPath := filepath.Join(ecr.profilePath, "History")
	tempPath := "/tmp/chrome_downloads_copy.db"
	
	if err := ecr.copyFile(historyPath, tempPath); err != nil {
		return nil, fmt.Errorf("failed to copy history file: %v", err)
	}
	defer os.Remove(tempPath)

	db, err := sql.Open("sqlite3", tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open downloads database: %v", err)
	}
	defer db.Close()

	query := `
		SELECT id, target_path, url, start_time, total_bytes, received_bytes
		FROM downloads 
		ORDER BY start_time DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query downloads: %v", err)
	}
	defer rows.Close()

	var downloads []DownloadEntry
	for rows.Next() {
		var entry DownloadEntry
		var startTime int64

		err := rows.Scan(&entry.ID, &entry.TargetPath, &entry.URL, &startTime, &entry.TotalBytes, &entry.ReceivedBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan download row: %v", err)
		}

		// Convert Chrome timestamp
		if startTime > 0 {
			// Chrome's WebKit timestamp: microseconds since 1601-01-01
			const webkitToUnixOffset = 11644473600000000 // microseconds between 1601 and 1970
			unixMicroseconds := startTime - webkitToUnixOffset
			if unixMicroseconds > 0 {
				entry.StartTime = time.Unix(unixMicroseconds/1000000, (unixMicroseconds%1000000)*1000)
			} else {
				entry.StartTime = time.Time{} // Invalid timestamp
			}
		} else {
			entry.StartTime = time.Time{} // Zero time for invalid timestamps
		}

		downloads = append(downloads, entry)
	}

	return downloads, nil
}

// GetSearchQueries extracts search queries from browsing history
func (ecr *ExtendedChromeReader) GetSearchQueries(limit int) ([]SearchResult, error) {
	history, err := ecr.GetHistory(1000) // Get more history to analyze
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %v", err)
	}

	searchQueries := make(map[string]*SearchResult)
	
	for _, entry := range history {
		query := extractSearchQuery(entry.URL)
		if query != "" {
			if existing, exists := searchQueries[query]; exists {
				existing.Count++
				if entry.LastVisited.After(existing.LastUsed) {
					existing.LastUsed = entry.LastVisited
				}
			} else {
				searchQueries[query] = &SearchResult{
					Query:    query,
					Count:    1,
					LastUsed: entry.LastVisited,
				}
			}
		}
	}

	// Convert map to slice and sort by count
	var results []SearchResult
	for _, result := range searchQueries {
		results = append(results, *result)
	}

	// Simple sorting by count (descending)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Count > results[i].Count {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// extractSearchQuery extracts search query from URL
func extractSearchQuery(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Google search
	if strings.Contains(u.Host, "google.") {
		return u.Query().Get("q")
	}
	
	// Bing search
	if strings.Contains(u.Host, "bing.com") {
		return u.Query().Get("q")
	}
	
	// Yahoo search
	if strings.Contains(u.Host, "yahoo.com") || strings.Contains(u.Host, "yahoo.co.jp") {
		return u.Query().Get("p")
	}
	
	// DuckDuckGo
	if strings.Contains(u.Host, "duckduckgo.com") {
		return u.Query().Get("q")
	}

	return ""
}

// GetBrowsingPatterns analyzes browsing patterns by hour of day
func (ecr *ExtendedChromeReader) GetBrowsingPatterns() (map[int]int, error) {
	history, err := ecr.GetHistory(1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %v", err)
	}

	hourlyPattern := make(map[int]int)
	
	for _, entry := range history {
		hour := entry.LastVisited.Hour()
		hourlyPattern[hour]++
	}

	return hourlyPattern, nil
}

// copyFile copies a file from source to destination
func (ecr *ExtendedChromeReader) copyFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found at: %s", src)
	}

	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	err = os.WriteFile(dst, sourceData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file: %v", err)
	}

	return nil
}

// PrintDownloads prints download history in a formatted way
func PrintDownloads(downloads []DownloadEntry) {
	fmt.Println("\n=== Chrome ダウンロード履歴 ===")
	fmt.Printf("%-5s %-40s %-15s %-20s %s\n", "ID", "ファイル名", "サイズ (MB)", "ダウンロード日時", "URL")
	fmt.Println(strings.Repeat("-", 120))

	for _, entry := range downloads {
		fileName := filepath.Base(entry.TargetPath)
		if len(fileName) > 37 {
			fileName = fileName[:34] + "..."
		}
		
		url := entry.URL
		if len(url) > 40 {
			url = url[:37] + "..."
		}

		sizeMB := float64(entry.TotalBytes) / (1024 * 1024)

		fmt.Printf("%-5d %-40s %-15.2f %-20s %s\n", 
			entry.ID, 
			fileName, 
			sizeMB,
			entry.StartTime.Format("2006/01/02 15:04:05"),
			url)
	}
}

// PrintSearchQueries prints search queries in a formatted way
func PrintSearchQueries(queries []SearchResult) {
	fmt.Println("\n=== よく検索されるクエリ ===")
	fmt.Printf("%-50s %-8s %s\n", "検索クエリ", "回数", "最終使用日時")
	fmt.Println(strings.Repeat("-", 80))

	for _, query := range queries {
		queryText := query.Query
		if len(queryText) > 47 {
			queryText = queryText[:44] + "..."
		}

		fmt.Printf("%-50s %-8d %s\n", 
			queryText, 
			query.Count,
			query.LastUsed.Format("2006/01/02 15:04:05"))
	}
}

// PrintBrowsingPatterns prints hourly browsing patterns
func PrintBrowsingPatterns(patterns map[int]int) {
	fmt.Println("\n=== 時間別ブラウジングパターン ===")
	fmt.Println("時間 | アクティビティ")
	fmt.Println(strings.Repeat("-", 30))

	for hour := 0; hour < 24; hour++ {
		count := patterns[hour]
		bar := strings.Repeat("█", count/5) // Scale down for display
		fmt.Printf("%2d時 | %s (%d)\n", hour, bar, count)
	}
}
