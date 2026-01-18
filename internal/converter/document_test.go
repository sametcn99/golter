package converter

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/go-pdf/fpdf"
)

func TestDocumentConverter_SupportedExtensions(t *testing.T) {
	c := &DocumentConverter{}

	// Test SupportedSourceExtensions
	srcExts := c.SupportedSourceExtensions()
	if len(srcExts) == 0 {
		t.Error("SupportedSourceExtensions returned empty list")
	}
	expectedSrc := []string{".pdf", ".md", ".html", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".csv", ".xlsx", ".xls"}
	for _, exp := range expectedSrc {
		found := false
		for _, got := range srcExts {
			if got == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedSourceExtensions missing %s", exp)
		}
	}

	// Test SupportedTargetFormats
	for _, ext := range expectedSrc {
		targets := c.SupportedTargetFormats(ext)
		if len(targets) == 0 {
			t.Errorf("SupportedTargetFormats(%s) returned empty list", ext)
		}
	}

	// Test unsupported target formats
	if c.SupportedTargetFormats(".txt") != nil {
		t.Error("SupportedTargetFormats(.txt) should return nil")
	}
}

func TestDocumentConverter_Convert_Error(t *testing.T) {
	c := &DocumentConverter{}

	// Test non-existent file
	err := c.Convert("non_existent.md", "out.html", Options{})
	if err == nil {
		t.Error("Convert should fail for non-existent file")
	}

	// Test unsupported conversion
	_ = c.Convert("test.md", "out.jpg", Options{}) // File existence check is first, so create dummy
	if err := os.WriteFile("test.md", []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.Remove("test.md")

	err = c.Convert("test.md", "out.jpg", Options{})
	if err == nil {
		t.Error("Convert should fail for unsupported conversion")
	} else if !strings.Contains(err.Error(), "unsupported conversion") {
		t.Errorf("Expected unsupported conversion error, got: %v", err)
	}
}

func TestDocumentConverter_Name(t *testing.T) {
	c := &DocumentConverter{}
	if !strings.Contains(c.Name(), "Document Converter") {
		t.Errorf("Name() = %v, want it to contain 'Document Converter'", c.Name())
	}
}

func TestDocumentConverter_CanConvert(t *testing.T) {
	c := &DocumentConverter{}
	tests := []struct {
		src    string
		target string
		want   bool
	}{
		// PDF
		{".pdf", ".md", true},
		{".pdf", ".pdf", true},
		{".pdf", ".html", false},

		// Markdown
		{".md", ".pdf", true},
		{".md", ".html", true},
		{".md", ".docx", true},
		{".md", ".epub", true},
		{".md", ".txt", false},

		// HTML
		{".html", ".md", true},
		{".html", ".docx", true},
		{".html", ".epub", true},
		{".html", ".pdf", false},

		// DOCX
		{".docx", ".md", true},
		{".docx", ".html", true},
		{".docx", ".txt", true},
		{".docx", ".pdf", false},

		// CSV/Excel
		{".csv", ".xlsx", true},
		{".csv", ".xls", true},
		{".xlsx", ".csv", true},
		{".xls", ".csv", true},
		{".csv", ".pdf", false},

		// EPUB
		{".epub", ".pdf", true},
		{".epub", ".md", true},
		{".epub", ".html", true},
		{".epub", ".mobi", true},
		{".epub", ".azw3", true},
		{".epub", ".csv", false},

		// Ebook formats
		{".mobi", ".epub", true},
		{".mobi", ".pdf", true},
		{".azw", ".azw3", true},
		{".azw3", ".html", true},
		{".fb2", ".md", true},
		{".fb2", ".csv", false},
	}

	for _, tt := range tests {
		if got := c.CanConvert(tt.src, tt.target); got != tt.want {
			t.Errorf("CanConvert(%q, %q) = %v, want %v", tt.src, tt.target, got, tt.want)
		}
	}
}

func TestDocumentConverter_SupportedTargetFormats(t *testing.T) {
	c := &DocumentConverter{}

	tests := []struct {
		src  string
		want []string
	}{
		{".pdf", []string{".md", ".pdf"}},
		{".md", []string{".html", ".pdf", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2"}},
		{".csv", []string{".xlsx"}},
		{".docx", []string{".md", ".html", ".txt"}},
		{".epub", []string{".pdf", ".md", ".html", ".mobi", ".azw", ".azw3", ".fb2", ".txt"}},
		{".mobi", []string{".epub", ".azw", ".azw3", ".fb2", ".pdf", ".html", ".txt", ".md"}},
	}

	for _, tt := range tests {
		got := c.SupportedTargetFormats(tt.src)
		if len(got) != len(tt.want) {
			t.Errorf("SupportedTargetFormats(%q) length = %v, want %v", tt.src, len(got), len(tt.want))
			continue
		}

		// Simple check for existence
		for _, w := range tt.want {
			found := false
			for _, g := range got {
				if g == w {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SupportedTargetFormats(%q) missing %v", tt.src, w)
			}
		}
	}
}

// Global test helpers

func createTestPDF(t *testing.T, path string) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello World")
	if err := pdf.OutputFileAndClose(path); err != nil {
		t.Fatalf("failed to create test pdf: %v", err)
	}
}

func createTestEPUB(t *testing.T, path string) {
	e := epub.NewEpub("Test Title")
	e.SetAuthor("Test Author")
	_, err := e.AddSection("<h1>Hello World</h1>", "Section 1", "", "")
	if err != nil {
		t.Fatalf("failed to add section to epub: %v", err)
	}
	if err := e.Write(path); err != nil {
		t.Fatalf("failed to create test epub: %v", err)
	}
}

func ensureEbookConvertInPath(t *testing.T) bool {
	binName := "ebook-convert"
	if runtime.GOOS == "windows" {
		binName = "ebook-convert.exe"
	}

	if _, err := exec.LookPath(binName); err == nil {
		return true
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\\Program Files\\Calibre2\\ebook-convert.exe`,
			`C:\\Program Files (x86)\\Calibre2\\ebook-convert.exe`,
		}
	} else {
		candidates = []string{
			"/usr/bin/ebook-convert",
			"/usr/local/bin/ebook-convert",
			"/snap/bin/ebook-convert",
			"/opt/homebrew/bin/ebook-convert",
		}
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			dir := filepath.Dir(p)
			current := os.Getenv("PATH")
			if !strings.Contains(current, dir) {
				if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+current); err != nil {
					t.Logf("failed to update PATH for ebook-convert: %v", err)
					return false
				}
			}
			return true
		}
	}

	return false
}

func ensurePandocInPath(t *testing.T) bool {
	binName := "pandoc"
	if runtime.GOOS == "windows" {
		binName = "pandoc.exe"
	}

	if _, err := exec.LookPath(binName); err == nil {
		return true
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\\Program Files\\Pandoc\\pandoc.exe`,
			`C:\\Program Files (x86)\\Pandoc\\pandoc.exe`,
		}
	} else {
		candidates = []string{
			"/usr/bin/pandoc",
			"/usr/local/bin/pandoc",
			"/opt/homebrew/bin/pandoc",
		}
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			dir := filepath.Dir(p)
			current := os.Getenv("PATH")
			if !strings.Contains(current, dir) {
				if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+current); err != nil {
					t.Logf("failed to update PATH for pandoc: %v", err)
					return false
				}
			}
			return true
		}
	}

	return false
}
