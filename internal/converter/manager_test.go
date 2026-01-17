package converter

import (
	"reflect"
	"sort"
	"testing"
)

// MockConverter is a mock implementation of the Converter interface for testing.
type MockConverter struct {
	name             string
	supportedSources []string
	supportedTargets map[string][]string
	convertFunc      func(src, target string, opts Options) error
}

func (m *MockConverter) Name() string {
	return m.name
}

func (m *MockConverter) CanConvert(srcExt, targetExt string) bool {
	targets, ok := m.supportedTargets[srcExt]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == targetExt {
			return true
		}
	}
	return false
}

func (m *MockConverter) SupportedSourceExtensions() []string {
	return m.supportedSources
}

func (m *MockConverter) SupportedTargetFormats(srcExt string) []string {
	return m.supportedTargets[srcExt]
}

func (m *MockConverter) Convert(src, target string, opts Options) error {
	if m.convertFunc != nil {
		return m.convertFunc(src, target, opts)
	}
	return nil
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("NewManager returned nil")
	}
	if len(m.converters) != 0 {
		t.Errorf("NewManager should have 0 converters, got %d", len(m.converters))
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager()
	c := &MockConverter{name: "Test"}
	m.Register(c)

	if len(m.converters) != 1 {
		t.Errorf("Expected 1 converter, got %d", len(m.converters))
	}
	if m.converters[0] != c {
		t.Error("Registered converter does not match")
	}
}

func TestManager_FindConverter(t *testing.T) {
	m := NewManager()
	c1 := &MockConverter{
		name:             "C1",
		supportedSources: []string{".jpg"},
		supportedTargets: map[string][]string{".jpg": {".png"}},
	}
	c2 := &MockConverter{
		name:             "C2",
		supportedSources: []string{".mp4"},
		supportedTargets: map[string][]string{".mp4": {".gif"}},
	}
	m.Register(c1)
	m.Register(c2)

	tests := []struct {
		name      string
		srcExt    string
		targetExt string
		want      Converter
		wantErr   bool
	}{
		{"Found C1", ".jpg", ".png", c1, false},
		{"Found C1 case insensitive", "JPG", "PNG", c1, false},
		{"Found C1 without dot", "jpg", "png", c1, false},
		{"Found C2", ".mp4", ".gif", c2, false},
		{"Not Found", ".jpg", ".gif", nil, true},
		{"Unknown Source", ".txt", ".png", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.FindConverter(tt.srcExt, tt.targetExt)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindConverter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindConverter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetSupportedTargetFormats(t *testing.T) {
	m := NewManager()
	c1 := &MockConverter{
		supportedTargets: map[string][]string{
			".jpg": {".png", ".webp"},
		},
	}
	c2 := &MockConverter{
		supportedTargets: map[string][]string{
			".jpg": {".bmp"}, // Another converter supporting jpg
		},
	}
	m.Register(c1)
	m.Register(c2)

	got := m.GetSupportedTargetFormats(".jpg")
	sort.Strings(got)
	want := []string{".bmp", ".png", ".webp"}
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetSupportedTargetFormats() = %v, want %v", got, want)
	}

	gotEmpty := m.GetSupportedTargetFormats(".txt")
	if len(gotEmpty) != 0 {
		t.Errorf("GetSupportedTargetFormats() for unknown ext should be empty, got %v", gotEmpty)
	}
}

func TestManager_SupportedExtensions(t *testing.T) {
	m := NewManager()
	c1 := &MockConverter{
		supportedSources: []string{".jpg", ".png"},
	}
	c2 := &MockConverter{
		supportedSources: []string{".mp4"},
	}
	m.Register(c1)
	m.Register(c2)

	got := m.SupportedExtensions()
	sort.Strings(got)
	want := []string{".jpg", ".mp4", ".png"}
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("SupportedExtensions() = %v, want %v", got, want)
	}
}
