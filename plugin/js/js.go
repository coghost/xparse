package js

import (
	"fmt"
	"strings"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

type Response struct {
	RefinedString string
}

const (
	inputKey  = "raw"
	outputKey = "refined"
)

func Eval(code string, raw string) (*Response, error) {
	jsVM := otto.New()

	err := jsVM.Set("raw", raw)
	if err != nil {
		return nil, fmt.Errorf("cannot set jsvm with input-key(%s): %w", inputKey, err)
	}

	_, err = jsVM.Run(code)
	if err != nil {
		return nil, fmt.Errorf("cannot run code: %w", err)
	}

	refined, err := jsVM.Get("refined")
	if err != nil {
		return nil, fmt.Errorf("cannot get output-key(%s): %w", outputKey, err)
	}

	output, err := refined.ToString()
	if err != nil {
		return nil, fmt.Errorf("cannot convert response(%v) to string: %w", refined, err)
	}

	return &Response{
		RefinedString: strings.TrimSpace(output),
	}, nil
}
