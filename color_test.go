package check //nolint:testpackage // Testing unexported identifiers.

import (
	"testing"
)

func TestWantColor(tt *testing.T) {
	tt.Parallel()

	// env is a convenience builder for getenv closures in test cases.
	env := func(kv ...string) func(string) string {
		return func(key string) string {
			for i := 0; i < len(kv); i += 2 {
				if kv[i] == key {
					return kv[i+1]
				}
			}
			return ""
		}
	}

	cases := []struct {
		name   string
		getenv func(string) string
		isTTY  bool
		want   bool
	}{
		// --- NO_COLOR wins over everything ---
		{
			name:   "NO_COLOR alone",
			getenv: env("NO_COLOR", "1"),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "NO_COLOR overrides CLICOLOR_FORCE",
			getenv: env("NO_COLOR", "1", "CLICOLOR_FORCE", "1"),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "NO_COLOR overrides FORCE_COLOR",
			getenv: env("NO_COLOR", "1", "FORCE_COLOR", "1"),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "NO_COLOR overrides GO_TEST_COLOR",
			getenv: env("NO_COLOR", "1", "GO_TEST_COLOR", "1"),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "NO_COLOR overrides tty",
			getenv: env("NO_COLOR", "1"),
			isTTY:  true,
			want:   false,
		},

		// --- CLICOLOR_FORCE ---
		{
			name:   "CLICOLOR_FORCE=1 enables color",
			getenv: env("CLICOLOR_FORCE", "1"),
			isTTY:  false,
			want:   true,
		},
		{
			name:   "CLICOLOR_FORCE=0 does not enable color",
			getenv: env("CLICOLOR_FORCE", "0"),
			isTTY:  false,
			want:   false,
		},
		{
			name:   "CLICOLOR_FORCE=1 without tty but with TERM enables color",
			getenv: env("CLICOLOR_FORCE", "1", "TERM", "xterm-256color"),
			isTTY:  false,
			want:   true,
		},

		// --- FORCE_COLOR ---
		{
			name:   "FORCE_COLOR=1 enables color",
			getenv: env("FORCE_COLOR", "1"),
			isTTY:  false,
			want:   true,
		},
		{
			name:   "FORCE_COLOR=0 does not enable color",
			getenv: env("FORCE_COLOR", "0"),
			isTTY:  false,
			want:   false,
		},
		{
			name:   "FORCE_COLOR=2 enables color (any non-zero)",
			getenv: env("FORCE_COLOR", "2"),
			isTTY:  false,
			want:   true,
		},
		{
			name:   "FORCE_COLOR wins over CLICOLOR_FORCE=0",
			getenv: env("FORCE_COLOR", "1", "CLICOLOR_FORCE", "0"),
			isTTY:  false,
			want:   true,
		},

		// --- GO_TEST_COLOR (legacy) ---
		{
			name:   "GO_TEST_COLOR=1 enables color",
			getenv: env("GO_TEST_COLOR", "1"),
			isTTY:  false,
			want:   true,
		},
		{
			name:   "GO_TEST_COLOR=0 enables color (legacy accepts zero)",
			getenv: env("GO_TEST_COLOR", "0"),
			isTTY:  false,
			want:   true,
		},
		{
			name:   "GO_TEST_COLOR=1 enables color even when FORCE_COLOR=0",
			getenv: env("GO_TEST_COLOR", "1", "FORCE_COLOR", "0"),
			isTTY:  false,
			want:   true, // FORCE_COLOR=0 is treated as "not set"; GO_TEST_COLOR still works
		},

		// --- Terminal auto-detection ---
		{
			name:   "tty with TERM=xterm-256color",
			getenv: env("TERM", "xterm-256color"),
			isTTY:  true,
			want:   true,
		},
		{
			name:   "tty with TERM=alacritty",
			getenv: env("TERM", "alacritty"),
			isTTY:  true,
			want:   true,
		},
		{
			name:   "tty with TERM=foot",
			getenv: env("TERM", "foot"),
			isTTY:  true,
			want:   true,
		},
		{
			name:   "not tty even with good TERM",
			getenv: env("TERM", "xterm-256color"),
			isTTY:  false,
			want:   false,
		},
		{
			name:   "TERM=dumb disables color",
			getenv: env("TERM", "dumb"),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "empty TERM disables color",
			getenv: env("TERM", ""),
			isTTY:  true,
			want:   false,
		},
		{
			name:   "no TERM and not tty",
			getenv: env(),
			isTTY:  false,
			want:   false,
		},
	}

	for _, v := range cases {
		tt.Run(v.name, func(tt *testing.T) {
			tt.Parallel()
			t := T(tt)
			t.Equal(wantColor(v.getenv, v.isTTY), v.want)
		})
	}
}
