package check_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"
	"unsafe"

	"github.com/powerman/check"
)

func init() {
	time.Local = time.UTC
}

type (
	myInt    int
	myString string
	myStruct struct {
		i int
		s string
	}
)

var (
	// Zero values for standard types.
	zBool       bool
	zInt        int
	zInt8       int8
	zInt16      int16
	zInt32      int32
	zInt64      int64
	zUint       uint
	zUint8      uint8
	zUint16     uint16
	zUint32     uint32
	zUint64     uint64
	zUintptr    uintptr
	zFloat32    float32
	zFloat64    float64
	zArray0     [0]int
	zArray1     [1]int
	zChan       chan int
	zFunc       func()
	zIface      interface{}
	zMap        map[int]int
	zSlice      []int
	zString     string
	zStruct     struct{}
	zUnsafe     unsafe.Pointer // don't like to import unsafe
	zBoolPtr    *bool
	zIntPtr     *int
	zInt8Ptr    *int8
	zInt16Ptr   *int16
	zInt32Ptr   *int32
	zInt64Ptr   *int64
	zUintPtr    *uint
	zUint8Ptr   *uint8
	zUint16Ptr  *uint16
	zUint32Ptr  *uint32
	zUint64Ptr  *uint64
	zUintptrPtr *uintptr
	zFloat32Ptr *float32
	zFloat64Ptr *float64
	zArray0Ptr  *[0]int
	zArray1Ptr  *[1]int
	zChanPtr    *chan int
	zFuncPtr    *func()
	zIfacePtr   *interface{}
	zMapPtr     *map[int]int
	zSlicePtr   *[]int
	zStringPtr  *string
	zStructPtr  *struct{}
	zUnsafePtr  *unsafe.Pointer // don't like to import unsafe
	// Zero values for named types.
	zMyInt    myInt
	zMyString myString
	zJSON     json.RawMessage
	zJSONPtr  *json.RawMessage
	zTime     time.Time
	// Initialized but otherwise zero-like values.
	vChan              = make(chan int)
	vFunc              = func() {}
	vIface interface{} = zIntPtr
	vMap               = make(map[int]int)
	vSlice             = make([]int, 0)
	// Non-zero values.
	xBool       bool             = true
	xInt        int              = -42
	xInt8       int8             = -8
	xInt16      int16            = -16
	xInt32      int32            = -32
	xInt64      int64            = -64
	xUint       uint             = 42
	xUint8      uint8            = 8
	xUint16     uint16           = 16
	xUint32     uint32           = 32
	xUint64     uint64           = 64
	xUintptr    uintptr          = 0xDEADBEEF
	xFloat32    float32          = -3.2
	xFloat64    float64          = 6.4
	xArray1     [1]int           = [1]int{-1}
	xChan       chan int         = make(chan int, 1)
	xFunc       func()           = func() { panic(nil) }
	xIface      io.Reader        = os.Stdin
	xMap                         = map[int]int{2: -2, 3: -3, 5: -5}
	xSlice                       = []int{3, 5, 8}
	xString     string           = "<nil>"
	xStruct                      = myStruct{i: 10, s: "ten"}
	xUnsafe                      = unsafe.Pointer(&xUintptr) // don't like to import unsafe
	xBoolPtr    *bool            = &xBool
	xIntPtr     *int             = &xInt
	xInt8Ptr    *int8            = &xInt8
	xInt16Ptr   *int16           = &xInt16
	xInt32Ptr   *int32           = &xInt32
	xInt64Ptr   *int64           = &xInt64
	xUintPtr    *uint            = &xUint
	xUint8Ptr   *uint8           = &xUint8
	xUint16Ptr  *uint16          = &xUint16
	xUint32Ptr  *uint32          = &xUint32
	xUint64Ptr  *uint64          = &xUint64
	xUintptrPtr *uintptr         = &xUintptr
	xFloat32Ptr *float32         = &xFloat32
	xFloat64Ptr *float64         = &xFloat64
	xArray1Ptr  *[1]int          = &xArray1
	xChanPtr    *chan int        = &xChan
	xFuncPtr    *func()          = &xFunc
	xIfacePtr   *io.Reader       = &xIface
	xMapPtr     *map[int]int     = &xMap
	xSlicePtr   *[]int           = &xSlice
	xStringPtr  *string          = &xString
	xStructPtr  *myStruct        = &xStruct
	xUnsafePtr  *unsafe.Pointer  = &xUnsafe // don't like to import unsafe
	xMyInt      myInt            = 31337
	xMyString   myString         = "x"
	xJSON       json.RawMessage  = []byte(`{"s":"ten","i":10}`)
	xJSONPtr    *json.RawMessage = &xJSON
	xTime       time.Time        = time.Now()
	xTimeEST    time.Time        = xTime.In(func() *time.Location { loc, _ := time.LoadLocation("EST"); return loc }())
)

