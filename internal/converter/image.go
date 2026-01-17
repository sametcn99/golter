package converter

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"runtime"
	"strings"

	"github.com/chai2010/webp"
)

// ImageConverter handles image format conversions with optimization
type ImageConverter struct{}

func (c *ImageConverter) Name() string {
	return "Image Converter"
}

func (c *ImageConverter) isSupported(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

func (c *ImageConverter) CanConvert(srcExt, targetExt string) bool {
	return c.isSupported(srcExt) && c.isSupported(targetExt)
}

func (c *ImageConverter) SupportedSourceExtensions() []string {
	return []string{".jpg", ".jpeg", ".png", ".webp"}
}

func (c *ImageConverter) SupportedTargetFormats(srcExt string) []string {
	if !c.isSupported(srcExt) {
		return nil
	}
	return []string{".jpg", ".png", ".webp"}
}

func (c *ImageConverter) Convert(src, target string, opts Options) error {
	// Open source file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	// Parse quality option
	quality := parseQuality(opts)

	// Create output file
	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode to target format with optimized settings
	targetLower := strings.ToLower(target)
	switch {
	case strings.HasSuffix(targetLower, ".png"):
		encoder := png.Encoder{
			CompressionLevel: getPNGCompressionLevel(quality),
		}
		if err := encoder.Encode(outFile, img); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
		return nil

	case strings.HasSuffix(targetLower, ".jpg"), strings.HasSuffix(targetLower, ".jpeg"):
		if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
		return nil

	case strings.HasSuffix(targetLower, ".webp"):
		// WebP is excellent for compression
		if err := webp.Encode(outFile, img, &webp.Options{
			Quality:  float32(quality),
			Lossless: quality >= 95,
		}); err != nil {
			return fmt.Errorf("failed to encode WebP: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported target format: %s", target)
	}
}

// parseQuality extracts and normalizes quality from options
func parseQuality(opts Options) int {
	quality := 80 // Default
	if q, ok := opts["quality"].(string); ok {
		switch {
		case strings.Contains(q, "High"):
			quality = 92
		case strings.Contains(q, "Balanced"), strings.Contains(q, "Medium"):
			quality = 75
		case strings.Contains(q, "Compact"), strings.Contains(q, "Low"):
			quality = 55
		}
	}
	return quality
}

// getPNGCompressionLevel returns the appropriate PNG compression level
func getPNGCompressionLevel(quality int) png.CompressionLevel {
	switch {
	case quality >= 90:
		return png.NoCompression
	case quality >= 70:
		return png.DefaultCompression
	default:
		return png.BestCompression
	}
}

func init() {
	// Optimize for multi-core processing
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Register image decoders
	_ = jpeg.Decode
	_ = png.Decode
	_ = webp.Decode
}
