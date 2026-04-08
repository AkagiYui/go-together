// Package main provides terminal UI functionality.
package main

import (
	"fmt"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI handles the terminal user interface.
type UI struct {
	app          *tview.Application
	flex         *tview.Flex
	statsView    *tview.TextView
	downloadView *tview.TextView
	logView      *tview.TextView
}

// NewUI creates a new UI instance.
func NewUI() *UI {
	app := tview.NewApplication()

	statsView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() { app.Draw() })
	statsView.SetBorder(true).SetTitle(" Traffic Statistics ").SetTitleAlign(tview.AlignLeft)

	downloadView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() { app.Draw() })
	downloadView.SetBorder(true).SetTitle(" Download Status ").SetTitleAlign(tview.AlignLeft)

	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetScrollable(true).
		SetChangedFunc(func() { app.Draw() })
	logView.SetBorder(true).SetTitle(" Logs ").SetTitleAlign(tview.AlignLeft)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(statsView, 8, 1, false).
		AddItem(downloadView, 6, 1, false).
		AddItem(logView, 0, 1, false)

	return &UI{
		app:          app,
		flex:         flex,
		statsView:    statsView,
		downloadView: downloadView,
		logView:      logView,
	}
}

// Run starts the UI application.
func (ui *UI) Run() error {
	ui.app.SetRoot(ui.flex, true).EnableMouse(false)
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
			ui.app.Stop()
			return nil
		}
		return event
	})
	return ui.app.Run()
}

// Stop stops the UI application.
func (ui *UI) Stop() {
	ui.app.Stop()
}

// UpdateStats updates the traffic statistics display.
func (ui *UI) UpdateStats(ifaceName string, ifaceRx, ifaceTx, totalRx, totalTx uint64, dlRatio, ulRatio, targetRatio float64) {
	statsText := fmt.Sprintf(
		"Interface: [yellow]%s[white]\n\n"+
			"Download: [green]%s[white] (占本网卡 %.2f%%)  保持至少: [cyan]%.2f%%[white]\n"+
			"Upload:   [green]%s[white] (占本网卡 %.2f%%)\n\n"+
			"Total Download: [blue]%s[white]  Total Upload: [blue]%s[white]",
		ifaceName,
		FormatBytes(ifaceRx), dlRatio*100, targetRatio*100,
		FormatBytes(ifaceTx), ulRatio*100,
		FormatBytes(totalRx),
		FormatBytes(totalTx),
	)
	ui.statsView.SetText(statsText)
}

// UpdateDownload updates the download status display.
func (ui *UI) UpdateDownload(isDownloading bool, currentURL string, downloaded uint64, dryRun bool) {
	var text string
	if isDownloading {
		mode := "Downloading"
		if dryRun {
			mode = "[yellow]DRY RUN[white]"
		}
		text = fmt.Sprintf("%s from:\n[green]%s[white]\n\nDownloaded: [cyan]%s[white]",
			mode, currentURL, FormatBytes(downloaded))
	} else {
		text = "[gray]Idle[white]"
	}
	ui.downloadView.SetText(text)
}

// AppendLog appends a log message to the log view.
func (ui *UI) AppendLog(message string) {
	timestamp := "???" // timestamp removed for simplicity
	current := ui.logView.GetText(false)
	ui.logView.SetText(current + fmt.Sprintf("[%s] %s\n", timestamp, message))
	// Scroll to bottom
	ui.logView.ScrollToEnd()
}

// ClearLog clears the log view.
func (ui *UI) ClearLog() {
	ui.logView.SetText("")
}
