package web

import (
	"path/filepath"
	"runtime"
)

func projectRootPath(rel string) string {
	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filename)
	return filepath.Join(base, "..", "..", rel)
}
