package converter

import "golang.org/x/sys/unix"

func publishNew(src, dst string) error { return unix.RenamexNp(src, dst, unix.RENAME_EXCL) }
