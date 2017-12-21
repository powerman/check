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
	t.Must(t.Nil(nil))
	t.NotNil(false)
	t.Run("Equal", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Equal(nil, nil)
		t.Equal(true, true)
		t.Equal(false, false)
		t.Equal(0, 0)
		t.Equal(3.14, 3.14)
		t.Equal("", "")
		t.Equal("msg", "msg")
		t.Equal(t, t)
	})
	t.Run("NotEqual", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.NotEqual(nil, true)
		t.NotEqual(false, nil)
		t.NotEqual(int32(0), int64(0))
		t.NotEqual(0, 0.0)
		t.NotEqual("", "msg")
		t.NotEqual(t, tt)
		t.NotEqual(&testing.T{}, &testing.T{})
	})
	t.BytesEqual(nil, nil)
	t.BytesEqual([]byte(nil), nil)
	t.BytesEqual([]byte{}, nil)
	t.BytesEqual([]byte{}, []byte(nil))
	t.BytesEqual([]byte{0}, []byte{0})
	t.NotBytesEqual([]byte{0}, nil)
	t.NotBytesEqual([]byte{0}, []byte{})
	t.NotBytesEqual([]byte{0}, []byte{0, 0})
	t.DeepEqual(t, t)
	t.DeepEqual(tt, tt)
	t.DeepEqual(tt, t.T)
	t.DeepEqual(nil, nil)
	t.DeepEqual(42, 42)
	t.DeepEqual([]byte{2, 5}, []byte{2, 5})
	t.DeepEqual(&testing.T{}, &testing.T{})
	t.NotDeepEqual(int64(42), int32(42))
	t.NotDeepEqual(t, tt)
	t.Match("", `^$`)
	t.Match("", `^.*$`)
	t.Match(" ", `^\s+$`)
	t.Match("World", `(?i)w`)
	t.Match("World", regexp.MustCompile(`(?i)w`))
	t.Match(io.ErrClosedPipe, "closed pipe")
	var err error
	t.Match(err, "^$")
	t.NotMatch(" ", `^$`)
	t.NotMatch("", `^\s+$`)
	t.NotMatch("World", `w`)
	t.NotMatch("World", regexp.MustCompile(`(?-i)w`))
	t.True(1 < 2)
	t.False(2 < 1)
	t.Run("Contains", func(tt *testing.T) {
		t := check.T{tt}
		t.Parallel()
		t.Contains("something", "thing")
		t.Contains([]int{2, 4, 6}, 4)
		t.Contains([3]*testing.T{nil, tt, nil}, tt)
		t.Contains(map[*testing.T]int{nil: 2, tt: 5}, tt)
		t.Contains(map[string]int{"something": 2, "thing": 5}, "thing")
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
		t.NotZero(true)
		t.NotZero(-1)
		t.NotZero(0.000000001)
		t.NotZero(" ")
		t.NotZero("0")
		t.NotZero([2]int{0, 1})
		t.NotZero(make(chan int))
		rw = os.Stdout
		t.NotZero(rw)
		t.NotZero(map[string]int{"": 0})
		t.NotZero(new(int))
		t.NotZero(make([]int, 1))
		t.NotZero(time.Now())
		t.NotZero(testing.T{})
	})
	t.Run("Len", func(tt *testing.T) {
		t := check.T{tt}
		c := make(chan int, 5)
		t.Parallel()
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
	t.Err(io.EOF, io.EOF)
	t.Err(io.EOF, errors.New("EOF"))
	t.Err(err, nil)
	err = &net.OpError{}
	t.Err(err, &net.OpError{})
	err = (*net.OpError)(nil)
	t.NotErr(err, nil)
	t.Panic(func() { panic(nil) }, "")
	t.Panic(func() { panic(nil) }, "^<nil>$")
	t.Panic(func() { panic("") }, regexp.MustCompile("^$"))
	t.Panic(func() { panic("oops") }, "(?i)Oops")
	t.Panic(func() { panic(t) }, "^check.T{")
}
