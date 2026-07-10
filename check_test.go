//nolint:err113,errname // It's just a test.
package check_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/powerman/check"
)

type (
	myInt    int
	myString string
	myStruct struct {
		i int
		s string
	}
	myError     struct{ s string }
	plainError  struct{ msg string }
	targetErr   struct{ msg string }
	ptrFieldErr struct{ msg string }
	fieldError  struct {
		ns  string
		tag string
	}
	fieldErrors []*fieldError
	causeError  struct {
		msg string
		err error
	}
)

func (e myError) Error() string         { return e.s }
func (e *plainError) Error() string     { return e.msg }
func (e *targetErr) Error() string      { return e.msg }
func (e *ptrFieldErr) Error() string    { return e.msg }
func (e *fieldError) Error() string     { return e.ns + ":" + e.tag }
func (e *fieldError) Namespace() string { return e.ns }
func (e *fieldError) Tag() string       { return e.tag }
func (fieldErrors) Error() string       { return "fieldErrors" }

func (e *causeError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}
func (e *causeError) Cause() error  { return e.err }
func (e *causeError) Unwrap() error { return e.err }

func pkgerrorsNew() error {
	return &plainError{msg: "EOF"}
}

func pkgerrorsWithStack(err error) error {
	return &causeError{err: err}
}

func pkgerrorsWrap(err error, msg string) error {
	return &causeError{msg: msg, err: err}
}

var (
	// Zero values for standard types.
	zBool    bool
	zInt     int
	zInt8    int8
	zInt16   int16
	zInt32   int32
	zInt64   int64
	zUint    uint
	zUint8   uint8
	zUint16  uint16
	zUint32  uint32
	zUint64  uint64
	zUintptr uintptr
	zFloat32 float32
	zFloat64 float64
	zArray0  [0]int
	zArray1  [1]int
	zChan    chan int
	zFunc    func()
	zIface   any
	zMap     map[int]int
	zSlice   []int
	zString  string
	zStruct  struct{}
	// zUnsafe     unsafe.Pointer // don't like to import unsafe.
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
	zIfacePtr   *any
	zMapPtr     *map[int]int
	zSlicePtr   *[]int
	zStringPtr  *string
	zStructPtr  *struct{}
	// zUnsafePtr  *unsafe.Pointer // don't like to import unsafe
	// Zero values for named types.
	zMyInt    myInt
	zMyString myString
	zJSON     json.RawMessage
	zJSONPtr  *json.RawMessage
	zTime     time.Time
	// Initialized but otherwise zero-like values.
	vChan      = make(chan int)
	vFunc      = func() {}
	vIface any = zIntPtr
	vMap       = make(map[int]int)
	vSlice     = make([]int, 0)
	// Non-zero values.
	xBool              = true
	xInt               = -42
	xInt8    int8      = -8
	xInt16   int16     = -16
	xInt32   int32     = -32
	xInt64   int64     = -64
	xUint    uint      = 42
	xUint8   uint8     = 8
	xUint16  uint16    = 16
	xUint32  uint32    = 32
	xUint64  uint64    = 64
	xUintptr uintptr   = 0xDEADBEEF
	xFloat32 float32   = -3.2
	xFloat64           = 6.4
	xArray1            = [1]int{-1}
	xChan              = make(chan int, 1)
	xFunc              = func() { panic(nil) }
	xIface   io.Reader = os.Stdin
	xMap               = map[int]int{2: -2, 3: -3, 5: -5}
	xSlice             = []int{3, 5, 8}
	xString            = "<nil>"
	xStruct            = myStruct{i: 10, s: "ten"}
	// xUnsafe                      = unsafe.Pointer(&xUintptr) // don't like to import unsafe.
	xBoolPtr    = &xBool
	xIntPtr     = &xInt
	xInt8Ptr    = &xInt8
	xInt16Ptr   = &xInt16
	xInt32Ptr   = &xInt32
	xInt64Ptr   = &xInt64
	xUintPtr    = &xUint
	xUint8Ptr   = &xUint8
	xUint16Ptr  = &xUint16
	xUint32Ptr  = &xUint32
	xUint64Ptr  = &xUint64
	xUintptrPtr = &xUintptr
	xFloat32Ptr = &xFloat32
	xFloat64Ptr = &xFloat64
	xArray1Ptr  = &xArray1
	xChanPtr    = &xChan
	xFuncPtr    = &xFunc
	xIfacePtr   = &xIface
	xMapPtr     = &xMap
	xSlicePtr   = &xSlice
	xStringPtr  = &xString
	xStructPtr  = &xStruct
	// xUnsafePtr  *unsafe.Pointer  = &xUnsafe // don't like to import unsafe.
	xMyInt    myInt           = 31337
	xMyString myString        = "xyz"
	xJSON     json.RawMessage = []byte(`{"s":"ten","i":10}`)
	xJSONPtr                  = &xJSON
	xTime                     = time.Now()
	xTimeEST                  = xTime.In(func() *time.Location { loc, _ := time.LoadLocation("EST"); return loc }())
)

func TestTODO(tt *testing.T) {
	tt.Parallel()
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

func TestError(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t = t.TODO()
	t.Error()
	t.Error("message")
	t.Error("format: %q", "message")
}

func TestErrorf(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t.TODO().Errorf("message %d", 1)
	t.TODO().Errorf("format: %s", "message")
}

func TestFatal(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t.TODO().Fatal()
	t.TODO().Fatal("message")
	t.TODO().Fatal("format: %s", "message")
}

func TestFatalf(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t.TODO().Fatalf("message")
	t.TODO().Fatalf("format: %s", "message")
}

func TestMustAll(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt).MustAll()
	t.Nil(nil)
	t.NotNil(false)
	t.TODO().Nil(false)
	t.TODO().NotNil(nil)
}

func TestMust(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t.Must(t.Nil(nil))
	t.Must(t.NotNil(false))
}

func bePositive(_ *check.TB, actual any) bool {
	return actual.(int) > 0
}

func beEqual(_ *check.TB, actual, expected any) bool {
	return actual == expected
}

func TestCheckerShould(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	t.Should(bePositive, 42, "custom check!!!")
	t.Panic(func() { t.Should(bePositive, "42", "bad arg type") })
	t.TODO().Should(func(_ *check.TB, _ any) bool { return false }, 42)
	t.Should(beEqual, 123, 123)
	t.TODO().Should(beEqual, 123, 124)
	t.Panic(func() { t.Should(func() {}, nil) })
	t.Panic(func() { t.Should(bePositive) })
	t.Panic(func() { t.Should(beEqual, nil) })
}

