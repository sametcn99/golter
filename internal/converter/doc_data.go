package converter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/clbanning/mxj/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

// DocDataConverter handles structured data format conversions.
type DocDataConverter struct{}

func (c *DocDataConverter) Name() string {
	return "Doc Data Converter"
}

func (c *DocDataConverter) isSupported(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".json", ".yaml", ".yml", ".xml", ".toml", ".csv", ".xlsx", ".xls":
		return true
	default:
		return false
	}
}

func (c *DocDataConverter) CanConvert(srcExt, targetExt string) bool {
	srcExt = normalizeExt(srcExt)
	targetExt = normalizeExt(targetExt)

	if !c.isSupported(srcExt) || !c.isSupported(targetExt) {
		return false
	}

	switch srcExt {
	case ".json":
		return targetExt == ".yaml" || targetExt == ".yml" || targetExt == ".xml" || targetExt == ".csv" || targetExt == ".xlsx" || targetExt == ".xls"
	case ".yaml", ".yml":
		return targetExt == ".json" || targetExt == ".toml"
	case ".toml":
		return targetExt == ".yaml" || targetExt == ".yml"
	case ".xml":
		return targetExt == ".json"
	case ".csv":
		return targetExt == ".json"
	case ".xlsx", ".xls":
		return targetExt == ".json"
	}

	return false
}

func (c *DocDataConverter) SupportedSourceExtensions() []string {
	return []string{".json", ".yaml", ".yml", ".xml", ".toml", ".csv", ".xlsx", ".xls"}
}

func (c *DocDataConverter) SupportedTargetFormats(srcExt string) []string {
	srcExt = normalizeExt(srcExt)
	switch srcExt {
	case ".json":
		return []string{".yaml", ".yml", ".xml", ".csv", ".xlsx", ".xls"}
	case ".yaml", ".yml":
		return []string{".json", ".toml"}
	case ".toml":
		return []string{".yaml", ".yml"}
	case ".xml":
		return []string{".json"}
	case ".csv":
		return []string{".json"}
	case ".xlsx", ".xls":
		return []string{".json"}
	}
	return nil
}

func (c *DocDataConverter) Convert(src, target string, opts Options) error {
	_ = opts

	srcExt := normalizeExt(filepath.Ext(src))
	targetExt := normalizeExt(filepath.Ext(target))

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", src)
	}

	switch srcExt {
	case ".json":
		switch targetExt {
		case ".yaml", ".yml":
			return c.convertJSONToYAML(src, target)
		case ".xml":
			return c.convertJSONToXML(src, target)
		case ".csv":
			return c.convertJSONToCSV(src, target)
		case ".xlsx", ".xls":
			return c.convertJSONToExcel(src, target)
		}
	case ".yaml", ".yml":
		switch targetExt {
		case ".json":
			return c.convertYAMLToJSON(src, target)
		case ".toml":
			return c.convertYAMLToTOML(src, target)
		}
	case ".toml":
		if targetExt == ".yaml" || targetExt == ".yml" {
			return c.convertTOMLToYAML(src, target)
		}
	case ".xml":
		if targetExt == ".json" {
			return c.convertXMLToJSON(src, target)
		}
	case ".csv":
		if targetExt == ".json" {
			return c.convertCSVToJSON(src, target)
		}
	case ".xlsx", ".xls":
		if targetExt == ".json" {
			return c.convertExcelToJSON(src, target)
		}
	}

	return fmt.Errorf("unsupported conversion: %s to %s", srcExt, targetExt)
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, item := range v {
			out[k] = normalizeValue(item)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, item := range v {
			out[fmt.Sprint(k)] = normalizeValue(item)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = normalizeValue(item)
		}
		return out
	default:
		return v
	}
}

