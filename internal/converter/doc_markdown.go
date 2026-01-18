package converter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goepub "github.com/bmaupin/go-epub"
	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark"
)

func (c *DocumentConverter) convertMarkdownToHTML(src, target string) error {
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(source, &buf); err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Wrap in basic HTML structure
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        pre {
            background: #f4f4f4;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
        code {
            background: #f4f4f4;
            padding: 2px 5px;
            border-radius: 3px;
        }
        img { max-width: 100%%; }
    </style>
</head>
<body>
%s
</body>
</html>`, filepath.Base(src), buf.String())

	if err := os.WriteFile(target, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToPDF(src, target string) error {
	// Read markdown source
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Convert to HTML first
	var htmlBuf bytes.Buffer
	if err := goldmark.Convert(source, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Create PDF
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.SetMargins(20, 20, 20)
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "", 12)

	_, lineHt := pdfDoc.GetFontSize()
	html := pdfDoc.HTMLBasicNew()
	html.Write(lineHt, htmlBuf.String())

	if err := pdfDoc.OutputFileAndClose(target); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToEPUB(src, target string) error {
	// Read and convert markdown to HTML
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := goldmark.Convert(source, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Create EPUB
	title := strings.TrimSuffix(filepath.Base(src), ".md")
	e := goepub.NewEpub(title)
	e.SetAuthor("Golter Converter")

	_, err = e.AddSection(htmlBuf.String(), "Chapter 1", "", "")
	if err != nil {
		return fmt.Errorf("failed to add content to EPUB: %w", err)
	}

	if err := e.Write(target); err != nil {
		return fmt.Errorf("failed to write EPUB file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToEbook(src, target string, opts Options) error {
	if strings.EqualFold(filepath.Ext(target), ".epub") {
		return c.convertMarkdownToEPUB(src, target)
	}

	tempEPUB, cleanup, err := tempPathWithExt("golter_ebook_epub", ".epub")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertMarkdownToEPUB(src, tempEPUB); err != nil {
		return err
	}

	return c.convertEbookWithCalibre(tempEPUB, target, opts)
}
