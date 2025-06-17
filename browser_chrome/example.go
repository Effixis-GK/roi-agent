package main

import (
	"fmt"
	"log"
	"strings"
)

// この例では拡張機能を使用してより詳細な分析を行います
func runExtendedAnalysis() {
	reader := NewExtendedChromeReader()

	fmt.Println("🔍 Chrome Browser Analysis Tool")
	fmt.Println(strings.Repeat("=", 50))

	// 1. 基本的なブラウジング履歴
	fmt.Println("\n📖 Getting browsing history...")
	history, err := reader.GetHistory(15)
	if err != nil {
		log.Printf("Warning: Could not read history: %v\n", err)
	} else {
		PrintHistory(history)
	}

	// 2. トップドメイン
	fmt.Println("\n🌐 Getting top domains...")
	domains, err := reader.GetTopDomains(8)
	if err != nil {
		log.Printf("Warning: Could not read domains: %v\n", err)
	} else {
		PrintTopDomains(domains)
	}

	// 3. ダウンロード履歴
	fmt.Println("\n💾 Getting download history...")
	downloads, err := reader.GetDownloads(10)
	if err != nil {
		log.Printf("Warning: Could not read downloads: %v\n", err)
	} else {
		PrintDownloads(downloads)
	}

	// 4. 検索クエリ分析
	fmt.Println("\n🔎 Analyzing search queries...")
	queries, err := reader.GetSearchQueries(10)
	if err != nil {
		log.Printf("Warning: Could not analyze search queries: %v\n", err)
	} else {
		PrintSearchQueries(queries)
	}

	// 5. ブラウジングパターン
	fmt.Println("\n📊 Analyzing browsing patterns...")
	patterns, err := reader.GetBrowsingPatterns()
	if err != nil {
		log.Printf("Warning: Could not analyze browsing patterns: %v\n", err)
	} else {
		PrintBrowsingPatterns(patterns)
	}

	fmt.Println("\n✅ Analysis complete!")
}

// この関数を使用するには、main.goのmain関数内で runExtendedAnalysis() を呼び出してください
