package check

import (
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

	fmt.Println(reporting.OpenJson)
	json.NewEncoder(os.Stdout).Encode(report)
	fmt.Println(",")
	fmt.Println(reporting.CloseJson)
}
