package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentConverter_Office(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_office_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocumentConverter{}

	// CSV/Excel Tests
	csvPath := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(csvPath, []byte("Name,Age\nA,1\nB,2"), 0644); err != nil {
		t.Fatalf("failed to write csv file: %v", err)
	}

	t.Run("CSV->XLSX", func(t *testing.T) {
		target := filepath.Join(tmpDir, "test.xlsx")
		if err := c.Convert(csvPath, target, Options{}); err != nil {
			t.Errorf("Convert(CSV->XLSX) failed: %v", err)
		}
	})

	t.Run("XLSX->CSV", func(t *testing.T) {
		xlsxPath := filepath.Join(tmpDir, "test.xlsx")
		// Ensure previous step created the file
		if _, err := os.Stat(xlsxPath); err != nil {
			t.Skip("test.xlsx not found, skipping reverse conversion")
		}
		targetCsv := filepath.Join(tmpDir, "back_to.csv")
		if err := c.Convert(xlsxPath, targetCsv, Options{}); err != nil {
			t.Errorf("Convert(XLSX->CSV) failed: %v", err)
		}
	})
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
	t.Run("MD->DOCX", func(t *testing.T) {
		docxPath := filepath.Join(tmpDir, "out.docx")
		if err := c.Convert(mdPath, docxPath, Options{}); err != nil {
			t.Fatalf("Convert(MD->DOCX) failed: %v", err)
		}
		if _, err := os.Stat(docxPath); os.IsNotExist(err) {
			t.Fatalf("Target DOCX not created: %s", docxPath)
		}
	})

	// DOCX -> MD
	t.Run("DOCX->MD", func(t *testing.T) {
		docxPath := filepath.Join(tmpDir, "out.docx")
		// Ensure previous step created the file
		if _, err := os.Stat(docxPath); err != nil {
			t.Skip("out.docx not found, skipping reverse conversion")
		}
		backToMD := filepath.Join(tmpDir, "back.md")
		if err := c.Convert(docxPath, backToMD, Options{}); err != nil {
			t.Fatalf("Convert(DOCX->MD) failed: %v", err)
		}
		if _, err := os.Stat(backToMD); os.IsNotExist(err) {
			t.Fatalf("Target MD not created: %s", backToMD)
		}
	})
}
