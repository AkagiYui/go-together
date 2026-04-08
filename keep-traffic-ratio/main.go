// Package main is a traffic ratio keeper that monitors network traffic and downloads files to maintain a specified ratio.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

func main() {
	// Parse command line arguments
	ifaceName := flag.String("interface", "", "Network interface name to monitor (required)")
	targetRatio := flag.Float64("ratio", 0.0, "Target download traffic ratio (0.0-1.0, required)")
	dryRun := flag.Bool("dryrun", false, "Dry run mode - don't actually download")
	noUI := flag.Bool("no-ui", false, "Disable terminal UI (print to stdout instead)")
	flag.Parse()

	if *ifaceName == "" || *targetRatio == 0.0 {
		fmt.Println("Usage: keep-traffic-ratio -interface <interface> -ratio <ratio> [-dryrun] [-no-ui]")
		fmt.Println("\nExample:")
		fmt.Println("  keep-traffic-ratio -interface eth0 -ratio 0.5")
		fmt.Println("  keep-traffic-ratio -interface eth0 -ratio 0.5 -dryrun")
		fmt.Println("  keep-traffic-ratio -interface eth0 -ratio 0.5 -dryrun -no-ui")
		os.Exit(1)
	}

	// Verify interface exists
	counters, err := net.IOCounters(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting network interfaces: %v\n", err)
		os.Exit(1)
	}
	interfaceExists := false
	for _, c := range counters {
		if c.Name == *ifaceName {
			interfaceExists = true
			break
		}
	}
	if !interfaceExists {
		fmt.Fprintf(os.Stderr, "Error: interface '%s' not found\n", *ifaceName)
		fmt.Println("\nAvailable interfaces:")
		for _, c := range counters {
			fmt.Printf("  - %s\n", c.Name)
		}
		os.Exit(1)
	}

	// Fetch initial values
	values, err := FetchValues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching values: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched %d traffic URLs\n", len(values))

	// Initialize components
	monitor := NewTrafficMonitor(*ifaceName)
	downloader := NewDownloader(*dryRun)

	var ui *UI
	if !*noUI {
		ui = NewUI()
		// Start UI in a goroutine
		go func() {
			if err := ui.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}()
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if ui != nil {
			ui.Stop()
		}
		os.Exit(0)
	}()

	// Random source for URL selection
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Main loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update traffic statistics
			if err := monitor.Update(); err != nil {
				if ui != nil {
					ui.AppendLog(fmt.Sprintf("[red]Error updating stats: %v[white]", err))
				} else {
					fmt.Fprintf(os.Stderr, "Error updating stats: %v\n", err)
				}
				continue
			}

			ifaceRx, ifaceTx, totalRx, totalTx := monitor.GetStats()
			dlRatio := monitor.GetDownloadRatio()
			ulRatio := monitor.GetUploadRatio()

			// Update display
			if ui != nil {
				ui.UpdateStats(*ifaceName, ifaceRx, ifaceTx, totalRx, totalTx, dlRatio, ulRatio, *targetRatio)
				ui.UpdateDownload(downloader.IsDownloading(), downloader.GetCurrentURL(), downloader.GetDownloaded(), *dryRun)
			} else {
				// Print to stdout in no-ui mode
				fmt.Printf("\rInterface: %s | Download: %s (%.2f%%) | 保持至少: %.2f%% | Upload: %s | Total: %s / %s | Downloading: %v",
					*ifaceName,
					FormatBytes(ifaceRx), dlRatio*100,
					*targetRatio*100,
					FormatBytes(ifaceTx),
					FormatBytes(totalRx), FormatBytes(totalTx),
					downloader.IsDownloading(),
				)
				if downloader.IsDownloading() {
					fmt.Printf(" | Downloaded: %s", FormatBytes(downloader.GetDownloaded()))
				}
				fmt.Println()
			}

			// Check if we need to download
			if dlRatio < *targetRatio && !downloader.IsDownloading() {
				// Select a random URL
				url := values[randSrc.Intn(len(values))]
				if ui != nil {
					ui.AppendLog(fmt.Sprintf("Starting download from: %s", url))
				} else {
					fmt.Printf("\nStarting download from: %s\n", url)
				}
				go func(u string) {
					if err := downloader.Download(u); err != nil {
						if ui != nil {
							ui.AppendLog(fmt.Sprintf("[red]Download error: %v[white]", err))
						} else {
							fmt.Printf("Download error: %v\n", err)
						}
					} else {
						if ui != nil {
							ui.AppendLog(fmt.Sprintf("[green]Download completed: %s[white]", u))
						} else {
							fmt.Printf("Download completed: %s\n", u)
						}
					}
				}(url)
			}
		}
	}
}
