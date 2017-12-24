package check

import (
	"fmt"
	"path/filepath"
	"reflect"
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

func myStack(myfile, file string) bool {
	return filepath.Dir(myfile) == filepath.Dir(file) && !strings.HasSuffix(file, "_test.go")
}

func callerFileLines() (file string, line int, funcLine int) {
	pc, file, line, ok := runtime.Caller(0)
	myfile := file
	for stack := 1; ok && myStack(myfile, file); stack++ {
		pc, file, line, ok = runtime.Caller(stack)
	}
	if f := runtime.FuncForPC(pc); f != nil {
		_, funcLine = f.FileLine(f.Entry())
	}
	return file, line, funcLine
}

// funcName returns f's name without package name.
func funcName(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	if i := strings.LastIndex(name, "/"); i != -1 {
		name = name[i+1:]
	}
	if i := strings.Index(name, "."); i != -1 {
		name = name[i+1:]
	}
	return name
}

func format(msg ...interface{}) string {
	if len(msg) > 1 {
		return fmt.Sprintf(msg[0].(string), msg[1:]...)
	}
	return fmt.Sprint(msg...)
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
