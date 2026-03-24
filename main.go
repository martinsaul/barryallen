package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
)

const serviceName = "BarryAllen"
const serviceDesc = "Barry Allen Speed Test Service - runs periodic internet speed tests"

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "install":
			err := installService()
			if err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
			fmt.Println("Service installed successfully.")
			return
		case "uninstall", "remove":
			err := removeService()
			if err != nil {
				log.Fatalf("Failed to remove service: %v", err)
			}
			fmt.Println("Service removed successfully.")
			return
		case "start":
			err := startService()
			if err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
			fmt.Println("Service started.")
			return
		case "stop":
			err := stopService()
			if err != nil {
				log.Fatalf("Failed to stop service: %v", err)
			}
			fmt.Println("Service stopped.")
			return
		case "run":
			// Run a single speed test and print results (useful for testing)
			result, err := runSpeedTest(nil)
			if err != nil {
				log.Fatalf("Speed test failed: %v", err)
			}
			fmt.Printf("Server:    %s (%s)\n", result.ServerName, result.ServerHost)
			fmt.Printf("Latency:   %.2f ms\n", result.LatencyMs)
			fmt.Printf("Download:  %.2f Mbps\n", result.DownloadMbps)
			fmt.Printf("Upload:    %.2f Mbps\n", result.UploadMbps)
			return
		case "help", "-h", "--help":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// Default: run as Windows service
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to detect service mode: %v", err)
	}

	if isService {
		err = svc.Run(serviceName, &barryAllenService{})
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
	} else {
		fmt.Println("Not running as a Windows service. Use 'barryallen run' for a single test.")
		fmt.Println()
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Barry Allen - Internet Speed Test Service")
	fmt.Println()
	fmt.Println("Usage: barryallen <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install     Install as a Windows service")
	fmt.Println("  uninstall   Remove the Windows service")
	fmt.Println("  start       Start the service")
	fmt.Println("  stop        Stop the service")
	fmt.Println("  run         Run a single speed test (for testing)")
	fmt.Println("  help        Show this help message")
}
