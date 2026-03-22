package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	speedTestInterval = 5 * time.Minute
	dataDir           = `C:\speedtest`
	csvFile           = `C:\speedtest\speedtest.csv`
	logFile           = `C:\speedtest\barryallen.log`
)

var csvHeader = []string{
	"timestamp", "server_name", "server_host", "latency_ms",
	"download_mbps", "upload_mbps", "server_id",
}

type barryAllenService struct{}

func (s *barryAllenService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Set up file logging
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		elog, _ := eventlog.Open(serviceName)
		if elog != nil {
			elog.Error(1, fmt.Sprintf("Failed to create data dir: %v", err))
			elog.Close()
		}
		return true, 1
	}

	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return true, 1
	}
	defer lf.Close()
	logger := log.New(lf, "", log.LstdFlags)

	// Ensure CSV has header
	if err := ensureCSVHeader(); err != nil {
		logger.Printf("ERROR: Failed to ensure CSV header: %v", err)
		return true, 1
	}

	logger.Println("Barry Allen service started")
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	ticker := time.NewTicker(speedTestInterval)
	defer ticker.Stop()

	// Run first test immediately
	go func() {
		runAndRecord(logger)
	}()

	for {
		select {
		case <-ticker.C:
			runAndRecord(logger)
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				logger.Println("Barry Allen service stopping")
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			}
		}
	}
}

func runAndRecord(logger *log.Logger) {
	logger.Println("Starting speed test...")
	result, err := runSpeedTest()
	if err != nil {
		logger.Printf("ERROR: Speed test failed: %v", err)
		return
	}

	logger.Printf("Speed test complete: %.2f/%.2f Mbps (down/up), %.2f ms latency, server: %s",
		result.DownloadMbps, result.UploadMbps, result.LatencyMs, result.ServerName)

	if err := appendCSV(result); err != nil {
		logger.Printf("ERROR: Failed to write CSV: %v", err)
	}
}

func ensureCSVHeader() error {
	if err := os.MkdirAll(filepath.Dir(csvFile), 0755); err != nil {
		return err
	}

	info, err := os.Stat(csvFile)
	if os.IsNotExist(err) || (err == nil && info.Size() == 0) {
		f, err := os.Create(csvFile)
		if err != nil {
			return err
		}
		defer f.Close()
		w := csv.NewWriter(f)
		if err := w.Write(csvHeader); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	}
	return err
}

func appendCSV(r *SpeedTestResult) error {
	f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	record := []string{
		r.Timestamp.Format(time.RFC3339),
		r.ServerName,
		r.ServerHost,
		fmt.Sprintf("%.2f", r.LatencyMs),
		fmt.Sprintf("%.2f", r.DownloadMbps),
		fmt.Sprintf("%.2f", r.UploadMbps),
		r.ServerID,
	}
	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
