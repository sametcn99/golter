package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentConverter_EPUB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_epub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}
	epubPath := filepath.Join(tmpDir, "test.epub")
	createTestEPUB(t, epubPath)

	t.Run("EPUB->MD", func(t *testing.T) {
		target := filepath.Join(tmpDir, "epub_out.md")
		if err := c.Convert(epubPath, target, Options{}); err != nil {
			t.Errorf("Convert(EPUB->MD) failed: %v", err)
		}
	})

	t.Run("EPUB->HTML", func(t *testing.T) {
		target := filepath.Join(tmpDir, "epub_out.html")
		if err := c.Convert(epubPath, target, Options{}); err != nil {
			t.Errorf("Convert(EPUB->HTML) failed: %v", err)
		}
	})

	t.Run("EPUB->PDF", func(t *testing.T) {
		target := filepath.Join(tmpDir, "epub_out.pdf")
		if err := c.Convert(epubPath, target, Options{}); err != nil {
			t.Errorf("Convert(EPUB->PDF) failed: %v", err)
		}
	})
}

func TestDocumentConverter_Convert_Integration_EbookCalibre(t *testing.T) {
	if !ensureEbookConvertInPath(t) {
		t.Skip("ebook-convert not found, skipping ebook conversion tests")
	}

	tmpDir, err := os.MkdirTemp("", "golter_ebook_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}

	// Create a source EPUB
	epubPath := filepath.Join(tmpDir, "test.epub")
	createTestEPUB(t, epubPath)

	t.Run("EPUB->MOBI", func(t *testing.T) {
		mobiPath := filepath.Join(tmpDir, "test.mobi")
		if err := c.Convert(epubPath, mobiPath, Options{}); err != nil {
			t.Fatalf("Convert(EPUB->MOBI) failed: %v", err)
		}
		if _, err := os.Stat(mobiPath); os.IsNotExist(err) {
			t.Fatalf("Target MOBI not created: %s", mobiPath)
		}
	})

	t.Run("MOBI->EPUB", func(t *testing.T) {
		mobiPath := filepath.Join(tmpDir, "test.mobi")
		backToEPUB := filepath.Join(tmpDir, "back.epub")
		if err := c.Convert(mobiPath, backToEPUB, Options{}); err != nil {
			t.Fatalf("Convert(MOBI->EPUB) failed: %v", err)
		}
		if _, err := os.Stat(backToEPUB); os.IsNotExist(err) {
			t.Fatalf("Target EPUB not created: %s", backToEPUB)
		}
	})

	t.Run("EPUB->AZW3", func(t *testing.T) {
		azw3Path := filepath.Join(tmpDir, "test.azw3")
		if err := c.Convert(epubPath, azw3Path, Options{}); err != nil {
			t.Fatalf("Convert(EPUB->AZW3) failed: %v", err)
		}
		if _, err := os.Stat(azw3Path); os.IsNotExist(err) {
			t.Fatalf("Target AZW3 not created: %s", azw3Path)
		}
	})
}

func TestDocumentConverter_Convert_EbookFromMarkdownAndHTML(t *testing.T) {
	if !ensureEbookConvertInPath(t) {
		t.Skip("ebook-convert not found, skipping ebook conversion tests")
	}

	tmpDir, err := os.MkdirTemp("", "golter_ebook_src_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}

	t.Run("MD->MOBI", func(t *testing.T) {
		mdPath := filepath.Join(tmpDir, "src.md")
		if err := os.WriteFile(mdPath, []byte("# Title\n\nHello ebook"), 0644); err != nil {
			t.Fatalf("failed to write md: %v", err)
		}
		mobiPath := filepath.Join(tmpDir, "md_out.mobi")
		if err := c.Convert(mdPath, mobiPath, Options{}); err != nil {
			t.Fatalf("Convert(MD->MOBI) failed: %v", err)
		}
		if _, err := os.Stat(mobiPath); os.IsNotExist(err) {
			t.Fatalf("Target MOBI not created: %s", mobiPath)
		}
	})

	t.Run("HTML->AZW3", func(t *testing.T) {
		htmlPath := filepath.Join(tmpDir, "src.html")
		if err := os.WriteFile(htmlPath, []byte("<html><body><h1>Title</h1><p>Hello ebook</p></body></html>"), 0644); err != nil {
			t.Fatalf("failed to write html: %v", err)
		}
		azw3Path := filepath.Join(tmpDir, "html_out.azw3")
		if err := c.Convert(htmlPath, azw3Path, Options{}); err != nil {
			t.Fatalf("Convert(HTML->AZW3) failed: %v", err)
		}
		if _, err := os.Stat(azw3Path); os.IsNotExist(err) {
			t.Fatalf("Target AZW3 not created: %s", azw3Path)
		}
	})
}
