package converter

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/chai2010/webp"
)

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
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	quality := 80
	if q, ok := opts["quality"].(string); ok {
		if strings.HasPrefix(q, "High") {
			quality = 90
		} else if strings.HasPrefix(q, "Medium") {
			quality = 75
		} else if strings.HasPrefix(q, "Low") {
			quality = 50
		}
	}

	targetLower := strings.ToLower(target)
	if strings.HasSuffix(targetLower, ".png") {
		return png.Encode(outFile, img)
	} else if strings.HasSuffix(targetLower, ".jpg") || strings.HasSuffix(targetLower, ".jpeg") {
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	} else if strings.HasSuffix(targetLower, ".webp") {
		return webp.Encode(outFile, img, &webp.Options{Quality: float32(quality)})
	}

	return fmt.Errorf("unsupported target format: %s", target)
}

func init() {
	_ = jpeg.Decode
	_ = png.Decode
	_ = webp.Decode
}