func TestCheckerNilTrue(tt *testing.T) {
	tt.Parallel()
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
	// // t.True(zUnsafe == nil)
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
	// // t.True(zUnsafePtr == nil)
	t.True(zMyInt == 0)
	t.True(zMyString == "")
	t.True(zJSON == nil)
	t.True(zJSONPtr == nil)
	t.True(zTime.Equal(time.Time{}))
	t.False(vChan == nil)
	t.False(vFunc == nil)
	t.False(vIface == nil)
	t.False(vMap == nil)
	t.False(vSlice == nil)

	// Subtle case when t.Nil() differs from == nil.
	vIfaceNil := vIface
	t.Nil(vIfaceNil)
	t.False(vIfaceNil == nil)
	vIfaceNil = nil
	t.Nil(vIfaceNil)
	t.True(vIfaceNil == nil)

	cases := []struct {
		equalNil bool
		isNil    bool
		actual   any
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
		// {false, false, zUnsafe},
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
		// {false, true, zUnsafePtr},
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
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	cases := []struct {
		comparable bool
		actual     any
		actual2    any
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
	t.False(xTime == xTimeEST) //nolint:revive,staticcheck // Need == instead of Equal() here.
	t.Equal(xTime, xTimeEST)
	t.EQ(xTime, xTimeEST)
	t.DeepEqual(xTime, xTimeEST)
	todo.NotEqual(xTime, xTimeEST)
	todo.NE(xTime, xTimeEST)
	todo.NotDeepEqual(xTime, xTimeEST)

	// Equal not match or panic, DeepEqual match.
	type notComparable struct {
		s  string
		is []int
	}
	cases = []struct {
		comparable bool
		actual     any
		actual2    any
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

func TestCheckerBytesEqual(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	cases := []struct {
		equal    bool
		actual   []byte
		expected []byte
	}{
		{true, nil, nil},
		{true, []byte(nil), []byte(nil)},
		{true, []byte{}, []byte{}},
		{true, []byte(nil), nil},
		{true, []byte{}, nil},
		{true, []byte(nil), []byte{}},
		{true, []byte{0}, []byte{0}},
		{false, []byte{0}, nil},
		{false, []byte{0}, []byte(nil)},
		{false, []byte{0}, []byte{}},
		{false, []byte{0}, []byte{0, 0}},
	}
	for _, v := range cases {
		if v.equal {
			t.BytesEqual(v.actual, v.expected)
			todo.NotBytesEqual(v.actual, v.expected)
		} else {
			todo.BytesEqual(v.actual, v.expected)
			t.NotBytesEqual(v.actual, v.expected)
		}
	}
}

func TestCheckerMatch(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	types := []struct {
		actual   bool
		expected bool
		zero     any
	}{
		{true, false, nil},
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
		{false, false, zChan},
		{false, false, zFunc},
		{false, false, zIface},
		{false, false, zMap},
		{false, false, zSlice},
		{true, true, zString},
		{false, false, zStruct},
		{false, false, zBoolPtr},
		{false, false, zIntPtr},
		{false, false, zInt8Ptr},
		{false, false, zInt16Ptr},
		{false, false, zInt32Ptr},
		{false, false, zInt64Ptr},
		{false, false, zUintPtr},
		{false, false, zUint8Ptr},
		{false, false, zUint16Ptr},
		{false, false, zUint32Ptr},
		{false, false, zUint64Ptr},
		{false, false, zUintptrPtr},
		{false, false, zFloat32Ptr},
		{false, false, zFloat64Ptr},
		{false, false, zArray0Ptr},
		{false, false, zArray1Ptr},
		{false, false, zChanPtr},
		{false, false, zFuncPtr},
		{false, false, zIfacePtr},
		{false, false, zMapPtr},
		{false, false, zSlicePtr},
		{false, false, zStringPtr},
		{false, false, zStructPtr},
		{false, false, zMyInt},
		{true, false, zMyString},
		{true, false, zJSON},
		{false, false, zJSONPtr},
		{true, false, zTime},
		{true, false, time.Sunday},
		{true, false, errors.New("")},
		{true, false, []byte(nil)},
		{true, false, []rune(nil)},
		{true, true, regexp.MustCompile("")}, // it's also a Stringer
		{false, false, (*regexp.Regexp)(nil)},
		{false, false, regexp.Regexp{}},
	}
	for i, va := range types {
		for j, ve := range types {
			msg := fmt.Sprintf("case %d/%d: %#v, %#v", i, j, va.zero, ve.zero)
			switch va.zero.(type) {
			case nil:
				todo.Match(va.zero, ve.zero, msg)
			default:
				if va.actual && ve.expected {
					t.Match(va.zero, ve.zero, msg)
				} else {
					t.Panic(func() { t.Match(va.zero, ve.zero) }, msg)
				}
			}
		}
	}

	cases := []struct {
		actual        any
		regexMatch    any
		regexNotMatch any
	}{
		{"", `^$`, `.`},
		{myString("Test"), regexp.MustCompile(`st$`), regexp.MustCompile(`ST$`)},
		{[]byte(nil), `^$`, `nil`},
		{[]byte("Test"), regexp.MustCompile(`st$`), regexp.MustCompile(`ST$`)},
		{[]rune(nil), `^$`, `nil`},
		{[]rune("Test"), regexp.MustCompile(`st$`), regexp.MustCompile(`ST$`)},
		{zTime, `00:00:00`, `01:01:01`},
		{time.Sunday, regexp.MustCompile(`^Sun`), regexp.MustCompile(`Sun$`)},
		{errors.New(""), `^$`, `nil`},
		{io.EOF, regexp.MustCompile(`^EO`), regexp.MustCompile(`EO$`)},
	}
	for _, v := range cases {
		t.Match(v.actual, v.regexMatch)
		todo.Match(v.actual, v.regexNotMatch)
	}

	// No value for .regexMatch.
	todo.Match(nil, ``)
	todo.Match(nil, regexp.MustCompile(``))
	t.NotMatch(nil, ``)
	t.NotMatch(nil, regexp.MustCompile(``))
}

func TestCheckerContains(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	failures := []struct {
		panic    bool
		actual   any
		expected any
	}{
		{true, nil, nil},
		{true, zBool, zBool},
		{true, zInt, zInt},
		{true, zInt8, zInt8},
		{true, zInt16, zInt16},
		{true, zInt32, zInt32},
		{true, zInt64, zInt64},
		{true, zUint, zUint},
		{true, zUint8, zUint8},
		{true, zUint16, zUint16},
		{true, zUint32, zUint32},
		{true, zUint64, zUint64},
		{true, zUintptr, zUintptr},
		{true, zFloat32, zFloat32},
		{true, zFloat64, zFloat64},
		{true, zArray0, zBool},
		{false, zArray0, xInt},
		{true, zArray1, zBool},
		{false, zArray1, xInt},
		{true, zChan, zChan},
		{true, zFunc, zFunc},
		{true, zIface, zIface},
		{true, zMap, zBool},
		{false, zMap, xInt},
		{true, zSlice, zBool},
		{false, zSlice, xInt},
		{true, zString, zBool},
		{false, zString, xString},
		{true, zStruct, zStruct},
		{true, zBoolPtr, zBoolPtr},
		{true, zIntPtr, zIntPtr},
		{true, zInt8Ptr, zInt8Ptr},
		{true, zInt16Ptr, zInt16Ptr},
		{true, zInt32Ptr, zInt32Ptr},
		{true, zInt64Ptr, zInt64Ptr},
		{true, zUintPtr, zUintPtr},
		{true, zUint8Ptr, zUint8Ptr},
		{true, zUint16Ptr, zUint16Ptr},
		{true, zUint32Ptr, zUint32Ptr},
		{true, zUint64Ptr, zUint64Ptr},
		{true, zUintptrPtr, zUintptrPtr},
		{true, zFloat32Ptr, zFloat32Ptr},
		{true, zFloat64Ptr, zFloat64Ptr},
		{true, zArray0Ptr, zArray0Ptr},
		{true, zArray1Ptr, zArray1Ptr},
		{true, zChanPtr, zChanPtr},
		{true, zFuncPtr, zFuncPtr},
		{true, zIfacePtr, zIfacePtr},
		{true, zMapPtr, zMapPtr},
		{true, zSlicePtr, zSlicePtr},
		{true, zStringPtr, zStringPtr},
		{true, zStructPtr, zStructPtr},
		{true, zMyInt, zMyInt},
		{true, zMyString, zBool},
		{false, zMyString, xString},
		{true, zJSON, zBool},
		{false, zJSON, xUint8},
		{true, zJSONPtr, zJSONPtr},
		{true, zTime, zTime},
	}
	for i, v := range failures {
		msg := fmt.Sprintf("case %d: %#v, %#v", i, v.actual, v.expected)
		if v.panic {
			t.Panic(func() { t.Contains(v.actual, v.expected) }, msg)
		} else {
			t.NotContains(v.actual, v.expected, msg)
		}
	}

	t.Contains("", "")
	t.Contains("Test", "")
	t.Contains(myString("Test"), "es")
	t.Contains([...]time.Time{zTime, xTime, xTimeEST}, xTime)
	t.Contains([]*time.Time{&zTime, &xTime, &xTimeEST}, &xTime)
	t.Contains([]byte("Test"), byte('e'))
	t.Contains([]rune("Test"), 'e')
	t.Contains(map[int]string{2: "two", 5: "five", 10: "ten"}, "five")
	t.Contains(map[string]int{"two": 2, "five": 5, "ten": 10}, 5)
	t.NotContains(map[string]int{"two": 2, "five": 5, "ten": 10}, 0)
}

func TestCheckerHasKey(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	failures := []struct {
		panic    bool
		actual   any
		expected any
	}{
		{true, nil, nil},
		{true, zBool, zBool},
		{true, zInt, zInt},
		{true, zInt8, zInt8},
		{true, zInt16, zInt16},
		{true, zInt32, zInt32},
		{true, zInt64, zInt64},
		{true, zUint, zUint},
		{true, zUint8, zUint8},
		{true, zUint16, zUint16},
		{true, zUint32, zUint32},
		{true, zUint64, zUint64},
		{true, zUintptr, zUintptr},
		{true, zFloat32, zFloat32},
		{true, zFloat64, zFloat64},
		{true, zArray0, zArray0},
		{true, zArray1, zArray1},
		{true, zChan, zChan},
		{true, zFunc, zFunc},
		{true, zIface, zIface},
		{true, zMap, zBool},
		{false, zMap, zInt},
		{true, zSlice, zSlice},
		{true, zString, zString},
		{true, zStruct, zStruct},
		{true, zBoolPtr, zBoolPtr},
		{true, zIntPtr, zIntPtr},
		{true, zInt8Ptr, zInt8Ptr},
		{true, zInt16Ptr, zInt16Ptr},
		{true, zInt32Ptr, zInt32Ptr},
		{true, zInt64Ptr, zInt64Ptr},
		{true, zUintPtr, zUintPtr},
		{true, zUint8Ptr, zUint8Ptr},
		{true, zUint16Ptr, zUint16Ptr},
		{true, zUint32Ptr, zUint32Ptr},
		{true, zUint64Ptr, zUint64Ptr},
		{true, zUintptrPtr, zUintptrPtr},
		{true, zFloat32Ptr, zFloat32Ptr},
		{true, zFloat64Ptr, zFloat64Ptr},
		{true, zArray0Ptr, zArray0Ptr},
		{true, zArray1Ptr, zArray1Ptr},
		{true, zChanPtr, zChanPtr},
		{true, zFuncPtr, zFuncPtr},
		{true, zIfacePtr, zIfacePtr},
		{true, zMapPtr, zMapPtr},
		{true, zSlicePtr, zSlicePtr},
		{true, zStringPtr, zStringPtr},
		{true, zStructPtr, zStructPtr},
		{true, zMyInt, zMyInt},
		{true, zMyString, zMyString},
		{true, zJSON, zJSON},
		{true, zJSONPtr, zJSONPtr},
		{true, zTime, zTime},
	}
	for i, v := range failures {
		msg := fmt.Sprintf("case %d: %#v, %#v", i, v.actual, v.expected)
		if v.panic {
			t.Panic(func() { t.HasKey(v.actual, v.expected) }, msg)
		} else {
			t.NotHasKey(v.actual, v.expected, msg)
		}
	}

	t.HasKey(map[int]string{2: "two", 5: "five", 10: "ten"}, 5)
	t.HasKey(map[string]int{"two": 2, "five": 5, "ten": 10}, "five")
	t.NotHasKey(map[string]int{"two": 2, "five": 5, "ten": 10}, "")
}

func TestCheckerSortEqual(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	type point struct{ x, y int }

	// Basic reordering.
	t.SortEqual([]int{1, 2, 3}, []int{3, 1, 2})
	todo.NotSortEqual([]int{1, 2, 3}, []int{3, 1, 2})

	// Duplicates counted (multiset, not set) equality.
	todo.SortEqual([]int{1, 1, 2}, []int{1, 2, 2})
	t.NotSortEqual([]int{1, 1, 2}, []int{1, 2, 2})
	t.SortEqual([]int{1, 1, 2}, []int{2, 1, 1})
	todo.NotSortEqual([]int{1, 1, 2}, []int{2, 1, 1})

	// Nil and empty slices are equal (unlike DeepEqual).
	t.SortEqual([]int(nil), []int{})
	todo.NotSortEqual([]int(nil), []int{})
	t.SortEqual([]int(nil), []int(nil))
	t.SortEqual([]int{}, []int{})

	// Different length.
	todo.SortEqual([]int{1, 2}, []int{1, 2, 3})
	t.NotSortEqual([]int{1, 2}, []int{1, 2, 3})

	// Arrays.
	t.SortEqual([3]int{1, 2, 3}, [3]int{3, 2, 1})

	// Unsortable elements are fine: matching does not sort them.
	t.SortEqual(
		[]point{{1, 2}, {3, 4}},
		[]point{{3, 4}, {1, 2}},
	)
	todo.SortEqual(
		[]point{{1, 2}, {3, 4}},
		[]point{{3, 4}, {1, 3}},
	)
	t.NotSortEqual(
		[]point{{1, 2}, {3, 4}},
		[]point{{3, 4}, {1, 3}},
	)

	// []any with mixed element types, compared like DeepEqual.
	t.SortEqual([]any{1, "a", true}, []any{true, 1, "a"})

	// time.Time elements compared using .Equal (DeepEqual semantics), not ==.
	t.SortEqual([]time.Time{xTime}, []time.Time{xTimeEST})

	// Panics on non-slice/array actual or expected (incl. untyped nil).
	t.PanicMatch(func() { t.SortEqual(42, []int{1}) }, "actual is not a slice or array")
	t.PanicMatch(func() { t.SortEqual(nil, []int{1}) }, "actual is not a slice or array")
	t.PanicMatch(func() { t.SortEqual([]int{1}, 42) }, "expected is not a slice or array")
	t.PanicMatch(func() { t.SortEqual([]int{1}, nil) }, "expected is not a slice or array")
}

func TestCheckerSubset(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	// Slice subset, multiset semantics, order-insensitive.
	t.Subset([]int{1, 2, 3}, []int{2, 1})
	todo.NotSubset([]int{1, 2, 3}, []int{2, 1})
	t.Subset([]int{1, 2, 3}, []int{})
	t.Subset([]int{1, 2, 3}, []int(nil))

	// Diverges from testify: duplicates in expected are counted,
	// so [1] is NOT a subset of [1,1] and [1,1] IS a subset of [1,1,2].
	todo.Subset([]int{1}, []int{1, 1})
	t.NotSubset([]int{1}, []int{1, 1})
	t.Subset([]int{1, 1, 2}, []int{1, 1})

	// Missing element.
	todo.Subset([]int{1, 2}, []int{3})
	t.NotSubset([]int{1, 2}, []int{3})

	// Map subset: every expected key exists in actual with an equal value.
	t.Subset(
		map[string]int{"a": 1, "b": 2, "c": 3},
		map[string]int{"a": 1, "b": 2},
	)
	// Missing key.
	todo.Subset(
		map[string]int{"a": 1},
		map[string]int{"a": 1, "b": 2},
	)
	t.NotSubset(
		map[string]int{"a": 1},
		map[string]int{"a": 1, "b": 2},
	)
	// Value mismatch.
	todo.Subset(
		map[string]int{"a": 1, "b": 99},
		map[string]int{"a": 1, "b": 2},
	)
	t.NotSubset(
		map[string]int{"a": 1, "b": 99},
		map[string]int{"a": 1, "b": 2},
	)
	// Empty/nil expected map is a subset of anything.
	t.Subset(map[string]int{"a": 1}, make(map[string]int))
	t.Subset(map[string]int{"a": 1}, map[string]int(nil))

	// Kind mismatches panic.
	t.PanicMatch(func() { t.Subset(make(map[string]int), []int{1}) }, "actual is not a slice or array")
	t.PanicMatch(func() { t.Subset([]int{1}, make(map[string]int)) }, "actual is not a map")
	t.PanicMatch(func() { t.Subset([]int{1}, 42) }, "expected is not a slice, array or map")
	t.PanicMatch(func() { t.Subset([]int{1}, nil) }, "expected is not a slice, array or map")
}

// TestElemEqualUsesRegisteredChecker verifies SortEqual/Subset element comparison
// consults the EqualChecker registry (see RegisterEqualChecker) before falling back
// to DeepEqual, same as DeepEqual/NotDeepEqual do.
//
// Deliberately not parallel: it mutates the process-global equal checker registry,
// so it must run to completion (incl. its Cleanup-based reset) before any
// parallel test in this package touches DeepEqual/SortEqual/Subset.
//
//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestElemEqualUsesRegisteredChecker(tt *testing.T) {
	t := check.T(tt)
	t.Cleanup(check.ResetEqualCheckers)

	type fakeElem struct{ id int }

	// Claims only fakeElem pairs and treats them as always equal - this would
	// be false under plain DeepEqual, proving the checker (not DeepEqual) decided.
	check.RegisterEqualChecker(func(actual, expected any) (equal, ok bool) {
		_, okA := actual.(fakeElem)
		_, okB := expected.(fakeElem)
		if !okA || !okB {
			return false, false
		}
		return true, true
	})

	t.SortEqual([]fakeElem{{1}, {2}}, []fakeElem{{3}, {4}})
	t.Subset([]fakeElem{{1}}, []fakeElem{{99}})
}

func TestFileDirExists(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	t.Must(t.Nil(os.WriteFile(file, nil, 0o600)))
	missing := filepath.Join(dir, "missing")

	t.FileExists(file)
	todo.NotFileExists(file)
	t.NotFileExists(dir)
	todo.FileExists(dir)
	t.NotFileExists(missing)
	todo.FileExists(missing)

	t.DirExists(dir)
	todo.NotDirExists(dir)
	t.NotDirExists(file)
	todo.DirExists(file)
	t.NotDirExists(missing)
	todo.DirExists(missing)
}

func TestCheckerZero(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	cases := []struct {
		zero    any
		notzero any
	}{
		{zBool, xBool},
		{zInt, xInt},
		{zInt8, xInt8},
		{zInt16, xInt16},
		{zInt32, xInt32},
		{zInt64, xInt64},
		{zUint, xUint},
		{zUint8, xUint8},
		{zUint16, xUint16},
		{zUint32, xUint32},
		{zUint64, xUint64},
		{zUintptr, xUintptr},
		{zFloat32, xFloat32},
		{zFloat64, xFloat64},
		{zArray0, xArray1},
		{zArray1, xArray1},
		{zChan, xChan},
		{zFunc, xFunc},
		{zIface, xIface},
		{zMap, xMap},
		{zSlice, xSlice},
		{zString, xString},
		{zStruct, xStruct},
		{zBoolPtr, xBoolPtr},
		{zIntPtr, xIntPtr},
		{zInt8Ptr, xInt8Ptr},
		{zInt16Ptr, xInt16Ptr},
		{zInt32Ptr, xInt32Ptr},
		{zInt64Ptr, xInt64Ptr},
		{zUintPtr, xUintPtr},
		{zUint8Ptr, xUint8Ptr},
		{zUint16Ptr, xUint16Ptr},
		{zUint32Ptr, xUint32Ptr},
		{zUint64Ptr, xUint64Ptr},
		{zUintptrPtr, xUintptrPtr},
		{zFloat32Ptr, xFloat32Ptr},
		{zFloat64Ptr, xFloat64Ptr},
		{zArray0Ptr, xArray1Ptr},
		{zArray1Ptr, xArray1Ptr},
		{zChanPtr, xChanPtr},
		{zFuncPtr, xFuncPtr},
		{zIfacePtr, xIfacePtr},
		{zMapPtr, xMapPtr},
		{zSlicePtr, xSlicePtr},
		{zStringPtr, xStringPtr},
		{zStructPtr, xStructPtr},
		{zMyInt, xMyInt},
		{zMyString, xMyString},
		{zJSON, xJSON},
		{zJSONPtr, xJSONPtr},
		{zTime, xTime},
		{nil, vChan},
		{nil, vFunc},
		{vIface, xIface},
		{nil, vMap},
		{nil, vSlice},
		{[0][]int{}, [1][]int{{1}}},
		{[2][]int{nil, nil}, [2][]int{nil, {}}},
		{[2][2][2]int{1: {1: {1: 0}}}, [2][2][2]int{1: {1: {1: 1}}}},
	}
	for i, v := range cases {
		msg := fmt.Sprintf("case %d: %#v, %#v", i, v.zero, v.notzero)
		t.Zero(v.zero, msg)
		todo.Zero(v.notzero, msg)
		t.NotZero(v.notzero, msg)
		todo.NotZero(v.zero, msg)
	}

	t.Zero(nil)
	todo.NotZero(nil)
}

func TestCheckerLen(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)

	cases := []struct {
		panic  bool
		actual any
		len    int
	}{
		{true, nil, 0},
		{true, zBool, 0},
		{true, zInt, 0},
		{true, zInt8, 0},
		{true, zInt16, 0},
		{true, zInt32, 0},
		{true, zInt64, 0},
		{true, zUint, 0},
		{true, zUint8, 0},
		{true, zUint16, 0},
		{true, zUint32, 0},
		{true, zUint64, 0},
		{true, zUintptr, 0},
		{true, zFloat32, 0},
		{true, zFloat64, 0},
		{false, zArray0, 1},
		{false, zArray1, 0},
		{false, zChan, 1},
		{true, zFunc, 0},
		{true, zIface, 0},
		{false, zMap, 1},
		{false, zSlice, 1},
		{false, zString, 1},
		{true, zStruct, 0},
		{true, zBoolPtr, 0},
		{true, zIntPtr, 0},
		{true, zInt8Ptr, 0},
		{true, zInt16Ptr, 0},
		{true, zInt32Ptr, 0},
		{true, zInt64Ptr, 0},
		{true, zUintPtr, 0},
		{true, zUint8Ptr, 0},
		{true, zUint16Ptr, 0},
		{true, zUint32Ptr, 0},
		{true, zUint64Ptr, 0},
		{true, zUintptrPtr, 0},
		{true, zFloat32Ptr, 0},
		{true, zFloat64Ptr, 0},
		// {true, zArray0Ptr, 0},
		// {true, zArray1Ptr, 0},
		{true, zChanPtr, 0},
		{true, zFuncPtr, 0},
		{true, zIfacePtr, 0},
		{true, zMapPtr, 0},
		{true, zSlicePtr, 0},
		{true, zStringPtr, 0},
		{true, zStructPtr, 0},
		{true, zMyInt, 0},
		{false, zMyString, 1},
		{false, zJSON, 1},
		{true, zJSONPtr, 0},
		{true, zTime, 0},
	}
	for _, v := range cases {
		t.Run("", func(tt *testing.T) {
			tt.Parallel()
			t := check.T(tt)
			todo := t.TODO()

			if v.panic {
				t.Panic(func() { t.Len(v.actual, v.len) })
			} else {
				todo.Len(v.actual, v.len)
				t.NotLen(v.actual, v.len)
			}
		})
	}

	todo := t.TODO()

	t.Len(zArray0, 0)
	t.Len(zArray1, 1)

	c := make(chan int, 5)
	t.Len(c, 0)
	todo.NotLen(c, 0)
	c <- 42
	t.Len(c, 1)
	todo.NotLen(c, 1)

	m := make(map[string]int, 10)
	t.Len(m, 0)
	m["one"] = 1
	m["ten"] = 10
	t.Len(m, 2)

	t.Len(json.RawMessage("тест"), 8)
	t.Len([]rune("тест"), 4)

	t.Len(myString("test"), 4)
	t.Len("тест", 8)
}

func TestCheckerOrdered(t *testing.T) {
	t.Parallel()
	cases := []struct {
		panic bool
		min   any
		mid   any
		max   any
	}{
		{true, nil, nil, nil},
		{true, zBool, xBool, xBool},
		{false, xInt, xInt + 1, xInt + 2},
		{false, xInt8, xInt8 + 1, xInt8 + 2},
		{false, xInt16, xInt16 + 1, xInt16 + 2},
		{false, xInt32, xInt32 + 1, xInt32 + 2},
		{false, xInt64, xInt64 + 1, xInt64 + 2},
		{false, xUint, xUint + 1, xUint + 2},
		{false, xUint8, xUint8 + 1, xUint8 + 2},
		{false, xUint16, xUint16 + 1, xUint16 + 2},
		{false, xUint32, xUint32 + 1, xUint32 + 2},
		{false, xUint64, xUint64 + 1, xUint64 + 2},
		{false, xUintptr, xUintptr + 1, xUintptr + 2},
		{false, xFloat32, xFloat32 + 1, xFloat32 + 2},
		{false, xFloat64, xFloat64 + 1, xFloat64 + 2},
		{true, zArray0, zArray0, zArray0},
		{true, zArray1, xArray1, xArray1},
		{true, zChan, xChan, xChan},
		{true, zFunc, xFunc, xFunc},
		{true, zIface, xIface, xIface},
		{true, zMap, xMap, xMap},
		{true, zSlice, xSlice, xSlice},
		{false, xString, xString + "1", xString + "2"},
		{true, zStruct, xStruct, xStruct},
		{true, zBoolPtr, xBoolPtr, xBoolPtr},
		{true, zIntPtr, xIntPtr, xIntPtr},
		{true, zInt8Ptr, xInt8Ptr, xInt8Ptr},
		{true, zInt16Ptr, xInt16Ptr, xInt16Ptr},
		{true, zInt32Ptr, xInt32Ptr, xInt32Ptr},
		{true, zInt64Ptr, xInt64Ptr, xInt64Ptr},
		{true, zUintPtr, xUintPtr, xUintPtr},
		{true, zUint8Ptr, xUint8Ptr, xUint8Ptr},
		{true, zUint16Ptr, xUint16Ptr, xUint16Ptr},
		{true, zUint32Ptr, xUint32Ptr, xUint32Ptr},
		{true, zUint64Ptr, xUint64Ptr, xUint64Ptr},
		{true, zUintptrPtr, xUintptrPtr, xUintptrPtr},
		{true, zFloat32Ptr, xFloat32Ptr, xFloat32Ptr},
		{true, zFloat64Ptr, xFloat64Ptr, xFloat64Ptr},
		{true, zArray0Ptr, zArray0Ptr, zArray0Ptr},
		{true, zArray1Ptr, xArray1Ptr, xArray1Ptr},
		{true, zChanPtr, xChanPtr, xChanPtr},
		{true, zFuncPtr, xFuncPtr, xFuncPtr},
		{true, zIfacePtr, xIfacePtr, xIfacePtr},
		{true, zMapPtr, xMapPtr, xMapPtr},
		{true, zSlicePtr, xSlicePtr, xSlicePtr},
		{true, zStringPtr, xStringPtr, xStringPtr},
		{true, zStructPtr, xStructPtr, xStructPtr},
		{false, xMyInt, xMyInt + 1, xMyInt + 2},
		{false, xMyString, xMyString + "1", xMyString + "2"},
		{true, xJSON, xJSON, xJSON},
		{true, xJSONPtr, xJSONPtr, xJSONPtr},
		{false, xTime, xTime.Add(time.Millisecond), xTime.Add(time.Second)},
	}

	t.Run("Less", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()
		for _, v := range cases {
			actual, expected := v.min, v.max
			if v.panic {
				t.Panic(func() { t.Less(actual, expected) })
				t.Panic(func() { t.LT(actual, expected) })
				t.Panic(func() { t.LessOrEqual(actual, expected) })
				t.Panic(func() { t.LE(actual, expected) })
			} else {
				t.Less(actual, expected)
				t.LT(actual, expected)
				t.LessOrEqual(actual, expected)
				t.LessOrEqual(actual, actual)
				t.LE(actual, expected)
				t.LE(actual, actual)

				actual, expected = expected, actual
				todo.Less(actual, expected)
				todo.LT(actual, expected)
				todo.LessOrEqual(actual, expected)
				todo.LE(actual, expected)
			}
		}
	})

	t.Run("Greater", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()
		for _, v := range cases {
			actual, expected := v.min, v.max
			if v.panic {
				t.Panic(func() { t.Greater(actual, expected) })
				t.Panic(func() { t.GT(actual, expected) })
				t.Panic(func() { t.GreaterOrEqual(actual, expected) })
				t.Panic(func() { t.GE(actual, expected) })
			} else {
				todo.Greater(actual, expected)
				todo.GT(actual, expected)
				todo.GreaterOrEqual(actual, expected)
				todo.GE(actual, expected)

				actual, expected = expected, actual
				t.Greater(actual, expected)
				t.GT(actual, expected)
				t.GreaterOrEqual(actual, expected)
				t.GreaterOrEqual(actual, actual)
				t.GE(actual, expected)
				t.GE(actual, actual)
			}
		}
	})

	t.Run("Between", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()
		for _, v := range cases {
			minimum, middle, maximum := v.min, v.mid, v.max
			if v.panic {
				t.Panic(func() { t.Between(middle, minimum, maximum) })
				t.Panic(func() { t.BetweenOrEqual(middle, minimum, maximum) })
				t.Panic(func() { t.NotBetween(minimum, middle, maximum) })
				t.Panic(func() { t.NotBetweenOrEqual(minimum, middle, maximum) })
			} else {
				t.Between(middle, minimum, maximum)
				t.BetweenOrEqual(middle, minimum, maximum)
				t.BetweenOrEqual(middle, middle, maximum)
				t.BetweenOrEqual(middle, minimum, middle)
				todo.NotBetween(middle, minimum, maximum)
				todo.NotBetweenOrEqual(middle, minimum, maximum)
				todo.NotBetweenOrEqual(middle, middle, maximum)
				todo.NotBetweenOrEqual(middle, minimum, middle)
				t.NotBetween(minimum, middle, maximum)
				t.NotBetween(maximum, minimum, middle)
				t.NotBetweenOrEqual(minimum, middle, maximum)
				t.NotBetweenOrEqual(maximum, minimum, middle)
				todo.Between(minimum, middle, maximum)
				todo.Between(maximum, minimum, middle)
				todo.BetweenOrEqual(minimum, middle, maximum)
				todo.BetweenOrEqual(maximum, minimum, middle)
			}
		}
	})
}

