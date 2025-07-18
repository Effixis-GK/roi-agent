package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ROI Agent Data Sender - Enhanced with 10-minute intervals")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  data-sender process                 # Process and send current 10-minute interval")
		fmt.Println("  data-sender test                    # Test configuration and connection")
		fmt.Println("  data-sender status                  # Show current status and configuration")
		fmt.Println("  data-sender logs [limit]            # Show recent transmission logs (default: 10)")
		fmt.Println("  data-sender cleanup                 # Cleanup old files (data, transmission, logs)")
		fmt.Println("  data-sender set-interval <minutes>  # Set transmission interval (1-1440 minutes)")
		fmt.Println("  data-sender env-example             # Create .env.example file")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  data-sender process                 # Send data for the current interval")
		fmt.Println("  data-sender test                    # Test if data transmission works")
		fmt.Println("  data-sender logs 20                 # Show last 20 transmission attempts")
		fmt.Println("  data-sender set-interval 5          # Set interval to 5 minutes")
		fmt.Println("")
		fmt.Println("Environment Variables (.env file):")
		fmt.Println("  ROI_AGENT_BASE_URL         # Server base URL")
		fmt.Println("  ROI_AGENT_API_KEY          # API authentication key")
		fmt.Println("  ROI_AGENT_INTERVAL_MINUTES # Transmission interval in minutes (default: 10)")
		return
	}

	sender := NewDataSender()
	command := os.Args[1]

	switch command {
	case "process":
		if err := sender.processCurrentInterval(); err != nil {
			log.Fatalf("Error processing current interval: %v", err)
		}
	case "test":
		sender.TestConnection()
	case "status":
		sender.ShowStatus()
	case "logs":
		limit := 10
		if len(os.Args) >= 3 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil {
				limit = l
			}
		}
		sender.ShowTransmissionLogs(limit)
	case "cleanup":
		if err := sender.CleanupOldFiles(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		} else {
			fmt.Println("Cleanup completed")
		}
	case "env-example":
		if err := sender.CreateEnvExample(); err != nil {
			log.Printf("Error creating .env.example: %v", err)
		} else {
			fmt.Println(".env.example file created")
		}
	case "set-interval":
		if len(os.Args) < 3 {
			fmt.Println("Usage: data-sender set-interval <minutes>")
			fmt.Println("Valid range: 1-1440 minutes (1 minute to 24 hours)")
			return
		}
		interval, err := strconv.Atoi(os.Args[2])
		if err != nil || interval < 1 || interval > 1440 {
			fmt.Println("Error: Invalid interval. Must be between 1 and 1440 minutes.")
			return
		}
		sender.SetTransmissionInterval(interval)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
