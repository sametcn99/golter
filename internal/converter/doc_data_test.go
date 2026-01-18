package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDocDataConverter_SupportedExtensions(t *testing.T) {
	c := &DocDataConverter{}

	expectedSrc := []string{".json", ".yaml", ".yml", ".xml", ".toml", ".csv", ".xlsx", ".xls"}
	srcExts := c.SupportedSourceExtensions()
	if len(srcExts) == 0 {
		t.Error("SupportedSourceExtensions returned empty list")
	}
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

	for _, ext := range expectedSrc {
		targets := c.SupportedTargetFormats(ext)
		if len(targets) == 0 {
			t.Errorf("SupportedTargetFormats(%s) returned empty list", ext)
		}
	}
}

func TestDocDataConverter_Conversions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "golter_data_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &DocDataConverter{}

	jsonPath := filepath.Join(tmpDir, "sample.json")
	jsonPayload := []byte(`[{"name":"Ada","age":42},{"name":"Bob","age":35}]`)
	if err := os.WriteFile(jsonPath, jsonPayload, 0644); err != nil {
		t.Fatalf("failed to write json file: %v", err)
	}

	t.Run("JSON->YAML", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "sample.yaml")
		if err := c.Convert(jsonPath, yamlPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->YAML) failed: %v", err)
		}
	})

	t.Run("YAML->JSON", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "sample.yaml")
		if err := c.Convert(jsonPath, yamlPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->YAML) failed: %v", err)
		}

		jsonFromYaml := filepath.Join(tmpDir, "from_yaml.json")
		if err := c.Convert(yamlPath, jsonFromYaml, Options{}); err != nil {
			t.Fatalf("Convert(YAML->JSON) failed: %v", err)
		}
	})

	t.Run("JSON->XML", func(t *testing.T) {
		xmlPath := filepath.Join(tmpDir, "sample.xml")
		if err := c.Convert(jsonPath, xmlPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->XML) failed: %v", err)
		}
	})

	t.Run("XML->JSON", func(t *testing.T) {
		xmlPath := filepath.Join(tmpDir, "sample.xml")
		if err := c.Convert(jsonPath, xmlPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->XML) failed: %v", err)
		}

		jsonFromXml := filepath.Join(tmpDir, "from_xml.json")
		if err := c.Convert(xmlPath, jsonFromXml, Options{}); err != nil {
			t.Fatalf("Convert(XML->JSON) failed: %v", err)
		}
	})

	t.Run("JSON->CSV", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "sample.csv")
		if err := c.Convert(jsonPath, csvPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->CSV) failed: %v", err)
		}
	})

	t.Run("CSV->JSON", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "sample.csv")
		if err := c.Convert(jsonPath, csvPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->CSV) failed: %v", err)
		}

		jsonFromCSV := filepath.Join(tmpDir, "from_csv.json")
		if err := c.Convert(csvPath, jsonFromCSV, Options{}); err != nil {
			t.Fatalf("Convert(CSV->JSON) failed: %v", err)
		}

		data, err := os.ReadFile(jsonFromCSV)
		if err != nil {
			t.Fatalf("failed to read json output: %v", err)
		}
		var payload interface{}
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("invalid json output: %v", err)
		}
	})

	t.Run("JSON->XLSX", func(t *testing.T) {
		xlsxPath := filepath.Join(tmpDir, "sample.xlsx")
		if err := c.Convert(jsonPath, xlsxPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->XLSX) failed: %v", err)
		}
	})

	t.Run("XLSX->JSON", func(t *testing.T) {
		xlsxPath := filepath.Join(tmpDir, "sample.xlsx")
		if err := c.Convert(jsonPath, xlsxPath, Options{}); err != nil {
			t.Fatalf("Convert(JSON->XLSX) failed: %v", err)
		}

		jsonFromXlsx := filepath.Join(tmpDir, "from_xlsx.json")
		if err := c.Convert(xlsxPath, jsonFromXlsx, Options{}); err != nil {
			t.Fatalf("Convert(XLSX->JSON) failed: %v", err)
		}
	})

	t.Run("YAML->TOML", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "sample.yaml")
		yamlPayload := []byte("name: Ada\nage: 42\n")
		if err := os.WriteFile(yamlPath, yamlPayload, 0644); err != nil {
			t.Fatalf("failed to write yaml file: %v", err)
		}

		tomlPath := filepath.Join(tmpDir, "sample.toml")
		if err := c.Convert(yamlPath, tomlPath, Options{}); err != nil {
			t.Fatalf("Convert(YAML->TOML) failed: %v", err)
		}
	})

	t.Run("TOML->YAML", func(t *testing.T) {
		tomlPath := filepath.Join(tmpDir, "sample.toml")
		tomlPayload := []byte("name = \"Ada\"\nage = 42\n")
		if err := os.WriteFile(tomlPath, tomlPayload, 0644); err != nil {
			t.Fatalf("failed to write toml file: %v", err)
		}

		yamlFromToml := filepath.Join(tmpDir, "from_toml.yaml")
		if err := c.Convert(tomlPath, yamlFromToml, Options{}); err != nil {
			t.Fatalf("Convert(TOML->YAML) failed: %v", err)
		}

		if _, err := os.Stat(yamlFromToml); err != nil {
			t.Fatalf("expected output file not found: %v", err)
		}
	})
}
