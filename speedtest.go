package main

import (
	"fmt"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

// SpeedTestResult holds the results of a single speed test.
type SpeedTestResult struct {
	Timestamp    time.Time
	ServerName   string
	ServerHost   string
	ServerID     string
	LatencyMs    float64
	DownloadMbps float64
	UploadMbps   float64
}

func runSpeedTest() (*SpeedTestResult, error) {
	client := speedtest.New()

	serverList, err := client.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("fetching servers: %w", err)
	}

	if len(serverList) == 0 {
		return nil, fmt.Errorf("no speed test servers found")
	}

	// Pick the closest server
	server := serverList[0]

	if err := server.PingTest(nil); err != nil {
		return nil, fmt.Errorf("ping test: %w", err)
	}

	if err := server.DownloadTest(); err != nil {
		return nil, fmt.Errorf("download test: %w", err)
	}

	if err := server.UploadTest(); err != nil {
		return nil, fmt.Errorf("upload test: %w", err)
	}

	return &SpeedTestResult{
		Timestamp:    time.Now(),
		ServerName:   server.Name,
		ServerHost:   server.Host,
		ServerID:     server.ID,
		LatencyMs:    float64(server.Latency) / float64(time.Millisecond),
		DownloadMbps: server.DLSpeed.Mbps(),
		UploadMbps:   server.ULSpeed.Mbps(),
	}, nil
}
