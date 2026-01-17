package converter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	goepub "github.com/bmaupin/go-epub"
	"github.com/go-pdf/fpdf"
	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/xuri/excelize/v2"
	"github.com/yuin/goldmark"
)

// DocumentConverter handles document format conversions
type DocumentConverter struct{}

var ebookExtensions = map[string]struct{}{
	".epub": {},
	".mobi": {},
	".azw":  {},
	".azw3": {},
	".fb2":  {},
}

func isEbookExt(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	_, ok := ebookExtensions[ext]
	return ok
}

func tempPathWithExt(prefix, ext string) (string, func(), error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	return filepath.Join(dir, "temp"+ext), cleanup, nil
}

func (c *DocumentConverter) Name() string {
	return "Document Converter"
}

func (c *DocumentConverter) CanConvert(srcExt, targetExt string) bool {
	srcExt = strings.ToLower(srcExt)
	targetExt = strings.ToLower(targetExt)

	if !strings.HasPrefix(srcExt, ".") {
		srcExt = "." + srcExt
	}
	if !strings.HasPrefix(targetExt, ".") {
		targetExt = "." + targetExt
	}

	switch srcExt {
	case ".pdf":
		return targetExt == ".md" || targetExt == ".pdf"
	case ".md":
		return targetExt == ".pdf" || targetExt == ".html" || targetExt == ".docx" || isEbookExt(targetExt)
	case ".html":
		return targetExt == ".md" || targetExt == ".docx" || isEbookExt(targetExt)
	case ".docx":
		return targetExt == ".md" || targetExt == ".html" || targetExt == ".txt"
	case ".csv":
		return targetExt == ".xlsx" || targetExt == ".xls"
	case ".xlsx", ".xls":
		return targetExt == ".csv"
	case ".epub", ".mobi", ".azw", ".azw3", ".fb2":
		if isEbookExt(targetExt) {
			return srcExt != targetExt
		}
		return targetExt == ".pdf" || targetExt == ".md" || targetExt == ".html" || targetExt == ".txt"
	}
	return false
}

func (c *DocumentConverter) SupportedSourceExtensions() []string {
	return []string{".pdf", ".md", ".html", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".csv", ".xlsx", ".xls"}
}