func TestTODO(tt *testing.T) {
	t := check.T(tt)
	// Normal tests.
	t.True(true)
	// If you need to mark just one/few broken tests:
	t.TODO().True(false)
	t.True(true)
	// If there are several broken tests mixed with working ones:
	todo := t.TODO()
	t.True(true)
	todo.True(false)
	t.True(true)
	if todo.True(false) {
		panic("never here")
	}
	// If all tests below this point are broken:
	t = t.TODO()
	t.True(false)
	// Second TODO() doesn't switch it off:
	t = t.TODO()
	t.True(false)
}

func TestMust(tt *testing.T) {
	t := check.T(tt)
	t.Must(t.Nil(nil))
	t.Must(t.NotNil(false))
}

func bePositive(_ *check.C, actual interface{}) bool {
	return actual.(int) > 0
}

func beEqual(_ *check.C, actual, expected interface{}) bool {
	return actual == expected
}

func TestCheckerShould(tt *testing.T) {
	t := check.T(tt)
	t.Should(bePositive, 42, "custom check!!!")
	t.Panic(func() { t.Should(bePositive, "42", "bad arg type") })
	t.TODO().Should(func(_ *check.C, _ interface{}) bool { return false }, 42)
	t.Should(beEqual, 123, 123)
	t.TODO().Should(beEqual, 123, 124)
}

