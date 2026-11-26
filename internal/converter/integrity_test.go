package converter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

func put(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSafePublication(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "a.md"), filepath.Join(dir, "a.html")
	put(t, src, "# content")
	c := &DocumentConverter{}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- ConvertNew(t.Context(), c, src, dst, Options{}) }()
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("%d concurrent publishers succeeded", successes)
	}
	assertText(t, dst, "<h1>content</h1>")
	for _, cancelled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		err := stagedOutput(ctx, dst, func(temp string) error {
			put(t, temp, "partial")
			if cancelled {
				cancel()
				return nil
			}
			return errors.New("encode failed")
		}, publishNew)
		cancel()
		if err == nil {
			t.Fatal("published partial output")
		}
		assertText(t, dst, "<h1>content</h1>")
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 2 {
		t.Fatalf("unexpected leftovers %v %v", files, err)
	}
}

func TestCancelledImage(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "image.png"), filepath.Join(dir, "image.jpg")
	createTestImage(t, src)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c := &ImageConverter{}
	if err := c.Convert(ctx, src, dst, Options{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled image appeared", err)
	}
	put(t, dst, "original")
	if err := c.Convert(ctx, src, dst, Options{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	assertText(t, dst, "original")
}

type failingCSV struct{ writeErr, closeErr error }

func (w failingCSV) Write(p []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return len(p), nil
}
func (w failingCSV) Close() error { return w.closeErr }

func TestCSVWriteErrors(t *testing.T) {
	failure := errors.New("disk full")
	for _, writer := range []failingCSV{{writeErr: failure}, {closeErr: failure}} {
		if err := writeCSV(writer, [][]string{{"small buffered row"}}); !errors.Is(err, failure) {
			t.Fatalf("write error lost: %v", err)
		}
	}
}

func TestExactNumbers(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "ids.json")
	const id = "9007199254740993"
	put(t, src, `[{"id":`+id+`}]`)
	c := &DocDataConverter{}
	for _, ext := range []string{".csv", ".xlsx", ".yaml", ".xml"} {
		t.Run(ext, func(t *testing.T) {
			dst := filepath.Join(dir, "ids"+ext)
			if err := c.Convert(t.Context(), src, dst, Options{}); err != nil {
				t.Fatal(err)
			}
			switch ext {
			case ".xlsx":
				f, e := excelize.OpenFile(dst)
				if e != nil {
					t.Fatal(e)
				}
				defer f.Close()
				value, e := f.GetCellValue("Sheet1", "A2")
				if e != nil || value != id {
					t.Fatalf("cell %q, %v", value, e)
				}
			case ".yaml":
				b, e := os.ReadFile(dst)
				if e != nil {
					t.Fatal(e)
				}
				var rows []map[string]any
				if e = yaml.Unmarshal(b, &rows); e != nil || len(rows) != 1 || fmt.Sprint(rows[0]["id"]) != id {
					t.Fatalf("YAML %s (%v)", b, e)
				}
				if _, isString := rows[0]["id"].(string); isString {
					t.Fatal("ID lost numeric type")
				}
			default:
				assertText(t, dst, id)
			}
		})
	}
	put(t, src, `{"id":1} {"id":2}`)
	if e := c.Convert(t.Context(), src, filepath.Join(dir, "bad.csv"), Options{}); e == nil {
		t.Fatal("accepted trailing JSON")
	}
}

func TestColumnConflicts(t *testing.T) {
	d := &DocumentConverter{}
	c := &DocDataConverter{}
	for _, header := range []string{"id,id", ",col1"} {
		t.Run(header, func(t *testing.T) {
			dir := t.TempDir()
			csv := filepath.Join(dir, "table.csv")
			xlsx := filepath.Join(dir, "table.xlsx")
			dst := filepath.Join(dir, "table.json")
			put(t, csv, header+"\nA,B\n")
			put(t, dst, "original")
			if err := d.Convert(t.Context(), csv, xlsx, Options{}); err != nil {
				t.Fatal(err)
			}
			for _, src := range []string{csv, xlsx} {
				err := c.Convert(t.Context(), src, dst, Options{})
				if err == nil || !strings.Contains(err.Error(), "duplicate column") {
					t.Fatal(err)
				}
				assertText(t, dst, "original")
			}
		})
	}
}

func TestExcelFormats(t *testing.T) {
	for _, c := range []Converter{&DocumentConverter{}, &DocDataConverter{}} {
		if slices.Contains(c.SupportedSourceExtensions(), ".xls") {
			t.Fatal("legacy XLS advertised")
		}
		for _, ext := range c.SupportedSourceExtensions() {
			if slices.Contains(c.SupportedTargetFormats(ext), ".xls") {
				t.Fatal("legacy XLS output advertised")
			}
		}
		if err := c.Convert(t.Context(), filepath.Join("testdata", "excel97.xls"), filepath.Join(t.TempDir(), "result.csv"), Options{}); err == nil || !strings.Contains(err.Error(), "unsupported conversion") {
			t.Fatal(err)
		}
	}
}

func TestFallbackPDF(t *testing.T) {
	reader := requireTool(t, "pdftotext")
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	d := &DocumentConverter{}
	const text = "Çağrı ŞİĞÜ ölümsüz"
	md, book, pdf := filepath.Join(dir, "kitap.md"), filepath.Join(dir, "kitap.epub"), filepath.Join(dir, "kitap.pdf")
	existing := filepath.Join(dir, "kitap_temp.html")
	put(t, existing, "existing")
	put(t, md, "**"+text+"**")
	if err := d.Convert(t.Context(), md, book, Options{}); err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{md, book} {
		if err := d.Convert(t.Context(), src, pdf, Options{}); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(reader, "-enc", "UTF-8", pdf, "-").CombinedOutput()
		if err != nil || !strings.Contains(string(out), text) {
			t.Fatalf("%s text: %q (%v)", src, out, err)
		}
		assertText(t, existing, "existing")
	}
	// Make the final write fail after the intermediate HTML has been created.
	if err := os.Remove(pdf); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(pdf, 0700); err != nil {
		t.Fatal(err)
	}
	if err := d.Convert(t.Context(), book, pdf, Options{}); err == nil {
		t.Fatal("wrote PDF over directory")
	}
	assertText(t, existing, "existing")
}

func TestAdvertisedEbookConversions(t *testing.T) {
	calibre := requireTool(t, "ebook-convert")
	pdfReader := requireTool(t, "pdftotext")
	dir := t.TempDir()
	book := filepath.Join(dir, "input.epub")
	createTestEPUB(t, book)
	d := &DocumentConverter{}
	for _, ext := range []string{".epub", ".mobi", ".azw", ".azw3", ".fb2", ".md", ".html"} {
		t.Run(ext, func(t *testing.T) {
			src := filepath.Join(t.TempDir(), "input"+ext)
			switch ext {
			case ".epub":
				src = book
			case ".md":
				put(t, src, "# Hello World")
			case ".html":
				put(t, src, "<html><body><p>Hello World</p></body></html>")
			default:
				fixture := src
				if ext == ".azw" {
					fixture = src + ".mobi"
				}
				if out, err := exec.Command(calibre, book, fixture).CombinedOutput(); err != nil {
					t.Fatalf("fixture: %s (%v)", out, err)
				}
				if fixture != src {
					if err := os.Rename(fixture, src); err != nil {
						t.Fatal(err)
					}
				}
			}
			if d.CanConvert(ext, ".azw") {
				t.Fatal("AZW output advertised")
			}
			for _, target := range d.SupportedTargetFormats(ext) {
				t.Run(target, func(t *testing.T) {
					dst := filepath.Join(t.TempDir(), "output"+target)
					if err := d.Convert(t.Context(), src, dst, Options{}); err != nil {
						t.Fatal(err)
					}
					switch target {
					case ".txt", ".md", ".html":
						assertText(t, dst, "Hello World")
					case ".pdf":
						out, err := exec.Command(pdfReader, dst, "-").CombinedOutput()
						if err != nil || !strings.Contains(strings.Join(strings.Fields(string(out)), " "), "Hello World") {
							t.Fatalf("PDF text %q: %v", out, err)
						}
					case ".docx":
						out, err := exec.Command(requireTool(t, "pandoc"), dst, "-t", "plain").CombinedOutput()
						if err != nil || !strings.Contains(string(out), "Hello World") {
							t.Fatalf("DOCX: %s, %v", out, err)
						}
					default:
						txt := dst + ".txt"
						if out, err := exec.Command(calibre, dst, txt).CombinedOutput(); err != nil {
							t.Fatalf("read back: %s (%v)", out, err)
						}
						assertText(t, txt, "Hello World")
					}
				})
			}
		})
	}
}

func TestJSONTableRoundTrip(t *testing.T) {
	c := &DocDataConverter{}
	for _, ext := range []string{".csv", ".xlsx"} {
		dir := t.TempDir()
		src := filepath.Join(dir, "source.json")
		table := filepath.Join(dir, "table"+ext)
		dst := filepath.Join(dir, "out.json")
		put(t, src, `[{"name":"Ada","id":9007199254740993},{"name":"Bob","id":42}]`)
		if err := c.Convert(t.Context(), src, table, Options{}); err != nil {
			t.Fatal(err)
		}
		if err := c.Convert(t.Context(), table, dst, Options{}); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(dst)
		if err != nil {
			t.Fatal(err)
		}
		var rows []struct{ Name, ID string }
		if err = json.Unmarshal(b, &rows); err != nil || len(rows) != 2 || rows[0].Name != "Ada" || rows[0].ID != "9007199254740993" || rows[1].Name != "Bob" || rows[1].ID != "42" {
			t.Fatalf("%s: %s, %v", ext, b, err)
		}
	}
}
