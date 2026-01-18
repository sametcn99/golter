package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DocumentConverter handles document format conversions
type DocumentConverter struct{}

var ebookExtensions = map[string]struct{}{
	".epub": {},
	".mobi": {},
	".azw":  {},
	".azw3": {},
	".fb2":  {},
}

func isEbookExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	_, ok := ebookExtensions[ext]
	return ok
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
	srcExt = strings.ToLower(srcExt)
	targetExt = strings.ToLower(targetExt)

	if !strings.HasPrefix(srcExt, ".") {
		srcExt = "." + srcExt
	}
	if !strings.HasPrefix(targetExt, ".") {
		targetExt = "." + targetExt
	}

	switch srcExt {
	case ".pdf":
		return targetExt == ".md" || targetExt == ".pdf"
	case ".md":
		return targetExt == ".pdf" || targetExt == ".html" || targetExt == ".docx" || isEbookExt(targetExt)
	case ".html":
		return targetExt == ".md" || targetExt == ".docx" || isEbookExt(targetExt)
	case ".docx":
		return targetExt == ".md" || targetExt == ".html" || targetExt == ".txt"
	case ".csv":
		return targetExt == ".xlsx" || targetExt == ".xls"
	case ".xlsx", ".xls":
		return targetExt == ".csv"
	case ".epub", ".mobi", ".azw", ".azw3", ".fb2":
		if isEbookExt(targetExt) {
			return srcExt != targetExt
		}
		return targetExt == ".pdf" || targetExt == ".md" || targetExt == ".html" || targetExt == ".txt"
	}
	return false
}

func (c *DocumentConverter) SupportedSourceExtensions() []string {
	return []string{".pdf", ".md", ".html", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".csv", ".xlsx", ".xls"}
}

func (c *DocumentConverter) SupportedTargetFormats(srcExt string) []string {
	srcExt = strings.ToLower(srcExt)
	if !strings.HasPrefix(srcExt, ".") {
		srcExt = "." + srcExt
	}

	switch srcExt {
	case ".pdf":
		return []string{".md", ".pdf"} // .pdf -> .pdf implies compression
	case ".md":
		return []string{".html", ".pdf", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2"}
	case ".html":
		return []string{".md", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2"}
	case ".docx":
		return []string{".md", ".html", ".txt"}
	case ".csv":
		return []string{".xlsx"}
	case ".xlsx", ".xls":
		return []string{".csv"}
	case ".epub":
		return []string{".pdf", ".md", ".html", ".mobi", ".azw", ".azw3", ".fb2", ".txt"}
	case ".mobi", ".azw", ".azw3", ".fb2":
		allTargets := []string{".epub", ".mobi", ".azw", ".azw3", ".fb2", ".pdf", ".html", ".txt", ".md"}
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

func (c *DocumentConverter) Convert(src, target string, opts Options) error {
	srcExt := strings.ToLower(filepath.Ext(src))
	targetExt := strings.ToLower(filepath.Ext(target))

	// Validate source file exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", src)
	}

	switch srcExt {
	case ".pdf":
		if targetExt == ".md" {
			return c.convertPDFToMarkdown(src, target)
		} else if targetExt == ".pdf" {
			return c.compressPDF(src, target)
		}
	case ".md":
		if targetExt == ".html" {
			return c.convertMarkdownToHTML(src, target)
		} else if targetExt == ".pdf" {
			return c.convertMarkdownToPDF(src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertMarkdownToEPUB(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertMarkdownToEbook(src, target, opts)
		}
	case ".html":
		if targetExt == ".md" {
			return c.convertHTMLToMarkdown(src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertHTMLToEPUB(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertHTMLToEbook(src, target, opts)
		}
	case ".docx":
		if targetExt == ".md" || targetExt == ".html" || targetExt == ".txt" {
			return c.convertWithPandoc(src, target, opts)
		}
	case ".csv":
		if targetExt == ".xlsx" || targetExt == ".xls" {
			return c.convertCSVToExcel(src, target)
		}
	case ".xlsx", ".xls":
		if targetExt == ".csv" {
			return c.convertExcelToCSV(src, target)
		}
	case ".epub":
		if targetExt == srcExt {
			break
		}
		if targetExt == ".md" {
			return c.convertEPUBToMarkdown(src, target)
		} else if targetExt == ".html" {
			return c.convertEPUBToHTML(src, target)
		} else if targetExt == ".pdf" {
			return c.convertEPUBToPDF(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertEbookWithCalibre(src, target, opts)
		} else if targetExt == ".txt" {
			return c.convertEbookWithCalibre(src, target, opts)
		}
	case ".mobi", ".azw", ".azw3", ".fb2":
		if targetExt == srcExt {
			break
		}
		if targetExt == ".md" {
			return c.convertEbookToMarkdown(src, target, opts)
		} else if targetExt == ".html" || targetExt == ".pdf" || targetExt == ".txt" || isEbookExt(targetExt) {
			return c.convertEbookWithCalibre(src, target, opts)
		}
	}

	return fmt.Errorf("unsupported conversion: %s to %s", srcExt, targetExt)
}