func TestCheckerApprox(t *testing.T) {
	t.Parallel()
	cases := []struct {
		panic    bool
		actual   any
		expected any
		delta    any
		smape    float64
	}{
		{true, nil, nil, nil, 0},
		{true, zBool, xBool, xBool, 0},
		{false, xInt, xInt + 5, 7, 10.0},
		{false, xInt8, xInt8 + 5, 7, 50.0},
		{false, xInt16, xInt16 + 5, 7, 20.0},
		{false, xInt32, xInt32 + 5, 7, 10.0},
		{false, xInt64, xInt64 + 5, 7, 5.0},
		{false, xUint, xUint + 5, uint(7), 6.0},
		{false, xUint8, xUint8 + 5, uint(7), 30.0},
		{false, xUint16, xUint16 + 5, uint(7), 20.0},
		{false, xUint32, xUint32 + 5, uint(7), 10.0},
		{false, xUint64, xUint64 + 5, uint(7), 5.0},
		{false, xUintptr, xUintptr + 5, uint(7), 0.0000001},
		{false, xFloat32, xFloat32 - 5, 7.0, 50.0},
		{false, xFloat64, xFloat64 + 5, 7.0, 33.0},
		{true, zArray0, zArray0, zArray0, 0},
		{true, zArray1, xArray1, xArray1, 0},
		{true, zChan, xChan, xChan, 0},
		{true, zFunc, xFunc, xFunc, 0},
		{true, zIface, xIface, xIface, 0},
		{true, zMap, xMap, xMap, 0},
		{true, zSlice, xSlice, xSlice, 0},
		{true, xString, xString, xString, 0},
		{true, zStruct, xStruct, xStruct, 0},
		{true, zBoolPtr, xBoolPtr, xBoolPtr, 0},
		{true, zIntPtr, xIntPtr, xIntPtr, 0},
		{true, zInt8Ptr, xInt8Ptr, xInt8Ptr, 0},
		{true, zInt16Ptr, xInt16Ptr, xInt16Ptr, 0},
		{true, zInt32Ptr, xInt32Ptr, xInt32Ptr, 0},
		{true, zInt64Ptr, xInt64Ptr, xInt64Ptr, 0},
		{true, zUintPtr, xUintPtr, xUintPtr, 0},
		{true, zUint8Ptr, xUint8Ptr, xUint8Ptr, 0},
		{true, zUint16Ptr, xUint16Ptr, xUint16Ptr, 0},
		{true, zUint32Ptr, xUint32Ptr, xUint32Ptr, 0},
		{true, zUint64Ptr, xUint64Ptr, xUint64Ptr, 0},
		{true, zUintptrPtr, xUintptrPtr, xUintptrPtr, 0},
		{true, zFloat32Ptr, xFloat32Ptr, xFloat32Ptr, 0},
		{true, zFloat64Ptr, xFloat64Ptr, xFloat64Ptr, 0},
		{true, zArray0Ptr, zArray0Ptr, zArray0Ptr, 0},
		{true, zArray1Ptr, xArray1Ptr, xArray1Ptr, 0},
		{true, zChanPtr, xChanPtr, xChanPtr, 0},
		{true, zFuncPtr, xFuncPtr, xFuncPtr, 0},
		{true, zIfacePtr, xIfacePtr, xIfacePtr, 0},
		{true, zMapPtr, xMapPtr, xMapPtr, 0},
		{true, zSlicePtr, xSlicePtr, xSlicePtr, 0},
		{true, zStringPtr, xStringPtr, xStringPtr, 0},
		{true, zStructPtr, xStructPtr, xStructPtr, 0},
		{false, xMyInt, xMyInt + 5, 7, 0.01},
		{true, xMyString, xMyString, xMyString, 0},
		{true, xJSON, xJSON, xJSON, 0},
		{true, xJSONPtr, xJSONPtr, xJSONPtr, 0},
		{false, xTime, xTime.Add(5 * time.Second), 7 * time.Second, 0},
	}

	t.Run("Delta", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()
		for _, v := range cases {
			if v.panic {
				t.Panic(func() { t.InDelta(v.actual, v.expected, v.delta) })
				t.Panic(func() { t.NotInDelta(v.actual, v.expected, v.delta) })
			} else {
				t.InDelta(v.actual, v.expected, v.delta)
				t.InDelta(v.expected, v.actual, v.delta)
				todo.NotInDelta(v.actual, v.expected, v.delta)
				todo.NotInDelta(v.expected, v.actual, v.delta)
				t.NotInDelta(v.actual, v.expected, half(v.delta))
				t.NotInDelta(v.expected, v.actual, half(v.delta))
				todo.InDelta(v.actual, v.expected, half(v.delta))
				todo.InDelta(v.expected, v.actual, half(v.delta))
			}
		}
	})

	t.Run("SMAPE", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()
		for _, v := range cases {
			if v.panic || v.smape == 0 {
				t.Panic(func() { t.InSMAPE(v.actual, v.expected, v.smape) })
				t.Panic(func() { t.NotInSMAPE(v.actual, v.expected, v.smape) })
			} else {
				t.InSMAPE(v.actual, v.expected, v.smape)
				t.InSMAPE(v.expected, v.actual, v.smape)
				todo.NotInSMAPE(v.actual, v.expected, v.smape)
				todo.NotInSMAPE(v.expected, v.actual, v.smape)
				t.NotInSMAPE(v.actual, v.expected, half(v.smape).(float64))
				t.NotInSMAPE(v.expected, v.actual, half(v.smape).(float64))
				todo.InSMAPE(v.actual, v.expected, half(v.smape).(float64))
				todo.InSMAPE(v.expected, v.actual, half(v.smape).(float64))
			}
		}

		t.InSMAPE(0, 0, 0.5)
		t.InSMAPE(0.0, 0.0, 0.5)
	})
}

