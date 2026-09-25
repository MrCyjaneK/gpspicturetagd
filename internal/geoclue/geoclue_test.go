package geoclue

import "testing"

func TestBetter(t *testing.T) {
	lo := Fix{Level: 1, Horizontal: 100}
	hi := Fix{Level: 3, Horizontal: 50}
	if !Better(hi, lo) || Better(lo, hi) {
		t.Fatal("level should win")
	}
	near := Fix{Level: 3, Horizontal: 10}
	far := Fix{Level: 3, Horizontal: 80}
	if !Better(near, far) || Better(far, near) {
		t.Fatal("tighter accuracy should win at same level")
	}
}
