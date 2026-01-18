package converter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-pdf/fpdf"
	"github.com/taylorskalyo/goreader/epub"
)

func (c *DocumentConverter) convertEbookToMarkdown(src, target string, opts Options) error {
	tempHTML, cleanup, err := tempPathWithExt("golter_ebook_html", ".html")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertEbookWithCalibre(src, tempHTML, opts); err != nil {
		return err
	}

	return c.convertHTMLToMarkdown(tempHTML, target)
}

func (c *DocumentConverter) convertEbookWithCalibre(src, target string, opts Options) error {
	_, err := exec.LookPath("ebook-convert")
	if err != nil {
		return fmt.Errorf("ebook-convert not found: please install Calibre to convert ebook formats (https://calibre-ebook.com)")
	}

	args := []string{src, target}
	if extra, ok := opts["ebookArgs"].([]string); ok && len(extra) > 0 {
		args = append(args, extra...)
	} else if extraStr, ok := opts["ebookArgs"].(string); ok && strings.TrimSpace(extraStr) != "" {
		args = append(args, strings.Fields(extraStr)...)
	}

	cmd := exec.Command("ebook-convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebook-convert failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToMarkdown(src, target string) error {
	rc, err := epub.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer rc.Close()

	if len(rc.Rootfiles) == 0 {
		return fmt.Errorf("no rootfiles found in EPUB")
	}

	book := rc.Rootfiles[0]
	var contentBuilder strings.Builder
	converter := md.NewConverter("", true, nil)

	// Iterate through spine items
	for _, item := range book.Spine.Itemrefs {
		if item.Item == nil {
			continue
		}

		// Open the file from the EPUB
		f, err := item.Item.Open()
		if err != nil {
			continue
		}

		b, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}

		// Convert HTML content to Markdown
		markdown, err := converter.ConvertString(string(b))
		if err != nil {
			continue
		}

		contentBuilder.WriteString(markdown)
		contentBuilder.WriteString("\n\n---\n\n")
	}

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToHTML(src, target string) error {
	rc, err := epub.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer rc.Close()

	if len(rc.Rootfiles) == 0 {
		return fmt.Errorf("no rootfiles found in EPUB")
	}

	book := rc.Rootfiles[0]
	var contentBuilder strings.Builder

	contentBuilder.WriteString("<!DOCTYPE html><html><body>")

	// Iterate through spine items
	for _, item := range book.Spine.Itemrefs {
		if item.Item == nil {
			continue
		}

		f, err := item.Item.Open()
		if err != nil {
			continue
		}

		b, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}

		// Simple concatenation of body content would be better, but full HTML concatenation is easier for now
		// Ideally we should strip <html>, <head>, <body> tags and just take the inner content
		// For simplicity, we just append the whole thing, browsers handle nested html tags somewhat okay-ish
		// or better: just append the raw content.
		contentBuilder.Write(b)
		contentBuilder.WriteString("<hr>")
	}

	contentBuilder.WriteString("</body></html>")

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToPDF(src, target string) error {
	// First convert to HTML
	tempHTML := strings.TrimSuffix(target, filepath.Ext(target)) + "_temp.html"
	if err := c.convertEPUBToHTML(src, tempHTML); err != nil {
		return err
	}
	defer os.Remove(tempHTML)

	// Then convert HTML to PDF (using existing logic logic, but we need to read the temp file)
	// We can reuse convertMarkdownToPDF logic but starting from HTML

	// Read HTML source
	source, err := os.ReadFile(tempHTML)
	if err != nil {
		return fmt.Errorf("failed to read temp HTML file: %w", err)
	}

	// Create PDF
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.SetMargins(20, 20, 20)
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "", 12)

	_, lineHt := pdfDoc.GetFontSize()
	html := pdfDoc.HTMLBasicNew()
	html.Write(lineHt, string(source))

	if err := pdfDoc.OutputFileAndClose(target); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}