func TestCheckerApproxOverflow(t *testing.T) {
	t.Parallel()

	// Regression tests for integer overflow in isInDelta:
	// the old e-d/e+d range check wrapped at uint/int64 extremes.

	type overflowCase struct {
		actual   any
		expected any
		delta    any
		inDelta  bool // whether InDelta should return true
	}

	overflowCases := []overflowCase{
		// Unsigned: basic cases with wrap-around risk
		{uint(0), uint(0), uint(1), true},
		{uint(1), uint(0), uint(1), true},
		{uint64(math.MaxUint64), uint64(math.MaxUint64), uint64(1), true},
		{uint64(math.MaxUint64), uint64(0), uint64(1), false},
		// Signed: MaxInt64 boundaries
		{int64(math.MaxInt64), int64(math.MaxInt64 - 1), int64(2), true},
		{int64(math.MinInt64), int64(math.MinInt64 + 1), int64(2), true},
		// Signed: opposite signs — large actual distance, small delta
		{int64(math.MaxInt64), int64(math.MinInt64), int64(1), false},
		{int64(math.MinInt64), int64(math.MaxInt64), int64(1), false},
	}

	for _, v := range overflowCases {
		t.Run("", func(tt *testing.T) {
			tt.Parallel()
			c := check.T(tt)
			if v.inDelta {
				c.InDelta(v.actual, v.expected, v.delta)
				c.InDelta(v.expected, v.actual, v.delta)
			} else {
				c.NotInDelta(v.actual, v.expected, v.delta)
				c.NotInDelta(v.expected, v.actual, v.delta)
			}
		})
	}
}

