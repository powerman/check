# check [![GoDoc](https://godoc.org/github.com/powerman/check?status.svg)](http://godoc.org/github.com/powerman/check) [![CircleCI](https://circleci.com/gh/powerman/check.svg?style=svg)](https://circleci.com/gh/powerman/check)

Checkers for use with Go `testing` package (testify/assert done right).

```go
import "github.com/powerman/check"

func TestSomething(tt *testing.T) {
    t := check.T{tt}
    t.Equal(2, 2)
    t.Log("You can use new t just like usual *testing.T")
    t.Run("Subtests are supported too", func(tt *testing.T) {
        t := check.T{tt}
        t.Parallel()
        t.NotEqual(2, 3)
        obj, err := NewObj()
        if t.Nil(err) {
            t.Match(obj.field, `^\d+$`)
        }
    })
}
```

## Installation

Require [Go 1.9](https://golang.org/doc/go1.9#test-helper).

```
go get github.com/powerman/check
```

## TODO

- color output when stdout on terminal
- nice diff for: complex data structures / multiline strings / json
- count skipped tests
