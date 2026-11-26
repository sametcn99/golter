//go:build linux || darwin

package converter

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestCancelDuringImageRead(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "reading.png")
	if err := unix.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	valid := filepath.Join(dir, "fixture.png")
	createTestImage(t, valid)
	input, err := os.Open(valid)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	dst := filepath.Join(dir, "previous.webp")
	put(t, dst, "original")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- (&ImageConverter{}).Convert(ctx, fifo, dst, Options{Quality: "High"}) }()
	// Opening the writer succeeds only once the converter has opened the input.
	writer, err := os.OpenFile(fifo, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, err = io.Copy(writer, input)
	closeErr := writer.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("image worker did not finish")
	}
	assertText(t, dst, "original")
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 3 {
		t.Fatalf("unexpected temporary output: %v, %v", files, err)
	}
}
