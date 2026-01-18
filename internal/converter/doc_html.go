package converter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	goepub "github.com/bmaupin/go-epub"
)

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
