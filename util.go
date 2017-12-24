package check

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func callerTestFileLines() (file string, line int, funcLine int) {
	pc, file, line, ok := runtime.Caller(0)
	myfile := file
	for stack := 1; ok && samePackage(myfile, file); stack++ {
		pc, file, line, ok = runtime.Caller(stack)
	}
	if f := runtime.FuncForPC(pc); f != nil {
		_, funcLine = f.FileLine(f.Entry())
	}
	return file, line, funcLine
}

func samePackage(basefile, file string) bool {
	return filepath.Dir(basefile) == filepath.Dir(file) && !strings.HasSuffix(file, "_test.go")
}

func callerFuncName(stack int) string {
	pc, _, _, _ := runtime.Caller(stack + 1)
	return strings.TrimPrefix(funcNameAt(pc), "(*T).")
}

func funcName(f interface{}) string {
	return funcNameAt(reflect.ValueOf(f).Pointer())
}

func funcNameAt(pc uintptr) string {
	name := "<unknown>"
	if f := runtime.FuncForPC(pc); f != nil {
		name = f.Name()
		if i := strings.LastIndex(name, "/"); i != -1 {
			name = name[i+1:]
		}
		if i := strings.Index(name, "."); i != -1 {
			name = name[i+1:]
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