func half(v any) any {
	if v, ok := v.(time.Duration); ok {
		return v / 2
	}
	switch val := reflect.ValueOf(v); val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() / 2
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() / 2
	case reflect.Float32, reflect.Float64:
		return val.Float() / 2
	case reflect.Complex128, reflect.Complex64, // ???
		// No meaningful "half":
		reflect.Array, reflect.Slice, reflect.Map, reflect.Struct, reflect.Bool, reflect.String,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Invalid,
		reflect.Pointer, reflect.UnsafePointer:
	}
	panic(fmt.Sprintf("can't get half from %#v", v))
}

func TestCheckerSubstring(t *testing.T) {
	t.Parallel()
	cases := []struct {
		panic  bool
		actual any
		prefix string
		suffix string
	}{
		{true, xBool, "", ""},
		{true, xInt, "", ""},
		{true, xInt8, "", ""},
		{true, xInt16, "", ""},
		{true, xInt32, "", ""},
		{true, xInt64, "", ""},
		{true, xUint, "", ""},
		{true, xUint8, "", ""},
		{true, xUint16, "", ""},
		{true, xUint32, "", ""},
		{true, xUint64, "", ""},
		{true, xUintptr, "", ""},
		{true, xFloat32, "", ""},
		{true, xFloat64, "", ""},
		{true, zArray0, "", ""},
		{true, xArray1, "", ""},
		{true, xChan, "", ""},
		{true, xFunc, "", ""},
		{true, xIface, "", ""},
		{true, xMap, "", ""},
		{true, xSlice, "", ""},
		{false, xString, "<ni", "il>"},
		{true, xStruct, "", ""},
		{true, xBoolPtr, "", ""},
		{true, xIntPtr, "", ""},
		{true, xInt8Ptr, "", ""},
		{true, xInt16Ptr, "", ""},
		{true, xInt32Ptr, "", ""},
		{true, xInt64Ptr, "", ""},
		{true, xUintPtr, "", ""},
		{true, xUint8Ptr, "", ""},
		{true, xUint16Ptr, "", ""},
		{true, xUint32Ptr, "", ""},
		{true, xUint64Ptr, "", ""},
		{true, xUintptrPtr, "", ""},
		{true, xFloat32Ptr, "", ""},
		{true, xFloat64Ptr, "", ""},
		{true, zArray0Ptr, "", ""},
		{true, xArray1Ptr, "", ""},
		{true, xChanPtr, "", ""},
		{true, xFuncPtr, "", ""},
		{true, xIfacePtr, "", ""},
		{true, xMapPtr, "", ""},
		{true, xSlicePtr, "", ""},
		{true, xStringPtr, "", ""},
		{true, xStructPtr, "", ""},
		{true, xMyInt, "", ""},
		{false, xMyString, "xy", "yz"},
		{false, xJSON, "{", "}"},
		{true, xJSONPtr, "", ""},
		{false, zTime, "0001-01-01", "UTC"},
		{false, []byte("String"), "Str", "ing"},
		{false, []rune("Symbol"), "Sym", "bol"},
		{false, time.Sunday, "Sun", "day"},
		{false, io.EOF, "EO", "OF"},
	}

	substrings := []struct {
		prefix any
		suffix any
	}{
		{time.Sunday.String(), time.Monday.String()},
		{[]byte(time.Sunday.String()), []byte(time.Monday.String())},
		{[]rune(time.Sunday.String()), []rune(time.Monday.String())},
		{time.Sunday, time.Monday},
		{errors.New(time.Sunday.String()), errors.New(time.Monday.String())},
	}

	t.Run("HasPrefix", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		for i, v := range cases {
			msg := fmt.Sprintf("case %d: %#v, %#v, %#v", i, v.actual, v.prefix, v.suffix)
			if v.panic {
				t.Panic(func() { t.HasPrefix(v.actual, v.prefix) }, msg)
				t.Panic(func() { t.NotHasPrefix(v.actual, v.prefix) }, msg)
				t.Panic(func() { t.HasPrefix("", v.actual) }, msg)
				t.Panic(func() { t.NotHasPrefix("", v.actual) }, msg)
			} else {
				t.HasPrefix(v.actual, v.prefix, msg)
				todo.HasPrefix(v.actual, v.suffix, msg)
				t.NotHasPrefix(v.actual, v.suffix, msg)
				todo.NotHasPrefix(v.actual, v.prefix, msg)
			}
		}

		for _, v := range substrings {
			t.HasPrefix("Sunday Monday", v.prefix)
			todo.NotHasPrefix("Sunday Monday", v.prefix)
		}

		todo.HasPrefix(nil, "")
		t.NotHasPrefix(nil, "")
		todo.HasPrefix("", nil)
		t.NotHasPrefix("", nil)

		t.HasPrefix("", "")
		todo.NotHasPrefix("", "")
		t.HasPrefix("x", "")
		todo.NotHasPrefix("x", "")
	})

	t.Run("HasSuffix", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		for i, v := range cases {
			msg := fmt.Sprintf("case %d: %#v, %#v, %#v", i, v.actual, v.suffix, v.suffix)
			if v.panic {
				t.Panic(func() { t.HasSuffix(v.actual, v.suffix) }, msg)
				t.Panic(func() { t.NotHasSuffix(v.actual, v.suffix) }, msg)
				t.Panic(func() { t.HasSuffix("", v.actual) }, msg)
				t.Panic(func() { t.NotHasSuffix("", v.actual) }, msg)
			} else {
				t.HasSuffix(v.actual, v.suffix, msg)
				todo.HasSuffix(v.actual, v.prefix, msg)
				t.NotHasSuffix(v.actual, v.prefix, msg)
				todo.NotHasSuffix(v.actual, v.suffix, msg)
			}
		}

		for _, v := range substrings {
			t.HasSuffix("Sunday Monday", v.suffix)
			todo.NotHasSuffix("Sunday Monday", v.suffix)
		}

		todo.HasSuffix(nil, "")
		t.NotHasSuffix(nil, "")
		todo.HasSuffix("", nil)
		t.NotHasSuffix("", nil)

		t.HasSuffix("", "")
		todo.NotHasSuffix("", "")
		t.HasSuffix("x", "")
		todo.NotHasSuffix("x", "")
	})
}

