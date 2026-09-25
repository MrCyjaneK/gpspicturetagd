package tag

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gpspicturetagd/internal/seen"
)

func writeTinyJPEG(t *testing.T, path string, mtime time.Time) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.Set(0, 0, color.RGBA{R: 1, A: 255})
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 70}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestFileTagsAndPreservesMtime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pic.jpg")
	mtime := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
	writeTinyJPEG(t, path, mtime)

	pos := FromFix(60.17, 24.94, 8, true, time.Now())
	ok, reason, err := File(path, pos, time.Hour, nil)
	if err != nil || !ok || reason != "" {
		t.Fatalf("File: ok=%t reason=%q err=%v", ok, reason, err)
	}
	if !HasGPS(path) {
		t.Fatal("expected GPS after tag")
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st.ModTime().Equal(mtime) {
		t.Fatalf("mtime %v, want %v", st.ModTime(), mtime)
	}

	ok, reason, err = File(path, pos, time.Hour, nil)
	if err != nil || ok || reason != "already tagged" {
		t.Fatalf("retarget: ok=%t reason=%q err=%v", ok, reason, err)
	}
}

func TestCheckTooOld(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.jpg")
	writeTinyJPEG(t, path, time.Now().Add(-3*time.Hour))
	if got := Check(path, time.Hour, nil); got != "too old" {
		t.Fatalf("Check = %q", got)
	}
}

func TestCheckCacheSkipsEXIF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pic.jpg")
	writeTinyJPEG(t, path, time.Now())
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	c := seen.Open(filepath.Join(dir, "seen"))
	c.Remember(path, st.ModTime().UnixNano(), st.Size())
	if got := Check(path, time.Hour, c); got != "already tagged" {
		t.Fatalf("cached Check = %q", got)
	}
}
