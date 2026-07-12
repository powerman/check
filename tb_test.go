package check_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/powerman/check"
)

var errBoom = errors.New("boom")

// ansiRE strips ANSI color codes, since check colors the "Checker:" line's
// checker name when GO_TEST_COLOR/a color terminal makes check emit them.
var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

// fakeTB is a minimal [testing.TB] double used to observe how check reacts
// to a failed check, without making a real (sub)test fail: unlike a real
// *testing.T, it just records what happened instead of also propagating
// failure up to the actual test binary's exit code.
type fakeTB struct {
	testing.TB

	name          string
	errorfCalls   int
	msgs          []string
	failNowCalled bool
}

func (*fakeTB) Helper()        {}
func (f *fakeTB) Name() string { return f.name }

func (f *fakeTB) Errorf(format string, args ...any) {
	f.errorfCalls++
	f.msgs = append(f.msgs, ansiRE.ReplaceAllString(fmt.Sprintf(format, args...), ""))
}

func (f *fakeTB) FailNow() { f.failNowCalled = true }

func TestNewSoft(tt *testing.T) {
	tt.Parallel()
	fake := &fakeTB{name: "fakeTestNewSoft"}
	t := check.New(fake)
	var reachedAfter bool
	t.Nil(errBoom) // Fails, but New is soft: execution continues.
	reachedAfter = true

	realT := check.T(tt)
	realT.Equal(fake.errorfCalls, 1)
	realT.False(fake.failNowCalled)
	realT.True(reachedAfter)
}

func TestMustConstructor(tt *testing.T) {
	tt.Parallel()
	fake := &fakeTB{name: "fakeTestMustConstructor"}
	t := check.Must(fake)
	t.Nil(errBoom) // Fails: Must stops the test right here (fake.FailNow is a no-op).

	realT := check.T(tt)
	realT.Equal(fake.errorfCalls, 1)
	realT.True(fake.failNowCalled)
}

func TestMustWorksWithBenchmark(tt *testing.T) {
	tt.Parallel()
	res := testing.Benchmark(BenchmarkSmoke)
	t := check.T(tt)
	t.True(res.N > 0)
}

func BenchmarkSmoke(b *testing.B) {
	t := check.Must(b)
	for b.Loop() {
		t.Equal(1+1, 2)
	}
}

func FuzzSmoke(f *testing.F) {
	f.Add(2)
	f.Fuzz(func(tt *testing.T, n int) {
		t := check.Must(tt)
		t.Equal(n+0, n)
	})
}

func TestTBTODOMustAll(tt *testing.T) {
	tt.Parallel()
	t := check.Must(tt).TODO()
	t.True(false) // Swapped by TODO: counted as passed.

	tAll := check.New(tt).MustAll()
	tAll.Nil(nil)
	tAll.TODO().Nil(false)
}

func TestTBShould(tt *testing.T) {
	tt.Parallel()
	t := check.New(tt)
	t.Should(func(_ *check.TB, actual any) bool {
		return actual.(int) > 0 //nolint:forcetypeassert // Test-only.
	}, 42, "custom check via TB")
}

func TestMergeContext(tt *testing.T) {
	tt.Parallel()

	type keyA struct{}
	type keyB struct{}

	t := check.New(tt)
	base := context.WithValue(t.Context(), keyA{}, "from-base")

	merged := t.MergeContext(base)
	t.Equal(merged.Context().Value(keyA{}), "from-base")
	t.Nil(merged.Context().Value(keyB{}))

	// The original t is untouched.
	t.Nil(t.Context().Value(keyA{}))
}

func TestMergeContextValuePrecedence(tt *testing.T) {
	tt.Parallel()

	type key struct{}

	t := check.New(tt)
	tWithValue := t.MergeContext(context.WithValue(t.Context(), key{}, "current"))
	merged := tWithValue.MergeContext(context.WithValue(t.Context(), key{}, "new"))

	// Values are looked up in the newly merged ctx first.
	t.Equal(merged.Context().Value(key{}), "new")
}

func TestMergeContextCancellation(tt *testing.T) {
	tt.Parallel()
	t := check.New(tt)

	baseCtx, baseCancel := context.WithCancel(t.Context())
	tt.Cleanup(baseCancel)
	merged := t.MergeContext(baseCtx)

	select {
	case <-merged.Context().Done():
		t.Error("merged context should not be done yet")
	default:
	}

	baseCancel()

	select {
	case <-merged.Context().Done():
	case <-time.After(time.Second):
		t.Error("merged context should be cancelled after base is cancelled")
	}
	t.Err(merged.Context().Err(), context.Canceled)
}

// TestCheckerNameInReport is a regression test for the reported name of a
// failed checker in the "Checker:" line: it must be the bare method name
// (e.g. "Nil"), not a receiver-qualified one like "(*checks).Nil" - the
// checker methods live on the unexported *checks type (promoted into both
// TB and C), and Error/Errorf/Fatal/Fatalf additionally exist as explicit
// overrides on *TB itself, so callerFuncName must strip whichever
// "(*Receiver)." prefix is present, not just one hardcoded receiver name.
// Fatal/Fatalf are safe to exercise directly here because fakeTB's
// FailNow is a no-op (New(fake).TB is the fake, so TB.Fatal's
// t.TB.FailNow() never reaches a real *testing.T). See
// tb_internal_test.go for the *C-path (Error/Errorf) half of this
// regression, which needs package-internal access to construct a *C
// around a fake testing.TB - *C.Fatal/Fatalf always call the real,
// wrapped *testing.T.FailNow() (not the fake), so they aren't safe to
// exercise directly from a test.
func TestCheckerNameInReport(tt *testing.T) {
	tt.Parallel()

	fake := &fakeTB{name: "fakeCheckerNameInReport"}
	t := check.New(fake)
	t.Nil("not-nil-value") // Checker method, lives on *checks.
	t.Equal(1, 2)          // Checker method, lives on *checks.
	t.Error("direct error")
	t.Errorf("direct %s", "errorf")
	t.Fatal("direct fatal") //nolint:revive // fakeTB.FailNow is a no-op: execution continues.
	t.Fatalf("direct %s", "fatalf")

	realT := check.T(tt)
	realT.Equal(len(fake.msgs), 6)
	realT.Match(fake.msgs[0], `Checker:  Nil\n`)
	realT.Match(fake.msgs[1], `Checker:  Equal\n`)
	realT.Match(fake.msgs[2], `Checker:  Error\n`)
	realT.Match(fake.msgs[3], `Checker:  Errorf\n`)
	realT.Match(fake.msgs[4], `Checker:  Fatal\n`)
	realT.Match(fake.msgs[5], `Checker:  Fatalf\n`)
}
