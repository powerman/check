package check

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

func caller(stack int) (name string) {
	if pc, _, _, ok := runtime.Caller(stack + 1); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			name = f.Name()
		}
		if pos := strings.LastIndex(name, "."); pos != -1 {
			name = name[pos+1:]
		}
	}
	return name
}

func callerFileLines(stack int) (testFile string, funcLine int, testLine int) {
	if pc, file, line, ok := runtime.Caller(stack + 1); ok {
		testFile = file
		testLine = line
		if f := runtime.FuncForPC(pc); f != nil {
			_, funcLine = f.FileLine(f.Entry())
		}
	}
	return
}

func format(msg ...interface{}) string {
	if len(msg) > 1 {
		return fmt.Sprintf(msg[0].(string), msg[1:]...)
	}
	return fmt.Sprint(msg...)
}

func match(actual *interface{}, regex interface{}) bool {
	if *actual == nil {
		return false
	}
	if err, _ := (*actual).(error); err != nil {
		*actual = err.Error()
	} else if stringer, _ := (*actual).(fmt.Stringer); stringer != nil {
		*actual = stringer.String()
	}
	if pattern, ok := regex.(string); ok {
		regex = regexp.MustCompile(pattern)
	}
	return regex.(*regexp.Regexp).MatchString((*actual).(string))
}

func contains(actual, expected interface{}) (found bool) {
	actualV := reflect.ValueOf(actual)
	switch reflect.TypeOf(actual).Kind() {
	case reflect.String:
		found = strings.Contains(actual.(string), expected.(string))
	case reflect.Map:
		found = actualV.MapIndex(reflect.ValueOf(expected)).IsValid()
	case reflect.Slice, reflect.Array:
		for i := 0; i < actualV.Len() && !found; i++ {
			found = found || actualV.Index(i).Interface() == expected
		}
	}
	return found
}

func zero(actual interface{}) bool {
	if actual == nil {
		return true
	} else if typ := reflect.TypeOf(actual); typ.Comparable() {
		return actual == reflect.Zero(typ).Interface()
	} else if typ.Kind() != reflect.Struct {
		return reflect.ValueOf(actual).Len() == 0
	}
	return false
}

func less(actual, expected interface{}) bool {
	switch v1, v2 := reflect.ValueOf(actual), reflect.ValueOf(expected); v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() < v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() < v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() < v2.Float()
	case reflect.String:
		return v1.String() < v2.String()
	}
	panic("actual is not a number or string")
}

func greater(actual, expected interface{}) bool {
	switch v1, v2 := reflect.ValueOf(actual), reflect.ValueOf(expected); v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() > v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() > v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() > v2.Float()
	case reflect.String:
		return v1.String() > v2.String()
	}
	panic("actual is not a number or string")
}

// TODO Use in future checks.
// func normJSON(s string) string {
// 	var v interface{}
// 	if json.Unmarshal([]byte(s), &v) != nil {
// 		return s
// 	}
// 	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
// 		return string(b)
// 	}
// 	return s
// }