func TestCheckerNilTrue(tt *testing.T) {
	t := check.T(tt)
	todo := t.TODO()

	// Ensure expected values
	t.Equal(zBool, false) // gometalinter hates zBool==false
	t.True(zInt == 0)
	t.True(zInt8 == 0)
	t.True(zInt16 == 0)
	t.True(zInt32 == 0)
	t.True(zInt64 == 0)
	t.True(zUint == 0)
	t.True(zUint8 == 0)
	t.True(zUint16 == 0)
	t.True(zUint32 == 0)
	t.True(zUint64 == 0)
	t.True(zUintptr == 0)
	t.True(zFloat32 == 0)
	t.True(zFloat64 == 0)
	t.True(zArray0 == [0]int{})
	t.True(zArray1 == [1]int{})
	t.True(zChan == nil)
	t.True(zFunc == nil)
	t.True(zIface == nil)
	t.True(zMap == nil)
	t.True(zSlice == nil)
	t.True(zString == "")
	t.True(zStruct == struct{}{})
	t.True(zUnsafe == nil)
	t.True(zBoolPtr == nil)
	t.True(zIntPtr == nil)
	t.True(zInt8Ptr == nil)
	t.True(zInt16Ptr == nil)
	t.True(zInt32Ptr == nil)
	t.True(zInt64Ptr == nil)
	t.True(zUintPtr == nil)
	t.True(zUint8Ptr == nil)
	t.True(zUint16Ptr == nil)
	t.True(zUint32Ptr == nil)
	t.True(zUint64Ptr == nil)
	t.True(zUintptrPtr == nil)
	t.True(zFloat32Ptr == nil)
	t.True(zFloat64Ptr == nil)
	t.True(zArray0Ptr == nil)
	t.True(zArray1Ptr == nil)
	t.True(zChanPtr == nil)
	t.True(zFuncPtr == nil)
	t.True(zIfacePtr == nil)
	t.True(zMapPtr == nil)
	t.True(zSlicePtr == nil)
	t.True(zStringPtr == nil)
	t.True(zStructPtr == nil)
	t.True(zUnsafePtr == nil)
	t.True(zMyInt == 0)
	t.True(zMyString == "")
	t.True(zJSON == nil)
	t.True(zJSONPtr == nil)
	t.True(zTime == time.Time{})
	t.False(vChan == nil)
	t.False(vFunc == nil)
	t.False(vIface == nil)
	t.False(vMap == nil)
	t.False(vSlice == nil)

	// Subtle case when t.Nil() differs from == nil.
	zIface = zIntPtr
	t.Nil(zIface)
	t.False(zIface == nil)
	zIface = nil
	t.Nil(zIface)
	t.True(zIface == nil)

	cases := []struct {
		equalNil bool
		isNil    bool
		actual   interface{}
	}{
		{true, true, nil},
		{false, false, zBool},
		{false, false, zInt},
		{false, false, zInt8},
		{false, false, zInt16},
		{false, false, zInt32},
		{false, false, zInt64},
		{false, false, zUint},
		{false, false, zUint8},
		{false, false, zUint16},
		{false, false, zUint32},
		{false, false, zUint64},
		{false, false, zUintptr},
		{false, false, zFloat32},
		{false, false, zFloat64},
		{false, false, zArray0},
		{false, false, zArray1},
		{false, true, zChan},
		{false, true, zFunc},
		{true, true, zIface},
		{false, true, zMap},
		{false, true, zSlice},
		{false, false, zString},
		{false, false, zStruct},
		{false, false, zUnsafe},
		{false, true, zBoolPtr},
		{false, true, zIntPtr},
		{false, true, zInt8Ptr},
		{false, true, zInt16Ptr},
		{false, true, zInt32Ptr},
		{false, true, zInt64Ptr},
		{false, true, zUintPtr},
		{false, true, zUint8Ptr},
		{false, true, zUint16Ptr},
		{false, true, zUint32Ptr},
		{false, true, zUint64Ptr},
		{false, true, zUintptrPtr},
		{false, true, zFloat32Ptr},
		{false, true, zFloat64Ptr},
		{false, true, zArray0Ptr},
		{false, true, zArray1Ptr},
		{false, true, zChanPtr},
		{false, true, zFuncPtr},
		{false, true, zIfacePtr},
		{false, true, zMapPtr},
		{false, true, zSlicePtr},
		{false, true, zStringPtr},
		{false, true, zStructPtr},
		{false, true, zUnsafePtr},
		{false, false, zMyInt},
		{false, false, zMyString},
		{false, true, zJSON},
		{false, true, zJSONPtr},
		{false, false, zTime},
		{false, false, vChan},
		{false, false, vFunc},
		{false, true, vIface}, // WARNING false-positive (documented)
		{false, false, vMap},
		{false, false, vSlice},
	}
	for i, v := range cases {
		msg := fmt.Sprintf("case %d: %#v", i, v.actual)
		if v.equalNil {
			t.True(v.actual == nil, msg)
		} else {
			t.False(v.actual == nil, msg)
		}
		if v.isNil {
			t.Nil(v.actual, msg)
			todo.NotNil(v.actual, msg)
		} else {
			todo.Nil(v.actual, msg)
			t.NotNil(v.actual, msg)
		}
	}
}

