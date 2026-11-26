package tui

import (
	"context"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/sametcn99/golter/internal/converter"
)

func key(m *Model, k tea.KeyType, s string) tea.Cmd {
	updated, cmd := m.Update(tea.KeyMsg{Type: k, Runes: []rune(s)})
	*m = updated.(Model)
	return cmd
}

func TestSharedActions(t *testing.T) {
	m := NewModel(t.TempDir())
	for _, tc := range []struct {
		files    []string
		targets  []string
		compress bool
	}{
		{[]string{"a.md", "b.csv"}, nil, false},
		{[]string{"a.csv", "b.md"}, nil, false},
		{[]string{"a.json", "b.csv"}, []string{".xlsx"}, false},
		{[]string{"a.pdf"}, []string{".md", ".pdf"}, true},
		{[]string{"a.gif", "b.mp4"}, []string{".avi", ".gif", ".mkv", ".mov", ".mp4", ".webm"}, true},
	} {
		formats, actions := selectionOptions(tc.files, m.manager)
		if !slices.Equal(formats, tc.targets) || slices.Contains(actions, iconCompress+"  Compress Files") != tc.compress {
			t.Fatalf("%v: %v %v", tc.files, formats, actions)
		}
	}
	for _, ext := range []string{".json", ".xml", ".yaml", ".yml", ".toml"} {
		if getFileType(ext) != FileTypeDocument {
			t.Fatal(ext)
		}
	}
	for _, name := range []string{"a.md", "b.csv"} {
		if err := os.WriteFile(filepath.Join(m.selector.currentDir, name), []byte("file"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	m.selector.loadFiles()
	key(&m, tea.KeyRunes, "a")
	key(&m, tea.KeyRunes, "c")
	key(&m, tea.KeyEnter, "")
	if m.state != StateSelectingAction || !strings.Contains(m.View(), "no common") {
		t.Fatal(m.View())
	}
	key(&m, tea.KeyEsc, "")
	if m.state != StateSelecting {
		t.Fatal(m.state)
	}
}

func TestFilteredSelectionAndResize(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.md", "query.md"} {
		if e := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	m := NewModel(dir)
	key(&m, tea.KeyRunes, "/")
	key(&m, tea.KeyRunes, "q")
	key(&m, tea.KeyRunes, "c")
	if m.state != StateSelecting || m.selector.list.FilterInput.Value() != "qc" {
		t.Fatal("filter captured by shortcut")
	}
	m.selector.list.SetFilterText("query.md")
	cmd := key(&m, tea.KeySpace, "")
	if cmd == nil {
		t.Fatal("filter not refreshed")
	}
	next, _ := m.Update(cmd())
	m = next.(Model)
	if !m.selector.list.SelectedItem().(item).selected {
		t.Fatal("missing selection in filtered view")
	}
	m.selector.list.ResetFilter()
	items := m.selector.list.Items()
	if len(items) != 3 || !items[0].(item).isDir || items[1].(item).selected || !items[2].(item).selected {
		t.Fatal(items)
	}
	for _, width := range []int{10, 12, 40, 80} {
		for _, p := range []string{".", "/", "Çığ", "/çok/uzun/界/🙂/bir/dizin/İstanbul"} {
			next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 20})
			m = next.(Model)
			m.selector.currentDir = p
			if !utf8.ValidString(m.View()) {
				t.Fatal("invalid UTF-8 view")
			}
			short := truncatePath(p, width-10)
			if !utf8.ValidString(short) || ansi.StringWidth(short) > width-10 {
				t.Fatalf("%d %q", width, short)
			}
		}
	}
}

type gatedXML struct {
	converter.DocDataConverter
	started chan struct{}
	release chan struct{}
}

func (g *gatedXML) Convert(ctx context.Context, src, dst string, opts converter.Options) error {
	if filepath.Base(src) == "b.xml" {
		close(g.started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-g.release:
		}
	}
	return g.DocDataConverter.Convert(ctx, src, dst, opts)
}

func TestKeyboardBatch(t *testing.T) {
	for _, scenario := range []string{"success", "existing output", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"a.xml", "b.xml"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte("<root>value</root>"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			m := NewModel(dir)
			g := &gatedXML{started: make(chan struct{}), release: make(chan struct{})}
			m.manager = converter.NewManager()
			m.manager.Register(g)
			key(&m, tea.KeyRunes, "a")
			key(&m, tea.KeyRunes, "c")
			key(&m, tea.KeyEnter, "")
			if m.state != StateSelectingFormat || !strings.Contains(m.View(), ".json") {
				t.Fatal(m.View())
			}
			if scenario == "existing output" {
				if e := os.WriteFile(filepath.Join(dir, "a.json"), []byte("original"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			cmd := key(&m, tea.KeyEnter, "")
			defer m.cancelCtx()
			messages := make(chan tea.Msg, 2)
			for _, worker := range cmd().(tea.BatchMsg) {
				go func() { messages <- worker() }()
			}
			select {
			case <-g.started:
			case <-time.After(3 * time.Second):
				t.Fatal("second worker did not start")
			}
			receive := func() {
				select {
				case msg := <-messages:
					next, _ := m.Update(msg)
					m = next.(Model)
				case <-time.After(3 * time.Second):
					t.Fatal("worker did not report")
				}
			}
			receive()
			if m.state != StateConverting || !strings.Contains(m.View(), "1 / 2") {
				t.Fatal(m.View())
			}
			if scenario == "cancel" {
				key(&m, tea.KeyCtrlC, "")
			} else {
				close(g.release)
			}
			receive()
			if m.state != StateDone {
				t.Fatal(m.state)
			}
			if scenario == "cancel" {
				if _, e := os.Stat(filepath.Join(dir, "b.json")); !errors.Is(e, os.ErrNotExist) || !strings.Contains(m.output, "context canceled") {
					t.Fatal(m.output, e)
				}
			}
			if scenario == "existing output" {
				b, e := os.ReadFile(filepath.Join(dir, "a.json"))
				if e != nil || string(b) != "original" || !strings.Contains(m.output, "1/2") {
					t.Fatal(string(b), m.output, e)
				}
			}
			if scenario == "success" && !strings.Contains(m.output, "Successfully converted 2") {
				t.Fatal(m.output)
			}
			key(&m, tea.KeyEsc, "")
			if m.state != StateSelecting || len(m.selector.SelectedFiles()) != 0 {
				t.Fatal(m.state)
			}
			key(&m, tea.KeyRunes, "q")
			key(&m, tea.KeyRunes, "n")
			if m.state != StateSelecting {
				t.Fatal(m.state)
			}
		})
	}
}

func TestCompressionAndCollisions(t *testing.T) {
	jobs := planBatch([]string{"photo.jpg", "photo.webp"}, ".png")
	for _, j := range jobs {
		if j.err == nil || !strings.Contains(j.err.Error(), "multiple inputs") {
			t.Fatal(j)
		}
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "image.png")
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	f, e := os.Create(src)
	if e != nil {
		t.Fatal(e)
	}
	if e = png.Encode(f, img); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	m := NewModel(dir)
	key(&m, tea.KeyRunes, "a")
	key(&m, tea.KeyRunes, "c")
	key(&m, tea.KeyDown, "")
	key(&m, tea.KeyEnter, "")
	if m.state != StateSelectingQuality || !strings.Contains(m.View(), "compression quality") {
		t.Fatal(m.View())
	}
	key(&m, tea.KeyDown, "")
	cmd := key(&m, tea.KeyEnter, "")
	next, _ := m.Update(cmd())
	m = next.(Model)
	if m.state != StateDone || m.err != nil {
		t.Fatal(m.err, m.state)
	}
	output, e := os.Open(filepath.Join(dir, "image_compressed.png"))
	if e != nil {
		t.Fatal(e)
	}
	defer output.Close()
	result, e := png.Decode(output)
	if e != nil || result.Bounds() != img.Bounds() {
		t.Fatal(result, e)
	}
}

func TestDefaultDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	m := NewModel("")
	cwd, e := os.Getwd()
	if e != nil || m.selector.currentDir != cwd {
		t.Fatal(m.selector.currentDir, cwd, e)
	}
	if m.selector.list.FilterState() != list.Unfiltered {
		t.Fatal("unexpected filter")
	}
}
