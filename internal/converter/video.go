package converter

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// VideoConverter handles video format conversions using ffmpeg
type VideoConverter struct{}

func (c *VideoConverter) Name() string {
	return "Video Converter (ffmpeg)"
}

func (c *VideoConverter) isSupported(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".webm" || ext == ".gif" || ext == ".mov"
}

func (c *VideoConverter) CanConvert(srcExt, targetExt string) bool {
	return c.isSupported(srcExt) && c.isSupported(targetExt)
}

func (c *VideoConverter) SupportedSourceExtensions() []string {
	return []string{".mp4", ".avi", ".mkv", ".webm", ".gif", ".mov"}
}

func (c *VideoConverter) SupportedTargetFormats(srcExt string) []string {
	if !c.isSupported(srcExt) {
		return nil
	}
	return []string{".mp4", ".avi", ".mkv", ".webm", ".gif", ".mov"}
}

func (c *VideoConverter) Convert(src, target string, opts Options) error {
	// Check if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found: please install ffmpeg to convert videos (https://ffmpeg.org)")
	}

	// Parse quality setting
	quality := parseVideoQuality(opts)

	// Build ffmpeg arguments
	args := buildFFmpegArgs(src, target, quality)

	// Execute ffmpeg
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// videoQuality holds encoding parameters
type videoQuality struct {
	crf     string
	preset  string
	audioBr string
}

// parseVideoQuality extracts quality settings from options
func parseVideoQuality(opts Options) videoQuality {
	q := videoQuality{
		crf:     "23",     // Default CRF (balanced)
		preset:  "medium", // Default preset
		audioBr: "192k",   // Default audio bitrate
	}

	if qualityStr, ok := opts["quality"].(string); ok {
		switch {
		case strings.Contains(qualityStr, "High"):
			q.crf = "18"
			q.preset = "slow"
			q.audioBr = "256k"
		case strings.Contains(qualityStr, "Balanced"), strings.Contains(qualityStr, "Medium"):
			q.crf = "23"
			q.preset = "medium"
			q.audioBr = "192k"
		case strings.Contains(qualityStr, "Compact"), strings.Contains(qualityStr, "Low"):
			q.crf = "28"
			q.preset = "fast"
			q.audioBr = "128k"
		}
	}

	return q
}

// buildFFmpegArgs constructs optimized ffmpeg arguments
func buildFFmpegArgs(src, target string, quality videoQuality) []string {
	targetLower := strings.ToLower(target)

	// Base arguments: input, overwrite, hide banner
	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-i", src,
	}

	// Use hardware acceleration if available
	args = append(args, "-threads", fmt.Sprintf("%d", runtime.NumCPU()))

	// Format-specific encoding
	switch {
	case strings.HasSuffix(targetLower, ".gif"):
		// GIF encoding with optimized palette
		args = append(args,
			"-vf", "fps=15,scale=480:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
			"-loop", "0",
		)

	case strings.HasSuffix(targetLower, ".webm"):
		// VP9 encoding for WebM
		args = append(args,
			"-c:v", "libvpx-vp9",
			"-crf", quality.crf,
			"-b:v", "0",
			"-c:a", "libopus",
			"-b:a", quality.audioBr,
		)

	case strings.HasSuffix(targetLower, ".mp4"):
		// H.264 encoding for MP4
		args = append(args,
			"-c:v", "libx264",
			"-crf", quality.crf,
			"-preset", quality.preset,
			"-c:a", "aac",
			"-b:a", quality.audioBr,
			"-movflags", "+faststart", // Enable streaming
		)

	case strings.HasSuffix(targetLower, ".mkv"):
		// H.265/HEVC for MKV (better compression)
		args = append(args,
			"-c:v", "libx265",
			"-crf", quality.crf,
			"-preset", quality.preset,
			"-c:a", "aac",
			"-b:a", quality.audioBr,
		)

	case strings.HasSuffix(targetLower, ".avi"):
		// MPEG-4 for AVI compatibility
		args = append(args,
			"-c:v", "mpeg4",
			"-q:v", quality.crf,
			"-c:a", "mp3",
			"-b:a", quality.audioBr,
		)

	case strings.HasSuffix(targetLower, ".mov"):
		// ProRes for MOV (high quality)
		args = append(args,
			"-c:v", "libx264",
			"-crf", quality.crf,
			"-preset", quality.preset,
			"-c:a", "aac",
			"-b:a", quality.audioBr,
		)

	default:
		// Generic encoding
		args = append(args,
			"-crf", quality.crf,
		)
	}

	// Add output file
	args = append(args, target)

	return args
}
