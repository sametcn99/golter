//go:build !windows && !darwin && !linux

package converter

import "os"

func publishNew(src, dst string) error { return os.Link(src, dst) }