func (c *DocDataConverter) convertJSONToYAML(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read json file: %w", err)
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	payload = normalizeValue(payload)
	out, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertYAMLToJSON(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read yaml file: %w", err)
	}

	var payload interface{}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse yaml: %w", err)
	}

	payload = normalizeValue(payload)
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertYAMLToTOML(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read yaml file: %w", err)
	}

	var payload interface{}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse yaml: %w", err)
	}

	payload = normalizeValue(payload)
	out, err := toml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal toml: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write toml file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertTOMLToYAML(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read toml file: %w", err)
	}

	var payload map[string]interface{}
	if err := toml.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse toml: %w", err)
	}

	payload = normalizeValue(payload).(map[string]interface{})
	out, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertJSONToXML(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read json file: %w", err)
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	payload = normalizeValue(payload)
	var xmlBytes []byte
	switch v := payload.(type) {
	case map[string]interface{}:
		xmlBytes, err = mxj.Map(v).XmlIndent("", "  ")
	default:
		xmlBytes, err = mxj.Map{"root": v}.XmlIndent("", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal xml: %w", err)
	}

	if err := os.WriteFile(target, xmlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write xml file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertXMLToJSON(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read xml file: %w", err)
	}

	m, err := mxj.NewMapXml(data)
	if err != nil {
		return fmt.Errorf("failed to parse xml: %w", err)
	}

	payload := normalizeValue(m)
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertJSONToCSV(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read json file: %w", err)
	}

	headers, rows, err := jsonToRows(data)
	if err != nil {
		return err
	}

	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write csv header: %w", err)
		}
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write csv row: %w", err)
		}
	}

	return nil
}

func (c *DocDataConverter) convertCSVToJSON(src, target string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse csv file: %w", err)
	}
	if len(records) == 0 {
		return fmt.Errorf("csv file is empty")
	}

	headers := normalizeHeaders(records[0])
	data := make([]map[string]string, 0, len(records)-1)
	for _, row := range records[1:] {
		entry := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(row) {
				entry[header] = row[i]
			} else {
				entry[header] = ""
			}
		}
		data = append(data, entry)
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertJSONToExcel(src, target string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read json file: %w", err)
	}

	headers, rows, err := jsonToRows(data)
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close excel file: %v\n", err)
		}
	}()

	sheetName := "Sheet1"

	for colIdx, header := range headers {
		cellRef, err := excelize.CoordinatesToCellName(colIdx+1, 1)
		if err != nil {
			return fmt.Errorf("failed to create cell reference: %w", err)
		}
		if err := f.SetCellValue(sheetName, cellRef, header); err != nil {
			return fmt.Errorf("failed to set header cell: %w", err)
		}
	}

	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cellRef, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err != nil {
				return fmt.Errorf("failed to create cell reference: %w", err)
			}
			if err := f.SetCellValue(sheetName, cellRef, value); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	if err := f.SaveAs(target); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}

func (c *DocDataConverter) convertExcelToJSON(src, target string) error {
	f, err := excelize.OpenFile(src)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close excel file: %v\n", err)
		}
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("excel file has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return fmt.Errorf("failed to read excel sheet: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("excel sheet is empty")
	}

	headers := normalizeHeaders(rows[0])
	data := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		entry := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(row) {
				entry[header] = row[i]
			} else {
				entry[header] = ""
			}
		}
		data = append(data, entry)
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(target, out, 0644); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}

func jsonToRows(data []byte) ([]string, [][]string, error) {
	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, nil, fmt.Errorf("failed to parse json: %w", err)
	}

	payload = normalizeValue(payload)

	switch v := payload.(type) {
	case []interface{}:
		return rowsFromSlice(v)
	case map[string]interface{}:
		return rowsFromMap(v)
	default:
		return []string{"value"}, [][]string{{formatValue(v)}}, nil
	}
}

func rowsFromSlice(items []interface{}) ([]string, [][]string, error) {
	headersSet := map[string]struct{}{}
	rows := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			rows = append(rows, m)
			for key := range m {
				headersSet[key] = struct{}{}
			}
		} else {
			rows = append(rows, map[string]interface{}{"value": item})
			headersSet["value"] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headersSet))
	for key := range headersSet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	outRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = formatValue(row[header])
		}
		outRows = append(outRows, record)
	}

	return headers, outRows, nil
}

func rowsFromMap(item map[string]interface{}) ([]string, [][]string, error) {
	headers := make([]string, 0, len(item))
	for key := range item {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	record := make([]string, len(headers))
	for i, header := range headers {
		record[i] = formatValue(item[header])
	}

	return headers, [][]string{record}, nil
}

func normalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, header := range headers {
		if strings.TrimSpace(header) == "" {
			normalized[i] = fmt.Sprintf("col%d", i+1)
		} else {
			normalized[i] = header
		}
	}
	return normalized
}

func formatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		if bytes.HasPrefix(encoded, []byte("\"")) {
			var out string
			if err := json.Unmarshal(encoded, &out); err == nil {
				return out
			}
		}
		return string(encoded)
	}
}
