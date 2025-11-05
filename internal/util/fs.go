package util

import (
	"os"
	"path/filepath"
)

// MustMkdirAll ensures a directory exists, ignoring errors if already there.
func MustMkdirAll(p string) {
	_ = os.MkdirAll(p, 0o755)
}

// FileStem returns a filename without its extension.
func FileStem(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
