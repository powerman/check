package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"
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

var (
	typString = reflect.TypeOf("")
	typBytes  = reflect.TypeOf([]byte(nil))
)

func toString(i interface{}) (s string, multiline bool) {

	switch v := i.(type) {
	case nil:
		return fmt.Sprintf("(%T) %#[1]v", i), false

	case json.RawMessage:
		var buf bytes.Buffer
		json.Indent(&buf, []byte(v), "", "  ")
		s, multiline = buf.String(), true

	case *json.RawMessage:
		if v == nil {
			return fmt.Sprintf("(%T) %#[1]v", i), false
		}
		var buf bytes.Buffer
		json.Indent(&buf, []byte(*v), "", "  ")
		s, multiline = buf.String(), true

	case string:
		s, multiline = quote(v)

	case []rune:
		s, multiline = quote(string(v))

	case fmt.Stringer:
		// TODO handle nil?
		s, multiline = quote(v.String())

	default:
		if !reflect.Indirect(reflect.ValueOf(i)).Type().ConvertibleTo(typBytes) {
			return fmt.Sprintf("(%T) %#[1]v", i), true
		}

		val := reflect.Indirect(reflect.ValueOf(i))
		typ := val.Type()
		if typ.Kind() == reflect.String {
			s, multiline = quote(val.String())
		} else if typ.Elem().Kind() == reflect.Int32 {
			s, multiline = quote(val.Convert(typString).String())
		} else {
			buf := val.Convert(typBytes).Bytes()
			if utf8.Valid(buf) {
				s, multiline = quote(string(buf))
			} else {
				return fmt.Sprintf("(%T) [% [1]X]", i, buf), true
			}
		}
	}

	if multiline {
		return fmt.Sprintf("(%T) '\n%s\n'", i, s), multiline
	}
	return fmt.Sprintf("(%T) '%s'", i, s), multiline
}

// quote like %#v, except keep \n and " unquoted for readability.
func quote(s string) (quoted string, multiline bool) {
	r := []rune(strconv.Quote(s))
	q := r[:0]
	var esc bool
	for _, c := range r[1 : len(r)-1] {
		if esc {
			esc = false
			switch c {
			case 'n':
				c = '\n'
				multiline = true
			case '"':
			default:
				q = append(q, '\\')
			}
		} else if c == '\\' {
			esc = true
			continue
		}
		q = append(q, c)
	}
	return string(q), multiline
}
