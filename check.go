package check

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

var typString = reflect.TypeOf("")

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
	actualDump := newDump(actual)
	t.Errorf("%s\nChecker:  %s%s%s\nActual:   %s%s%s\n\n",
		format(msg...),
		ansiYellow, checker, ansiReset,
		ansiRed, actualDump, ansiReset,
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
	switch f := anyCheckFunc.(type) {
	case func(t *T, actual interface{}) bool:
		return t.should1(f, args...)
	case func(t *T, actual, expected interface{}) bool:
		return t.should2(f, args...)
	default:
		panic("anyCheckFunc is not a CheckFunc1 or CheckFunc2")
	}
}

func (t *T) should1(f CheckFunc1, args ...interface{}) bool {
	t.Helper()
	if len(args) < 1 {
		panic("not enough params for " + funcName(f))
	}
	actual, msg := args[0], args[1:]
	if f(t, actual) {
		return pass(t.T)
	}
	return t.fail1("Should "+funcName(f), actual, msg...)
}

func (t *T) should2(f CheckFunc2, args ...interface{}) bool {
	t.Helper()
	if len(args) < 2 {
		panic("not enough params for " + funcName(f))
	}
	actual, expected, msg := args[0], args[1], args[2:]
	if f(t, actual, expected) {
		return pass(t.T)
	}
	return t.fail2("Should "+funcName(f), actual, expected, msg...)
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
	if isNil(actual) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

func isNil(actual interface{}) bool {
	val := reflect.ValueOf(actual)
	return actual == nil || val.Kind() == reflect.Ptr && val.IsNil()
}

// NotNil checks for actual != nil.
//
// See Nil about subtle case in check logic.
func (t *T) NotNil(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isNil(actual) {
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
//
// Note: For time.Time it uses actual.Equal(expected) instead.
func (t *T) Equal(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if isEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

func isEqual(actual, expected interface{}) bool {
	switch actual := actual.(type) {
	case time.Time:
		return actual.Equal(expected.(time.Time))
	}
	return actual == expected
}

// EQ is a synonym for Equal.
func (t *T) EQ(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.Equal(actual, expected, msg...)
}

// NotEqual checks for actual != expected.
func (t *T) NotEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// NE is a synonym for NotEqual.
func (t *T) NE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.NotEqual(actual, expected, msg...)
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
//   - []byte       - will match with string(actual)
//   - []rune       - will match with string(actual)
//   - fmt.Stringer - will match with actual.String()
//   - error        - will match with actual.Error()
//   - nil          - will not match (even with empty regex)
func (t *T) Match(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if isMatch(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// isMatch updates actual to be a real string used for matching, to make
// dump easier to understand, but this result in losing type information.
func isMatch(actual *interface{}, regex interface{}) bool {
	switch v := (*actual).(type) {
	case nil:
		return false
	case error:
		*actual = v.Error()
	case fmt.Stringer:
		*actual = v.String()
	default:
		*actual = reflect.ValueOf(*actual).Convert(typString).Interface()
	}

	s := (*actual).(string)

	switch v := regex.(type) {
	case *regexp.Regexp:
		return v.MatchString(s)
	case string:
		return regexp.MustCompile(v).MatchString(s)
	}
	panic("regex is not a *regexp.Regexp or string")
}

// NotMatch checks for !regex.MatchString(actual).
//
// See Match about supported actual/regex types and check logic.
func (t *T) NotMatch(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isMatch(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// Contains checks is actual contains substring/element expected.
//
// Element of array/slice/map is checked using == expected.
//
// Type of expected depends on type of actual:
//   - if actual is a string, then expected should be a string
//   - if actual is an array, then expected should have array's element type
//   - if actual is a slice,  then expected should have slice's element type
//   - if actual is a map,    then expected should have map's value type
//
// Hint: In a map it looks for a value, if you need to look for a key -
// use HasKey instead.
func (t *T) Contains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if isContains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

func isContains(actual, expected interface{}) (found bool) {
	switch valActual := reflect.ValueOf(actual); valActual.Kind() {
	case reflect.String:
		strActual := valActual.Convert(typString).Interface().(string)
		valExpected := reflect.ValueOf(expected)
		if valExpected.Kind() != reflect.String {
			panic("expected underlying type is not a string")
		}
		strExpected := valExpected.Convert(typString).Interface().(string)
		found = strings.Contains(strActual, strExpected)

	case reflect.Map:
		if valActual.Type().Elem() != reflect.TypeOf(expected) {
			panic("expected type not match actual element type")
		}
		keys := valActual.MapKeys()
		for i := 0; i < len(keys) && !found; i++ {
			found = valActual.MapIndex(keys[i]).Interface() == expected
		}

	case reflect.Slice, reflect.Array:
		if valActual.Type().Elem() != reflect.TypeOf(expected) {
			panic("expected type not match actual element type")
		}
		for i := 0; i < valActual.Len() && !found; i++ {
			found = valActual.Index(i).Interface() == expected
		}

	default:
		panic("actual is not a string, array, slice or map")
	}
	return found
}

// NotContains checks is actual not contains substring/element expected.
//
// See Contains about supported actual/expected types and check logic.
func (t *T) NotContains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isContains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// HasKey checks is actual has key expected.
func (t *T) HasKey(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if hasKey(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

func hasKey(actual, expected interface{}) bool {
	return reflect.ValueOf(actual).MapIndex(reflect.ValueOf(expected)).IsValid()
}

// NotHasKey checks is actual has no key expected.
func (t *T) NotHasKey(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !hasKey(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// Zero checks is actual is zero value of it's type.
func (t *T) Zero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if isZero(actual) {
		return pass(t.T)
	}
	return t.fail1("", actual, msg...)
}

func isZero(actual interface{}) bool {
	if actual == nil {
		return true
	} else if typ := reflect.TypeOf(actual); typ.Comparable() {
		// Not Func, Map, Slice, Array with non-comparable
		// elements, Struct with non-comparable fields.
		return actual == reflect.Zero(typ).Interface()
	} else if typ.Kind() == reflect.Map ||
		typ.Kind() == reflect.Slice ||
		typ.Kind() == reflect.Array {
		return reflect.ValueOf(actual).Len() == 0
	}
	// Func, Struct with non-comparable fields.
	return false
}

// NotZero checks is actual is not zero value of it's type.
func (t *T) NotZero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isZero(actual) {
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

	if isMatch(&panicVal, regex) {
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

	if !isMatch(&panicVal, regex) {
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
//   - time.Time
func (t *T) Less(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if isLess(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

func isLess(actual, expected interface{}) bool {
	switch v1, v2 := reflect.ValueOf(actual), reflect.ValueOf(expected); v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() < v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() < v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() < v2.Float()
	case reflect.String:
		return v1.String() < v2.String()
	default:
		if actualTime, ok := actual.(time.Time); ok {
			return actualTime.Before(expected.(time.Time))
		}
	}
	panic("actual is not a number, string or time.Time")
}

// LT is a synonym for Less.
func (t *T) LT(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.Less(actual, expected, msg...)
}

// LessOrEqual checks for actual <= expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
//   - time.Time
func (t *T) LessOrEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isGreater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

func isGreater(actual, expected interface{}) bool {
	switch v1, v2 := reflect.ValueOf(actual), reflect.ValueOf(expected); v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() > v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() > v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() > v2.Float()
	case reflect.String:
		return v1.String() > v2.String()
	default:
		if actualTime, ok := actual.(time.Time); ok {
			return actualTime.After(expected.(time.Time))
		}
	}
	panic("actual is not a number, string or time.Time")
}

// LE is a synonym for LessOrEqual.
func (t *T) LE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.LessOrEqual(actual, expected, msg...)
}

// Greater checks for actual > expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
//   - time.Time
func (t *T) Greater(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if isGreater(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// GT is a synonym for Greater.
func (t *T) GT(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.Greater(actual, expected, msg...)
}

// GreaterOrEqual checks for actual >= expected.
//
// Both actual and expected must be either:
//   - signed integers
//   - unsigned integers
//   - floats
//   - strings
//   - time.Time
func (t *T) GreaterOrEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !isLess(actual, expected) {
		return pass(t.T)
	}
	return t.fail2("", actual, expected, msg...)
}

// GE is a synonym for GreaterOrEqual.
func (t *T) GE(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	return t.GreaterOrEqual(actual, expected, msg...)
}
