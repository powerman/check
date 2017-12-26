package check

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
)

type counter struct {
	name         string
	passed       int
	forged       int
	failed       int
	force        bool
	passedDigits int
	forgedDigits int
	failedDigits int
}

func (c counter) report() {
	var passed, forged, failed string
	if c.passed != 0 || c.force {
		passed = fmt.Sprintf("%s%*d passed%s", ansiGreen, c.passedDigits, c.passed, ansiReset)
	}
	if c.forged != 0 || c.force {
		color := ansiYellow
		if c.forged == 0 {
			color = ansiReset
		}
		forged = fmt.Sprintf("%s%*d todo%s", color, c.forgedDigits, c.forged, ansiReset)
	}
	if c.failed != 0 || c.force {
		color := ansiRed
		if c.failed == 0 {
			color = ansiReset
		}
		failed = fmt.Sprintf("%s%*d failed%s", color, c.failedDigits, c.failed, ansiReset)
	}
	fmt.Printf("  checks:  %10s  %7s  %10s\t%s\n", passed, forged, failed, c.name)
}

var stats = struct {
	sync.Mutex
	counter map[*testing.T]*counter
}{sync.Mutex{}, map[*testing.T]*counter{}}

// Report output statistics about passed/failed checks.
// It should be called from TestMain after m.Run(), for ex.:
//
//	func TestMain(m *testing.M) {
//		code := m.Run()
//		check.Report()
//		os.Exit(code)
//	}
//
// If this is all you need - just use TestMain instead.
func Report() {
	stats.Lock()
	defer stats.Unlock()

	total := counter{name: "(total)", force: true}
	ts := make([]*testing.T, 0, len(stats.counter))
	for t := range stats.counter {
		ts = append(ts, t)
		total.passed += stats.counter[t].passed
		total.forged += stats.counter[t].forged
		total.failed += stats.counter[t].failed
	}

	total.passedDigits = digits(total.passed)
	total.forgedDigits = digits(total.forged)
	total.failedDigits = digits(total.failed)

	sort.Slice(ts, func(a, b int) bool { return ts[a].Name() < ts[b].Name() })
	for _, t := range ts {
		if testing.Verbose() {
			stats.counter[t].passedDigits = total.passedDigits
			stats.counter[t].forgedDigits = total.forgedDigits
			stats.counter[t].failedDigits = total.failedDigits
			stats.counter[t].report()
		}
	}
	total.report()
}

// TestMain provides same default implementation as used by testing
// package with extra Report call to output statistics. Usage:
//
//	func TestMain(m *testing.M) { check.TestMain(m) }
func TestMain(m *testing.M) {
	code := m.Run()
	Report()
	os.Exit(code)
}
