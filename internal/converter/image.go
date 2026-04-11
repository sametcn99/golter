package converter

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"slices"
	"strings"

	"github.com/chai2010/webp"
)

// ImageConverter handles image format conversions with optimization
type ImageConverter struct{}

func (c *ImageConverter) Name() string {
	return "Image Converter"
}

func (c *ImageConverter) isSupported(ext string) bool {
	ext = normalizeExt(ext)
	return slices.Contains(c.SupportedSourceExtensions(), ext)
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

func (c *ImageConverter) Convert(ctx context.Context, src, target string, opts Options) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- c.doConvert(src, target, opts)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("image conversion cancelled: %w", ctx.Err())
	case err := <-errCh:
		return err
	}
}

func (c *ImageConverter) doConvert(src, target string, opts Options) error {
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
	q := opts.Quality
	if q != "" {
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
	// Register image decoders to prevent unused import errors if they were imported just for registration
	// Though typically `import _ "image/jpeg"` is used, this makes it explicit here.
	_ = jpeg.Decode
	_ = png.Decode
	_ = webp.Decode
}
