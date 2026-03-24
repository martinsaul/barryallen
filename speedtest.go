package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
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

const maxServersToTry = 3

func runSpeedTest() (*SpeedTestResult, error) {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     false,
			TLSClientConfig:      &tls.Config{},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	client := speedtest.New(speedtest.WithDoer(httpClient))

	serverList, err := client.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("fetching servers: %w", err)
	}

	if len(serverList) == 0 {
		return nil, fmt.Errorf("no speed test servers found")
	}

	limit := maxServersToTry
	if len(serverList) < limit {
		limit = len(serverList)
	}

	var lastErr error
	for i := 0; i < limit; i++ {
		server := serverList[i]

		if err := server.PingTest(nil); err != nil {
			lastErr = fmt.Errorf("server %s: ping test: %w", server.Name, err)
			continue
		}

		if err := server.DownloadTest(); err != nil {
			lastErr = fmt.Errorf("server %s: download test: %w", server.Name, err)
			continue
		}

		if err := server.UploadTest(); err != nil {
			lastErr = fmt.Errorf("server %s: upload test: %w", server.Name, err)
			continue
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

	return nil, fmt.Errorf("all %d servers failed, last error: %w", limit, lastErr)
}
