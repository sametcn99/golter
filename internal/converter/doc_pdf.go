package converter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func (c *DocumentConverter) convertPDFToMarkdown(src, target string) error {
	f, r, err := pdf.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	reader, err := r.GetPlainText()
	if err != nil {
		return fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	_, err = buf.ReadFrom(reader)
	if err != nil {
		return fmt.Errorf("failed to read PDF content: %w", err)
	}

	// Add markdown formatting
	content := buf.String()
	if content == "" {
		return fmt.Errorf("no text content found in PDF (might be image-based)")
	}

	// Basic markdown formatting
	content = "# " + filepath.Base(src) + "\n\n" + content

	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) compressPDF(src, target string) error {
	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.OPTIMIZE

	if err := api.OptimizeFile(src, target, conf); err != nil {
		return fmt.Errorf("failed to compress PDF: %w", err)
	}

	return nil
}
