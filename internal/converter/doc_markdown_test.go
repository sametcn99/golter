package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentConverter_Markdown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_md_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}
	mdPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdPath, []byte("# Hello\n\nTest content."), 0644); err != nil {
		t.Fatalf("failed to write md file: %v", err)
	}

	t.Run("MD->HTML", func(t *testing.T) {
		target := filepath.Join(tmpDir, "md_out.html")
		if err := c.Convert(mdPath, target, Options{}); err != nil {
			t.Errorf("Convert(MD->HTML) failed: %v", err)
		}
	})
	t.Run("MD->PDF", func(t *testing.T) {
		target := filepath.Join(tmpDir, "md_out.pdf")
		if err := c.Convert(mdPath, target, Options{}); err != nil {
			t.Errorf("Convert(MD->PDF) failed: %v", err)
		}
	})
	t.Run("MD->EPUB", func(t *testing.T) {
		target := filepath.Join(tmpDir, "md_out.epub")
		if err := c.Convert(mdPath, target, Options{}); err != nil {
			t.Errorf("Convert(MD->EPUB) failed: %v", err)
		}
	})
}
