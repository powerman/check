package check

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	_ "github.com/smartystreets/goconvey/convey"
)

func TestDump(tt *testing.T) {
	t := T(tt)

	type (
		myBool    bool
		myInt     int
		myInt32   int32
		myInt64   int64
		myUint    uint
		myUint64  uint64
		myByte    byte
		myRune    rune
		myUintptr uintptr
		myFloat64 float64
		myString  string
		myRunes   []rune
		myBytes   []byte
		myStruct  struct {
			i int
			s string
		}
	)
	var (
		j    = json.RawMessage(`[{"key":"one","value":1},{"key":"two","value":2}]`)
		jnil *json.RawMessage
	)

	cases := []struct {
		improved bool
		i        interface{}
	}{
		{true, nil},
		{false, true},
		{false, myBool(true)},
		{false, -42},
		{false, myInt(-42)},
		{false, int32(-32)},
		{false, myInt32(-32)},
		{false, int64(-64)},
		{false, myInt64(-64)},
		{false, uint(42)},
		{false, myUint(42)},
		{false, uint64(64)},
		{false, myUint64(64)},
		{true, byte(10)},
		{true, myByte(10)},
		{true, byte(255)},
		{true, myByte(255)},
		{true, rune(0)},
		{true, myRune(0)},
		{true, ' '},
		{true, myRune(' ')},
		{true, ' '},
		{true, myRune(' ')},
		{true, '\n'},
		{true, myRune('\n')},
		{true, '€'},
		{true, myRune('€')},
		{false, uintptr(0)},
		{false, myUintptr(0)},
		{false, uintptr(42)},
		{false, myUintptr(42)},
		{false, 0.0},
		{false, myFloat64(0.0)},
		{false, time.Monday},
		{false, [0]int{}},
		{false, [2]int{}},
		{false, []int(nil)},
		{false, []int{}},
		{false, []int{1: 0}},
		{false, chan int(nil)},
		{false, make(chan int)},
		{false, (chan<- int)(make(chan int, 2))},
		{false, (func())(nil)},
		{false, func(i int) int { return 0 }},
		{false, io.EOF},
		{false, map[int]int(nil)},
		{false, map[int]int{2: 0}},
		{false, make(map[int]int, 2)},
		{false, (*int)(nil)},
		{true, ""},
		{true, myString("")},
		{true, " "},
		{true, myString(" ")},
		{true, "\\`'\""},
		{true, myString("\\`'\"")},
		{true, "€"},
		{true, myString("€")},
		{true, "\x01\x02\x03\n\xff\xff"},
		{true, myString("\x01\x02\x03\n\xff\xff")},
		{true, "line1\nline2"},
		{true, myString("line1\nline2")},
		{false, []byte(nil)},
		{false, myBytes(nil)},
		{true, []byte{}},
		{true, myBytes{}},
		{false, []byte("\x01\x02\x03\n\xff\xff")},
		{false, myBytes("\x01\x02\x03\n\xff\xff")},
		{true, []byte("line1\nvery long line2")},
		{true, myBytes("line1\nvery long line2")},
		{true, j},
		{true, myBytes(j)},
		{false, jnil},
		{true, &j},
		{true, []rune{}},
		{true, myRunes{}},
		{true, []rune{0, ' ', ' ', '\n', '€'}},
		{true, myRunes{0, ' ', ' ', '\n', '€'}},
		{false, time.Time{}},
		{false, time.Now()},
		{false, struct {
			i int
			s string
		}{0, ""}},
		{false, myStruct{0, ""}},
	}

	for _, v := range cases {
		old, new := spewCfg.Sdump(v.i), newDump(v.i).String()
		if v.improved {
			t.NotEqual(new, old)
		} else {
			t.Equal(new, old)
		}
	}
}
