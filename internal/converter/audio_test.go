package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudioConverter_SupportedExtensions(t *testing.T) {
	c := &AudioConverter{}

	// Test SupportedSourceExtensions
	srcExts := c.SupportedSourceExtensions()
	if len(srcExts) == 0 {
		t.Error("SupportedSourceExtensions returned empty list")
	}
	expectedSrc := []string{".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac"}
	for _, exp := range expectedSrc {
		found := false
		for _, got := range srcExts {
			if got == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedSourceExtensions missing %s", exp)
		}
	}

	// Test SupportedTargetFormats
	for _, ext := range expectedSrc {
		targets := c.SupportedTargetFormats(ext)
		if len(targets) == 0 {
			t.Errorf("SupportedTargetFormats(%s) returned empty list", ext)
		}
	}

	// Test unsupported target formats
	if c.SupportedTargetFormats(".txt") != nil {
		t.Error("SupportedTargetFormats(.txt) should return nil")
	}
}

func TestAudioConverter_Convert_Error(t *testing.T) {
	c := &AudioConverter{}

	// Test non-existent file
	err := c.Convert("non_existent.wav", "out.mp3", Options{})
	if err == nil {
		t.Error("Convert should fail for non-existent file")
	}
}

func TestAudioConverter_Convert_Integration_Exhaustive(t *testing.T) {
	// Check for ffmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping integration test")
	}

	tmpDir, err := os.MkdirTemp("", "golter_audio_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &AudioConverter{}
	formats := []string{".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac"}
	qualities := []string{"High", "Balanced", "Low"}

	// Generate source files
	for _, ext := range formats {
		srcPath := filepath.Join(tmpDir, "src"+ext)
		// Generate 0.5s stereo audio
		cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "sine=frequency=1000:duration=0.5", "-ac", "2", "-y", srcPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Logf("Failed to generate %s: %v\nOutput: %s", ext, err, out)
			continue
		}
	}

	for _, srcExt := range formats {
		srcPath := filepath.Join(tmpDir, "src"+srcExt)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		for _, targetExt := range formats {
			if srcExt == targetExt {
				continue
			}

			for _, quality := range qualities {
				testName := fmt.Sprintf("%s->%s (%s)", srcExt, targetExt, quality)
				t.Run(testName, func(t *testing.T) {
					targetName := fmt.Sprintf("out_%s_%s%s", srcExt[1:], quality, targetExt)
					targetPath := filepath.Join(tmpDir, targetName)

					opts := Options{"quality": quality}
					err := c.Convert(srcPath, targetPath, opts)
					if err != nil {
						t.Errorf("Convert failed: %v", err)
						return
					}

					if _, err := os.Stat(targetPath); os.IsNotExist(err) {
						t.Errorf("Target file not created: %s", targetPath)
					}
				})
			}
		}
	}
}

func TestAudioConverter_Name(t *testing.T) {
	c := &AudioConverter{}
	if !strings.Contains(c.Name(), "Audio Converter") {
		t.Errorf("Name() = %v, want it to contain 'Audio Converter'", c.Name())
	}
}

func TestAudioConverter_CanConvert(t *testing.T) {
	c := &AudioConverter{}
	tests := []struct {
		src    string
		target string
		want   bool
	}{
		{".mp3", ".wav", true},
		{".wav", ".mp3", true},
		{".flac", ".m4a", true},
		{".txt", ".mp3", false},
		{".mp3", ".txt", false},
	}

	for _, tt := range tests {
		if got := c.CanConvert(tt.src, tt.target); got != tt.want {
			t.Errorf("CanConvert(%q, %q) = %v, want %v", tt.src, tt.target, got, tt.want)
		}
	}
}

func TestParseAudioQuality(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want audioQuality
	}{
		{"Default", Options{}, audioQuality{bitrate: "192k", sampleRate: "44100"}},
		{"High", Options{"quality": "High"}, audioQuality{bitrate: "320k", sampleRate: "48000"}},
		{"Compact", Options{"quality": "Compact"}, audioQuality{bitrate: "128k", sampleRate: "44100"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAudioQuality(tt.opts)
			if got != tt.want {
				t.Errorf("parseAudioQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildAudioFFmpegArgs(t *testing.T) {
	q := audioQuality{bitrate: "192k", sampleRate: "44100"}

	tests := []struct {
		name   string
		src    string
		target string
		check  func([]string) bool
	}{
		{
			"MP3 Conversion", "input.wav", "output.mp3",
			func(args []string) bool {
				return contains(args, "-c:a", "libmp3lame")
			},
		},
		{
			"FLAC Conversion", "input.mp3", "output.flac",
			func(args []string) bool {
				return contains(args, "-c:a", "flac")
			},
		},
		{
			"WAV Conversion", "input.mp3", "output.wav",
			func(args []string) bool {
				return contains(args, "-c:a", "pcm_s16le")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildAudioFFmpegArgs(tt.src, tt.target, q)
			if !tt.check(args) {
				t.Errorf("buildAudioFFmpegArgs() args = %v, failed check", args)
			}
		})
	}
}
