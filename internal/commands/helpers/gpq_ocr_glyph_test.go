package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewGlyphsDecodedFromPixels verifies the added templates (è, š) are
// matched directly by the pixel decoder, independent of name reconciliation.
// An empty member roster forces reconcileName to return the literal decode.
func TestNewGlyphsDecodedFromPixels(t *testing.T) {
	font, err := LoadGPQFont()
	if err != nil {
		t.Fatalf("LoadGPQFont: %v", err)
	}
	cases := []struct {
		file string
		want string // substring the raw pixel decode must contain
	}{
		{"2.png", "Kagètsu"},
		{"6.png", "Mišs"},
	}
	for _, c := range cases {
		data, err := os.ReadFile(filepath.Join(gpqTestsDir, c.file))
		if err != nil {
			t.Fatalf("read %s: %v", c.file, err)
		}
		got, err := ParseSmallImage(data, nil, font)
		if err != nil {
			t.Fatalf("ParseSmallImage %s: %v", c.file, err)
		}
		found := false
		for name := range got {
			if strings.Contains(name, c.want) {
				found = true
				break
			}
		}
		if !found {
			keys := make([]string, 0, len(got))
			for k := range got {
				keys = append(keys, k)
			}
			t.Errorf("%s: raw pixel decode did not contain %q; got keys=%v", c.file, c.want, keys)
		}
	}
}
