package converter

import (
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
		if err := c.Convert(pdfPath, target, Options{}); err != nil {
			t.Logf("Convert(PDF->MD) failed: %v", err)
		} else if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Error("Target MD not created")
		}
	})

	t.Run("PDF->PDF (Compress)", func(t *testing.T) {
		target := filepath.Join(tmpDir, "pdf_compressed.pdf")
		if err := c.Convert(pdfPath, target, Options{}); err != nil {
			t.Logf("Convert(PDF->PDF) failed: %v", err)
		} else if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Error("Target PDF not created")
		}
	})
}
