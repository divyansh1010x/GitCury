package utils

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"
)

// CreativeLoader provides a more engaging progress display with various animations
type CreativeLoader struct {
	message       string
	isActive      bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	writer        io.Writer
	hideInQuiet   bool
	animationType AnimationType
	currentPhase  string
	startTime     time.Time
	mu            sync.RWMutex
}

// AnimationType defines the type of animation to use
type AnimationType int

const (
	SpinnerAnimation AnimationType = iota
	DotsAnimation
	BarAnimation
	BrailleAnimation
	GitAnimation
	ProcessingAnimation
)

// Animation frames for different types
var (
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	dotsFrames    = []string{"   ", ".  ", ".. ", "..."}
	barFrames     = []string{"▱▱▱▱▱", "▰▱▱▱▱", "▰▰▱▱▱", "▰▰▰▱▱", "▰▰▰▰▱", "▰▰▰▰▰", "▱▰▰▰▰", "▱▱▰▰▰", "▱▱▱▰▰", "▱▱▱▱▰"}
	brailleFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	gitFrames     = []string{"🌱", "🌿", "🌳", "📝", "✨", "🚀"}
	processFrames = []string{"📂", "🔍", "⚙️", "🔧", "✅", "🎉"}
)

// Creative messages for different phases
var creativeMessages = map[string][]string{
	"analyzing": {
		"Analyzing code changes",
		"Examining file modifications",
		"Understanding your work",
		"Parsing repository state",
		"Inspecting changes",
	},
	"generating": {
		"Crafting commit messages",
		"Generating descriptions",
		"Creating summaries",
		"Composing commit text",
		"Building messages",
	},
	"clustering": {
		"Grouping related files",
		"Organizing changes",
		"Clustering modifications",
		"Categorizing updates",
		"Arranging file groups",
	},
	"processing": {
		"Processing files",
		"Working through changes",
		"Handling modifications",
		"Processing updates",
		"Managing file operations",
	},
	"finalizing": {
		"Finalizing results",
		"Completing operations",
		"Wrapping up",
		"Finishing touches",
		"Nearly done",
	},
}

// NewCreativeLoader creates a new creative loader
func NewCreativeLoader(message string, animationType AnimationType) *CreativeLoader {
	ctx, cancel := context.WithCancel(context.Background())

	return &CreativeLoader{
		message:       message,
		isActive:      false,
		ctx:           ctx,
		cancel:        cancel,
		writer:        os.Stdout,
		hideInQuiet:   true,
		animationType: animationType,
		currentPhase:  "processing",
	}
}

// SetWriter sets the output writer
func (cl *CreativeLoader) SetWriter(writer io.Writer) *CreativeLoader {
	cl.writer = writer
	return cl
}

// SetHideInQuiet sets whether to hide in quiet mode
func (cl *CreativeLoader) SetHideInQuiet(hide bool) *CreativeLoader {
	cl.hideInQuiet = hide
	return cl
}

// SetPhase updates the current phase for dynamic messages
func (cl *CreativeLoader) SetPhase(phase string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.currentPhase = phase
}

// Start begins the creative loader animation
func (cl *CreativeLoader) Start() {
	cl.mu.Lock()
	if cl.isActive {
		cl.mu.Unlock()
		return
	}
	cl.isActive = true
	cl.startTime = time.Now()
	cl.mu.Unlock()

	cl.wg.Add(1)
	go cl.animate()
}

// Stop stops the creative loader animation
func (cl *CreativeLoader) Stop() {
	cl.mu.Lock()
	if !cl.isActive {
		cl.mu.Unlock()
		return
	}
	cl.isActive = false
	cl.mu.Unlock()

	cl.cancel()
	cl.wg.Wait()

	// Clear the line properly
	if !cl.hideInQuiet || !IsQuietMode() {
		fmt.Fprintf(cl.writer, "\r%s\r", clearLine(100))
	}
}

// UpdateMessage updates the loader message
func (cl *CreativeLoader) UpdateMessage(message string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.message = message
}