func (c *DocumentConverter) SupportedTargetFormats(srcExt string) []string {
	srcExt = strings.ToLower(srcExt)
	if !strings.HasPrefix(srcExt, ".") {
		srcExt = "." + srcExt
	}

	switch srcExt {
	case ".pdf":
		return []string{".md", ".pdf"} // .pdf -> .pdf implies compression
	case ".md":
		return []string{".html", ".pdf", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2"}
	case ".html":
		return []string{".md", ".docx", ".epub", ".mobi", ".azw", ".azw3", ".fb2"}
	case ".docx":
		return []string{".md", ".html", ".txt"}
	case ".csv":
		return []string{".xlsx"}
	case ".xlsx", ".xls":
		return []string{".csv"}
	case ".epub":
		return []string{".pdf", ".md", ".html", ".mobi", ".azw", ".azw3", ".fb2", ".txt"}
	case ".mobi", ".azw", ".azw3", ".fb2":
		allTargets := []string{".epub", ".mobi", ".azw", ".azw3", ".fb2", ".pdf", ".html", ".txt", ".md"}
		filtered := make([]string, 0, len(allTargets))
		for _, t := range allTargets {
			if t == srcExt {
				continue
			}
			filtered = append(filtered, t)
		}
		return filtered
	}
	return nil
}

func (c *DocumentConverter) Convert(src, target string, opts Options) error {
	srcExt := strings.ToLower(filepath.Ext(src))
	targetExt := strings.ToLower(filepath.Ext(target))

	// Validate source file exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", src)
	}

	switch srcExt {
	case ".pdf":
		if targetExt == ".md" {
			return c.convertPDFToMarkdown(src, target)
		} else if targetExt == ".pdf" {
			return c.compressPDF(src, target)
		}
	case ".md":
		if targetExt == ".html" {
			return c.convertMarkdownToHTML(src, target)
		} else if targetExt == ".pdf" {
			return c.convertMarkdownToPDF(src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertMarkdownToEPUB(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertMarkdownToEbook(src, target, opts)
		}
	case ".html":
		if targetExt == ".md" {
			return c.convertHTMLToMarkdown(src, target)
		} else if targetExt == ".docx" {
			return c.convertWithPandoc(src, target, opts)
		} else if targetExt == ".epub" {
			return c.convertHTMLToEPUB(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertHTMLToEbook(src, target, opts)
		}
	case ".docx":
		if targetExt == ".md" || targetExt == ".html" || targetExt == ".txt" {
			return c.convertWithPandoc(src, target, opts)
		}
	case ".csv":
		if targetExt == ".xlsx" || targetExt == ".xls" {
			return c.convertCSVToExcel(src, target)
		}
	case ".xlsx", ".xls":
		if targetExt == ".csv" {
			return c.convertExcelToCSV(src, target)
		}
	case ".epub":
		if targetExt == srcExt {
			break
		}
		if targetExt == ".md" {
			return c.convertEPUBToMarkdown(src, target)
		} else if targetExt == ".html" {
			return c.convertEPUBToHTML(src, target)
		} else if targetExt == ".pdf" {
			return c.convertEPUBToPDF(src, target)
		} else if isEbookExt(targetExt) {
			return c.convertEbookWithCalibre(src, target, opts)
		} else if targetExt == ".txt" {
			return c.convertEbookWithCalibre(src, target, opts)
		}
	case ".mobi", ".azw", ".azw3", ".fb2":
		if targetExt == srcExt {
			break
		}
		if targetExt == ".md" {
			return c.convertEbookToMarkdown(src, target, opts)
		} else if targetExt == ".html" || targetExt == ".pdf" || targetExt == ".txt" || isEbookExt(targetExt) {
			return c.convertEbookWithCalibre(src, target, opts)
		}
	}

	return fmt.Errorf("unsupported conversion: %s to %s", srcExt, targetExt)
}

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

func (c *DocumentConverter) convertMarkdownToHTML(src, target string) error {
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(source, &buf); err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Wrap in basic HTML structure
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        pre {
            background: #f4f4f4;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
        code {
            background: #f4f4f4;
            padding: 2px 5px;
            border-radius: 3px;
        }
        img { max-width: 100%%; }
    </style>
</head>
<body>
%s
</body>
</html>`, filepath.Base(src), buf.String())

	if err := os.WriteFile(target, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToPDF(src, target string) error {
	// Read markdown source
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Convert to HTML first
	var htmlBuf bytes.Buffer
	if err := goldmark.Convert(source, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Create PDF
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.SetMargins(20, 20, 20)
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "", 12)

	_, lineHt := pdfDoc.GetFontSize()
	html := pdfDoc.HTMLBasicNew()
	html.Write(lineHt, htmlBuf.String())

	if err := pdfDoc.OutputFileAndClose(target); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertHTMLToMarkdown(src, target string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open HTML file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(b))
	if err != nil {
		return fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	if err := os.WriteFile(target, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToEPUB(src, target string) error {
	// Read and convert markdown to HTML
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := goldmark.Convert(source, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Create EPUB
	title := strings.TrimSuffix(filepath.Base(src), ".md")
	e := goepub.NewEpub(title)
	e.SetAuthor("Golter Converter")

	_, err = e.AddSection(htmlBuf.String(), "Chapter 1", "", "")
	if err != nil {
		return fmt.Errorf("failed to add content to EPUB: %w", err)
	}

	if err := e.Write(target); err != nil {
		return fmt.Errorf("failed to write EPUB file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertHTMLToEPUB(src, target string) error {
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	title := strings.TrimSuffix(filepath.Base(src), ".html")
	e := goepub.NewEpub(title)
	e.SetAuthor("Golter Converter")

	_, err = e.AddSection(string(source), "Chapter 1", "", "")
	if err != nil {
		return fmt.Errorf("failed to add content to EPUB: %w", err)
	}

	if err := e.Write(target); err != nil {
		return fmt.Errorf("failed to write EPUB file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertMarkdownToEbook(src, target string, opts Options) error {
	if strings.EqualFold(filepath.Ext(target), ".epub") {
		return c.convertMarkdownToEPUB(src, target)
	}

	tempEPUB, cleanup, err := tempPathWithExt("golter_ebook_epub", ".epub")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertMarkdownToEPUB(src, tempEPUB); err != nil {
		return err
	}

	return c.convertEbookWithCalibre(tempEPUB, target, opts)
}

func (c *DocumentConverter) convertHTMLToEbook(src, target string, opts Options) error {
	if strings.EqualFold(filepath.Ext(target), ".epub") {
		return c.convertHTMLToEPUB(src, target)
	}

	tempEPUB, cleanup, err := tempPathWithExt("golter_ebook_epub", ".epub")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertHTMLToEPUB(src, tempEPUB); err != nil {
		return err
	}

	return c.convertEbookWithCalibre(tempEPUB, target, opts)
}

func (c *DocumentConverter) convertEbookToMarkdown(src, target string, opts Options) error {
	tempHTML, cleanup, err := tempPathWithExt("golter_ebook_html", ".html")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := c.convertEbookWithCalibre(src, tempHTML, opts); err != nil {
		return err
	}

	return c.convertHTMLToMarkdown(tempHTML, target)
}

func (c *DocumentConverter) convertEbookWithCalibre(src, target string, opts Options) error {
	_, err := exec.LookPath("ebook-convert")
	if err != nil {
		return fmt.Errorf("ebook-convert not found: please install Calibre to convert ebook formats (https://calibre-ebook.com)")
	}

	args := []string{src, target}
	if extra, ok := opts["ebookArgs"].([]string); ok && len(extra) > 0 {
		args = append(args, extra...)
	} else if extraStr, ok := opts["ebookArgs"].(string); ok && strings.TrimSpace(extraStr) != "" {
		args = append(args, strings.Fields(extraStr)...)
	}

	cmd := exec.Command("ebook-convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebook-convert failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *DocumentConverter) convertWithPandoc(src, target string, opts Options) error {
	_, err := exec.LookPath("pandoc")
	if err != nil {
		return fmt.Errorf("pandoc not found: please install Pandoc to convert DOCX formats (https://pandoc.org)")
	}

	args := []string{src, "-o", target}
	if extra, ok := opts["pandocArgs"].([]string); ok && len(extra) > 0 {
		args = append(args, extra...)
	} else if extraStr, ok := opts["pandocArgs"].(string); ok && strings.TrimSpace(extraStr) != "" {
		args = append(args, strings.Fields(extraStr)...)
	}

	cmd := exec.Command("pandoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// convertCSVToExcel converts a CSV file to Excel format
func (c *DocumentConverter) convertCSVToExcel(src, target string) error {
	// Open CSV file
	csvFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer csvFile.Close()

	// Read CSV content
	reader := csv.NewReader(csvFile)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close excel file: %v\n", err)
		}
	}()

	// Get the default sheet name
	sheetName := "Sheet1"

	// Write data to Excel
	for rowIdx, row := range records {
		for colIdx, cell := range row {
			cellRef, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err != nil {
				return fmt.Errorf("failed to create cell reference: %w", err)
			}
			if err := f.SetCellValue(sheetName, cellRef, cell); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	// Style the header row (first row) if there's data
	if len(records) > 0 && len(records[0]) > 0 {
		// Create a bold style for header
		headerStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#E0E0E0"},
				Pattern: 1,
			},
		})
		if err == nil {
			// Apply style to header row
			for colIdx := range records[0] {
				cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
				f.SetCellStyle(sheetName, cellRef, cellRef, headerStyle)
			}
		}

		// Auto-fit column widths (approximate)
		for colIdx, cell := range records[0] {
			colName, _ := excelize.ColumnNumberToName(colIdx + 1)
			width := float64(len(cell)) * 1.2
			if width < 10 {
				width = 10
			}
			if width > 50 {
				width = 50
			}
			f.SetColWidth(sheetName, colName, colName, width)
		}
	}

	// Save the Excel file
	if err := f.SaveAs(target); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}

// convertExcelToCSV converts an Excel file to CSV format
func (c *DocumentConverter) convertExcelToCSV(src, target string) error {
	// Open Excel file
	f, err := excelize.OpenFile(src)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close excel file: %v\n", err)
		}
	}()

	// Get the first sheet name
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("Excel file has no sheets")
	}
	sheetName := sheets[0]

	// Get all rows from the first sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read Excel sheet: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("Excel sheet is empty")
	}

	// Create CSV file
	csvFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer csvFile.Close()

	// Write CSV content
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToMarkdown(src, target string) error {
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

	// Iterate through spine items
	for _, item := range book.Spine.Itemrefs {
		if item.Item == nil {
			continue
		}

		// Open the file from the EPUB
		f, err := item.Item.Open()
		if err != nil {
			continue
		}

		b, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}

		// Convert HTML content to Markdown
		markdown, err := converter.ConvertString(string(b))
		if err != nil {
			continue
		}

		contentBuilder.WriteString(markdown)
		contentBuilder.WriteString("\n\n---\n\n")
	}

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToHTML(src, target string) error {
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

	contentBuilder.WriteString("<!DOCTYPE html><html><body>")

	// Iterate through spine items
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

		// Simple concatenation of body content would be better, but full HTML concatenation is easier for now
		// Ideally we should strip <html>, <head>, <body> tags and just take the inner content
		// For simplicity, we just append the whole thing, browsers handle nested html tags somewhat okay-ish
		// or better: just append the raw content.
		contentBuilder.Write(b)
		contentBuilder.WriteString("<hr>")
	}

	contentBuilder.WriteString("</body></html>")

	if err := os.WriteFile(target, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func (c *DocumentConverter) convertEPUBToPDF(src, target string) error {
	// First convert to HTML
	tempHTML := strings.TrimSuffix(target, filepath.Ext(target)) + "_temp.html"
	if err := c.convertEPUBToHTML(src, tempHTML); err != nil {
		return err
	}
	defer os.Remove(tempHTML)

	// Then convert HTML to PDF (using existing logic logic, but we need to read the temp file)
	// We can reuse convertMarkdownToPDF logic but starting from HTML

	// Read HTML source
	source, err := os.ReadFile(tempHTML)
	if err != nil {
		return fmt.Errorf("failed to read temp HTML file: %w", err)
	}

	// Create PDF
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
