// Package check_test tests check with external dependencies
// that must stay out of the core go.mod.
//
// This module uses `replace github.com/powerman/check => ../`
// and is NEVER tagged or released.
package check_test

import (
	"io"
	"testing"

	"github.com/pkg/errors"

	"github.com/powerman/check"
)

func TestErrPkgErrors(tt *testing.T) {
	t := check.T(tt)

	// pkg/errors.Wrap with io.EOF.
	t.Err(errors.Wrap(io.EOF, "wrapped"), io.EOF)
	t.NotErr(errors.Wrap(io.EOF, "wrapped"), io.ErrUnexpectedEOF)

	// Double wrapped.
	t.Err(errors.Wrap(errors.Wrap(io.EOF, "outer"), "inner"), io.EOF)
	t.NotErr(errors.Wrap(errors.Wrap(io.EOF, "outer"), "inner"), io.ErrUnexpectedEOF)

	// pkg/errors.Wrap with nil returns nil.
	t.Err(nil, nil)
}

func TestNonPkgErrors(tt *testing.T) {
	t := check.T(tt)

	// Non-pkg/errors errors still work.
	t.Err(io.EOF, io.EOF)
	t.NotErr(io.EOF, io.ErrUnexpectedEOF)
}