// animate runs the animation loop
func (cl *CreativeLoader) animate() {
	defer cl.wg.Done()

	var frames []string
	var frameDuration time.Duration

	switch cl.animationType {
	case SpinnerAnimation:
		frames = spinnerFrames
		frameDuration = 80 * time.Millisecond
	case DotsAnimation:
		frames = dotsFrames
		frameDuration = 400 * time.Millisecond
	case BarAnimation:
		frames = barFrames
		frameDuration = 150 * time.Millisecond
	case BrailleAnimation:
		frames = brailleFrames
		frameDuration = 100 * time.Millisecond
	case GitAnimation:
		frames = gitFrames
		frameDuration = 600 * time.Millisecond
	case ProcessingAnimation:
		frames = processFrames
		frameDuration = 800 * time.Millisecond
	default:
		frames = spinnerFrames
		frameDuration = 80 * time.Millisecond
	}

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	frameIndex := 0
	messageIndex := 0
	messageChangeCounter := 0

	for {
		select {
		case <-cl.ctx.Done():
			return
		case <-ticker.C:
			if cl.hideInQuiet && IsQuietMode() {
				continue
			}

			cl.mu.RLock()
			if !cl.isActive {
				cl.mu.RUnlock()
				return
			}

			// Get current frame
			frame := frames[frameIndex]
			frameIndex = (frameIndex + 1) % len(frames)

			// Get dynamic message based on phase
			currentMessage := cl.message
			if messages, exists := creativeMessages[cl.currentPhase]; exists && len(messages) > 0 {
				// Change message every 8 animation cycles for smoother experience
				if messageChangeCounter%8 == 0 {
					messageIndex = rand.Intn(len(messages)) //nolint:gosec // Non-cryptographic use, display only
				}
				currentMessage = messages[messageIndex]
			}
			messageChangeCounter++

			// Create the animated line with consistent width
			elapsed := time.Since(cl.startTime)
			timeStr := formatDuration(elapsed)

			// Pad message to consistent width (40 characters) for smooth animation
			paddedMessage := padString(currentMessage, 40)

			var animatedLine string
			switch cl.animationType {
			case GitAnimation, ProcessingAnimation:
				animatedLine = fmt.Sprintf("\r%s %s [%s]", frame, paddedMessage, timeStr)
			case DotsAnimation:
				// For dots, show dots after padding
				baseMsg := padString(currentMessage, 37) // Leave room for dots
				animatedLine = fmt.Sprintf("\r🔄 %s%s [%s]", baseMsg, frame, timeStr)
			case BarAnimation:
				animatedLine = fmt.Sprintf("\r[%s] %s [%s]", frame, paddedMessage, timeStr)
			default:
				animatedLine = fmt.Sprintf("\r%s %s [%s]", frame, paddedMessage, timeStr)
			}

			// Clear any remaining characters and write the line
			fmt.Fprintf(cl.writer, "%s%s", animatedLine, clearEndOfLine())
			cl.mu.RUnlock()
		}
	}
}

// clearLine returns a string of spaces to clear a line of given length
func clearLine(length int) string {
	spaces := make([]byte, length)
	for i := range spaces {
		spaces[i] = ' '
	}
	return string(spaces)
}

// clearEndOfLine returns ANSI escape sequence to clear to end of line
func clearEndOfLine() string {
	return "\033[K"
}

// padString pads a string to a specific width, truncating if too long
func padString(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	format := fmt.Sprintf("%%-%ds", width)
	return fmt.Sprintf(format, s)
}

// Global creative loader instance for easy access
var globalLoader *CreativeLoader
var loaderMu sync.Mutex

// StartCreativeLoader starts a global creative loader
func StartCreativeLoader(message string, animationType AnimationType) {
	loaderMu.Lock()
	defer loaderMu.Unlock()

	if globalLoader != nil {
		globalLoader.Stop()
	}

	globalLoader = NewCreativeLoader(message, animationType)
	globalLoader.Start()
}

// UpdateCreativeLoaderMessage updates the global loader message
func UpdateCreativeLoaderMessage(message string) {
	loaderMu.Lock()
	defer loaderMu.Unlock()

	if globalLoader != nil {
		globalLoader.UpdateMessage(message)
	}
}

// UpdateCreativeLoaderPhase updates the global loader phase
func UpdateCreativeLoaderPhase(phase string) {
	loaderMu.Lock()
	defer loaderMu.Unlock()

	if globalLoader != nil {
		globalLoader.SetPhase(phase)
	}
}

// StopCreativeLoader stops the global creative loader
func StopCreativeLoader() {
	loaderMu.Lock()
	defer loaderMu.Unlock()

	if globalLoader != nil {
		globalLoader.Stop()
		globalLoader = nil
	}
}

// IsCreativeLoaderActive returns whether the global loader is active
func IsCreativeLoaderActive() bool {
	loaderMu.Lock()
	defer loaderMu.Unlock()

	return globalLoader != nil && globalLoader.isActive
}

// ShowCompletionMessage displays a completion message with appropriate styling
func ShowCompletionMessage(message string, success bool) {
	StopCreativeLoader()

	if IsQuietMode() {
		return
	}

	var icon string
	if success {
		icon = "✅"
	} else {
		icon = "❌"
	}

	fmt.Printf("\r%s %s\n", icon, message)
}
