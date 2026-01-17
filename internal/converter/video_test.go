package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVideoConverter_SupportedExtensions(t *testing.T) {
	c := &VideoConverter{}

	// Test SupportedSourceExtensions
	srcExts := c.SupportedSourceExtensions()
	if len(srcExts) == 0 {
		t.Error("SupportedSourceExtensions returned empty list")
	}
	expectedSrc := []string{".mp4", ".avi", ".mkv", ".webm", ".gif", ".mov"}
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

func TestVideoConverter_Convert_Error(t *testing.T) {
	c := &VideoConverter{}

	// Test non-existent file
	err := c.Convert("non_existent.mp4", "out.mkv", Options{})
	if err == nil {
		t.Error("Convert should fail for non-existent file")
	}
}

func TestVideoConverter_Convert_Integration_Exhaustive(t *testing.T) {
	// Check for ffmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping integration test")
	}

	tmpDir, err := os.MkdirTemp("", "golter_video_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &VideoConverter{}
	formats := []string{".mp4", ".avi", ".mkv", ".webm", ".mov"} // GIF is special, maybe test separately or include
	// GIF generation might be different or slow, but let's include it if possible.
	// ffmpeg testsrc to gif works.
	formats = append(formats, ".gif")

	qualities := []string{"High", "Balanced", "Low"}

	// Generate source files
	for _, ext := range formats {
		srcPath := filepath.Join(tmpDir, "src"+ext)
		// Generate 0.5s video to be fast
		cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.5:size=64x64:rate=10", "-y", srcPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Logf("Failed to generate %s: %v\nOutput: %s", ext, err, out)
			// Don't fail completely, just skip this source if ffmpeg fails (e.g. codec missing)
			continue
		}
	}

	for _, srcExt := range formats {
		srcPath := filepath.Join(tmpDir, "src"+srcExt)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue // Skip if generation failed
		}

		for _, targetExt := range formats {
			if srcExt == targetExt {
				continue // Skip same format conversion unless we want to test re-encoding (we do for compression)
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

func TestVideoConverter_Name(t *testing.T) {
	c := &VideoConverter{}
	if !strings.Contains(c.Name(), "Video Converter") {
		t.Errorf("Name() = %v, want it to contain 'Video Converter'", c.Name())
	}
}

func TestVideoConverter_CanConvert(t *testing.T) {
	c := &VideoConverter{}
	tests := []struct {
		src    string
		target string
		want   bool
	}{
		{".mp4", ".mkv", true},
		{".avi", ".mp4", true},
		{".gif", ".webm", true},
		{".mov", ".gif", true},
		{".txt", ".mp4", false},
		{".mp4", ".txt", false},
	}

	for _, tt := range tests {
		if got := c.CanConvert(tt.src, tt.target); got != tt.want {
			t.Errorf("CanConvert(%q, %q) = %v, want %v", tt.src, tt.target, got, tt.want)
		}
	}
}

func TestParseVideoQuality(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want videoQuality
	}{
		{"Default", Options{}, videoQuality{crf: "23", preset: "medium", audioBr: "192k"}},
		{"High", Options{"quality": "High"}, videoQuality{crf: "18", preset: "slow", audioBr: "256k"}},
		{"Balanced", Options{"quality": "Balanced"}, videoQuality{crf: "23", preset: "medium", audioBr: "192k"}},
		{"Compact", Options{"quality": "Compact"}, videoQuality{crf: "28", preset: "fast", audioBr: "128k"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVideoQuality(tt.opts)
			if got != tt.want {
				t.Errorf("parseVideoQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFFmpegArgs(t *testing.T) {
	q := videoQuality{crf: "23", preset: "medium", audioBr: "192k"}

	tests := []struct {
		name   string
		src    string
		target string
		check  func([]string) bool
	}{
		{
			"MP4 Conversion", "input.avi", "output.mp4",
			func(args []string) bool {
				return contains(args, "-c:v", "libx264") && contains(args, "-c:a", "aac")
			},
		},
		{
			"WebM Conversion", "input.mp4", "output.webm",
			func(args []string) bool {
				return contains(args, "-c:v", "libvpx-vp9") && contains(args, "-c:a", "libopus")
			},
		},
		{
			"GIF Conversion", "input.mp4", "output.gif",
			func(args []string) bool {
				// Check for palettegen filter
				for _, a := range args {
					if strings.Contains(a, "palettegen") {
						return true
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildFFmpegArgs(tt.src, tt.target, q)
			if !tt.check(args) {
				t.Errorf("buildFFmpegArgs() args = %v, failed check", args)
			}
		})
	}
}

func contains(slice []string, items ...string) bool {
	for i := 0; i <= len(slice)-len(items); i++ {
		match := true
		for j := 0; j < len(items); j++ {
			if slice[i+j] != items[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
