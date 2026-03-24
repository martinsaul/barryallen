package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

// SpeedTestResult holds the results of a single speed test.
type SpeedTestResult struct {
	Timestamp     time.Time
	ServerName    string
	ServerHost    string
	ServerID      string
	LatencyMs     float64
	DownloadMbps  float64
	UploadMbps    float64
	Status        string
	ServersTested string
}

const maxServersToTry = 10

type failedServer struct {
	id     string
	name   string
	reason string
}

func checkConnectivity() bool {
	targets := []string{"8.8.8.8:53", "1.1.1.1:53"}
	for _, target := range targets {
		conn, err := net.DialTimeout("tcp", target, 5*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func formatServersTested(failures []failedServer) string {
	var parts []string
	for _, f := range failures {
		parts = append(parts, fmt.Sprintf("%s (%s)", f.name, f.id))
	}
	return strings.Join(parts, "; ")
}

func runSpeedTest(blacklist *ServerBlacklist) (*SpeedTestResult, error) {
	serverList, err := speedtest.FetchServers()
	if err != nil {
		online := checkConnectivity()
		status := "offline"
		if online {
			status = "online"
		}
		return &SpeedTestResult{
			Timestamp: time.Now(),
			Status:    status,
		}, fmt.Errorf("fetching servers: %w", err)
	}

	if len(serverList) == 0 {
		online := checkConnectivity()
		status := "offline"
		if online {
			status = "online"
		}
		return &SpeedTestResult{
			Timestamp: time.Now(),
			Status:    status,
		}, fmt.Errorf("no speed test servers found")
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

		// A server succeeded — blacklist the ones that failed.
		if blacklist != nil {
			for _, f := range failures {
				blacklist.Strike(f.id, f.name, f.reason)
			}
		}

		return &SpeedTestResult{
			Timestamp:     time.Now(),
			ServerName:    server.Name,
			ServerHost:    server.Host,
			ServerID:      server.ID,
			LatencyMs:     float64(server.Latency) / float64(time.Millisecond),
			DownloadMbps:  dlMbps,
			UploadMbps:    ulMbps,
			Status:        "online",
			ServersTested: formatServersTested(failures),
		}, nil
	}

	// All servers failed — check actual connectivity.
	online := checkConnectivity()
	status := "offline"
	if online {
		status = "online"
		// Internet is up, so these servers are genuinely broken — blacklist them.
		if blacklist != nil {
			for _, f := range failures {
				blacklist.Strike(f.id, f.name, f.reason)
			}
		}
	}

	if tried == 0 {
		return &SpeedTestResult{
			Timestamp: time.Now(),
			Status:    status,
		}, fmt.Errorf("all servers are blacklisted")
	}

	return &SpeedTestResult{
		Timestamp:     time.Now(),
		Status:        status,
		ServersTested: formatServersTested(failures),
	}, fmt.Errorf("all %d servers failed, last error: %w", tried, lastErr)
}
