package converter

import (
	"context"
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

func (c *DocumentConverter) convertEbookToMarkdown(ctx context.Context, src, target string, opts Options) error {
	tempHTML, cleanup, err := tempPathWithExt("golter_ebook_html", ".html")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertEbookWithCalibre(ctx, src, tempHTML, opts); err != nil {
		return err
	}

	return c.convertHTMLToMarkdown(ctx, tempHTML, target)
}

func (c *DocumentConverter) convertEbookWithCalibre(ctx context.Context, src, target string, opts Options) error {
	_, err := exec.LookPath("ebook-convert")
	if err != nil {
		return fmt.Errorf("ebook-convert not found: please install Calibre to convert ebook formats (https://calibre-ebook.com)")
	}

	args := []string{src, target}
	if len(opts.EbookArgs) > 0 {
		args = append(args, opts.EbookArgs...)
	}

	cmd := exec.CommandContext(ctx, "ebook-convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebook-convert failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToMarkdown(ctx context.Context, src, target string) error {
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

	var processed, failed int

	for _, item := range book.Spine.Itemrefs {
		if item.Item == nil {
			continue
		}

		f, err := item.Item.Open()
		if err != nil {
			failed++
			continue
		}

		b, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			failed++
			continue
		}

		markdown, err := converter.ConvertString(string(b))
		if err != nil {
			failed++
			continue
		}

		if strings.TrimSpace(markdown) != "" {
			contentBuilder.WriteString(markdown)
			contentBuilder.WriteString("\n\n---\n\n")
			processed++
		}
	}

	if processed == 0 {
		return fmt.Errorf("failed to extract any content from EPUB (%d items failed)", failed)
	}

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToHTML(ctx context.Context, src, target string) error {
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

	contentBuilder.WriteString("<!DOCTYPE html>\n<html>\n<head><meta charset=\"UTF-8\"></head>\n<body>\n")

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

		// Extract only body content to avoid nested <html>/<head>/<body> tags
		body := extractBodyContent(string(b))
		contentBuilder.WriteString(body)
		contentBuilder.WriteString("\n<hr>\n")
	}

	contentBuilder.WriteString("</body>\n</html>")

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// extractBodyContent extracts the inner content of a <body> tag from an HTML document.
// If no body tag is found, returns the original content.
func extractBodyContent(html string) string {
	lower := strings.ToLower(html)

	bodyStart := strings.Index(lower, "<body")
	if bodyStart == -1 {
		return html
	}

	// Find the closing > of the <body ...> tag
	bodyTagEnd := strings.Index(lower[bodyStart:], ">")
	if bodyTagEnd == -1 {
		return html
	}
	contentStart := bodyStart + bodyTagEnd + 1

	bodyEnd := strings.LastIndex(lower, "</body>")
	if bodyEnd == -1 || bodyEnd <= contentStart {
		return html[contentStart:]
	}

	return strings.TrimSpace(html[contentStart:bodyEnd])
}

func (c *DocumentConverter) convertEPUBToPDF(ctx context.Context, src, target string) error {
	// Use Calibre ebook-convert as primary (produces best quality)
	if _, err := exec.LookPath("ebook-convert"); err == nil {
		return c.convertEbookWithCalibre(ctx, src, target, Options{})
	}

	// Fallback: EPUB → HTML → PDF using fpdf (limited tag support)
	tempHTML := strings.TrimSuffix(target, filepath.Ext(target)) + "_temp.html"
	if err := c.convertEPUBToHTML(ctx, src, tempHTML); err != nil {
		return err
	}
	defer os.Remove(tempHTML)

	source, err := os.ReadFile(tempHTML)
	if err != nil {
		return fmt.Errorf("failed to read temp HTML file: %w", err)
	}

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
