package converter

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
				_ = f.SetCellStyle(sheetName, cellRef, cellRef, headerStyle)
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
			_ = f.SetColWidth(sheetName, colName, colName, width)
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