func TestJSONEqual(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	cases := []struct {
		panic bool
		json  any
	}{
		{false, nil},
		{true, zBool},
		{true, zInt},
		{true, zInt8},
		{true, zInt16},
		{true, zInt32},
		{true, zInt64},
		{true, zUint},
		{true, zUint8},
		{true, zUint16},
		{true, zUint32},
		{true, zUint64},
		{true, zUintptr},
		{true, zFloat32},
		{true, zFloat64},
		{true, zArray0},
		{true, zArray1},
		{true, zChan},
		{true, zFunc},
		{false, zIface}, // nil
		{true, zMap},
		{true, zSlice},
		{false, zString},
		{true, zStruct},
		{true, zBoolPtr},
		{true, zIntPtr},
		{true, zInt8Ptr},
		{true, zInt16Ptr},
		{true, zInt32Ptr},
		{true, zInt64Ptr},
		{true, zUintPtr},
		{true, zUint8Ptr},
		{true, zUint16Ptr},
		{true, zUint32Ptr},
		{true, zUint64Ptr},
		{true, zUintptrPtr},
		{true, zFloat32Ptr},
		{true, zFloat64Ptr},
		{true, zArray0Ptr},
		{true, zArray1Ptr},
		{true, zChanPtr},
		{true, zFuncPtr},
		{true, zIfacePtr},
		{true, zMapPtr},
		{true, zSlicePtr},
		{true, zStringPtr},
		{true, zStructPtr},
		{true, zMyInt},
		{false, zMyString},
		{false, zJSON},
		{false, zJSONPtr},
		{true, zTime},
		{false, []byte(nil)},
		{false, []byte{}},
	}
	for i, v := range cases {
		if v.panic {
			t.Panic(func() { t.JSONEqual(v.json, `{}`, i) })
			t.Panic(func() { t.JSONEqual(`{}`, v.json) })
		} else {
			todo.JSONEqual(v.json, v.json)
		}
	}

	invalid := `{"a":1,"b":[2]`
	invalidRaw := json.RawMessage(invalid)
	todo.JSONEqual(invalid, invalid)
	todo.JSONEqual([]byte(invalid), []byte(invalid))
	todo.JSONEqual(&invalidRaw, invalid)
	todo.JSONEqual(&invalidRaw, invalid+"}")
	todo.JSONEqual(invalidRaw, []byte(invalid))
	t.JSONEqual(invalidRaw, invalidRaw)
	t.JSONEqual(&invalidRaw, &invalidRaw)
	t.JSONEqual(&invalidRaw, invalidRaw)
	t.JSONEqual(invalidRaw, &invalidRaw)

	validRaw := json.RawMessage(invalid + "}")
	valid := []any{
		`{ "b" : [ 2],"a" :1}  `,
		[]byte(`  { "b": [2 ],"a": 1}`),
		validRaw,
		&validRaw,
	}
	for _, actual := range valid {
		for _, expected := range valid {
			t.JSONEqual(actual, expected)
		}
	}
}

