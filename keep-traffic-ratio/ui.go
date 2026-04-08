// Package main provides terminal UI functionality using BubbleTea and bubblezone.
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

// UI styles
var (
	statsStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			Width(80)

	downloadStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			Width(80)

	logStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			Width(80).
			MaxHeight(10)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cyanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	blueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	grayStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	whiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

// MessageType defines different UI update messages
type MessageType int

const (
	UpdateStatsMsg MessageType = iota
	UpdateDownloadMsg
	AppendLogMsg
)

// UIMessage represents a message sent to the UI
type UIMessage struct {
	Type MessageType
	Data interface{}
}

// StatsData contains traffic statistics
type StatsData struct {
	IfaceName   string
	IfaceRx     uint64
	IfaceTx     uint64
	TotalRx     uint64
	TotalTx     uint64
	DlRatio     float64
	UlRatio     float64
	TargetRatio float64
}

// DownloadData contains download status
type DownloadData struct {
	IsDownloading bool
	CurrentURL    string
	Downloaded    uint64
	DryRun        bool
}

// LogData contains a log message
type LogData struct {
	Message string
}

// model represents the UI state
type model struct {
	stats    StatsData
	download DownloadData
	logs     []string
	ready    bool
	width    int
	height   int
}

// UI handles the terminal user interface
type UI struct {
	program *tea.Program
}

// NewUI creates a new UI instance
func NewUI() *UI {
	// Initialize the global zone manager
	zone.NewGlobal()

	m := model{
		logs: make([]string, 0, 100),
	}

	p := tea.NewProgram(&m)

	return &UI{
		program: p,
	}
}

// Run starts the UI application
func (ui *UI) Run() error {
	_, err := ui.program.Run()
	return err
}

// Stop stops the UI application
func (ui *UI) Stop() {
	ui.program.Quit()
	zone.Close()
}

// UpdateStats updates the traffic statistics display
func (ui *UI) UpdateStats(ifaceName string, ifaceRx, ifaceTx, totalRx, totalTx uint64, dlRatio, ulRatio, targetRatio float64) {
	ui.program.Send(UIMessage{
		Type: UpdateStatsMsg,
		Data: StatsData{
			IfaceName:   ifaceName,
			IfaceRx:     ifaceRx,
			IfaceTx:     ifaceTx,
			TotalRx:     totalRx,
			TotalTx:     totalTx,
			DlRatio:     dlRatio,
			UlRatio:     ulRatio,
			TargetRatio: targetRatio,
		},
	})
}

// UpdateDownload updates the download status display
func (ui *UI) UpdateDownload(isDownloading bool, currentURL string, downloaded uint64, dryRun bool) {
	ui.program.Send(UIMessage{
		Type: UpdateDownloadMsg,
		Data: DownloadData{
			IsDownloading: isDownloading,
			CurrentURL:    currentURL,
			Downloaded:    downloaded,
			DryRun:        dryRun,
		},
	})
}

// AppendLog appends a log message to the log view
func (ui *UI) AppendLog(message string) {
	ui.program.Send(UIMessage{
		Type: AppendLogMsg,
		Data: LogData{
			Message: message,
		},
	})
}

// ClearLog clears the log view
func (ui *UI) ClearLog() {
	ui.program.Send(UIMessage{
		Type: AppendLogMsg,
		Data: LogData{
			Message: "", // Empty message will clear logs
		},
	})
}

// Init initializes the model
func (m *model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case UIMessage:
		switch msg.Type {
		case UpdateStatsMsg:
			if data, ok := msg.Data.(StatsData); ok {
				m.stats = data
			}
		case UpdateDownloadMsg:
			if data, ok := msg.Data.(DownloadData); ok {
				m.download = data
			}
		case AppendLogMsg:
			if data, ok := msg.Data.(LogData); ok {
				if data.Message == "" {
					m.logs = make([]string, 0, 100)
				} else {
					timestamp := time.Now().Format("15:04:05")
					m.logs = append(m.logs, fmt.Sprintf("[%s] %s", timestamp, data.Message))
					// Keep only last 50 logs
					if len(m.logs) > 50 {
						m.logs = m.logs[len(m.logs)-50:]
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m *model) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion

	if !m.ready {
		view.SetContent("Initializing...")
		return view
	}

	// Stats section
	statsContent := fmt.Sprintf(
		"Interface: %s\n\n"+
			"Download: %s (占本网卡 %.2f%%)  保持至少: %s\n"+
			"Upload:   %s (占本网卡 %.2f%%)\n\n"+
			"Total Download: %s  Total Upload: %s",
		yellowStyle.Render(m.stats.IfaceName),
		greenStyle.Render(FormatBytes(m.stats.IfaceRx)), m.stats.DlRatio*100, cyanStyle.Render(fmt.Sprintf("%.2f%%", m.stats.TargetRatio*100)),
		greenStyle.Render(FormatBytes(m.stats.IfaceTx)), m.stats.UlRatio*100,
		blueStyle.Render(FormatBytes(m.stats.TotalRx)),
		blueStyle.Render(FormatBytes(m.stats.TotalTx)),
	)
	statsBox := zone.Mark("stats", statsStyle.Render(statsContent))

	// Download section
	var downloadContent string
	if m.download.IsDownloading {
		mode := "Downloading"
		if m.download.DryRun {
			mode = yellowStyle.Render("DRY RUN")
		}
		downloadContent = fmt.Sprintf("%s from:\n%s\n\nDownloaded: %s",
			mode, greenStyle.Render(m.download.CurrentURL), cyanStyle.Render(FormatBytes(m.download.Downloaded)))
	} else {
		downloadContent = grayStyle.Render("Idle")
	}
	downloadBox := zone.Mark("download", downloadStyle.Render(downloadContent))

	// Log section
	logContent := strings.Join(m.logs, "\n")
	if logContent == "" {
		logContent = grayStyle.Render("No logs yet...")
	}
	logBox := zone.Mark("logs", logStyle.Render(logContent))

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		statsBox,
		downloadBox,
		logBox,
	)

	view.SetContent(zone.Scan(content))
	return view
}
