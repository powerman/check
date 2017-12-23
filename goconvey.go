package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/smartystreets/goconvey/convey" // to setup -convey-json flag
	"github.com/smartystreets/goconvey/convey/reporting"
)

func printConveyJSON(actual, expected, failure string) {
	if !flags.detect().conveyJSON {
		return
	}

	testFile, funcLine, testLine := callerFileLines(3)
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
