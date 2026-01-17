package converter

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/chai2010/webp"
)

func createTestImage(t *testing.T, path string) {
	width := 10
	height := 10
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.White)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer f.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".png":
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("failed to encode png: %v", err)
		}
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(f, img, nil); err != nil {
			t.Fatalf("failed to encode jpeg: %v", err)
		}
	case ".webp":
		if err := webp.Encode(f, img, nil); err != nil {
			t.Fatalf("failed to encode webp: %v", err)
		}
	default:
		t.Fatalf("unsupported test image format: %s", ext)
	}
}

func TestImageConverter_Convert_Integration_Exhaustive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_img_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &ImageConverter{}
	formats := []string{".jpg", ".png", ".webp"}
	qualities := []string{"High", "Balanced", "Low"}

	for _, srcExt := range formats {
		srcName := "test" + srcExt
		srcPath := filepath.Join(tmpDir, srcName)
		createTestImage(t, srcPath)

		for _, targetExt := range formats {
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

func TestImageConverter_Name(t *testing.T) {
	c := &ImageConverter{}
	if c.Name() != "Image Converter" {
		t.Errorf("Name() = %v, want %v", c.Name(), "Image Converter")
	}
}

func TestImageConverter_CanConvert(t *testing.T) {
	c := &ImageConverter{}
	tests := []struct {
		src    string
		target string
		want   bool
	}{
		{".jpg", ".png", true},
		{".jpeg", ".webp", true},
		{".png", ".jpg", true},
		{".webp", ".jpeg", true},
		{".txt", ".png", false},
		{".jpg", ".txt", false},
		{".JPG", ".PNG", true}, // Case insensitive
	}

	for _, tt := range tests {
		if got := c.CanConvert(tt.src, tt.target); got != tt.want {
			t.Errorf("CanConvert(%q, %q) = %v, want %v", tt.src, tt.target, got, tt.want)
		}
	}
}

func TestImageConverter_SupportedSourceExtensions(t *testing.T) {
	c := &ImageConverter{}
	got := c.SupportedSourceExtensions()
	want := []string{".jpg", ".jpeg", ".png", ".webp"}

	if len(got) != len(want) {
		t.Errorf("SupportedSourceExtensions() length = %v, want %v", len(got), len(want))
	}

	// Check content
	m := make(map[string]bool)
	for _, v := range got {
		m[v] = true
	}
	for _, v := range want {
		if !m[v] {
			t.Errorf("SupportedSourceExtensions() missing %v", v)
		}
	}
}

func TestImageConverter_SupportedTargetFormats(t *testing.T) {
	c := &ImageConverter{}

	// Supported source
	got := c.SupportedTargetFormats(".jpg")
	want := []string{".jpg", ".png", ".webp"}
	if len(got) != len(want) {
		t.Errorf("SupportedTargetFormats(.jpg) length = %v, want %v", len(got), len(want))
	}

	// Unsupported source
	got = c.SupportedTargetFormats(".txt")
	if got != nil {
		t.Errorf("SupportedTargetFormats(.txt) = %v, want nil", got)
	}
}

func TestParseQuality(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want int
	}{
		{"Default", Options{}, 80},
		{"High", Options{"quality": "High"}, 92},
		{"Balanced", Options{"quality": "Balanced"}, 75},
		{"Medium", Options{"quality": "Medium"}, 75},
		{"Compact", Options{"quality": "Compact"}, 55},
		{"Low", Options{"quality": "Low"}, 55},
		{"Unknown", Options{"quality": "Super"}, 80},
		{"NotString", Options{"quality": 123}, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseQuality(tt.opts); got != tt.want {
				t.Errorf("parseQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPNGCompressionLevel(t *testing.T) {
	tests := []struct {
		quality int
		want    png.CompressionLevel
	}{
		{95, png.NoCompression},
		{90, png.NoCompression},
		{80, png.DefaultCompression},
		{70, png.DefaultCompression},
		{60, png.BestCompression},
		{0, png.BestCompression},
	}

	for _, tt := range tests {
		if got := getPNGCompressionLevel(tt.quality); got != tt.want {
			t.Errorf("getPNGCompressionLevel(%v) = %v, want %v", tt.quality, got, tt.want)
		}
	}
}
