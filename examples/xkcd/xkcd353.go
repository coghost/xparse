package main

import (
	"fmt"
	"path/filepath"

	"github.com/coghost/xparse"
	"github.com/gookit/goutil/fsutil"
)

type XkcdParser struct {
	*xparse.Parser
}

func NewXkcdParser(html, yml []byte) *XkcdParser {
	return &XkcdParser{
		xparse.NewHtmlParser(html, yml),
	}
}

func (p *XkcdParser) _refine_alt_alt(raw ...interface{}) interface{} {
	fmt.Println(raw[0])
	return raw[0]
}

func main() {
	home := xparse.GetProjectHome("xparse")
	rawHtml := fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd_353.html"))
	rawYaml := fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd.yaml"))

	xp := NewXkcdParser(rawHtml, rawYaml)
	xp.Refiners["_refine_alt_alt"] = xp._refine_alt_alt
	xp.DoParse()

	fmt.Println(xp.ParsedData)
}
