package contextx_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/powerman/check"
	"github.com/powerman/check/internal/contextx"
)

type ctxKey string

var errCause = errors.New("cause")

// closed reports whether ctx becomes done within a generous ceiling. The
// ceiling is never actually waited out when the test is correct: closed returns
// as soon as ctx is done.
func closed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(time.Second):
		return false
	}
}

// open reports whether ctx stays not-done for a brief window. It is used to
// assert the absence of a cancellation path, so a short window is sufficient.
func open(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Millisecond):
		return true
	}
}

func TestMergeValues(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	parent := context.WithValue(context.Background(), ctxKey("a"), "parent-a")
	parent = context.WithValue(parent, ctxKey("shared"), "parent-shared")
	extra := context.WithValue(context.Background(), ctxKey("b"), "extra-b")
	extra = context.WithValue(extra, ctxKey("shared"), "extra-shared")

	ctx := contextx.MergeValues(parent, extra)

	t.Equal(ctx.Value(ctxKey("a")), "parent-a")           // from parent
	t.Equal(ctx.Value(ctxKey("b")), "extra-b")            // fallback to extra
	t.Equal(ctx.Value(ctxKey("shared")), "parent-shared") // parent wins
	t.Nil(ctx.Value(ctxKey("missing")))
}

func TestMergeValues_cancellation_from_parent_only(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	parentCtx, cancelParent := context.WithCancel(context.Background())
	tt.Cleanup(cancelParent)
	extraCtx, cancelExtra := context.WithCancel(context.Background())
	tt.Cleanup(cancelExtra)

	ctx := contextx.MergeValues(parentCtx, extraCtx)

	cancelExtra()
	t.True(open(ctx)) // extra cancellation is not wired into a value merge

	cancelParent()
	t.True(closed(ctx))
	t.Err(ctx.Err(), context.Canceled)
}

func TestMergeCancel_via_parent(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	parent, cancelParent := context.WithCancel(context.Background())
	tt.Cleanup(cancelParent)

	// extra is non-cancellable, exercising AfterFunc's no-registration path.
	ctx, cancel := contextx.MergeCancel(parent, context.Background())
	tt.Cleanup(cancel)

	t.True(open(ctx))
	cancelParent()
	t.True(closed(ctx))
	t.Err(ctx.Err(), context.Canceled)
}

func TestMergeCancel_via_extra(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	extra, cancelExtra := context.WithCancelCause(context.Background())
	tt.Cleanup(func() { cancelExtra(context.Canceled) })

	ctx, cancel := contextx.MergeCancel(context.Background(), extra)
	tt.Cleanup(cancel)

	t.True(open(ctx))
	cancelExtra(errCause)
	t.True(closed(ctx))
	t.Err(ctx.Err(), context.Canceled)  // Err stays the canonical Canceled
	t.Err(context.Cause(ctx), errCause) // cause propagates from extra
}

// TestMergeCancel_via_func is the regression guard for the old hand-written
// implementation, whose returned cancel was a no-op when extra was
// non-cancellable. Here both contexts are non-cancellable, so only the returned
// cancel can close the merged context.
func TestMergeCancel_via_func(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	ctx, cancel := contextx.MergeCancel(context.Background(), context.Background())

	t.True(open(ctx))
	cancel()
	t.True(closed(ctx))
	t.Err(ctx.Err(), context.Canceled)
}

func TestMergeCancel_values_from_parent_only(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	parent := context.WithValue(context.Background(), ctxKey("a"), "parent-a")
	extra := context.WithValue(context.Background(), ctxKey("b"), "extra-b")

	ctx, cancel := contextx.MergeCancel(parent, extra)
	tt.Cleanup(cancel)

	t.Equal(ctx.Value(ctxKey("a")), "parent-a")
	t.Nil(ctx.Value(ctxKey("b"))) // extra contributes cancellation, not values
}

func TestMergeCancel_deadline(tt *testing.T) {
	tt.Parallel()

	base := time.Now()
	early := base.Add(time.Minute)
	late := base.Add(time.Hour)

	at := func(tt *testing.T, deadline time.Time) context.Context {
		tt.Helper()
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		tt.Cleanup(cancel)
		return ctx
	}
	mk := func(tt *testing.T, deadline *time.Time) context.Context {
		tt.Helper()
		if deadline == nil {
			return context.Background()
		}
		return at(tt, *deadline)
	}

	testCases := []struct {
		name          string
		parent, extra *time.Time
		want          *time.Time
	}{
		{"neither", nil, nil, nil},
		{"parent-only", &early, nil, &early},
		{"extra-only", nil, &early, &early},
		{"parent-earlier", &early, &late, &early},
		{"extra-earlier", &late, &early, &early},
	}
	for _, tc := range testCases {
		tt.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()
			t := check.T(tt)

			ctx, cancel := contextx.MergeCancel(mk(tt, tc.parent), mk(tt, tc.extra))
			tt.Cleanup(cancel)

			deadline, ok := ctx.Deadline()
			t.Equal(ok, tc.want != nil)
			if tc.want != nil {
				t.True(deadline.Equal(*tc.want))
			}
		})
	}
}

func TestMerge(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	parent := context.WithValue(context.Background(), ctxKey("a"), "parent-a")
	parent = context.WithValue(parent, ctxKey("shared"), "parent-shared")
	extra := context.WithValue(context.Background(), ctxKey("b"), "extra-b")
	extra = context.WithValue(extra, ctxKey("shared"), "extra-shared")

	parentCtx, cancelParent := context.WithCancel(parent)
	tt.Cleanup(cancelParent)
	extraCtx, cancelExtra := context.WithCancel(extra)
	tt.Cleanup(cancelExtra)

	ctx, cancel := contextx.Merge(parentCtx, extraCtx)
	tt.Cleanup(cancel)

	// Values come from both, with parent taking precedence on shared keys.
	t.Equal(ctx.Value(ctxKey("a")), "parent-a")
	t.Equal(ctx.Value(ctxKey("b")), "extra-b")
	t.Equal(ctx.Value(ctxKey("shared")), "parent-shared")

	// Cancelling extra cancels the merged context.
	t.True(open(ctx))
	cancelExtra()
	t.True(closed(ctx))
	t.Err(ctx.Err(), context.Canceled)
}
