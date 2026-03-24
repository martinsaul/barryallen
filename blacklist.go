package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var blacklistDurations = []time.Duration{
	6 * time.Hour,
	12 * time.Hour,
	24 * time.Hour,
	7 * 24 * time.Hour,
}

type BlacklistEntry struct {
	ServerID  string    `json:"server_id"`
	ServerName string   `json:"server_name"`
	Reason    string    `json:"reason"`
	Strikes   int       `json:"strikes"`
	BannedAt  time.Time `json:"banned_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ServerBlacklist struct {
	mu      sync.Mutex
	path    string
	Entries map[string]*BlacklistEntry `json:"entries"`
}

func NewServerBlacklist(dir string) *ServerBlacklist {
	bl := &ServerBlacklist{
		path:    filepath.Join(dir, "blacklist.json"),
		Entries: make(map[string]*BlacklistEntry),
	}
	bl.load()
	return bl
}

func (bl *ServerBlacklist) IsBlacklisted(serverID string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	entry, ok := bl.Entries[serverID]
	if !ok {
		return false
	}
	if time.Now().After(entry.ExpiresAt) {
		return false
	}
	return true
}

func (bl *ServerBlacklist) Strike(serverID, serverName, reason string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	entry, ok := bl.Entries[serverID]
	if !ok {
		entry = &BlacklistEntry{
			ServerID:   serverID,
			ServerName: serverName,
		}
		bl.Entries[serverID] = entry
	}

	// If previous ban expired, keep the strike count for escalation
	entry.Strikes++
	entry.Reason = reason
	entry.BannedAt = time.Now()

	idx := entry.Strikes - 1
	if idx >= len(blacklistDurations) {
		idx = len(blacklistDurations) - 1
	}
	entry.ExpiresAt = time.Now().Add(blacklistDurations[idx])

	bl.save()
}

func (bl *ServerBlacklist) load() {
	data, err := os.ReadFile(bl.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &bl.Entries)
}

func (bl *ServerBlacklist) save() {
	data, err := json.MarshalIndent(bl.Entries, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(bl.path, data, 0644)
}
