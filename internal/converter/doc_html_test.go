package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentConverter_HTML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_html_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}
	htmlPath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(htmlPath, []byte("<html><body><h1>Hello</h1></body></html>"), 0644); err != nil {
		t.Fatalf("failed to write html file: %v", err)
	}

	t.Run("HTML->MD", func(t *testing.T) {
		target := filepath.Join(tmpDir, "html_out.md")
		if err := c.Convert(htmlPath, target, Options{}); err != nil {
			t.Errorf("Convert(HTML->MD) failed: %v", err)
		}
	})
	t.Run("HTML->EPUB", func(t *testing.T) {
		target := filepath.Join(tmpDir, "html_out.epub")
		if err := c.Convert(htmlPath, target, Options{}); err != nil {
			t.Errorf("Convert(HTML->EPUB) failed: %v", err)
		}
	})
}
