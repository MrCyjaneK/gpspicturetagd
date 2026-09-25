package seen

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundTripAndInvalidate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "seen")
	file := filepath.Join(dir, "pic.jpg")
	if err := os.WriteFile(file, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	c := Open(p)
	if c.Tagged(file, st.ModTime().UnixNano(), st.Size()) {
		t.Fatal("empty cache should miss")
	}
	c.Remember(file, st.ModTime().UnixNano(), st.Size())
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	c2 := Open(p)
	if !c2.Tagged(file, st.ModTime().UnixNano(), st.Size()) {
		t.Fatal("expected hit after reload")
	}
	if c2.Tagged(file, st.ModTime().UnixNano(), st.Size()+1) {
		t.Fatal("size change should miss")
	}
	if c2.Tagged(file, st.ModTime().Add(time.Second).UnixNano(), st.Size()) {
		t.Fatal("mtime change should miss")
	}
}

func TestSavePrunesUnseen(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "seen")
	c := Open(p)
	c.Remember("/a.jpg", 1, 10)
	c.Remember("/b.jpg", 2, 20)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	c2 := Open(p)
	if !c2.Tagged("/a.jpg", 1, 10) {
		t.Fatal("want a")
	}
	if err := c2.Save(); err != nil {
		t.Fatal(err)
	}

	c3 := Open(p)
	if !c3.Tagged("/a.jpg", 1, 10) {
		t.Fatal("a should remain")
	}
	if c3.Tagged("/b.jpg", 2, 20) {
		t.Fatal("b was not live this scan")
	}
}

func TestSaveMode(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "seen")
	c := Open(p)
	c.Remember("/a.jpg", 1, 10)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", st.Mode().Perm())
	}
}
