package watch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkipsDirsAndNonJPEG(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("keep/a.jpg")
	mustWrite("keep/b.jpeg")
	mustWrite("keep/notes.txt")
	mustWrite("Screenshots/shot.jpg")
	mustWrite("Default/wall.jpg")
	mustWrite("keep/c.jpg.gpspicturetagd.tmp")

	got, err := New(root, []string{"Default", "Screenshots"}).Scan()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		filepath.Join(root, "keep/a.jpg"):  true,
		filepath.Join(root, "keep/b.jpeg"): true,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for _, p := range got {
		if !want[p] {
			t.Fatalf("unexpected %s in %v", p, got)
		}
	}
}

func TestIsJPEG(t *testing.T) {
	if !IsJPEG("x.JPG") || !IsJPEG("x.jpeg") || IsJPEG("x.png") || IsJPEG("x.jpg.gpspicturetagd.tmp") {
		t.Fatal("IsJPEG")
	}
}
