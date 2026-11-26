package converter

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// DocumentConverter handles document format conversions
type DocumentConverter struct{}

var ebookExtensions = []string{
	".epub",
	".mobi",
	".azw",
	".azw3",
	".fb2",
}

func isEbookExt(ext string) bool {
	ext = normalizeExt(ext)
	return slices.Contains(ebookExtensions, ext)
}

func tempPathWithExt(prefix, ext string) (string, func(), error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	return filepath.Join(dir, "temp"+ext), cleanup, nil
}

func (c *DocumentConverter) Name() string {
	return "Document Converter"
}

func (c *DocumentConverter) CanConvert(srcExt, targetExt string) bool {
	return slices.Contains(c.SupportedTargetFormats(srcExt), normalizeExt(targetExt))
}

func (c *DocumentConverter) SupportedSourceExtensions() []string {
	return []string{".pdf", ".md", ".html", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".csv", ".xlsx"}
}

func (c *DocumentConverter) SupportedTargetFormats(srcExt string) []string {
	srcExt = normalizeExt(srcExt)

	switch srcExt {
	case ".pdf":
		return []string{".md", ".pdf"} // .pdf -> .pdf implies compression
	case ".md":
		return []string{".html", ".pdf", ".docx", ".epub", ".mobi", ".azw3", ".fb2"}
	case ".html":
		return []string{".md", ".docx", ".epub", ".mobi", ".azw3", ".fb2"}
	case ".docx":
		return []string{".md", ".html", ".txt"}
	case ".csv":
		return []string{".xlsx"}
	case ".xlsx":
		return []string{".csv"}
	case ".epub":
		return []string{".pdf", ".md", ".html", ".mobi", ".azw3", ".fb2", ".txt"}
	case ".mobi", ".azw", ".azw3", ".fb2":
		allTargets := []string{".epub", ".mobi", ".azw3", ".fb2", ".pdf", ".html", ".txt", ".md"}
		filtered := make([]string, 0, len(allTargets))
		for _, t := range allTargets {
			if t == srcExt {
				continue
			}
			filtered = append(filtered, t)
		}
		return filtered
	}
	return nil
}

func (c *DocumentConverter) Convert(ctx context.Context, src, target string, opts Options) error {
	srcExt := normalizeExt(filepath.Ext(src))
	targetExt := normalizeExt(filepath.Ext(target))

	if !c.CanConvert(srcExt, targetExt) {
		return fmt.Errorf("unsupported conversion: %s to %s", srcExt, targetExt)
	}
	// Validate source file exists
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("source file not found: %s", src)
	}

	switch srcExt {
	case ".pdf":
		if targetExt == ".md" {
			return c.convertPDFToMarkdown(ctx, src, target)
		} else if targetExt == ".pdf" {
			return c.compressPDF(ctx, src, target)
		}
	case ".md":
		if targetExt == ".html" {
			return c.convertMarkdownToHTML(ctx, src, target)
		} else if targetExt == ".pdf" {
			return c.convertMarkdownToPDF(ctx, src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(ctx, src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertMarkdownToEPUB(ctx, src, target)
		} else if isEbookExt(targetExt) {
			return c.convertMarkdownToEbook(ctx, src, target, opts)
		}
	case ".html":
		if targetExt == ".md" {
			return c.convertHTMLToMarkdown(ctx, src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(ctx, src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertHTMLToEPUB(ctx, src, target)
		} else if isEbookExt(targetExt) {
			return c.convertHTMLToEbook(ctx, src, target, opts)
		}
	case ".docx":
		return c.convertWithPandoc(ctx, src, target, opts)
	case ".csv":
		return c.convertCSVToExcel(ctx, src, target)
	case ".xlsx":
		return c.convertExcelToCSV(ctx, src, target)
	case ".epub":
		switch targetExt {
		case ".md":
			return c.convertEPUBToMarkdown(ctx, src, target)
		case ".html":
			return c.convertEPUBToHTML(ctx, src, target)
		case ".pdf":
			return c.convertEPUBToPDF(ctx, src, target)
		default:
			return c.convertEbookWithCalibre(ctx, src, target, opts)
		}
	case ".mobi", ".azw", ".azw3", ".fb2":
		if targetExt == ".md" || targetExt == ".html" {
			return c.convertEbookToTextFormat(ctx, src, target, opts)
		}
		return c.convertEbookWithCalibre(ctx, src, target, opts)
	}
	return fmt.Errorf("unsupported conversion: %s to %s", srcExt, targetExt)
}
