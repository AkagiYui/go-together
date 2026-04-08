// Package main provides network traffic monitoring functionality.
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// TrafficMonitor monitors network traffic for a specific interface.
type TrafficMonitor struct {
	interfaceName string
	totalRx       uint64
	totalTx       uint64
	ifaceRx       uint64
	ifaceTx       uint64
	lastUpdate    time.Time
	mu            sync.RWMutex
}

// NewTrafficMonitor creates a new traffic monitor for the specified interface.
func NewTrafficMonitor(interfaceName string) *TrafficMonitor {
	return &TrafficMonitor{
		interfaceName: interfaceName,
		lastUpdate:    time.Now(),
	}
}

// Update updates the traffic counters.
func (tm *TrafficMonitor) Update() error {
	counters, err := net.IOCounters(true)
	if err != nil {
		return fmt.Errorf("failed to get network counters: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	var totalRx, totalTx uint64
	var ifaceRx, ifaceTx uint64

	for _, c := range counters {
		totalRx += c.BytesRecv
		totalTx += c.BytesSent
		if c.Name == tm.interfaceName {
			ifaceRx = c.BytesRecv
			ifaceTx = c.BytesSent
		}
	}

	tm.totalRx = totalRx
	tm.totalTx = totalTx
	tm.ifaceRx = ifaceRx
	tm.ifaceTx = ifaceTx
	tm.lastUpdate = time.Now()

	return nil
}

// GetStats returns the current traffic statistics.
func (tm *TrafficMonitor) GetStats() (ifaceRx, ifaceTx, totalRx, totalTx uint64) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.ifaceRx, tm.ifaceTx, tm.totalRx, tm.totalTx
}

// GetDownloadRatio returns the download traffic ratio of the monitored interface (download / total of this interface).
func (tm *TrafficMonitor) GetDownloadRatio() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	total := tm.ifaceRx + tm.ifaceTx
	if total == 0 {
		return 0
	}
	return float64(tm.ifaceRx) / float64(total)
}

// GetUploadRatio returns the upload traffic ratio of the monitored interface (upload / total of this interface).
func (tm *TrafficMonitor) GetUploadRatio() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	total := tm.ifaceRx + tm.ifaceTx
	if total == 0 {
		return 0
	}
	return float64(tm.ifaceTx) / float64(total)
}

// FormatBytes formats bytes to human readable string.
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
