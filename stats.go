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
	failed int
	force  bool
}

func (c counter) report() {
	var passed, failed string
	if c.passed != 0 || c.force {
		passed = fmt.Sprintf("%3d passed", c.passed)
	}
	if c.failed != 0 || c.force {
		failed = fmt.Sprintf("%3d failed", c.failed)
	}
	fmt.Printf("  checks:  %10s  %10s\t%s\n", passed, failed, c.name)
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

func pass(t *testing.T) bool {
	stats.Lock()
	defer stats.Unlock()

	if stats.counter[t] == nil {
		stats.counter[t] = &counter{name: t.Name()}
	}
	stats.counter[t].passed++
	return true
}

func fail(t *testing.T) bool {
	stats.Lock()
	defer stats.Unlock()

	if stats.counter[t] == nil {
		stats.counter[t] = &counter{name: t.Name()}
	}
	stats.counter[t].failed++
	return false
}
