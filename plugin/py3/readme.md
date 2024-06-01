py3

## use embed python

> if directly call py3.Exec failed.


```golang
package demo

import "github.com/kluctl/go-embed-python/python"


func CallEmbedAndTrim(code string, raw string) (*Response, error) {
	resp, err := callEmbed(code, raw)
	if err != nil {
		return nil, err
	}
	resp.RefinedString = strings.TrimSpace(resp.Stdout.String())
	return resp, nil
}

func callEmbed(code string, raw string) (*Response, error) {
	ep, err := python.NewEmbeddedPython("example")
	if err != nil {
		panic(err)
	}
	cmd := ep.PythonCmd("-c", code, raw)
	resp := &Response{}
	cmd.Stdout = &resp.Stdout
	cmd.Stderr = &resp.Stderr
	err = cmd.Run()
	return resp, err
}
```