func TestCheckerEqual(tt *testing.T) {
	t := check.T(tt)
	todo := t.TODO()

	cases := []struct {
		comparable bool
		actual     interface{}
		actual2    interface{}
	}{
		{true, zBool, xBool},
		{true, zInt, xInt},
		{true, zInt8, xInt8},
		{true, zInt16, xInt16},
		{true, zInt32, xInt32},
		{true, zInt64, xInt64},
		{true, zUint, xUint},
		{true, zUint8, xUint8},
		{true, zUint16, xUint16},
		{true, zUint32, xUint32},
		{true, zUint64, xUint64},
		{true, zUintptr, xUintptr},
		{true, zFloat32, xFloat32},
		{true, zFloat64, xFloat64},
		{true, zArray0, xArray1},
		{true, zArray1, xArray1},
		{true, zChan, xChan},
		{false, zFunc, xFunc},
		{true, zIface, xIface},
		{false, zMap, xMap},
		{false, zSlice, xSlice},
		{true, zString, xString},
		{true, zStruct, xStruct},
		{true, zBoolPtr, xBoolPtr},
		{true, zIntPtr, xIntPtr},
		{true, zInt8Ptr, xInt8Ptr},
		{true, zInt16Ptr, xInt16Ptr},
		{true, zInt32Ptr, xInt32Ptr},
		{true, zInt64Ptr, xInt64Ptr},
		{true, zUintPtr, xUintPtr},
		{true, zUint8Ptr, xUint8Ptr},
		{true, zUint16Ptr, xUint16Ptr},
		{true, zUint32Ptr, xUint32Ptr},
		{true, zUint64Ptr, xUint64Ptr},
		{true, zUintptrPtr, xUintptrPtr},
		{true, zFloat32Ptr, xFloat32Ptr},
		{true, zFloat64Ptr, xFloat64Ptr},
		{true, zArray0Ptr, xArray1Ptr},
		{true, zArray1Ptr, xArray1Ptr},
		{true, zChanPtr, xChanPtr},
		{true, zFuncPtr, xFuncPtr},
		{true, zIfacePtr, xIfacePtr},
		{true, zMapPtr, xMapPtr},
		{true, zSlicePtr, xSlicePtr},
		{true, zStringPtr, xStringPtr},
		{true, zStructPtr, xStructPtr},
		{true, zMyInt, xMyInt},
		{true, zMyString, xMyString},
		{false, zJSON, xJSON},
		{true, zJSONPtr, xJSONPtr},
		{true, zTime, xTime},
		{true, vChan, xChan},
		{false, vFunc, xFunc},
		{true, vIface, xIface},
		{false, vMap, xMap},
		{false, vSlice, xSlice},
		{true, "one\ntwo\nend", "one\nTWO\nend"},
		{true, io.EOF, io.ErrUnexpectedEOF},
		{true, t, tt},
		{true, int64(42), int32(42)},
		{false, []byte{}, []byte(nil)},
	}
	for _, v := range cases {
		if v.comparable {
			t.Equal(v.actual, v.actual)
			t.EQ(v.actual, v.actual)
			t.DeepEqual(v.actual, v.actual)
			todo.NotEqual(v.actual, v.actual)
			todo.NE(v.actual, v.actual)
			todo.NotDeepEqual(v.actual, v.actual)

			t.Equal(v.actual2, v.actual2)
			t.EQ(v.actual2, v.actual2)
			t.DeepEqual(v.actual2, v.actual2)
			todo.NotEqual(v.actual2, v.actual2)
			todo.NE(v.actual2, v.actual2)
			todo.NotDeepEqual(v.actual2, v.actual2)

			todo.Equal(v.actual, v.actual2)
			todo.EQ(v.actual, v.actual2)
			todo.DeepEqual(v.actual, v.actual2)
			t.NotEqual(v.actual, v.actual2)
			t.NE(v.actual, v.actual2)
			t.NotDeepEqual(v.actual, v.actual2)
		} else {
			t.Panic(func() { t.Equal(v.actual, v.actual) })
			t.Panic(func() { t.EQ(v.actual, v.actual) })
			t.Panic(func() { t.NotEqual(v.actual, v.actual) })
			t.Panic(func() { t.NE(v.actual, v.actual) })

			if reflect.TypeOf(v.actual).Kind() != reflect.Func {
				t.DeepEqual(v.actual, v.actual)
				todo.NotDeepEqual(v.actual, v.actual)
				t.DeepEqual(v.actual2, v.actual2)
				todo.NotDeepEqual(v.actual2, v.actual2)
				todo.DeepEqual(v.actual, v.actual2)
				t.NotDeepEqual(v.actual, v.actual2)
			}
		}
	}

	// No alternative value for .actual2.
	t.Equal(nil, nil)
	t.EQ(nil, nil)
	t.DeepEqual(nil, nil)
	todo.NotEqual(nil, nil)
	todo.NE(nil, nil)
	todo.NotDeepEqual(nil, nil)

	// Equal match, DeepEqual not match.
	t.False(xTime == xTimeEST)
	t.Equal(xTime, xTimeEST)
	t.EQ(xTime, xTimeEST)
	t.NotDeepEqual(xTime, xTimeEST)
	todo.NotEqual(xTime, xTimeEST)
	todo.NE(xTime, xTimeEST)
	todo.DeepEqual(xTime, xTimeEST)

	// Equal not match or panic, DeepEqual match.
	type notComparable struct {
		s  string
		is []int
	}
	cases = []struct {
		comparable bool
		actual     interface{}
		actual2    interface{}
	}{
		{true, io.EOF, errors.New("EOF")},
		{true, &testing.T{}, &testing.T{}},
		{false, []byte{2, 5}, []byte{2, 5}},
		{false, notComparable{"a", []int{3, 5}}, notComparable{"a", []int{3, 5}}},
	}
	for _, v := range cases {
		if v.comparable {
			t.False(v.actual == v.actual2)
			todo.Equal(v.actual, v.actual2)
			todo.EQ(v.actual, v.actual2)
			t.NotEqual(v.actual, v.actual2)
			t.NE(v.actual, v.actual2)
		}
		t.DeepEqual(v.actual, v.actual2)
		todo.NotDeepEqual(v.actual, v.actual2)
	}
}

