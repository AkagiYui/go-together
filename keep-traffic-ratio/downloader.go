// Package main provides download functionality.
package main

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// Downloader handles downloading traffic URLs.
type Downloader struct {
	dryRun        bool
	downloaded    atomic.Uint64
	isDownloading atomic.Bool
	currentURL    string
}

// NewDownloader creates a new downloader.
func NewDownloader(dryRun bool) *Downloader {
	return &Downloader{
		dryRun: dryRun,
	}
}

// Download downloads from the specified URL and returns the bytes downloaded.
func (d *Downloader) Download(url string) error {
	if d.isDownloading.Load() {
		return fmt.Errorf("already downloading")
	}

	d.isDownloading.Store(true)
	d.currentURL = url

	if d.dryRun {
		// Simulate download in dry run mode
		time.Sleep(1 * time.Second)
		d.downloaded.Add(1024 * 1024) // Simulate 1MB downloaded
		d.isDownloading.Store(false)
		d.currentURL = ""
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		d.isDownloading.Store(false)
		d.currentURL = ""
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.isDownloading.Store(false)
		d.currentURL = ""
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			d.downloaded.Add(uint64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.isDownloading.Store(false)
			d.currentURL = ""
			return fmt.Errorf("read error: %w", err)
		}
	}

	d.isDownloading.Store(false)
	d.currentURL = ""
	return nil
}

// GetDownloaded returns the total bytes downloaded.
func (d *Downloader) GetDownloaded() uint64 {
	return d.downloaded.Load()
}

// IsDownloading returns whether a download is in progress.
func (d *Downloader) IsDownloading() bool {
	return d.isDownloading.Load()
}

// GetCurrentURL returns the current downloading URL.
func (d *Downloader) GetCurrentURL() string {
	return d.currentURL
}
