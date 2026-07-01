package helpers

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
		var v interface{} // consume the value
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("json value decode: %v", err)
		}
	}
	return keys
}

// loadExpected reads every N.json and returns the parsed maps plus the union
// of all character names (used as the member roster for reconciliation).
func loadExpected(t *testing.T, n int) []map[string]int {
	t.Helper()
	expected := make([]map[string]int, n+1)
	for i := 1; i <= n; i++ {
		data, err := os.ReadFile(filepath.Join(gpqTestsDir, strconv.Itoa(i)+".json"))
		if err != nil {
			t.Fatalf("read %d.json: %v", i, err)
		}
		m := map[string]int{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal %d.json: %v", i, err)
		}
		expected[i] = m
	}
	return expected
}

func TestParseSmallImageAgainstProvided(t *testing.T) {
	const total = 12
	expected := loadExpected(t, total)

	memberSet := map[string]bool{}
	members := []string{}
	for i := 1; i <= total; i++ {
		for name := range expected[i] {
			if !memberSet[name] {
				memberSet[name] = true
				members = append(members, name)
			}
		}
	}

	font, err := LoadGPQFont()
	if err != nil {
		t.Fatalf("LoadGPQFont: %v", err)
	}

	for i := 1; i <= total; i++ {
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			imgData, err := os.ReadFile(filepath.Join(gpqTestsDir, strconv.Itoa(i)+".png"))
			if err != nil {
				t.Fatalf("read %d.png: %v", i, err)
			}
			entries, err := ParseSmallImage(imgData, members, font)
			if err != nil {
				t.Fatalf("ParseSmallImage %d.png: %v", i, err)
			}
			got := map[string]int{}
			for _, e := range entries {
				got[e.Name] = e.Score
			}
			want := expected[i]
			if len(got) != len(want) {
				t.Errorf("%d.png: got %d entries, want %d", i, len(got), len(want))
			}
			for name, score := range want {
				if got[name] != score {
					t.Errorf("%d.png: name %q got score %d, want %d", i, name, got[name], score)
				}
			}
			for name, score := range got {
				if _, ok := want[name]; !ok {
					t.Errorf("%d.png: unexpected name %q with score %d", i, name, score)
				}
			}

			// Row order must be preserved: parsed entry order should match the
			// key order in the expected JSON file (top-to-bottom leaderboard).
			rawJSON, err := os.ReadFile(filepath.Join(gpqTestsDir, strconv.Itoa(i)+".json"))
			if err != nil {
				t.Fatalf("read %d.json: %v", i, err)
			}
			wantOrder := orderedKeys(t, rawJSON)
			gotOrder := make([]string, len(entries))
			for j, e := range entries {
				gotOrder[j] = e.Name
			}
			if len(gotOrder) != len(wantOrder) {
				t.Fatalf("%d.png: got %d ordered entries, want %d", i, len(gotOrder), len(wantOrder))
			}
			for j := range wantOrder {
				if gotOrder[j] != wantOrder[j] {
					t.Errorf("%d.png: order mismatch at %d: got %q, want %q", i, j, gotOrder[j], wantOrder[j])
				}
			}
		})
	}
}
