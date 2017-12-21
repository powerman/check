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
