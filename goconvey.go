package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/smartystreets/goconvey/convey/reporting"
)

func printConveyJSON(actual, expected, failure string) {
	if !flags.detect().conveyJSON {
		return
	}

	testFile, testLine, funcLine := callerFileLines()
	report := reporting.ScopeResult{
		File: testFile,
		Line: funcLine,
		Assertions: []*reporting.AssertionResult{{
			File:     testFile,
			Line:     testLine,
			Expected: expected,
			Actual:   actual,
			Failure:  failure,
		}},
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, reporting.OpenJson)
	if json.NewEncoder(&buf).Encode(report) != nil {
		return
	}
	fmt.Fprintln(&buf, ",")
	fmt.Fprintln(&buf, reporting.CloseJson)
	_, _ = buf.WriteTo(os.Stdout)
}
