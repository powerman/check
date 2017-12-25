package check

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
)

type counter struct {
	name   string
	passed int
	forged int
	failed int
	force  bool
}

func (c counter) report() {
	var passed, forged, failed string
	if c.passed != 0 || c.force {
		passed = fmt.Sprintf("%s%3d passed%s", ansiGreen, c.passed, ansiReset)
	}
	if c.forged != 0 || c.force {
		color := ansiYellow
		if c.forged == 0 {
			color = ansiReset
		}
		forged = fmt.Sprintf("%s%2d todo%s", color, c.forged, ansiReset)
	}
	if c.failed != 0 || c.force {
		color := ansiRed
		if c.failed == 0 {
			color = ansiReset
		}
		failed = fmt.Sprintf("%s%3d failed%s", color, c.failed, ansiReset)
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
	}
	sort.Slice(ts, func(a, b int) bool { return ts[a].Name() < ts[b].Name() })
	for _, t := range ts {
		total.passed += stats.counter[t].passed
		total.forged += stats.counter[t].forged
		total.failed += stats.counter[t].failed
		if testing.Verbose() {
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
