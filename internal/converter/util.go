package converter

import (
	"strings"
)

// normalizeExt ensures an extension is lowercase and starts with a dot.
func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}