func TestHasType(tt *testing.T) {
	tt.Parallel()
	t := check.T(tt)
	todo := t.TODO()

	vs := []any{
		zBool,
		zInt,
		zInt8,
		zInt16,
		zInt32,
		zInt64,
		zUint,
		zUint8,
		zUint16,
		zUint32,
		zUint64,
		zUintptr,
		zFloat32,
		zFloat64,
		zArray0,
		zArray1,
		zChan,
		zFunc,
		zIface, // nil
		zMap,
		zSlice,
		zString,
		zStruct,
		zBoolPtr,
		zIntPtr,
		zInt8Ptr,
		zInt16Ptr,
		zInt32Ptr,
		zInt64Ptr,
		zUintPtr,
		zUint8Ptr,
		zUint16Ptr,
		zUint32Ptr,
		zUint64Ptr,
		zUintptrPtr,
		zFloat32Ptr,
		zFloat64Ptr,
		zArray0Ptr,
		zArray1Ptr,
		zChanPtr,
		zFuncPtr,
		zIfacePtr,
		zMapPtr,
		zSlicePtr,
		zStringPtr,
		zStructPtr,
		zMyInt,
		zMyString,
		zJSON,
		zJSONPtr,
		zTime,
	}
	for i, actual := range vs {
		for j, expected := range vs {
			if i == j {
				t.HasType(actual, expected)
				todo.NotHasType(actual, expected)
			} else {
				t.NotHasType(actual, expected)
				todo.HasType(actual, expected)
			}
		}
	}

	t.HasType(vChan, zChan)
	t.HasType(vFunc, zFunc)
	t.HasType(vIface, zIntPtr)
	t.HasType(vMap, zMap)
	t.HasType(vSlice, zSlice)
	var reader io.Reader
	t.HasType(reader, nil)
	t.HasType(&reader, (*io.Reader)(nil))
	t.NotHasType(&reader, nil)
	t.HasType(os.Stdin, (*os.File)(nil))
	t.NotHasType(os.Stdin, &reader)
	t.HasType(true, zBool)
	t.HasType(42, zInt)
	t.HasType("test", zString)
	t.HasType([]byte("test"), []byte(nil))
	t.HasType([]byte("test"), []byte{})
	t.HasType(new(int), zIntPtr)
	t.NotHasType(json.RawMessage([]byte("test")), []byte("test"))
}

func TestCheckers(t *testing.T) {
	t.Parallel()
	t.Run("Err", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)

		cases := []struct {
			err       bool
			deepEqual bool
			equal     bool
			actual    error
			expected  error
		}{
			{true, true, true, nil, nil},
			//nolint:dupword // Commented code.
			// {false, false, false, (*net.OpError)(nil), &net.OpError{}},
			{false, false, false, (*net.OpError)(nil), nil},
			{false, false, false, nil, (*net.OpError)(nil)},
			{true, true, true, (*net.OpError)(nil), (*net.OpError)(nil)},
			{true, true, false, &net.OpError{}, &net.OpError{}},
			{true, true, true, io.EOF, io.EOF},
			{true, true, false, io.EOF, errors.New("EOF")},
			{false, false, false, pkgerrorsNew(), io.EOF},
			{false, false, false, pkgerrorsNew(), errors.New("EOF")},
			{true, true, false, pkgerrorsNew(), pkgerrorsNew()},
			{true, false, false, pkgerrorsWithStack(io.EOF), io.EOF},
			{true, false, false, pkgerrorsWrap(io.EOF, "wrapped"), io.EOF},
			{true, false, false, pkgerrorsWrap(io.EOF, "wrapped"), errors.New("EOF")},
			{true, false, false, pkgerrorsWrap(pkgerrorsWrap(io.EOF, "wrapped"), "wrapped2"), io.EOF},
			{true, false, false, fmt.Errorf("wrapped: %w", io.EOF), io.EOF},
			{true, false, false, fmt.Errorf("wrapped: %w", io.EOF), errors.New("EOF")},
			{false, false, false, fmt.Errorf("wrapped: %w", io.EOF), &myError{"EOF"}},
			{false, false, false, fmt.Errorf("wrapped: %w", &myError{"EOF"}), io.EOF},
			{true, false, false, fmt.Errorf("wrapped[]: %w %w", io.EOF, &myError{"EOF"}), io.EOF},
			{true, false, false, fmt.Errorf("wrapped[]: %w %w", &myError{"EOF"}, io.EOF), io.EOF},
			{true, false, false, fmt.Errorf("wrapped[]: %w %w", &myError{"EOF"}, io.EOF), &myError{"EOF"}},
			{false, false, false, fmt.Errorf("wrapped[]: %w %w", io.EOF, &myError{"EOF"}), &myError{"EOF"}},
			{true, false, false, fmt.Errorf("wrapped2: %w", fmt.Errorf("wrapped: %w", io.EOF)), io.EOF},
			{true, false, false, fmt.Errorf("wrapped2: %w", pkgerrorsWrap(io.EOF, "wrapped")), io.EOF},
			{true, false, false, pkgerrorsWrap(fmt.Errorf("wrapped: %w", io.EOF), "wrapped2"), io.EOF},
			{
				true,
				false,
				false,
				pkgerrorsWrap(
					pkgerrorsWrap(
						fmt.Errorf("wrapped4: %w", fmt.Errorf("wrapped3: %w", pkgerrorsWrap(fmt.Errorf("wrapped: %w", io.EOF), "wrapped2"))),
						"wrapped5",
					),
					"wrapped6",
				),
				io.EOF,
			},
			{false, false, false, io.EOF, &myError{"EOF"}},
		}
		for _, v := range cases {
			t.Run("", func(tt *testing.T) {
				tt.Parallel()
				t := check.T(tt)
				todo := t.TODO()
				if v.err {
					t.Err(v.actual, v.expected)
					todo.NotErr(v.actual, v.expected)
				} else {
					todo.Err(v.actual, v.expected)
					t.NotErr(v.actual, v.expected)
				}
				if v.equal {
					t.Equal(v.actual, v.expected)
				} else {
					t.NotEqual(v.actual, v.expected)
				}
				if v.deepEqual {
					t.DeepEqual(v.actual, v.expected)
				} else {
					t.NotDeepEqual(v.actual, v.expected)
				}
			})
		}
	})

	t.Run("Panic", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		todo.Panic(func() {})
		t.NotPanic(func() {})

		t.Panic(func() { panic(nil) })       //nolint:govet // Testing nil panic.
		todo.NotPanic(func() { panic(nil) }) //nolint:govet // Testing nil panic.

		t.Panic(func() { panic("") })
		t.Panic(func() { panic("oops") })
		t.Panic(func() { panic(t) })
		todo.NotPanic(func() { panic("") })
		todo.NotPanic(func() { panic("oops") })
		todo.NotPanic(func() { panic(t) })
	})

	t.Run("PanicMatch", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		t.Panic(func() { t.PanicMatch(func() { panic(0) }, nil) })
		t.Panic(func() { t.PanicMatch(func() { panic(0) }, t) })
		t.NotPanic(func() { t.PanicMatch(func() { panic(0) }, `0`) })

		todo.PanicMatch(func() {}, ``)
		todo.PanicNotMatch(func() {}, ``)
		todo.PanicMatch(func() {}, `test`)
		todo.PanicNotMatch(func() {}, `test`)

		t.PanicMatch(func() { panic(nil) }, ``)                                     //nolint:govet // Testing nil panic.
		todo.PanicNotMatch(func() { panic(nil) }, ``)                               //nolint:govet // Testing nil panic.
		t.PanicMatch(func() { panic(nil) }, `panic called with nil argument`)       //nolint:govet // Testing nil panic.
		todo.PanicNotMatch(func() { panic(nil) }, `panic called with nil argument`) //nolint:govet // Testing nil panic.
		t.PanicNotMatch(func() { panic(nil) }, `test`)                              //nolint:govet // Testing nil panic.
		todo.PanicMatch(func() { panic(nil) }, `test`)                              //nolint:govet // Testing nil panic.

		t.PanicMatch(func() { panic("") }, regexp.MustCompile(`^$`))
		t.PanicMatch(func() { panic("oops") }, `(?i)Oops`)
		t.PanicMatch(func() { panic(t) }, `^&check.C{`)
		t.PanicNotMatch(func() { panic("") }, regexp.MustCompile(`.`))
		t.PanicNotMatch(func() { panic("oops") }, `(?-i)Oops`)
		todo.PanicNotMatch(func() { panic(t) }, `^&check.C{`)
	})

	t.Run("Implements", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)

		t.Implements(t, (*testing.TB)(nil))
		t.Implements(os.Stdin, (*io.Reader)(nil))
		t.Implements(os.Stdin, &xIface)
		t.Implements(*os.Stdin, (*io.Reader)(nil))
		t.Implements(time.Time{}, (*fmt.Stringer)(nil))
		t.Implements(&time.Time{}, (*fmt.Stringer)(nil))
		t.NotImplements(os.Stdin, (*fmt.Stringer)(nil))
		t.NotImplements(&os.Stdin, (*io.Reader)(nil))
		t.NotImplements(new(int), (*io.Reader)(nil))
	})

	t.Run("ErrIs", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		errA := errors.New("ERR_A")
		errB := errors.New("ERR_B")
		wrappedA := fmt.Errorf("wrapped: %w", errA)

		// errors.Is works through wrapping.
		t.ErrIs(wrappedA, errA)
		todo.NotErrIs(wrappedA, errA)
		t.NotErrIs(wrappedA, errB)
		todo.ErrIs(wrappedA, errB)
		// errors.Is does not compare by value.
		todo.ErrIs(errA, errors.New("ERR_A"))
		t.NotErrIs(errA, errors.New("ERR_A"))
		// Nil handling.
		t.ErrIs(nil, nil)
		todo.NotErrIs(nil, nil)
		todo.ErrIs(nil, io.EOF)
		t.NotErrIs(nil, io.EOF)
		todo.ErrIs(io.EOF, nil)
		t.NotErrIs(io.EOF, nil)
	})

	t.Run("ErrAs", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		errA := &targetErr{msg: "ERR_A"}
		errB := errors.New("ERR_B")
		wrappedA := fmt.Errorf("wrapped: %w", errA)

		// ErrAs finds the first error in the chain that matches the target type.
		var got1 *targetErr
		t.ErrAs(wrappedA, &got1)
		t.NotNil(got1)
		t.Equal(got1.msg, "ERR_A")

		var got2 *targetErr
		todo.ErrAs(errB, &got2)
		t.Nil(got2)
		t.NotErrAs(errB, &got2)

		// nil target panics (as errors.As does).
		t.Panic(func() { t.ErrAs(errA, nil) })
	})

	t.Run("NotErrAs", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		errB := errors.New("ERR_B")

		// NotErrAs passes when target type doesn't match.
		var got1 *targetErr
		todo.ErrAs(errB, &got1)
		t.Nil(got1)
		t.NotErrAs(errB, &got1)
		t.Nil(got1)

		// Nil actual: As(nil, target) is false.
		var got2 *targetErr
		todo.ErrAs(nil, &got2)
		t.Nil(got2)
	})

	t.Run("ErrChar", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		// Characterization of Err's value-comparison semantics
		// after the %#v → TypeOf+deepequal switch.
		//
		// This documents the behavior; it is not a gate.

		// Two errors.New instances with same text → equal (by type and
		// underlying string value).
		t.Err(errors.New("x"), errors.New("x"))
		todo.NotErr(errors.New("x"), errors.New("x"))

		// Wrapped fmt.Errorf → equal (errors.Is unwraps).
		wrapped := fmt.Errorf("outer: %w", errors.New("inner"))
		t.Err(wrapped, errors.New("inner"))
		todo.NotErr(wrapped, errors.New("inner"))

		// Pointer-field structs with same underlying value → equal
		// (bugfix: %#v previously compared by pointer address).
		p1 := &ptrFieldErr{msg: "same"}
		p2 := &ptrFieldErr{msg: "same"}
		t.Err(p1, p2)
		todo.NotErr(p1, p2)

		// Different field values → not equal.
		todo.Err(&ptrFieldErr{msg: "a"}, &ptrFieldErr{msg: "b"})
		t.NotErr(&ptrFieldErr{msg: "a"}, &ptrFieldErr{msg: "b"})

		// Nil handling.
		t.Err(nil, nil)
		todo.NotErr(nil, nil)
		todo.Err(nil, io.EOF)
		t.NotErr(nil, io.EOF)
	})

	t.Run("CheckFieldError", func(tt *testing.T) {
		tt.Parallel()
		t := check.T(tt)
		todo := t.TODO()

		// Single field errors.
		fe1 := &fieldError{ns: "ns", tag: "tag"}
		fe2 := &fieldError{ns: "ns", tag: "other"}
		fe3 := &fieldError{ns: "other", tag: "tag"}

		// BOTH sides are single fieldErr, same (Namespace, Tag).
		t.Err(fe1, fe1)
		todo.NotErr(fe1, fe1)
		// BOTH sides are single fieldErr, different Tag.
		todo.Err(fe1, fe2)
		t.NotErr(fe1, fe2)
		// BOTH sides are single fieldErr, different Namespace.
		todo.Err(fe1, fe3)
		t.NotErr(fe1, fe3)

		// Slice of field errors.
		slice1 := fieldErrors{
			{ns: "ns1", tag: "tag1"},
			{ns: "ns2", tag: "tag2"},
		}
		slice2 := fieldErrors{
			{ns: "ns1", tag: "tag1"},
			{ns: "ns2", tag: "tag2"},
		}
		slice3 := fieldErrors{
			{ns: "ns1", tag: "tag1"},
		}

		// Equal slices (same order).
		t.Err(slice1, slice2)
		todo.NotErr(slice1, slice2)
		// Order independence (SortEqual semantics).
		t.Err(slice1, fieldErrors{
			{ns: "ns2", tag: "tag2"},
			{ns: "ns1", tag: "tag1"},
		})
		todo.NotErr(slice1, fieldErrors{
			{ns: "ns2", tag: "tag2"},
			{ns: "ns1", tag: "tag1"},
		})
		// Duplicate elements with order independence.
		t.Err(fieldErrors{
			{ns: "n1", tag: "t1"},
			{ns: "n2", tag: "t2"},
			{ns: "n1", tag: "t1"},
		}, fieldErrors{
			{ns: "n1", tag: "t1"},
			{ns: "n1", tag: "t1"},
			{ns: "n2", tag: "t2"},
		})
		todo.NotErr(fieldErrors{
			{ns: "n1", tag: "t1"},
			{ns: "n2", tag: "t2"},
			{ns: "n1", tag: "t1"},
		}, fieldErrors{
			{ns: "n1", tag: "t1"},
			{ns: "n1", tag: "t1"},
			{ns: "n2", tag: "t2"},
		})
		// Different length→ not equal.
		todo.Err(slice1, slice3)
		t.NotErr(slice1, slice3)
		// Different content → not equal.
		todo.Err(slice1, fieldErrors{
			{ns: "ns1", tag: "tag1"},
			{ns: "ns2", tag: "DIFF"},
		})
		t.NotErr(slice1, fieldErrors{
			{ns: "ns1", tag: "tag1"},
			{ns: "ns2", tag: "DIFF"},
		})

		// Single vs one-element slice.
		t.Err(fe1, fieldErrors{{ns: "ns", tag: "tag"}})
		todo.NotErr(fe1, fieldErrors{{ns: "ns", tag: "tag"}})

		// Empty slices.
		t.Err(fieldErrors{}, fieldErrors{})
		todo.NotErr(fieldErrors{}, fieldErrors{})

		// Plain error to test mixed scenarios.
		plain := errors.New("plain")

		// Wrapped field error (unwrapping finds the fieldError).
		wrapped := fmt.Errorf("wrapped: %w", fe1)
		t.Err(wrapped, fe1)
		todo.NotErr(wrapped, fe1)
		// One side wrapped, other side plain → not claimed (ok=false).
		t.NotErr(wrapped, plain)
		t.NotErr(plain, wrapped)

		// Only one side has field errors → not claimed by CheckFieldError.
		// Err falls through to existing logic.
		t.NotErr(fe1, plain)
		t.NotErr(plain, fe1)
	})
}

