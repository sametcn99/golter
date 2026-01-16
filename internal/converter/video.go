package converter

import (
	"fmt"
	"os/exec"
	"strings"
)

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
		return fmt.Errorf("ffmpeg not found: please install ffmpeg to convert videos")
	}

	crf := "23" // Default
	if q, ok := opts["quality"].(string); ok {
		if strings.HasPrefix(q, "High") {
			crf = "23"
		} else if strings.HasPrefix(q, "Medium") {
			crf = "28"
		} else if strings.HasPrefix(q, "Low") {
			crf = "35"
		}
	}

	args := []string{"-i", src, "-y"}

	// Apply CRF for non-GIF formats as a generic compression method
	if !strings.HasSuffix(strings.ToLower(target), ".gif") {
		args = append(args, "-crf", crf)
	}

	args = append(args, target)

	// -y overwrites output file without asking
	cmd := exec.Command("ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %s, output: %s", err, string(output))
	}

	return nil
}
