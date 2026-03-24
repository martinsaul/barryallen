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

type failedServer struct {
	id   string
	name string
	reason string
}

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
	var failures []failedServer
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
			failures = append(failures, failedServer{server.ID, server.Name, "ping: " + err.Error()})
			continue
		}

		if err := server.DownloadTest(); err != nil {
			lastErr = fmt.Errorf("server %s (%s): download test: %w", server.Name, server.ID, err)
			failures = append(failures, failedServer{server.ID, server.Name, "download error: " + err.Error()})
			continue
		}

		dlMbps := server.DLSpeed.Mbps()
		if dlMbps <= 0 {
			lastErr = fmt.Errorf("server %s (%s): reported 0 download", server.Name, server.ID)
			failures = append(failures, failedServer{server.ID, server.Name, "reported 0 download speed"})
			continue
		}

		if err := server.UploadTest(); err != nil {
			lastErr = fmt.Errorf("server %s (%s): upload test: %w", server.Name, server.ID, err)
			failures = append(failures, failedServer{server.ID, server.Name, "upload error: " + err.Error()})
			continue
		}

		ulMbps := server.ULSpeed.Mbps()
		if ulMbps <= 0 {
			lastErr = fmt.Errorf("server %s (%s): reported 0 upload", server.Name, server.ID)
			failures = append(failures, failedServer{server.ID, server.Name, "reported 0 upload speed"})
			continue
		}

		// A server succeeded — blacklist the ones that failed during this run,
		// since the internet was clearly working.
		if blacklist != nil {
			for _, f := range failures {
				blacklist.Strike(f.id, f.name, f.reason)
			}
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

	// All servers failed — likely an internet outage, don't blacklist anyone.
	if tried == 0 {
		return nil, fmt.Errorf("all servers are blacklisted")
	}
	return nil, fmt.Errorf("all %d servers failed (possible connectivity issue), last error: %w", tried, lastErr)
}
