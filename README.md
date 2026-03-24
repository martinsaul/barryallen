# Barry Allen

A Windows service that runs internet speed tests every 5 minutes and logs results to CSV.

## Quick Install

Open PowerShell **as Administrator** and run:

```powershell
irm https://raw.githubusercontent.com/martinsaul/barryallen/master/install.ps1 | iex
```

Or download and run manually:

```powershell
# Download the installer
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/martinsaul/barryallen/master/install.ps1" -OutFile install.ps1

# Install (downloads latest release, installs service, starts it)
.\install.ps1

# For ARM64 Windows
.\install.ps1 -Arch arm64
```

## What It Does

- Runs as a Windows service (`BarryAllen`)
- Performs a speed test against the closest speedtest.net server every 5 minutes
- Logs results to `C:\speedtest\speedtest.csv`
- Auto-starts on boot, auto-restarts on failure

## CSV Format

| Column | Description |
|--------|-------------|
| timestamp | ISO 8601 timestamp |
| server_name | Speed test server name |
| server_host | Server hostname |
| latency_ms | Ping latency in milliseconds |
| download_mbps | Download speed in Mbps |
| upload_mbps | Upload speed in Mbps |
| server_id | Server identifier |
| status | Online/offline connectivity status |
| servers_tested | List of servers that failed |

## Manual Commands

```powershell
# Run a single speed test (no service needed)
C:\speedtest\barryallen.exe run

# Service management
C:\speedtest\barryallen.exe install
C:\speedtest\barryallen.exe start
C:\speedtest\barryallen.exe stop
C:\speedtest\barryallen.exe uninstall
```

## Uninstall

```powershell
.\install.ps1 -Uninstall
```

## Building from Source

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o barryallen.exe .
```

