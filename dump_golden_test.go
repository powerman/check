package check //nolint:testpackage // Testing unexported identifiers.

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update golden files")

type nested struct {
	Name    string
	Values  map[string]any
	Items   []int
	Pointer *string
}

// TestDumpGolden captures current dump/diff output as golden data
// and checks it remains byte-identical after vendoring.
func TestDumpGolden(tt *testing.T) {
	tt.Parallel()
	t := T(tt)

	s := "world"
	j := json.RawMessage(`{"key":"value"}`)

	vals := map[string]any{
		"nil":               nil,
		"byte":              byte(10),
		"rune":              '€',
		"multilineString":   "line1\nline2\nline3",
		"invalidUtf8Str":    "\x01\x02\x03\n\xff\xff",
		"utf8Bytes":         []byte("hello €"),
		"invalidUtf8Bytes":  []byte{0xff, 0xfe, 0x00},
		"jsonRawMessage":    j,
		"jsonRawMessagePtr": &j,
		"timeTime":          time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC),
		"nestedStruct": nested{
			Name:    "test",
			Values:  map[string]any{"a": 1, "b": true, "c": "hi"},
			Items:   []int{1, 2, 3},
			Pointer: &s,
		},
	}

	names := []string{
		"nil", "byte", "rune", "multilineString", "invalidUtf8Str",
		"utf8Bytes", "invalidUtf8Bytes", "jsonRawMessage",
		"jsonRawMessagePtr", "timeTime", "nestedStruct",
	}

	var b strings.Builder
	for _, name := range names {
		b.WriteString("=== newDump: ")
		b.WriteString(name)
		b.WriteString(" ===\n")
		b.WriteString(newDump(vals[name]).String())
	}

	// Also test diff between two similar but different values
	dumpA := newDump(nested{Name: "a", Items: []int{1, 2}})
	dumpB := newDump(nested{Name: "b", Items: []int{1, 3}})
	b.WriteString("=== diff between two nested structs ===\n")
	b.WriteString(dumpA.diff(dumpB))
	got := b.String()

	golden := filepath.Join("testdata", "dump_golden.txt")

	if *update {
		err := os.WriteFile(golden, []byte(got), 0o644)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	t.Equal(got, string(want))
}
