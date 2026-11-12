package check // Inspects the unexported stats map.

import "testing"

// namedTB is a minimal testing.TB double which only provides a name.
type namedTB struct {
	testing.TB

	name string
}

func (f *namedTB) Name() string { return f.name }

// Fuzzing runs every input with a new *testing.T with the same name:
// they must share a single stats entry instead of adding one per *testing.T.
func TestChecks_stat_sameName(tt *testing.T) {
	tt.Parallel()
	t := T(tt)

	const name = "TestChecks_stat_sameName/fake"
	for range 3 {
		c := &checks{tb: &namedTB{name: name}}
		c.pass()
		c.withTODO().pass()
		c.fail()
	}

	statsMu.Lock()
	stat := *stats[name]
	statsMu.Unlock()
	t.Equal(stat.passed.value, 3)
	t.Equal(stat.forged.value, 3)
	t.Equal(stat.failed.value, 3)
}
