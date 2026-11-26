package converter

import (
	"encoding/json"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	webpdecode "golang.org/x/image/webp"
)

func requireTool(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		if os.Getenv("GOLTER_REQUIRE_TOOLS") == "1" {
			t.Fatalf("required integration tool %s: %v", name, err)
		}
		t.Skipf("optional integration tool %s unavailable", name)
	}
	return path
}

func assertText(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), want) {
		t.Fatalf("%s: expected %q, got %q (%v)", path, want, data, err)
	}
}

func decodeImage(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if filepath.Ext(path) == ".webp" {
		img, err := webpdecode.Decode(file)
		if err != nil {
			t.Fatal(err)
		}
		return img
	}
	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func assertMedia(t *testing.T, path, kind string) {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-count_frames", "-show_entries", "stream=codec_type,codec_name,nb_read_frames:format=duration", "-of", "json", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	var probe struct {
		Streams []struct {
			Codec  string `json:"codec_name"`
			Kind   string `json:"codec_type"`
			Frames string `json:"nb_read_frames"`
		}
		Format struct{ Duration string }
	}
	if err = json.Unmarshal(out, &probe); err != nil {
		t.Fatal(err)
	}
	duration, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil || duration <= 0 {
		t.Fatalf("invalid duration: %s", out)
	}
	codecs := map[string]string{
		".mp3": "mp3", ".wav": "pcm_s16le", ".ogg": "vorbis",
		".flac": "flac", ".aac": "aac", ".m4a": "aac",
		".mp4": "h264", ".mov": "h264", ".avi": "mpeg4",
		".mkv": "hevc", ".webm": "vp9", ".gif": "gif",
	}
	for _, stream := range probe.Streams {
		frames, _ := strconv.Atoi(stream.Frames)
		if stream.Kind == kind && stream.Codec == codecs[filepath.Ext(path)] && frames > 0 {
			return
		}
	}
	t.Fatalf("no decodable %s stream: %s", kind, out)
}
