package watch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Watcher struct {
	root string
	skip map[string]bool
}

func New(root string, skipDirs []string) *Watcher {
	skip := map[string]bool{}
	for _, s := range skipDirs {
		skip[s] = true
	}
	return &Watcher{root: root, skip: skip}
}

func (wt *Watcher) skipped(path string) bool {
	rel, err := filepath.Rel(wt.root, path)
	if err != nil {
		return true
	}
	first, _, _ := strings.Cut(rel, string(os.PathSeparator))
	return wt.skip[first]
}

func IsJPEG(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if strings.HasSuffix(base, ".gpspicturetagd.tmp") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg"
}

func (wt *Watcher) Scan() ([]string, error) {
	var out []string
	err := filepath.WalkDir(wt.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != wt.root && wt.skipped(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if wt.skipped(path) || !IsJPEG(path) {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out, err
}
