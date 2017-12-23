package check_test

import (
	"errors"
	"io"
	"net"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/powerman/check"
)

func TestCheckers(tt *testing.T) {
	t := check.T{tt}

	var intPtr *int
	var empty interface{}
	var notEmpty interface{} = intPtr
	t.Must(t.Nil(nil))
	t.Nil(intPtr)
	t.Nil(empty)
	t.Nil(notEmpty) // see doc about why it works this way
	t.Must(t.NotNil(false))
	t.NotNil(uintptr(0))

	t.True(intPtr == nil)
	t.True(empty == nil)
	t.False(notEmpty == nil)

	equal := []struct{ actual, expected interface{} }{
		{nil, nil},
		{true, true},
		{false, false},
		{0, 0},
		{3.14, 3.14},
		{"", ""},
		{"msg", "msg"},
		{t, t},
	}
	notEqual := []struct{ actual, expected interface{} }{
		{nil, true},
		{false, nil},
		{int32(0), int64(0)},
		{0, 0.0},
		{"", "msg"},
		{t, tt},
		{&testing.T{}, &testing.T{}},
		{io.EOF, errors.New("EOF")},
	}
	t.Run("Equal", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		for _, v := range equal {
			t.Equal(v.actual, v.expected)
		}
	})
	t.Run("EQ", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		for _, v := range equal {
			t.EQ(v.actual, v.expected)
		}
	})
	t.Run("NotEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		for _, v := range notEqual {
			t.NotEqual(v.actual, v.expected)
		}
	})
	t.Run("NE", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		for _, v := range notEqual {
			t.NE(v.actual, v.expected)
		}
	})

	t.Run("BytesEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.BytesEqual(nil, nil)
		t.BytesEqual([]byte(nil), nil)
		t.BytesEqual([]byte{}, nil)
		t.BytesEqual([]byte{}, []byte(nil))
		t.BytesEqual([]byte{0}, []byte{0})
	})
	t.Run("NotBytesEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotBytesEqual([]byte{0}, nil)
		t.NotBytesEqual([]byte{0}, []byte{})
		t.NotBytesEqual([]byte{0}, []byte{0, 0})
	})

	t.Run("DeepEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.DeepEqual(t, t)
		t.DeepEqual(tt, tt)
		t.DeepEqual(tt, t.T)
		t.DeepEqual(nil, nil)
		t.DeepEqual(42, 42)
		t.DeepEqual([]byte{2, 5}, []byte{2, 5})
		t.DeepEqual(&testing.T{}, &testing.T{})
		t.DeepEqual(io.EOF, errors.New("EOF"))
	})
	t.Run("NotDeepEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotDeepEqual(int64(42), int32(42))
		t.NotDeepEqual([]byte{}, []byte(nil))
		t.NotDeepEqual(t, tt)
	})

	t.Run("Match", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Match("", `^$`)
		t.Match("", `^.*$`)
		t.Match(" ", `^\s+$`)
		t.Match("World", `(?i)w`)
		t.Match("World", regexp.MustCompile(`(?i)w`))
		t.Match(io.ErrClosedPipe, "closed pipe")
		t.Match(time.Time{}, "00:00:00")
	})
	t.Run("NotMatch", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		var err error
		t.NotMatch(err, "some error")
		t.NotMatch(" ", `^$`)
		t.NotMatch("", `^\s+$`)
		t.NotMatch("World", `w`)
		t.NotMatch("World", regexp.MustCompile(`(?-i)w`))
		t.NotMatch(time.Time{}, "23:59:00")
	})

	t.Run("Contains", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Contains("something", "thing")
		t.Contains([]int{2, 4, 6}, 4)
		t.Contains([3]*testing.T{nil, tt, nil}, tt)
		t.Contains(map[*testing.T]int{nil: 2, tt: 5}, tt)
		t.Contains(map[string]int{"something": 2, "thing": 5}, "thing")
	})
	t.Run("NotContains", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotContains("something", "Thing")
		t.NotContains([]int{2, 4, 6}, 3)
		t.NotContains([3]*testing.T{nil, tt, nil}, t)
		t.NotContains(map[*testing.T]int{nil: 2, tt: 5}, &testing.T{})
		t.NotContains(map[string]int{"something": 2, "thing": 5}, "some")
	})

	t.Run("Zero", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Zero(nil)
		t.Zero(false)
		t.Zero(0)
		t.Zero(int16(0))
		t.Zero(uintptr(0))
		t.Zero(0.0)
		t.Zero("")
		t.Zero([2]int{})
		var ch chan int
		t.Zero(ch)
		var rw io.ReadWriter
		t.Zero(rw)
		t.Zero(map[string]int{})
		t.Zero(make(map[string]int, 5))
		var ptr *int
		var i interface{} = ptr
		t.Zero(ptr)
		t.Zero(i)
		t.Zero([]int(nil))
		t.Zero([]int{})
		t.Zero(time.Time{})
		t.Zero(make([]int, 0, 5))
	})
	t.Run("NotZero", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotZero(true)
		t.NotZero(-1)
		t.NotZero(0.000000001)
		t.NotZero(" ")
		t.NotZero("0")
		t.NotZero([2]int{0, 1})
		t.NotZero(make(chan int))
		rw := os.Stdout
		t.NotZero(rw)
		t.NotZero(map[string]int{"": 0})
		t.NotZero(new(int))
		t.NotZero(make([]int, 1))
		t.NotZero(time.Now())
		t.NotZero(testing.T{})
	})

	t.Run("Len", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		c := make(chan int, 5)
		t.Len(c, 0)
		c <- 42
		t.Len(c, 1)
		t.Len([2]int{}, 2)
		var m map[string]int
		t.Len(m, 0)
		m = make(map[string]int, 10)
		t.Len(m, 0)
		m["one"] = 1
		t.Len(m, 1)
		t.Len([]int{3, 5}, 2)
		t.Len("cool", 4)
		t.Len("тест", 8)
		t.Len([]byte("тест"), 8)
		t.Len([]rune("тест"), 4)
	})
	t.Run("NotLen", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		c := make(chan int, 5)
		t.NotLen(c, 5)
		c <- 42
		t.NotLen(c, 0)
		t.NotLen([2]int{}, 0)
		var m map[string]int
		t.NotLen(m, 1)
		m = make(map[string]int, 10)
		t.NotLen(m, 10)
		m["one"] = 1
		t.NotLen(m, 0)
		t.NotLen([]int{3, 5}, 1)
		t.NotLen("cool", 3)
		t.NotLen("тест", 4)
		t.NotLen([]byte("тест"), 4)
		t.NotLen([]rune("тест"), 8)
	})

	t.Run("Err", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Err(io.EOF, io.EOF)
		t.Err(io.EOF, errors.New("EOF"))
		var err error
		t.Err(err, nil)
		err = &net.OpError{}
		t.Err(err, &net.OpError{})
	})
	t.Run("NotErr", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		err := (*net.OpError)(nil)
		t.NotErr(err, nil)
		t.NotErr(nil, io.EOF)
	})

	t.Run("Panic", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Panic(func() { panic(nil) })
		t.Panic(func() { panic("") })
		t.Panic(func() { panic("oops") })
		t.Panic(func() { panic(t) })
	})
	t.Run("NotPanic", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotPanic(func() {})
	})

	t.Run("PanicMatch", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.PanicMatch(func() { panic(nil) }, "")
		t.PanicMatch(func() { panic(nil) }, "^<nil>$")
		t.PanicMatch(func() { panic("") }, regexp.MustCompile("^$"))
		t.PanicMatch(func() { panic("oops") }, "(?i)Oops")
		t.PanicMatch(func() { panic(t) }, "^check.T{")
	})
	t.Run("PanicNotMatch", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.PanicNotMatch(func() { panic(nil) }, "^$")
		t.PanicNotMatch(func() { panic(nil) }, "^nil$")
		t.PanicNotMatch(func() { panic("") }, regexp.MustCompile("."))
		t.PanicNotMatch(func() { panic("oops") }, "(?-i)Oops")
		t.PanicNotMatch(func() { panic(t) }, "^*testing.T{")
	})
}
