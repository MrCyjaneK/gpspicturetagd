package seen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/renameio/v2"
)

type rec struct {
	mod  int64
	size int64
}

// Cache remembers JPEGs that already have GPS so later scans can skip EXIF reads.
// The on-disk file is a text list: "<mtime_ns> <size> <path>".
type Cache struct {
	path string
	all  map[string]rec
	live map[string]rec
}

func Open(path string) *Cache {
	c := &Cache{
		path: path,
		all:  make(map[string]rec),
		live: make(map[string]rec),
	}
	f, err := os.Open(path)
	if err != nil {
		return c
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		path, r, ok := parseLine(sc.Text())
		if ok {
			c.all[path] = r
		}
	}
	return c
}

func parseLine(line string) (string, rec, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", rec{}, false
	}
	a := strings.IndexByte(line, ' ')
	if a < 0 {
		return "", rec{}, false
	}
	rest := line[a+1:]
	b := strings.IndexByte(rest, ' ')
	if b < 0 {
		return "", rec{}, false
	}
	mod, err := strconv.ParseInt(line[:a], 10, 64)
	if err != nil {
		return "", rec{}, false
	}
	size, err := strconv.ParseInt(rest[:b], 10, 64)
	if err != nil {
		return "", rec{}, false
	}
	path := rest[b+1:]
	if path == "" {
		return "", rec{}, false
	}
	return path, rec{mod: mod, size: size}, true
}

func (c *Cache) Tagged(path string, mod int64, size int64) bool {
	if c == nil {
		return false
	}
	r, ok := c.all[path]
	if !ok || r.mod != mod || r.size != size {
		return false
	}
	c.live[path] = r
	return true
}

func (c *Cache) Remember(path string, mod int64, size int64) {
	if c == nil {
		return
	}
	r := rec{mod: mod, size: size}
	c.all[path] = r
	c.live[path] = r
}

func (c *Cache) Save() error {
	if c == nil || c.path == "" {
		return nil
	}
	var b strings.Builder
	for path, r := range c.live {
		fmt.Fprintf(&b, "%d %d %s\n", r.mod, r.size, path)
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	if err := renameio.WriteFile(c.path, []byte(b.String()), 0o600); err != nil {
		return err
	}
	return os.Chmod(c.path, 0o600)
}
