package tui

// Icons - Nerd Font compatible with fallbacks
const (
	// Navigation icons
	iconFolder       = "📁"
	iconFolderOpen   = "📂"
	iconFile         = "📄"
	iconBack         = "⬅️ "
	iconArrowRight   = "→"
	iconArrowDown    = "↓"
	iconChevronRight = "›"
	iconChevronDown  = "▾"

	// File type icons
	iconImage    = "🖼️ "
	iconVideo    = "🎬"
	iconAudio    = "🎵"
	iconDocument = "📄"
	iconPDF      = "📕"
	iconMarkdown = "📝"
	iconHTML     = "🌐"
	iconEPUB     = "📚"
	iconArchive  = "📦"
	iconGIF      = "🎞️ "
	iconCSV      = "📊"
	iconExcel    = "📗"

	// Status icons
	iconSelected    = "●"
	iconNotSelected = "○"
	iconSuccess     = "✅"
	iconError       = "❌"
	iconWarning     = "⚠️ "
	iconInfo        = "ℹ️ "
	iconSpinner     = "◐"
	iconLoading     = "⏳"
	iconDone        = "✓"

	// Action icons
	iconConvert  = "🔄"
	iconCompress = "📦"
	iconSettings = "⚙️ "
	iconQuit     = "🚪"
)

// Animated spinner frames
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Progress bar characters
const (
	progressBarFull  = "█"
	progressBarEmpty = "░"
	progressBarHead  = "▓"
)
