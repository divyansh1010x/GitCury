package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressReporter provides an interface for reporting progress of long-running operations
type ProgressReporter struct {
	total         int64
	current       int64
	start         time.Time
	lastUpdate    time.Time
	updateMu      sync.Mutex
	message       string
	finished      bool
	width         int
	writer        io.Writer
	hideInQuiet   bool
	updateFreq    time.Duration
	progressChar  string
	spinnerChars  []string
	spinnerPos    int
	spinnerActive bool
	spinnerTicker *time.Ticker
	spinnerDone   chan struct{}
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(total int64, message string) *ProgressReporter {
	return &ProgressReporter{
		total:        total,
		current:      0,
		start:        time.Now(),
		lastUpdate:   time.Now(),
		message:      message,
		finished:     false,
		width:        50,
		writer:       os.Stdout,
		hideInQuiet:  true,
		updateFreq:   200 * time.Millisecond,
		progressChar: "■",
		spinnerChars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		spinnerPos:   0,
	}
}

// NewIndeterminateProgressReporter creates a progress reporter for operations
// where the total amount of work is unknown
func NewIndeterminateProgressReporter(message string) *ProgressReporter {
	p := NewProgressReporter(-1, message)
	return p
}

// SetWidth sets the width of the progress bar
func (p *ProgressReporter) SetWidth(width int) *ProgressReporter {
	p.width = width
	return p
}

// SetWriter sets the io.Writer where progress updates are written
func (p *ProgressReporter) SetWriter(writer io.Writer) *ProgressReporter {
	p.writer = writer
	return p
}

// SetHideInQuiet sets whether the progress bar should be hidden in quiet mode
func (p *ProgressReporter) SetHideInQuiet(hide bool) *ProgressReporter {
	p.hideInQuiet = hide
	return p
}

// Start begins the progress reporting
func (p *ProgressReporter) Start() *ProgressReporter {
	p.updateMu.Lock()
	defer p.updateMu.Unlock()

	p.start = time.Now()
	p.lastUpdate = time.Now()
	p.finished = false

	// Start spinner for indeterminate progress
	if p.total < 0 {
		p.spinnerActive = true
		p.spinnerDone = make(chan struct{})
		p.spinnerTicker = time.NewTicker(100 * time.Millisecond)

		go func() {
			for {
				select {
				case <-p.spinnerTicker.C:
					p.updateMu.Lock()
					if p.spinnerActive {
						p.spinnerPos = (p.spinnerPos + 1) % len(p.spinnerChars)
						p.renderSpinner()
					}
					p.updateMu.Unlock()
				case <-p.spinnerDone:
					return
				}
			}
		}()
	} else {
		// Render the initial progress bar
		p.render()
	}

	return p
}

// Update updates the current progress
func (p *ProgressReporter) Update(current int64) {
	p.updateMu.Lock()
	defer p.updateMu.Unlock()

	if p.finished {
		return
	}

	p.current = current

	// Only update if enough time has passed since last update
	if time.Since(p.lastUpdate) >= p.updateFreq {
		if p.total < 0 {
			// Indeterminate progress uses spinner
			return
		}

		// Render the progress bar
		p.render()
		p.lastUpdate = time.Now()
	}
}

// UpdateMessage updates the message displayed with the progress bar
func (p *ProgressReporter) UpdateMessage(message string) {
	p.updateMu.Lock()
	defer p.updateMu.Unlock()

	if p.finished {
		return
	}

	p.message = message

	if p.total < 0 {
		p.renderSpinner()
	} else {
		p.render()
	}

	p.lastUpdate = time.Now()
}

// Increment increases the current progress by the specified amount
func (p *ProgressReporter) Increment(amount int64) {
	p.Update(p.current + amount)
}

// Done completes the progress reporting
func (p *ProgressReporter) Done() {
	p.updateMu.Lock()
	defer p.updateMu.Unlock()

	if p.finished {
		return
	}

	// If this is a spinner, stop it
	if p.total < 0 && p.spinnerActive {
		p.spinnerActive = false
		p.spinnerTicker.Stop()
		close(p.spinnerDone)

		// Clear the spinner line
		fmt.Fprintf(p.writer, "\r%s\r", strings.Repeat(" ", 80))
	} else if p.total >= 0 {
		// Make sure we show 100% completion
		p.current = p.total
		p.render()
	}

	fmt.Fprintln(p.writer)
	p.finished = true
}

// render displays the current progress bar
func (p *ProgressReporter) render() {
	if p.hideInQuiet && IsQuietMode() {
		return
	}

	percent := float64(p.current) / float64(p.total) * 100
	if percent > 100 {
		percent = 100
	}

	elapsed := time.Since(p.start)

	// Calculate ETA
	var etaStr string
	if p.current > 0 {
		itemsPerSecond := float64(p.current) / elapsed.Seconds()
		if itemsPerSecond > 0 {
			remainingItems := p.total - p.current
			etaSeconds := float64(remainingItems) / itemsPerSecond
			eta := time.Duration(etaSeconds) * time.Second
			if eta > 0 {
				etaStr = fmt.Sprintf(" ETA: %s", formatDuration(eta))
			}
		}
	}

	// Build progress bar
	width := p.width
	completed := int(float64(width) * float64(p.current) / float64(p.total))

	progressBar := strings.Repeat(p.progressChar, completed) + strings.Repeat("░", width-completed)

	// Format and print the progress line
	fmt.Fprintf(
		p.writer,
		"\r%s [%s] %.1f%% (%d/%d)%s",
		p.message,
		progressBar,
		percent,
		p.current,
		p.total,
		etaStr,
	)
}

// renderSpinner displays the spinner for indeterminate progress
func (p *ProgressReporter) renderSpinner() {
	if p.hideInQuiet && IsQuietMode() {
		return
	}

	spinner := p.spinnerChars[p.spinnerPos]
	elapsed := formatDuration(time.Since(p.start))

	fmt.Fprintf(
		p.writer,
		"\r%s %s [%s]",
		spinner,
		p.message,
		elapsed,
	)
}

// formatDuration formats a duration in a user-friendly way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// Global quiet mode flag
var quietMode bool
var quietModeOnce sync.Once
var quietModeMu sync.RWMutex

// SetQuietMode sets the global quiet mode flag
func SetQuietMode(quiet bool) {
	quietModeOnce.Do(func() {
		quietModeMu.Lock()
		quietMode = quiet
		quietModeMu.Unlock()
	})
}

// IsQuietMode returns whether quiet mode is enabled
func IsQuietMode() bool {
	quietModeMu.RLock()
	defer quietModeMu.RUnlock()
	return quietMode
}