type alwaysMatchChecker struct{}

func (alwaysMatchChecker) Error() string { return "always" }

//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestResetErrCheckers(tt *testing.T) {
	always := &alwaysMatchChecker{}

	// Register a custom checker that claims every pair.
	check.RegisterErrChecker(func(_, _ error) (equal, ok bool) {
		return true, true
	})

	// Custom checker makes Err claim any pair as equal.
	tt.Run("custom_claimed", func(tt *testing.T) {
		t := check.T(tt)
		t.Err(always, io.EOF)
	})

	// Reset removes the custom checker AND the default CheckFieldError.
	check.ResetErrCheckers()

	// After reset, Err uses built-in logic: different types → not equal.
	tt.Run("after_reset", func(tt *testing.T) {
		t := check.T(tt)
		todo := t.TODO()
		todo.Err(always, io.EOF)
		t.NotErr(always, io.EOF)
	})

	// Re-register CheckFieldError to restore default state.
	check.RegisterErrChecker(check.CheckFieldError)
}

// Fake types for probing-panic tests — they have ProtoReflect/GRPCStatus
// methods but are NOT real protobuf messages or gRPC status errors.
type fakeProto struct{}

func (fakeProto) ProtoReflect() int { return 0 }

type fakeGRPCErr struct{ msg string }

func (f fakeGRPCErr) Error() string { return f.msg }
func (fakeGRPCErr) GRPCStatus() int { return 0 }

//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestDeepEqualProtoPanic(tt *testing.T) {
	t := check.T(tt)

	// Without checkproto imported, proto messages trigger a panic.
	t.PanicMatch(func() { t.DeepEqual(fakeProto{}, fakeProto{}) },
		"import github.com/powerman/checkproto")

	// Mixed: one side has ProtoReflect.
	t.PanicMatch(func() { t.DeepEqual(fakeProto{}, "string") },
		"import github.com/powerman/checkproto")
	t.PanicMatch(func() { t.DeepEqual("string", fakeProto{}) },
		"import github.com/powerman/checkproto")

	// Nil does not panic (no method lookup on nil).
	t.DeepEqual(nil, nil)
}

//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestNotDeepEqualProtoPanic(tt *testing.T) {
	t := check.T(tt)

	t.PanicMatch(func() { t.NotDeepEqual(fakeProto{}, fakeProto{}) },
		"import github.com/powerman/checkproto")
}

//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestErrGRPCStatusPanic(tt *testing.T) {
	t := check.T(tt)

	err := &fakeGRPCErr{msg: "test"}

	// Without checkgrpc imported, gRPC status errors trigger a panic.
	t.PanicMatch(func() { t.Err(err, err) },
		"import github.com/powerman/checkgrpc")

	// Mixed: expected has GRPCStatus.
	t.PanicMatch(func() { t.Err(io.EOF, err) },
		"import github.com/powerman/checkgrpc")

	// Nil does not panic.
	t.Err(nil, nil)
}

//nolint:paralleltest // Modifies global registry, cannot run in parallel.
func TestNotErrGRPCStatusPanic(tt *testing.T) {
	t := check.T(tt)

	err := &fakeGRPCErr{msg: "test"}

	t.PanicMatch(func() { t.NotErr(err, err) },
		"import github.com/powerman/checkgrpc")
}
