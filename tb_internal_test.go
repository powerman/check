package check // Constructs *C around a fake testing.TB; needs the unexported checks field.

import (
	"fmt"
	"regexp"
	"testing"
)

// fakeReportTB is a minimal testing.TB double that only captures Errorf
// calls, used to inspect the exact "Checker:" text check reports.
type fakeReportTB struct {
	testing.TB

	msgs []string
}

func (*fakeReportTB) Helper()      {}
func (*fakeReportTB) Name() string { return "fakeReportTB" }
func (*fakeReportTB) FailNow()     {}
func (f *fakeReportTB) Errorf(format string, args ...any) {
	f.msgs = append(f.msgs, ansiTestRE.ReplaceAllString(fmt.Sprintf(format, args...), ""))
}

var ansiTestRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

// TestCheckerNameInReportViaC is the *C half of the callerFuncName
// regression test (see TestCheckerNameInReport in tb_test.go for the *TB
// half): Error/Errorf are explicit overrides on *C (needed to win the
// ambiguity with *testing.T's own methods), so they must also report
// their bare method name, not "(*C).Error" and friends. *C embeds a
// concrete *testing.T (not testing.TB), so this can't be built with a
// fake from outside the package - hence this being an internal test.
//
// Fatal/Fatalf aren't exercised here: unlike TB.Fatal (whose FailNow goes
// through the wrapped testing.TB, which can be a fake), C.Fatal always
// calls the real, embedded *testing.T's FailNow - not safe to trigger
// from within this test.
func TestCheckerNameInReportViaC(tt *testing.T) {
	tt.Parallel()

	fake := &fakeReportTB{}
	c := &C{checks: &checks{tb: fake}, T: tt}
	c.Error("direct error")
	c.Errorf("direct %s", "errorf")

	t := T(tt)
	t.Equal(len(fake.msgs), 2)
	t.Match(fake.msgs[0], `Checker:  Error\n`)
	t.Match(fake.msgs[1], `Checker:  Errorf\n`)
}
