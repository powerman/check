//go:build demo

// Package demo holds a deliberately failing test used to capture check's
// failure output (dump + diff) for the README. It is excluded from the
// normal build/test/lint by both the "demo" build tag and living under
// testdata/ (which "go build ./..." and golangci-lint skip by convention).
//
// Run it directly to see (and re-capture) the output:
//
//	go test -tags demo -v ./testdata/demo/
package demo

import (
	"testing"

	"github.com/powerman/check"
)

// TestMain enables the pass/fail/todo statistics line, so the demo output
// also shows what check.TestMain adds beyond plain go test.
func TestMain(m *testing.M) { check.TestMain(m) }

// Order is a stand-in for a real domain type. Customer/Items match between
// got/want on purpose - the point of the demo is that the Diff singles out
// the one field that's actually wrong instead of forcing you to eyeball two
// full dumps for what changed.
type Order struct {
	ID       string
	Customer string
	Total    int
	Items    []string
}

func TestDemoFailure(tt *testing.T) {
	t := check.New(tt)

	got := Order{ID: "A-1", Customer: "Ann", Total: 42, Items: []string{"pen", "cup"}}
	want := Order{ID: "A-1", Customer: "Ann", Total: 40, Items: []string{"pen", "cup"}}

	t.Equal(got.ID, want.ID, "id should be unchanged") // Passes.
	t.DeepEqual(got, want, "order total should match after checkout")
}
