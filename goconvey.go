package check

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var errNoGoConvey = errors.New("goconvey not detected")

// Types and constants for goconvey JSON reporter integration.
// These replicate the corresponding definitions from
// github.com/smartystreets/goconvey/convey/reporting
// to avoid importing the goconvey package.
const (
	goconveyOpenJSON  = ">->->OPEN-JSON->->->"
	goconveyCloseJSON = "<-<-<-CLOSE-JSON<-<-<"
)

type goconveyScopeResult struct {
	Title      string
	File       string
	Line       int
	Depth      int
	Assertions []*goconveyAssertionResult
	Output     string
}

type goconveyAssertionResult struct {
	File       string
	Line       int
	Expected   string
	Actual     string
	Failure    string
	Error      any
	StackTrace string
	Skipped    bool
}

func reportToGoConvey(actual, expected, failure string) error {
	if !flags.detect().conveyJSON {
		return errNoGoConvey
	}

	testFile, testLine, funcLine := callerTestFileLines()
	report := goconveyScopeResult{
		File: testFile,
		Line: funcLine,
		Assertions: []*goconveyAssertionResult{{
			File:     testFile,
			Line:     testLine,
			Expected: expected,
			Actual:   actual,
			Failure:  failure,
		}},
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, goconveyOpenJSON)
	//nolint:musttag // These structs match goconvey's wire format (no tags).
	err := json.NewEncoder(&buf).Encode(report)
	if err != nil {
		return err
	}
	fmt.Fprintln(&buf, ",")
	fmt.Fprintln(&buf, goconveyCloseJSON)
	_, err = buf.WriteTo(os.Stdout)
	return err
}
