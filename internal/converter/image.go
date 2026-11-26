package converter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"slices"
	"strings"

	webpdecode "golang.org/x/image/webp"

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
	return stagedOutput(ctx, target, func(temp string) error { return c.encode(ctx, src, temp, opts) }, os.Rename)
}

func (c *ImageConverter) encode(ctx context.Context, src, target string, opts Options) (err error) {
	// Open source file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image
	var img image.Image
	var format string
	input := bufio.NewReader(file)
	header, _ := input.Peek(12)
	if len(header) == 12 && string(header[:4]) == "RIFF" && string(header[8:]) == "WEBP" {
		format = "webp"
		img, err = webpdecode.Decode(input)
	} else {
		img, format, err = image.Decode(input)
	}
	if err != nil {
		return fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// Parse quality option
	quality := parseQuality(opts)

	// Create output file
	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { err = errors.Join(err, outFile.Close()) }()

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
		// libwebp takes straight RGBA bytes through its image.RGBA API.
		straight, ok := img.(*image.NRGBA)
		if !ok {
			straight = image.NewNRGBA(img.Bounds())
			draw.Draw(straight, straight.Bounds(), img, img.Bounds().Min, draw.Src)
		}
		pixels := &image.RGBA{Pix: straight.Pix, Stride: straight.Stride, Rect: straight.Rect}
		if err := webp.Encode(outFile, pixels, &webp.Options{
			Quality:  float32(quality),
			Lossless: quality >= 92,
			Exact:    quality >= 92,
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
