package js

import (
	"strings"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

type Response struct {
	RefinedString string
}

func Eval(code string, raw string) (*Response, error) {
	jsVM := otto.New()

	err := jsVM.Set("raw", raw)
	if err != nil {
		return nil, err
	}

	_, err = jsVM.Run(code)
	if err != nil {
		return nil, err
	}

	v, err := jsVM.Get("refined")
	if err != nil {
		return nil, err
	}

	val, err := v.ToString()
	if err != nil {
		return nil, err
	}

	return &Response{
		RefinedString: strings.TrimSpace(val),
	}, nil
}
