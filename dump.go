package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

var spewCfg = spew.ConfigState{
	Indent:                  "  ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	SpewKeys:                true,
}

type dump struct {
	orig         interface{}
	dump         string
	indirectType reflect.Type
}

// String returns dump of value given to newDump.
func (v *dump) String() string {
	return v.dump
}

func (v *dump) diff(expected *dump) string {
	if v.indirectType != expected.indirectType {
		return ""
	}
	if !strings.ContainsRune(v.dump, '\n') && !strings.ContainsRune(expected.dump, '\n') {
		return ""
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected.dump),
		B:        difflib.SplitLines(v.dump),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})
	return "\n\nDiff:\n" + diff
}

// newDump prepare i dump using spew.Sdump in most cases and custom
// improved dump for these cases:
// - nil: remove "(interface{})" prefix
// - byte: use 0xFF instead of decimal
// - rune: use quoted char instead of number for valid runes
// - string: use this instead of quoted single-line:
//   - valid utf8: don't quote ", show multiline strings on separate lines
//   - invalid utf8: use hexdump like for []byte
// - []byte: same as string instead of hexdump for valid utf8
// - []rune: use quoted char instead of number for valid runes in list
// - json.RawMessage: indent, then same as string
func newDump(i interface{}) *dump {
	d := &dump{orig: i}

	if i == nil {
		d.dump = "<nil>"
		return d
	}

	val := reflect.ValueOf(i)
	typ := reflect.TypeOf(i)
	kind := typ.Kind()
	if kind == reflect.Ptr {
		if val.IsNil() {
			d.dump = spewCfg.Sdump(i)
			return d
		}
		val = val.Elem()
		typ = typ.Elem()
		kind = typ.Kind()
	}
	d.indirectType = typ

	switch {
	case typ == reflect.TypeOf(json.RawMessage(nil)):
		v := val.Bytes()
		var buf bytes.Buffer
		json.Indent(&buf, v, "", "  ")
		d.dump = fmt.Sprintf("(%T) (len=%d) '\n%s\n'", i, len(v), buf.String())
		return d
	case kind == reflect.Uint8:
		v := byte(val.Uint())
		d.dump = fmt.Sprintf("(%T) 0x%02X", i, v)
		return d
	case kind == reflect.Int32:
		v := rune(val.Int())
		if utf8.ValidRune(v) {
			d.dump = fmt.Sprintf("(%T) %q", i, v)
			return d
		}
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Int32:
		valid := true
		for k := 0; k < val.Len() && valid; k++ {
			valid = valid && utf8.ValidRune(rune(val.Index(k).Int()))
		}
		if valid {
			d.dump = fmt.Sprintf("(%T) %q", i, i)
			return d
		}
	case kind == reflect.String:
		v := val.String()
		if utf8.ValidString(v) {
			d.dump = fmt.Sprintf("(%T) (len=%d) %s", i, len(v), quote(v))
		} else {
			d.dump = strings.Replace(spewCfg.Sdump([]byte(v)), "([]uint8)", fmt.Sprintf("(%T)", i), 1)
		}
		return d
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		v := val.Bytes()
		if len(v) > 0 && utf8.Valid(v) || len(v) == 0 && !val.IsNil() {
			d.dump = fmt.Sprintf("(%T) (len=%d) %s", i, len(v), quote(string(v)))
		} else {
			d.dump = spewCfg.Sdump(i)
		}
		return d
	}
	d.dump = spewCfg.Sdump(i)
	return d
}

// quote like %#v, except keep \n and " unquoted for readability.
func quote(s string) string {
	r := []rune(strconv.Quote(s))
	q := r[:0]
	var multiline, esc bool
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
	if multiline {
		return fmt.Sprintf("'\n%s\n'", string(q))
	}
	return fmt.Sprintf("'%s'", string(q))
}
