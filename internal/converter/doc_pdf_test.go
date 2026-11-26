package converter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentConverter_PDF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_pdf_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	createTestPDF(t, pdfPath)

	t.Run("PDF->MD", func(t *testing.T) {
		target := filepath.Join(tmpDir, "pdf_out.md")
		if err := c.Convert(context.Background(), pdfPath, target, Options{}); err != nil {
			t.Fatalf("Convert(PDF->MD) failed: %v", err)
		}
		assertText(t, target, "Hello World")
	})

	t.Run("PDF->PDF (Compress)", func(t *testing.T) {
		target := filepath.Join(tmpDir, "pdf_compressed.pdf")
		if err := c.Convert(context.Background(), pdfPath, target, Options{}); err != nil {
			t.Fatalf("Convert(PDF->PDF) failed: %v", err)
		}
		md := filepath.Join(tmpDir, "compressed.md")
		if err := c.Convert(context.Background(), target, md, Options{}); err != nil {
			t.Fatal(err)
		}
		assertText(t, md, "Hello World")
	})
}
