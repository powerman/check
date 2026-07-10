//nolint:testableexamples // No way to obtain a real *testing.T outside of `go test`.
package check_test

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/powerman/check"
)

// ExampleMust shows the recommended way to wrap a *testing.T:
// any failed check stops the test immediately, like testify/require.
func ExampleMust() {
	tt := new(testing.T)
	tt.Parallel()
	t := check.Must(tt)

	t.Equal(2+2, 4)
	t.Match("build-42", `^build-\d+$`)
}

// ExampleMust_tableDriven shows the usual table-driven pattern:
// wrap each subtest's own *testing.T inside tt.Run, not the outer one.
func ExampleMust_tableDriven() {
	tt := new(testing.T)
	tt.Parallel()
	t := check.Must(tt)
	t.True(true, "outer setup check")

	cases := []struct {
		name string
		got  int
		want int
	}{
		{"add one", 1 + 1, 2},
		{"add two", 2 + 2, 4},
	}
	for _, c := range cases {
		tt.Run(c.name, func(tt *testing.T) {
			tt.Parallel()
			t := check.Must(tt)
			t.Equal(c.got, c.want)
		})
	}
}

// ExampleNew shows the softer, testify/assert-like alternative to [Must]:
// a failed check doesn't stop the test,
// so guard dependent checks with the bool every checker returns.
func ExampleNew() {
	tt := new(testing.T)
	t := check.New(tt)

	obj, err := os.Open(os.DevNull)
	if t.Nil(err) {
		t.NotNil(obj)
		_ = obj.Close()
	}
}

// ExampleTB_TODO marks a known-broken check as expected-to-fail
// without disabling or deleting the test: it keeps running,
// and once the underlying defect is fixed this check starts failing again -
// a reminder to remove TODO.
func ExampleTB_TODO() {
	tt := new(testing.T)
	t := check.Must(tt)

	t.TODO().Equal(2+2, 5)
}

// ExampleTB_Should plugs a custom checker (bePositive, defined in check_test.go)
// into check's usual report/Must/TODO machinery,
// for checks not covered by any built-in checker.
func ExampleTB_Should() {
	tt := new(testing.T)
	t := check.Must(tt)

	t.Should(bePositive, 42, "custom check")
}

// Example_errorChecks contrasts the four ways to check an error,
// see package doc for when to use which.
func Example_errorChecks() {
	tt := new(testing.T)
	t := check.Must(tt)

	wrapped := fmt.Errorf("wrap: %w", io.EOF)

	// Err: unwraps to the root cause and compares by value.
	t.Err(wrapped, io.EOF)

	// ErrIs: pure errors.Is chain membership, no value comparison.
	t.ErrIs(wrapped, io.EOF)

	// ErrAs: extract the first error of a given type from the chain.
	var pathErr *fs.PathError
	if t.ErrAs(wrapped, &pathErr) {
		t.NotNil(pathErr)
	}

	// Match: check by error text against a regexp.
	t.Match(wrapped, `EOF$`)
}

// ExampleTB_MergeContext injects an application base context
// (e.g. one carrying a slog handler) into a test
// on top of the per-test cancellation/deadline [testing.TB.Context] already provides.
func ExampleTB_MergeContext() {
	tt := new(testing.T)
	t := check.Must(tt)

	type slogHandlerKey struct{}
	appCtx := context.WithValue(context.Background(), slogHandlerKey{}, "app-handler")

	t = t.MergeContext(appCtx)
	t.NotNil(t.Context().Value(slogHandlerKey{}))
}

// ExampleTB_SortEqual checks that two slices/arrays contain the same elements
// while ignoring their order.
func ExampleTB_SortEqual() {
	tt := new(testing.T)
	t := check.Must(tt)

	t.SortEqual([]int{1, 2, 3}, []int{3, 1, 2})
}
