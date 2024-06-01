package py3

import (
	"bytes"
	"os/exec"
	"strings"
)

const base = `
import sys
raw = sys.argv[1]
`

type Response struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer

	RefinedString string
}

func Eval(code string, raw string) (*Response, error) {
	resp, err := callExec(code, raw)
	if err != nil {
		return nil, err
	}

	resp.RefinedString = strings.TrimSpace(resp.Stdout.String())

	return resp, nil
}

func callExec(code string, raw string) (*Response, error) {
	code = base + code
	cmd := exec.Command("python", "-c", code, raw)

	eval := &Response{}
	cmd.Stdout = &eval.Stdout
	cmd.Stderr = &eval.Stderr

	err := cmd.Run()

	return eval, err
}
