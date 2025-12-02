package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	// ログを標準出力に設定（LaunchDaemonで/var/log/roiagent.logに出力するため）
	log.SetOutput(os.Stdout)
	
	if len(os.Args) < 2 {
		fmt.Println("ROI Agent Data Sender - Enhanced with remote configuration & auto-update")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  data-sender process                 # Process and send current interval data")
		fmt.Println("  data-sender register                # Send initial registration (device info + version)")
		fmt.Println("  data-sender fetch-config            # Fetch remote configuration from server")
		fmt.Println("  data-sender show-config             # Show current remote configuration")
		fmt.Println("  data-sender check-update            # Check for available updates")
		fmt.Println("  data-sender update                  # Check and install updates if available")
		fmt.Println("  data-sender test                    # Test configuration and connection")
		fmt.Println("  data-sender status                  # Show current status and configuration")
		fmt.Println("  data-sender logs [limit]            # Show recent transmission logs (default: 10)")
		fmt.Println("  data-sender cleanup                 # Cleanup old files (data, transmission, logs)")
		fmt.Println("  data-sender set-interval <minutes>  # Set transmission interval (1-1440 minutes)")
		fmt.Println("  data-sender env-example             # Create .env.example file")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  data-sender process                 # Send data for the current interval")
		fmt.Println("  data-sender fetch-config            # Fetch latest remote config")
		fmt.Println("  data-sender check-update            # Check if a new version is available")
		fmt.Println("  data-sender update                  # Install available updates")
		fmt.Println("  data-sender register                # Register device immediately after install")
		fmt.Println("  data-sender test                    # Test if data transmission works")
		fmt.Println("  data-sender logs 20                 # Show last 20 transmission attempts")
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
		// Check and apply remote config before processing
		sender.CheckAndApplyRemoteConfig()
		if err := sender.processCurrentInterval(); err != nil {
			log.Fatalf("Error processing current interval: %v", err)
		}
	case "register":
		if err := sender.SendInitialRegistration(); err != nil {
			log.Fatalf("Error sending initial registration: %v", err)
		}
		fmt.Println("✅ Initial registration sent successfully!")
	case "fetch-config":
		sender.FetchAndShowRemoteConfig()
	case "show-config":
		sender.ShowRemoteConfig()
	case "check-update":
		sender.CheckForUpdate()
	case "update":
		sender.CheckAndInstallUpdate()
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
