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

const maxServersToTry = 10

func runSpeedTest(blacklist *ServerBlacklist) (*SpeedTestResult, error) {
	serverList, err := speedtest.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("fetching servers: %w", err)
	}

	if len(serverList) == 0 {
		return nil, fmt.Errorf("no speed test servers found")
	}

	tried := 0
	var lastErr error
	for _, server := range serverList {
		if tried >= maxServersToTry {
			break
		}

		if blacklist != nil && blacklist.IsBlacklisted(server.ID) {
			continue
		}

		tried++
		client := speedtest.New()
		server.Context = client

		if err := server.PingTest(nil); err != nil {
			lastErr = fmt.Errorf("server %s (%s): ping test: %w", server.Name, server.ID, err)
			continue
		}

		if err := server.DownloadTest(); err != nil {
			lastErr = fmt.Errorf("server %s (%s): download test: %w", server.Name, server.ID, err)
			if blacklist != nil {
				blacklist.Strike(server.ID, server.Name, fmt.Sprintf("download error: %v", err))
			}
			continue
		}

		dlMbps := server.DLSpeed.Mbps()
		if dlMbps <= 0 {
			lastErr = fmt.Errorf("server %s (%s): reported 0 download", server.Name, server.ID)
			if blacklist != nil {
				blacklist.Strike(server.ID, server.Name, "reported 0 download speed")
			}
			continue
		}

		if err := server.UploadTest(); err != nil {
			lastErr = fmt.Errorf("server %s (%s): upload test: %w", server.Name, server.ID, err)
			if blacklist != nil {
				blacklist.Strike(server.ID, server.Name, fmt.Sprintf("upload error: %v", err))
			}
			continue
		}

		ulMbps := server.ULSpeed.Mbps()
		if ulMbps <= 0 {
			lastErr = fmt.Errorf("server %s (%s): reported 0 upload", server.Name, server.ID)
			if blacklist != nil {
				blacklist.Strike(server.ID, server.Name, "reported 0 upload speed")
			}
			continue
		}

		return &SpeedTestResult{
			Timestamp:    time.Now(),
			ServerName:   server.Name,
			ServerHost:   server.Host,
			ServerID:     server.ID,
			LatencyMs:    float64(server.Latency) / float64(time.Millisecond),
			DownloadMbps: dlMbps,
			UploadMbps:   ulMbps,
		}, nil
	}

	if tried == 0 {
		return nil, fmt.Errorf("all servers are blacklisted")
	}
	return nil, fmt.Errorf("all %d servers failed, last error: %w", tried, lastErr)
}
