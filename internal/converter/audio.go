package converter

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// AudioConverter handles audio format conversions using ffmpeg
type AudioConverter struct{}

func (c *AudioConverter) Name() string {
	return "Audio Converter (ffmpeg)"
}

func (c *AudioConverter) isSupported(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".mp3" || ext == ".wav" || ext == ".ogg" || ext == ".flac" || ext == ".m4a" || ext == ".aac"
}

func (c *AudioConverter) CanConvert(srcExt, targetExt string) bool {
	return c.isSupported(srcExt) && c.isSupported(targetExt)
}

func (c *AudioConverter) SupportedSourceExtensions() []string {
	return []string{".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac"}
}

func (c *AudioConverter) SupportedTargetFormats(srcExt string) []string {
	if !c.isSupported(srcExt) {
		return nil
	}
	return []string{".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac"}
}

func (c *AudioConverter) Convert(src, target string, opts Options) error {
	// Check if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found: please install ffmpeg to convert audio (https://ffmpeg.org)")
	}

	// Parse quality setting
	quality := parseAudioQuality(opts)

	// Build ffmpeg arguments
	args := buildAudioFFmpegArgs(src, target, quality)

	// Execute ffmpeg
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// audioQuality holds audio encoding parameters
type audioQuality struct {
	bitrate    string
	sampleRate string
}

// parseAudioQuality extracts quality settings from options
func parseAudioQuality(opts Options) audioQuality {
	q := audioQuality{
		bitrate:    "192k",  // Default bitrate
		sampleRate: "44100", // CD quality
	}

	if qualityStr, ok := opts["quality"].(string); ok {
		switch {
		case strings.Contains(qualityStr, "High"):
			q.bitrate = "320k"
			q.sampleRate = "48000"
		case strings.Contains(qualityStr, "Balanced"), strings.Contains(qualityStr, "Medium"):
			q.bitrate = "192k"
			q.sampleRate = "44100"
		case strings.Contains(qualityStr, "Compact"), strings.Contains(qualityStr, "Low"):
			q.bitrate = "128k"
			q.sampleRate = "44100"
		}
	}

	return q
}

// buildAudioFFmpegArgs constructs optimized ffmpeg arguments for audio
func buildAudioFFmpegArgs(src, target string, quality audioQuality) []string {
	targetLower := strings.ToLower(target)

	// Base arguments
	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-i", src,
		"-threads", fmt.Sprintf("%d", runtime.NumCPU()),
	}

	// Format-specific encoding
	switch {
	case strings.HasSuffix(targetLower, ".mp3"):
		// MP3 encoding with LAME
		args = append(args,
			"-c:a", "libmp3lame",
			"-b:a", quality.bitrate,
			"-ar", quality.sampleRate,
		)

	case strings.HasSuffix(targetLower, ".ogg"):
		// Vorbis encoding for OGG
		args = append(args,
			"-c:a", "libvorbis",
			"-b:a", quality.bitrate,
			"-ar", quality.sampleRate,
		)

	case strings.HasSuffix(targetLower, ".flac"):
		// FLAC is lossless - no bitrate needed
		args = append(args,
			"-c:a", "flac",
			"-ar", quality.sampleRate,
		)

	case strings.HasSuffix(targetLower, ".wav"):
		// PCM for WAV
		args = append(args,
			"-c:a", "pcm_s16le",
			"-ar", quality.sampleRate,
		)

	case strings.HasSuffix(targetLower, ".m4a"):
		// AAC for M4A
		args = append(args,
			"-c:a", "aac",
			"-b:a", quality.bitrate,
			"-ar", quality.sampleRate,
		)

	case strings.HasSuffix(targetLower, ".aac"):
		// AAC encoding
		args = append(args,
			"-c:a", "aac",
			"-b:a", quality.bitrate,
			"-ar", quality.sampleRate,
		)

	default:
		// Generic audio encoding
		args = append(args,
			"-b:a", quality.bitrate,
		)
	}

	// Add output file
	args = append(args, target)

	return args
}
