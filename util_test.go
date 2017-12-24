package check

import (
	"regexp"
	"testing"
)

func TestFormat(tt *testing.T) {
	t := T{tt}
	cases := []struct {
		args []interface{}
		want string
	}{
		{[]interface{}{}, ""},
		{[]interface{}{"msg"}, "msg"},
		{[]interface{}{"%s", "msg"}, "msg"},
		{[]interface{}{"one", "two"}, "one%!(EXTRA string=two)"},
		{[]interface{}{42}, "42"},
		{[]interface{}{regexp.MustCompile(".*")}, ".*"},
	}
	for i, v := range cases {
		t.Equal(format(v.args...), v.want, i)
	}
}

func TestCaller(tt *testing.T) {
	t := T{tt}
	t.Equal(callerFuncName(0), "TestCaller")
	t.Equal(callerFuncName(1000), "<unknown>")
}
