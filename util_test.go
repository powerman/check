package check //nolint:testpackage // Testing unexported identifiers.

import (
	"regexp"
	"testing"
)

func TestFormat(tt *testing.T) {
	t := T(tt)
	cases := []struct {
		args []any
		want string
	}{
		{[]any{}, ""},
		{[]any{"msg"}, "msg"},
		{[]any{"%s", "msg"}, "msg"},
		{[]any{"one", "two"}, "one%!(EXTRA string=two)"},
		{[]any{42}, "42"},
		{[]any{regexp.MustCompile(".*")}, ".*"},
	}
	for i, v := range cases {
		t.Equal(format(v.args...), v.want, i)
	}
}

func TestCaller(tt *testing.T) {
	t := T(tt)
	t.Equal(callerFuncName(0), "TestCaller")
	t.Equal(callerFuncName(1000), "<unknown>")
}
