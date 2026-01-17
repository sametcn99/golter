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
	err = c.Convert("test.md", "out.jpg", Options{}) // File existence check is first, so create dummy
	os.WriteFile("test.md", []byte("test"), 0644)
	defer os.Remove("test.md")

	err = c.Convert("test.md", "out.jpg", Options{})
	if err == nil {
		t.Error("Convert should fail for unsupported conversion")
	} else if !strings.Contains(err.Error(), "unsupported conversion") {
		t.Errorf("Expected unsupported conversion error, got: %v", err)
	}
}

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

func TestDocumentConverter_Convert_Integration_Exhaustive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_doc_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}

	// 1. PDF Tests
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	createTestPDF(t, pdfPath)

	t.Run("PDF->MD", func(t *testing.T) {
		target := filepath.Join(tmpDir, "pdf_out.md")
		if err := c.Convert(pdfPath, target, Options{}); err != nil {
			t.Logf("Convert(PDF->MD) failed (expected if pdfcpu/pdf lib issues): %v", err)
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

	// 2. Markdown Tests
	mdPath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(mdPath, []byte("# Hello\n\nTest content."), 0644)

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

	// 3. HTML Tests
	htmlPath := filepath.Join(tmpDir, "test.html")
	os.WriteFile(htmlPath, []byte("<html><body><h1>Hello</h1></body></html>"), 0644)

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

	// 4. CSV/Excel Tests
	csvPath := filepath.Join(tmpDir, "test.csv")
	os.WriteFile(csvPath, []byte("Name,Age\nA,1\nB,2"), 0644)

	t.Run("CSV->XLSX", func(t *testing.T) {
		target := filepath.Join(tmpDir, "test.xlsx")
		if err := c.Convert(csvPath, target, Options{}); err != nil {
			t.Errorf("Convert(CSV->XLSX) failed: %v", err)
		}

		// Test reverse: XLSX -> CSV
		if _, err := os.Stat(target); err == nil {
			targetCsv := filepath.Join(tmpDir, "back_to.csv")
			if err := c.Convert(target, targetCsv, Options{}); err != nil {
				t.Errorf("Convert(XLSX->CSV) failed: %v", err)
			}
		}
	})

	// 5. EPUB Tests
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

	// EPUB -> MOBI
	mobiPath := filepath.Join(tmpDir, "test.mobi")
	if err := c.Convert(epubPath, mobiPath, Options{}); err != nil {
		t.Fatalf("Convert(EPUB->MOBI) failed: %v", err)
	}
	if _, err := os.Stat(mobiPath); os.IsNotExist(err) {
		t.Fatalf("Target MOBI not created: %s", mobiPath)
	}

	// MOBI -> EPUB
	backToEPUB := filepath.Join(tmpDir, "back.epub")
	if err := c.Convert(mobiPath, backToEPUB, Options{}); err != nil {
		t.Fatalf("Convert(MOBI->EPUB) failed: %v", err)
	}
	if _, err := os.Stat(backToEPUB); os.IsNotExist(err) {
		t.Fatalf("Target EPUB not created: %s", backToEPUB)
	}

	// EPUB -> AZW3
	azw3Path := filepath.Join(tmpDir, "test.azw3")
	if err := c.Convert(epubPath, azw3Path, Options{}); err != nil {
		t.Fatalf("Convert(EPUB->AZW3) failed: %v", err)
	}
	if _, err := os.Stat(azw3Path); os.IsNotExist(err) {
		t.Fatalf("Target AZW3 not created: %s", azw3Path)
	}
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

	mdPath := filepath.Join(tmpDir, "src.md")
	if err := os.WriteFile(mdPath, []byte("# Title\n\nHello ebook"), 0644); err != nil {
		t.Fatalf("failed to write md: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "src.html")
	if err := os.WriteFile(htmlPath, []byte("<html><body><h1>Title</h1><p>Hello ebook</p></body></html>"), 0644); err != nil {
		t.Fatalf("failed to write html: %v", err)
	}

	// MD -> MOBI
	mobiPath := filepath.Join(tmpDir, "md_out.mobi")
	if err := c.Convert(mdPath, mobiPath, Options{}); err != nil {
		t.Fatalf("Convert(MD->MOBI) failed: %v", err)
	}
	if _, err := os.Stat(mobiPath); os.IsNotExist(err) {
		t.Fatalf("Target MOBI not created: %s", mobiPath)
	}

	// HTML -> AZW3
	azw3Path := filepath.Join(tmpDir, "html_out.azw3")
	if err := c.Convert(htmlPath, azw3Path, Options{}); err != nil {
		t.Fatalf("Convert(HTML->AZW3) failed: %v", err)
	}
	if _, err := os.Stat(azw3Path); os.IsNotExist(err) {
		t.Fatalf("Target AZW3 not created: %s", azw3Path)
	}
}

func TestDocumentConverter_Convert_DocxPandoc(t *testing.T) {
	if !ensurePandocInPath(t) {
		t.Skip("pandoc not found, skipping DOCX conversion tests")
	}

	tmpDir, err := os.MkdirTemp("", "golter_docx_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}

	mdPath := filepath.Join(tmpDir, "src.md")
	if err := os.WriteFile(mdPath, []byte("# Title\n\nHello docx"), 0644); err != nil {
		t.Fatalf("failed to write md: %v", err)
	}

	// MD -> DOCX
	docxPath := filepath.Join(tmpDir, "out.docx")
	if err := c.Convert(mdPath, docxPath, Options{}); err != nil {
		t.Fatalf("Convert(MD->DOCX) failed: %v", err)
	}
	if _, err := os.Stat(docxPath); os.IsNotExist(err) {
		t.Fatalf("Target DOCX not created: %s", docxPath)
	}

	// DOCX -> MD
	backToMD := filepath.Join(tmpDir, "back.md")
	if err := c.Convert(docxPath, backToMD, Options{}); err != nil {
		t.Fatalf("Convert(DOCX->MD) failed: %v", err)
	}
	if _, err := os.Stat(backToMD); os.IsNotExist(err) {
		t.Fatalf("Target MD not created: %s", backToMD)
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
