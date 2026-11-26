package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ConvertNew publishes a complete batch output only if its destination is still free.
func ConvertNew(ctx context.Context, c Converter, src, target string, opts Options) error {
	return stagedOutput(ctx, target, func(temp string) error { return c.Convert(ctx, src, temp, opts) }, publishNew)
}

func stagedOutput(ctx context.Context, target string, write func(string) error, publish func(string, string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(target), ".golter-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	temp := filepath.Join(dir, filepath.Base(target))
	if err = write(temp); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = publish(temp, target); err != nil {
		return fmt.Errorf("cannot publish %s: %w", target, err)
	}
	return nil
}
