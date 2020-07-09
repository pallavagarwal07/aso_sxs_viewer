package createwindow

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

type expectation struct {
	err  error
	want []string
}

func TestDisplayErrorArg(t *testing.T) {
	var expectations []expectation

	if runtime.GOOS == "darwin" {
		expectations = []expectation{
			{
				err:  fmt.Errorf(`normal text`),
				want: []string{"-e", `display dialog "normal text" with icon caution`},
			},
			{
				err:  fmt.Errorf(`text in "double qoutes"`),
				want: []string{"-e", `display dialog "text in \"double qoutes\"" with icon caution`},
			},
			{
				err: fmt.Errorf(`text with 	`),
				want: []string{"-e", `display dialog "text with \t" with icon caution`},
			},
		}

	}
	for _, e := range expectations {
		if out := displayErrorArg(e.err, false); !reflect.DeepEqual(out, e.want) {
			t.Errorf("Incorrect arguments. Got: %v, Want: %v.", out, e.want)
		}
	}
}
