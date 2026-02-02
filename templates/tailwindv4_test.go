package templates

import "testing"

func TestTailwindV4_IsDefined(t *testing.T) {
	// This is a tiny compile-time guard so editors/linters don't flag TailwindV4
	// as unused inside the templates package.
	if TailwindV4 == nil {
		t.Fatal("TailwindV4 template map is nil")
	}
}
