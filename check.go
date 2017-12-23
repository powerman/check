package check

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// T wraps *testing.T to make it convenient to call checkers in test.
//
// It's convenient to rename Test function's arg from t to something
// else, create wrapped variable with usual name t and use only t:
//
//	func TestSomething(tt *testing.T) {
//		t := check.T{tt}
//		// use only t in test and don't touch tt anymore
//	}
type T struct {
	*testing.T
}

func (t *T) fail0(msg ...interface{}) bool {
	t.Helper()
	t.Errorf("%s\nChecker:  %s%s%s\n\n",
		format(msg...),
		ansiYellow, caller(1), ansiReset,
	)
	return fail(t.T)
}

func (t *T) fail1(checker string, actual interface{}, msg ...interface{}) bool {
	if checker == "" {
		checker = caller(1)
	}
	t.Helper()
	t.Errorf("%s\nChecker:  %s%s%s\nActual:   %s%#v%s\n\n",
		format(msg...),
		ansiYellow, checker, ansiReset,
		ansiRed, actual, ansiReset,
	)
	return fail(t.T)
}

func (t *T) fail2(checker, actual, expected interface{}, msg ...interface{}) bool {
	if checker == "" {
		checker = caller(1)
	}
	t.Helper()
	actualDump := newDump(actual)
	expectedDump := newDump(expected)
	failure := fmt.Sprintf("%s\nChecker:  %s%s%s\nExpected: %s%s%sActual:   %s%s%s\n%s",
		format(msg...),
		ansiYellow, checker, ansiReset,
		ansiGreen, expectedDump, ansiReset,
		ansiRed, actualDump, ansiReset,
		colouredDiff(actualDump.diff(expectedDump)),
	)
	t.Errorf(failure)
	printConveyJSON(actualDump.String(), expectedDump.String(), failure)
	return fail(t.T)
}

func (t *T) fail2re(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if regex, ok := expected.(*regexp.Regexp); ok {
		expected = regex.String()
	}
	t.Errorf("%s\nChecker:  %s%s%s\nRegexp:   %s%#v%s\nActual:   %s%#v%s",
		format(msg...),
		ansiYellow, caller(1), ansiReset,
		ansiGreen, expected, ansiReset,
		ansiRed, actual, ansiReset,
	)
	return fail(t.T)
}

// Must interrupt test using t.FailNow if called with false value.
//
// This provides easy way to turn any check into assertion:
//
//   t.Must(t.Nil(err))
func (t *T) Must(continueTest bool) {
	t.Helper()
	if !continueTest {
		t.fail0()
		t.FailNow()
	}
	pass(t.T)
}

type (
	// CheckFunc1 is like Nil or Zero.
	CheckFunc1 func(t *T, actual interface{}) bool
	// CheckFunc2 is like Equal or Match.
	CheckFunc2 func(t *T, actual, expected interface{}) bool
)

var (
	typCheckFunc1 = reflect.TypeOf(CheckFunc1(nil))
	typCheckFunc2 = reflect.TypeOf(CheckFunc2(nil))
)

// Should use user-provided check function to do actual check.
//
// anyCheckFunc must have type CheckFunc1 or CheckFunc2. It should return
// it's name and true if check was successful. There is no need to call
// t.Error in anyCheckFunc - this will be done automatically when it
// returns.
//
// args must contain at least 1 element for CheckFunc1 and at least
// 2 elements for CheckFunc2.
// Rest of elements will be processed as usual msg ...interface{} param.
//
// Example:
//
//	func bePositive(_ *check.T, actual interface{}) bool {
//		return actual.(int) > 0
//	}
//	func TestCustomCheck(tt *testing.T) {
//		t := check.T{tt}
//		t.Should(bePositive, 42, "custom check!!!")
//	}
func (t *T) Should(anyCheckFunc interface{}, args ...interface{}) bool {
	t.Helper()
	checker := runtime.FuncForPC(reflect.ValueOf(anyCheckFunc).Pointer()).Name()
	if i := strings.LastIndex(checker, "."); i != -1 {
		checker = "Should" + checker[i:]
	}
	if f, ok := anyCheckFunc.(func(t *T, actual interface{}) bool); ok {
		if len(args) < 1 {
			panic(checker + " require 1 param")
		}
		actual, msg := args[0], args[1:]
		if f(t, actual) {
			return pass(t.T)
		}
		return t.fail1(checker, actual, msg...)
	} else if f, ok := anyCheckFunc.(func(t *T, actual, expected interface{}) bool); ok {
		if len(args) < 2 {
			panic(checker + " require 2 param")
		}
		actual, expected, msg := args[0], args[1], args[2:]
		if f(t, actual, expected) {
			return pass(t.T)
		}
		return t.fail2(checker, actual, expected, msg...)
	}
	panic("anyCheckFunc is not a CheckFunc1 or CheckFunc2")
}

