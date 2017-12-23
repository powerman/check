// Package check provide helpers to complement Go testing package.
//
// To use helpers just wrap each *testing.T in check.T:
//
//   func TestSomething(tt *testing.T) {
//       t := check.T{tt}
//       t.Equal(2, 2)
//       t.Log("You can use new t just like usual *testing.T")
//       t.Run("Subtests are supported too", func(tt *testing.T) {
//           t := check.T{tt}
//           t.Parallel()
//           t.NotEqual(2, 3)
//           obj, err := NewObj()
//           if t.Nil(err) {
//               t.Match(obj.field, `^\d+$`)
//           }
//       })
//   }
//
// To get optional statistics about executed checkers add this:
//
//   func TestMain(m *testing.M) { check.TestMain(m) }
package check

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// T wraps *testing.T to make it convenient to call checkers in tests.
type T struct {
	*testing.T
}

func (t *T) fail0(msg ...interface{}) bool {
	t.Helper()
	t.Errorf("%s\nChecker:  %s",
		format(msg...), caller(1))
	return fail(t.T)
}

func (t *T) fail1(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	t.Errorf("%s\nChecker:  %s\nActual:   %#v",
		format(msg...), caller(1), actual)
	return fail(t.T)
}

func (t *T) fail2(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	actualDump := newDump(actual)
	expectedDump := newDump(expected)
	failure := fmt.Sprintf("%s\nChecker:  %s\nExpected: %s\nActual:   %s%s",
		format(msg...), caller(1), expectedDump, actualDump, actualDump.diff(expectedDump))
	t.Errorf(failure)
	printConveyJSON(actualDump.String(), expectedDump.String(), failure)
	return fail(t.T)
}

func (t *T) fail2re(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if regex, ok := expected.(*regexp.Regexp); ok {
		expected = regex.String()
	}
	t.Errorf("%s\nChecker:  %s\nRegexp:   %#v\nActual:   %#v",
		format(msg...), caller(1), expected, actual)
	return fail(t.T)
}

// True checks for cond == true.
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

// Nil checks for actual == nil.
func (t *T) Nil(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual == nil {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// NotNil checks for actual != nil.
func (t *T) NotNil(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual != nil {
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
	return t.fail2(actual, expected, msg...)
}

// NotEqual checks for actual != expected.
func (t *T) NotEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if actual != expected {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// BytesEqual checks for bytes.Equal(actual, expected).
func (t *T) BytesEqual(actual, expected []byte, msg ...interface{}) bool {
	t.Helper()
	if bytes.Equal(actual, expected) {
		return pass(t.T)
	}
	return t.fail2(actual, expected, msg...)
}

// NotBytesEqual checks for !bytes.Equal(actual, expected).
func (t *T) NotBytesEqual(actual, expected []byte, msg ...interface{}) bool {
	t.Helper()
	if !bytes.Equal(actual, expected) {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// DeepEqual checks for reflect.DeepEqual(actual, expected).
func (t *T) DeepEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if reflect.DeepEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail2(actual, expected, msg...)
}

// NotDeepEqual checks for !reflect.DeepEqual(actual, expected).
func (t *T) NotDeepEqual(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// Match checks for regex.MatchString(actual).
// If actual is an error will match with actual.Error().
// If actual is nil it won't match anything.
// It will compile regex if given as string.
func (t *T) Match(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if match(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// NotMatch checks for !regex.MatchString(actual).
// If actual is an error will match with actual.Error().
// If actual is nil it won't match anything.
// It will compile regex if given as string.
func (t *T) NotMatch(actual, regex interface{}, msg ...interface{}) bool {
	t.Helper()
	if !match(&actual, regex) {
		return pass(t.T)
	}
	return t.fail2re(actual, regex, msg...)
}

// Contains checks is actual contains substring/element/key expected.
// Supported actual types are: string, array, slice, map.
func (t *T) Contains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if contains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2(actual, expected, msg...)
}

// NotContains checks is actual not contains substring/element/key expected.
// Supported actual types are: string, array, slice, map.
func (t *T) NotContains(actual, expected interface{}, msg ...interface{}) bool {
	t.Helper()
	if !contains(actual, expected) {
		return pass(t.T)
	}
	return t.fail2(actual, expected, msg...)
}

// Zero checks is actual is zero value of it's type.
func (t *T) Zero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if zero(actual) {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// NotZero checks is actual is not zero value of it's type.
func (t *T) NotZero(actual interface{}, msg ...interface{}) bool {
	t.Helper()
	if !zero(actual) {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// Len checks is len(actual) == expected.
func (t *T) Len(actual interface{}, expected int, msg ...interface{}) bool {
	t.Helper()
	l := reflect.ValueOf(actual).Len()
	if l == expected {
		return pass(t.T)
	}
	return t.fail2(l, expected, msg...)
}

// Err checks is actual error is the same as expected error.
// They may be a different objects, but must have same type and value.
func (t *T) Err(actual, expected error, msg ...interface{}) bool {
	t.Helper()
	if fmt.Sprintf("%#v", actual) == fmt.Sprintf("%#v", expected) {
		return pass(t.T)
	}
	return t.fail2(actual, expected, msg...)
}

// NotErr checks is actual error is not the same as expected error.
func (t *T) NotErr(actual, expected error, msg ...interface{}) bool {
	t.Helper()
	if fmt.Sprintf("%#v", actual) != fmt.Sprintf("%#v", expected) {
		return pass(t.T)
	}
	return t.fail1(actual, msg...)
}

// Panic checks is actual() panics and panic text match regex.
// In case of panic(nil) it will match like panic("<nil>").
func (t *T) Panic(actual func(), regex interface{}, msg ...interface{}) bool {
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

// Must interrupt test using t.FailNow if called with false value.
//
//   t.Must(t.Nil(err))
func (t *T) Must(continueTest bool) {
	t.Helper()
	if !continueTest {
		t.FailNow()
	}
}
