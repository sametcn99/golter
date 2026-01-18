package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type item struct {
	path     string
	isDir    bool
	info     os.FileInfo
	selected bool
}

func (i item) Title() string {
	// Parent directory item
	if i.info == nil && i.isDir {
		return iconBack + " .."
	}

	var prefix string
	if i.isDir {
		prefix = "  " // No checkbox for directories
	} else if i.selected {
		prefix = iconSelected
	} else {
		prefix = iconNotSelected
	}

	var icon string
	if i.isDir {
		icon = iconFolder
	} else {
		ext := strings.ToLower(filepath.Ext(i.info.Name()))
		icon = getFileIcon(ext)
	}

	name := i.info.Name()
	if i.isDir {
		return fmt.Sprintf("%s %s %s", prefix, icon, name)
	}

	// Add file size for files
	size := FormatSize(i.info.Size())
	return fmt.Sprintf("%s %s %s (%s)", prefix, icon, name, size)
}

func (i item) Description() string {
	return ""
}

func (i item) FilterValue() string {
	if i.info == nil {
		return ".."
	}
	return i.info.Name()
}
