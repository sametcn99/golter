package converter

import (
	"fmt"
	"sync"
)

// Manager handles the registration and retrieval of converters.
type Manager struct {
	mu         sync.RWMutex
	converters []Converter
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		converters: make([]Converter, 0),
	}
}

// Register adds a converter to the manager.
func (m *Manager) Register(c Converter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.converters = append(m.converters, c)
}

// FindConverter returns a converter that can handle the specified extensions.
func (m *Manager) FindConverter(srcExt, targetExt string) (Converter, error) {
	srcExt = normalizeExt(srcExt)
	targetExt = normalizeExt(targetExt)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.converters {
		if c.CanConvert(srcExt, targetExt) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no converter found for %s to %s", srcExt, targetExt)
}

// GetSupportedTargetFormats returns a list of supported target extensions for a given source extension.
func (m *Manager) GetSupportedTargetFormats(srcExt string) []string {
	targets := make(map[string]bool)
	srcExt = normalizeExt(srcExt)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.converters {
		for _, t := range c.SupportedTargetFormats(srcExt) {
			targets[t] = true
		}
	}

	result := make([]string, 0, len(targets))
	for t := range targets {
		result = append(result, t)
	}
	return result
}

// SupportedExtensions returns a list of supported source extensions.
func (m *Manager) SupportedExtensions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exts := make(map[string]bool)
	for _, c := range m.converters {
		for _, ext := range c.SupportedSourceExtensions() {
			exts[ext] = true
		}
	}

	result := make([]string, 0, len(exts))
	for ext := range exts {
		result = append(result, ext)
	}
	return result
}