func TestCheckers(t *testing.T) {
	time1 := time.Now()

	t.Run("BytesEqual", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.BytesEqual(nil, nil)
		t.BytesEqual([]byte(nil), nil)
		t.BytesEqual([]byte{}, nil)
		t.BytesEqual([]byte{}, []byte(nil))
		t.BytesEqual([]byte{0}, []byte{0})
	})
	t.Run("NotBytesEqual", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.NotBytesEqual([]byte{0}, nil)
		t.NotBytesEqual([]byte{0}, []byte{})
		t.NotBytesEqual([]byte{0}, []byte{0, 0})
	})

	t.Run("Match", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.Match("", `^$`)
		t.Match("", `^.*$`)
		t.Match(" ", `^\s+$`)
		t.Match("World", `(?i)w`)
		t.Match(myString("World"), `(?i)w`)
		t.Match([]byte("World"), `(?i)w`)
		t.Match([]rune("World"), `(?i)w`)
		t.Match("World", regexp.MustCompile(`(?i)w`))
		t.Match(io.ErrClosedPipe, "closed pipe")
		t.Match(time.Time{}, "00:00:00")
	})
	t.Run("NotMatch", func(tt *testing.T) {
		t := check.T(tt)
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
		t := check.T(tt)
		t.Parallel()
		t.Contains("something", "thing")
		t.Contains("something", myString("thing"))
		t.Contains(myString("something"), "thing")
		t.Contains(myString("something"), myString("thing"))
		t.Contains([]int{2, 4, 6}, 4)
		t.Contains([3]*testing.T{nil, tt, nil}, tt)
		t.Contains(map[*testing.T]int{nil: 2, tt: 5}, 5)
		t.Contains(map[int]string{2: "something", 5: "thing"}, "thing")
	})
	t.Run("NotContains", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.NotContains("something", "Thing")
		t.NotContains("something", myString("Thing"))
		t.NotContains(myString("something"), "Thing")
		t.NotContains(myString("something"), myString("Thing"))
		t.NotContains([]int{2, 4, 6}, 3)
		t.NotContains([3]*testing.T{nil, tt, nil}, &testing.T{})
		t.NotContains(map[*testing.T]int{nil: 2, tt: 5}, 0)
		t.NotContains(map[int]string{2: "something", 5: "thing"}, "some")
	})

	t.Run("HasKey", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.HasKey(map[*testing.T]int{nil: 2, tt: 5}, tt)
		t.HasKey(map[int]string{2: "something", 5: "thing"}, 5)
	})
	t.Run("NotHasKey", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.NotHasKey(map[*testing.T]int{nil: 2, tt: 5}, &testing.T{})
		t.NotHasKey(map[int]string{2: "something", 5: "thing"}, 0)
	})

	t.Run("Zero", func(tt *testing.T) {
		t := check.T(tt)
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
		t := check.T(tt)
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
		t := check.T(tt)
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
		t := check.T(tt)
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
		t := check.T(tt)
		t.Parallel()
		t.Err(io.EOF, io.EOF)
		t.Err(io.EOF, errors.New("EOF"))
		var err error
		t.Err(err, nil)
		err = &net.OpError{}
		t.Err(err, &net.OpError{})
	})
	t.Run("NotErr", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		err := (*net.OpError)(nil)
		t.NotErr(err, nil)
		t.NotErr(nil, io.EOF)
	})

	t.Run("Panic", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.Panic(func() { panic(nil) })
		t.Panic(func() { panic("") })
		t.Panic(func() { panic("oops") })
		t.Panic(func() { panic(t) })
	})
	t.Run("NotPanic", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.NotPanic(func() {})
	})

	t.Run("PanicMatch", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.PanicMatch(func() { panic(nil) }, "")
		t.PanicMatch(func() { panic(nil) }, "^<nil>$")
		t.PanicMatch(func() { panic("") }, regexp.MustCompile("^$"))
		t.PanicMatch(func() { panic("oops") }, "(?i)Oops")
		t.PanicMatch(func() { panic(t) }, "^&check.C{")
	})
	t.Run("PanicNotMatch", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.PanicNotMatch(func() { panic(nil) }, "^$")
		t.PanicNotMatch(func() { panic(nil) }, "^nil$")
		t.PanicNotMatch(func() { panic("") }, regexp.MustCompile("."))
		t.PanicNotMatch(func() { panic("oops") }, "(?-i)Oops")
		t.PanicNotMatch(func() { panic(t) }, "^*testing.T{")
	})

	less := []struct{ actual, expected interface{} }{
		{0, 1},
		{int8(-1), int8(0)},
		{'a', 'b'},
		{2 << 60, 2 << 61},
		{byte(254), byte(255)},
		{uint64(0), uint64(1)},
		{0.1, 0.2},
		{"a1", "a2"},
		{time1, time1.Add(time.Second)},
	}
	t.Run("Less+LT+LessOrEqual+LE", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		for _, v := range less {
			actual, expected := v.actual, v.expected
			t.Less(actual, expected)
			t.LT(actual, expected)
			t.LessOrEqual(actual, expected)
			t.LessOrEqual(actual, actual)
			t.LE(actual, expected)
			t.LE(actual, actual)
		}
	})
	t.Run("Greater+GT+GreaterOrEqual+GE", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		for _, v := range less {
			actual, expected := v.expected, v.actual
			t.Greater(actual, expected)
			t.GT(actual, expected)
			t.GreaterOrEqual(actual, expected)
			t.GreaterOrEqual(actual, actual)
			t.GE(actual, expected)
			t.GE(actual, actual)
		}
	})

	between := []struct{ min, mid, max interface{} }{
		{0, 1, 5},
		{int8(-1), int8(0), int8(1)},
		{'a', 'b', 'z'},
		{2 << 59, 2 << 60, 2 << 61},
		{byte(0), byte(254), byte(255)},
		{uint64(0), uint64(1), uint64(5)},
		{0.01, 0.1, 0.2},
		{"a1", "a2", "b"},
		{time1, time1.Add(time.Millisecond), time1.Add(time.Second)},
	}
	t.Run("Between+BetweenOrEqual+NotBetween+NotBetweenOrEqual", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		for _, v := range between {
			min, mid, max := v.min, v.mid, v.max
			t.Between(mid, min, max)
			t.BetweenOrEqual(mid, min, max)
			t.BetweenOrEqual(mid, mid, max)
			t.BetweenOrEqual(mid, min, mid)
			t.NotBetween(min, mid, max)
			t.NotBetween(max, min, mid)
			t.NotBetweenOrEqual(min, mid, max)
			t.NotBetweenOrEqual(max, min, mid)
		}
	})

	inDelta := []struct{ actual, expected, delta interface{} }{
		{-1, 0, 1},
		{byte(92), byte(100), byte(10)},
		{0.92, 1.0, 0.1},
		{time1, time1.Add(5 * time.Second), 7 * time.Second},
	}
	t.Run("InDelta+NotInDelta", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		for _, v := range inDelta {
			t.InDelta(v.actual, v.expected, v.delta)
			t.InDelta(v.expected, v.actual, v.delta)
			switch delta := v.delta.(type) {
			case int:
				t.NotInDelta(v.actual, v.expected, delta/2)
			case byte:
				t.NotInDelta(v.actual, v.expected, delta/2)
			case float64:
				t.NotInDelta(v.actual, v.expected, delta/2)
			case time.Duration:
				t.NotInDelta(v.actual, v.expected, delta/2)
			}
		}
	})

	inSMAPE := []struct {
		actual, expected interface{}
		smape            float64
	}{
		{-101, -100, 0.5},
		{-99, -100, 0.7},
		{byte(92), byte(100), 5},
		{0.92, 1.0, 5},
	}
	t.Run("InSMAPE+NotInSMAPE", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		for _, v := range inSMAPE {
			t.InSMAPE(v.actual, v.expected, v.smape)
			t.NotInSMAPE(v.actual, v.expected, v.smape/2)
		}
	})

	prefix := []struct{ actual, expected interface{} }{
		{myString("abcde"), []byte("ab")},
		{[]rune("abcde"), myString("ab")},
		{time.Time{}, errors.New("0001-01-01")},
	}
	t.Run("HasPrefix+NotHasPrefix", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.HasPrefix("", myString(""))
		for _, v := range prefix {
			actual, expected := v.actual, v.expected
			t.HasPrefix(actual, expected)
			t.NotHasPrefix(expected, actual)
		}
	})

	suffix := []struct{ actual, expected interface{} }{
		{myString("abcde"), []byte("de")},
		{[]rune("abcde"), myString("de")},
		{time.Time{}, errors.New("UTC")},
	}
	t.Run("HasSuffix+NotHasSuffix", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.HasSuffix("", myString(""))
		for _, v := range suffix {
			actual, expected := v.actual, v.expected
			t.HasSuffix(actual, expected)
			t.NotHasSuffix(expected, actual)
		}
	})

	t.Run("JSONEqual", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		t.JSONEqual(`{ "a" : [3, false],"z":42 }`, []byte(`{"z": 42,"a":[3  ,  false ]}`))
		t.JSONEqual(`true`, ` true `)
		raw := json.RawMessage(`true`)
		t.JSONEqual(&raw, raw)
	})

	t.Run("HasType+NotHasType", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		var reader io.Reader
		t.HasType(nil, nil)
		t.HasType(reader, nil)
		t.HasType(false, true)
		t.HasType(42, 0)
		t.HasType("test", "")
		t.HasType([]byte("test"), []byte(nil))
		t.HasType([]byte("test"), []byte{})
		t.HasType(&reader, (*io.Reader)(nil))
		t.HasType(os.Stdin, (*os.File)(nil))
		t.HasType(new(int), (*int)(nil))
		t.NotHasType(nil, (*int)(nil))
		t.NotHasType((*int)(nil), nil)
		t.NotHasType((*int)(nil), (*uint)(nil))
		t.NotHasType(&reader, nil)
		t.NotHasType(42, uint(42))
		t.NotHasType(0.0, 0)
		t.NotHasType(json.RawMessage([]byte("test")), []byte{})
	})

	t.Run("Implements+NotImplements", func(tt *testing.T) {
		t := check.T(tt)
		t.Parallel()
		var reader io.Reader = os.Stdin
		t.Implements(t, (*testing.TB)(nil))
		t.Implements(os.Stdin, (*io.Reader)(nil))
		t.Implements(os.Stdin, &reader)
		t.Implements(*os.Stdin, (*io.Reader)(nil))
		t.Implements(time.Time{}, (*fmt.Stringer)(nil))
		t.Implements(&time.Time{}, (*fmt.Stringer)(nil))
		t.NotImplements(os.Stdin, (*fmt.Stringer)(nil))
		t.NotImplements(&os.Stdin, (*io.Reader)(nil))
		t.NotImplements(new(int), (*io.Reader)(nil))
	})
}
