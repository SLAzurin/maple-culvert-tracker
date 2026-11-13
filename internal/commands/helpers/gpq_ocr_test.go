package helpers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

const gpqTestsDir = "../../../provided/gpq-tests"

// orderedKeys returns the object keys of a JSON object in their file order
// (encoding/json's map decoding does not preserve order).
func orderedKeys(t *testing.T, data []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil { // opening '{'
		t.Fatalf("json token: %v", err)
	}
	keys := []string{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("json key token: %v", err)
		}
		keys = append(keys, tok.(string))
		var v any // consume the value
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("json value decode: %v", err)
		}
	}
	return keys
}

// Render complete tables from the embedded font so score-column selection is
// exercised independently of the supplied screenshots' score distributions.
func TestParseSmallImageScoreColumn(t *testing.T) {
	font, err := LoadGPQFont()
	if err != nil {
		t.Fatal(err)
	}
	for _, score := range []int{0, 1, 42, 999, 123456} {
		t.Run(strconv.Itoa(score), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 620, 30))
			columns := []struct {
				x    int
				text string
			}{
				{4, "Alice"}, {110, "Hero"}, {220, "295"},
				{280, "Paper"}, {390, "10"}, {440, strconv.Itoa(score)}, {540, "1000"},
			}
			for _, column := range columns {
				x := column.x
				for _, r := range column.text {
					var glyph *gpqGlyph
					for i := range font.glyphs {
						if font.glyphs[i].r == r {
							glyph = &font.glyphs[i]
							break
						}
					}
					if glyph == nil {
						t.Fatalf("missing glyph %q", r)
					}
					for y, row := range glyph.bits {
						for dx, ink := range row {
							if ink {
								img.Set(x+dx, 8+y, color.White)
							}
						}
					}
					x += glyph.w + 1
				}
			}
			var data bytes.Buffer
			if err := png.Encode(&data, img); err != nil {
				t.Fatal(err)
			}
			got, err := ParseSmallImage(data.Bytes(), []string{"Alice"}, font)
			if err != nil {
				t.Fatal(err)
			}
			want := []ScoreEntry{}
			if score > 0 {
				want = append(want, ScoreEntry{Name: "Alice", Score: score})
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("entries = %#v, want %#v", got, want)
			}
		})
	}
}

func TestCleanDecodedName(t *testing.T) {
	for _, tc := range []struct{ decoded, want string }{
		{"Subzero(Subzero)", "Subzero"},
		{"FlamingJay..(FlamingJay)", "FlamingJay"},
		{"Kagètsu", "Kagètsu"},
		{"ShiningRo..", "ShiningRo"},
	} {
		t.Run(tc.decoded, func(t *testing.T) {
			if got := cleanDecodedName(tc.decoded); got != tc.want {
				t.Errorf("cleanDecodedName(%q) = %q, want %q", tc.decoded, got, tc.want)
			}
		})
	}
}

func TestParseSmallImageAgainstProvided(t *testing.T) {
	font, err := LoadGPQFont()
	if err != nil {
		t.Fatalf("LoadGPQFont: %v", err)
	}

	cases, err := filepath.Glob(filepath.Join(gpqTestsDir, "case-*"))
	if err != nil || len(cases) == 0 {
		t.Fatalf("find fixture cases: %v (found %d)", err, len(cases))
	}
	for _, caseDir := range cases {
		t.Run(filepath.Base(caseDir), func(t *testing.T) {
			roster, err := os.ReadFile(filepath.Join(caseDir, "members.json"))
			if err != nil {
				t.Fatal(err)
			}
			var members []string
			if err := json.Unmarshal(roster, &members); err != nil {
				t.Fatal(err)
			}
			images, err := filepath.Glob(filepath.Join(caseDir, "*.png"))
			if err != nil || len(images) == 0 {
				t.Fatalf("find fixture images: %v (found %d)", err, len(images))
			}
			for _, imagePath := range images {
				t.Run(filepath.Base(imagePath), func(t *testing.T) {
					imgData, err := os.ReadFile(imagePath)
					if err != nil {
						t.Fatal(err)
					}
					rawJSON, err := os.ReadFile(imagePath[:len(imagePath)-len(".png")] + ".json")
					if err != nil {
						t.Fatal(err)
					}
					var scores map[string]int
					if err := json.Unmarshal(rawJSON, &scores); err != nil {
						t.Fatal(err)
					}
					// Compare entries directly to verify scores, order, and duplicates.
					want := []ScoreEntry{}
					for _, name := range orderedKeys(t, rawJSON) {
						want = append(want, ScoreEntry{Name: name, Score: scores[name]})
					}
					got, err := ParseSmallImage(imgData, members, font)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("entries:\n got: %#v\nwant: %#v", got, want)
					}
				})
			}
		})
	}
}
