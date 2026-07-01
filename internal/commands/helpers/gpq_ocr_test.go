package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const gpqTestsDir = "../../../provided/gpq-tests"

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
			got, err := ParseSmallImage(imgData, members, font)
			if err != nil {
				t.Fatalf("ParseSmallImage %d.png: %v", i, err)
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
		})
	}
}