// Nil checks for actual == nil.
//
// There is one subtle difference between this check and Go `== nil` (if
// this surprises you then you should read
// https://golang.org/doc/faq#nil_error first):
//
//	var intPtr *int
//	var empty interface{}
//	var notEmpty interface{} = intPtr
//	t.True(intPtr == nil)   // TRUE
//	t.True(empty == nil)    // TRUE
//	t.True(notEmpty == nil) // FALSE
//
// When you call this function your actual value will be stored in
// interface{} argument, and this makes any typed nil pointer value `!=
// nil` inside this function (just like in example above happens with
// notEmpty variable).
//
// As it is very common case to check some typed pointer using Nil this
// check has to work around and detect nil even if usual `== nil` return
// false. But this has nasty side effect: if actual value already was of
// interface type and contains some typed nil pointer (which is usually
// bad thing and should be avoid) then Nil check will pass (which may be
// not what you want/expect):
//
//	t.Nil(nil)              // TRUE
//	t.Nil(intPtr)           // TRUE
//	t.Nil(empty)            // TRUE
//	t.Nil(notEmpty)         // WARNING: also TRUE!
//
// Second subtle case is less usual: uintptr(0) is sorta nil, but not
// really, so Nil(uintptr(0)) will fail. Nil(unsafe.Pointer(nil)) will
// also fail, for the same reason. Please do not use this and consider
// this behaviour undefined, because it may change in the future.
func (t *T) Nil(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	v := reflect.ValueOf(actual)
	if actual == nil || v.Kind() == reflect.Ptr && v.IsNil() {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// NotNil checks for actual != nil.
//
// See Nil about subtle case in check logic.
func (t *T) NotNil(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	v := reflect.ValueOf(actual)
	if !(actual == nil || v.Kind() == reflect.Ptr && v.IsNil()) {
		return pass(t.T)
	}
	return t.fail0(msg...)
}

// True checks for cond == true.
//
// This can be useful to use your own custom checks, but this way you
// won't get nice dump/diff for actual/expected values. You'll still have
// statistics about passed/failed checks and it's shorter than usual:
//
//	if !cond {
//		t.Errorf(msg...)
//	}
func (t *T) True(cond bool, msg ...interface{}) bool {
	t.Helper()
	if cond {
		return pass(t.T)
	}
	return t.fail0(msg...)
}

// False checks for cond == false.
func (t *T) False(cond bool, msg ...interface{}) bool {
	t.Helper()
	if !cond {
		return pass(t.T)
	}
	return t.fail0(msg...)
}

// Equal checks for actual == expected.
func (t *T) Equal(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual == expected {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// EQ is synonym for Equal (checks for actual == expected).
func (t *T) EQ(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual == expected {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NotEqual checks for actual != expected.
func (t *T) NotEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual != expected {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// NE is synonym for NotEqual (checks for actual != expected).
func (t *T) NE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual != expected {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// BytesEqual checks for bytes.Equal(actual, expected).
//
// Hint: BytesEqual([]byte{}, []byte(nil)) is true (unlike DeepEqual).
func (t *T) BytesEqual(actual, expected []byte, msg ...interface{}) bool {
	t.Helper()
	if bytes.Equal(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NotBytesEqual checks for !bytes.Equal(actual, expected).
//
// Hint: NotBytesEqual([]byte{}, []byte(nil)) is false (unlike NotDeepEqual).
func (t *T) NotBytesEqual(actual, expected []byte, msg ...interface{}) bool {
	t.Helper()
	if !bytes.Equal(actual, expected) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// DeepEqual checks for reflect.DeepEqual(actual, expected).
func (t *T) DeepEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if reflect.DeepEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NotDeepEqual checks for !reflect.DeepEqual(actual, expected).
func (t *T) NotDeepEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// Match checks for regex.MatchString(actual).
//
// Regex type can be either *regexp.Regexp or string.
//
// Actual type can be:
//   - string       - will match with actual
//   - fmt.Stringer - will match with actual.String()
//   - error        - will match with actual.Error()
//   - nil          - will not match (even with empty regex)
func (t *T) Match(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if match(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// NotMatch checks for !regex.MatchString(actual).
//
// See Match about supported actual/regex types and check logic.
func (t *T) NotMatch(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if !match(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// Contains checks is actual contains substring/element/key expected.
//
// Supported actual types are: string, array, slice, map.
//
// Type of expected depends on type of actual:
//   - if actual is a string, then expected should be a string
//   - if actual is an array, then expected should have array's element type
//   - if actual is a slice,  then expected should have slice's element type
//   - if actual is a map,    then expected should have map's key type
func (t *T) Contains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if contains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NotContains checks is actual not contains substring/element/key expected.
//
// See Contains about supported actual/expected types and check logic.
func (t *T) NotContains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !contains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// Zero checks is actual is zero value of it's type.
func (t *T) Zero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if zero(actual) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// NotZero checks is actual is not zero value of it's type.
func (t *T) NotZero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if !zero(actual) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// Len checks is len(actual) == expected.
func (t *T) Len(actual interface{}, expected int, msg ...interface{}) bool {
	t.Helper()
	l := reflect.ValueOf(actual).Len()
	if l == expected {
		return pass(t.T)
	}
	return t.fail2("", l, expected, msg...)
}

// NotLen checks is len(actual) != expected.
func (t *T) NotLen(actual interface{}, expected int, msg ...interface{}) bool {
	t.Helper()
	l := reflect.ValueOf(actual).Len()
	if l != expected {
		return pass(t.T)
	}
	return t.fail2("", l, expected, msg...)
}

// Err checks is actual error is the same as expected error.
//
// They may be a different instances, but must have same type and value.
//
// Checking for nil is okay, but using Nil(actual) instead is more clean.
func (t *T) Err(actual, expected error, msg ...interface{}) bool {
	t.Helper()
	if fmt.Sprintf("%#v", actual) == fmt.Sprintf("%#v", expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NotErr checks is actual error is not the same as expected error.
//
// They must have either different types or values (or one should be nil).
// Different instances with same type and value will be considered the
// same error, and so is both nil.
func (t *T) NotErr(actual, expected error, msg ...interface{}) bool {
	t.Helper()
	if fmt.Sprintf("%#v", actual) != fmt.Sprintf("%#v", expected) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

// Panic checks is actual() panics.
//
// It is able to detect panic(nil)… but you should try to avoid using this.
func (t *T) Panic(actual func(), msg ...interface{}) bool {
	t.Helper()
	var didPanic = true
	func() {
		defer func() { _ = recover() }()
		actual()
		didPanic = false
	}()
	if didPanic {
		return pass(t.T)
	}
	return t.fail0(msg...)
}

// NotPanic checks is actual() don't panics.
//
// It is able to detect panic(nil)… but you should try to avoid using this.
func (t *T) NotPanic(actual func(), msg ...interface{}) bool {
	t.Helper()
	var didPanic = true
	func() {
		defer func() { _ = recover() }()
		actual()
		didPanic = false
	}()
	if !didPanic {
		return pass(t.T)
	}
	return t.fail0(msg...)
}

// PanicMatch checks is actual() panics and panic text match regex.
//
// Regex type can be either *regexp.Regexp or string.
//
// In case of panic(nil) it will match like panic("<nil>").
func (t *T) PanicMatch(actual func(), regex interface{}, msg ...interface{}) bool {
	t.Helper()
	var panicVal interface{}
	var didPanic = true
	func() {
		defer func() { panicVal = recover() }()
		actual()
		didPanic = false
	}()
	if !didPanic {
		return t.fail0(msg...)
	}
	switch panicVal.(type) {
	case string, error:
	default:
		panicVal = fmt.Sprintf("%#v", panicVal)
	}
	if match(&panicVal, regex) {
		return pass(t.T)
	}
	return t.fail2re(panicVal, regex, msg...)
}

// PanicNotMatch checks is actual() panics and panic text not match regex.
//
// Regex type can be either *regexp.Regexp or string.
//
// In case of panic(nil) it will match like panic("<nil>").
func (t *T) PanicNotMatch(actual func(), regex interface{}, msg ...interface{}) bool {
	t.Helper()
	var panicVal interface{}
	var didPanic = true
	func() {
		defer func() { panicVal = recover() }()
		actual()
		didPanic = false
	}()
	if !didPanic {
		return t.fail0(msg...)
	}
	switch panicVal.(type) {
	case string, error:
	default:
		panicVal = fmt.Sprintf("%#v", panicVal)
	}
	if !match(&panicVal, regex) {
		return pass(t.T)
	}
	return t.fail2re(panicVal, regex, msg...)
}

// Less checks for actual < expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) Less(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if less(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// LT is synonym for Less (checks for actual < expected).
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) LT(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if less(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// LessOrEqual checks for actual <= expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) LessOrEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !greater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// LE is synonym for LessOrEqual (checks for actual <= expected).
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) LE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !greater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// Greater checks for actual > expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) Greater(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if greater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// GT is synonym for Greater (checks for actual > expected).
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) GT(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if greater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// GreaterOrEqual checks for actual >= expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) GreaterOrEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !less(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// GE is synonym for GreaterOrEqual (checks for actual >= expected).
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
func (t *T) GE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !less(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}
