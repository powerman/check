package check

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/smartystreets/goconvey/convey/reporting"
)

var errNoGoConvey = errors.New("goconvey not detected")

func reportToGoConvey(actual, expected, failure fmt.Stringer) error {
	if !flags.detect().conveyJSON {
		return errNoGoConvey
	}

	testFile, testLine, funcLine := callerTestFileLines()
	report := reporting.ScopeResult{
		File: testFile,
		Line: funcLine,
		Assertions: []*reporting.AssertionResult{{
			File:     testFile,
			Line:     testLine,
			Expected: expected.String(),
			Actual:   actual.String(),
			Failure:  failure.String(),
		}},
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, reporting.OpenJson)
	if err := json.NewEncoder(&buf).Encode(report); err != nil {
		return err
	}
	fmt.Fprintln(&buf, ",")
	fmt.Fprintln(&buf, reporting.CloseJson)
	_, err := buf.WriteTo(os.Stdout)
	return err
}
