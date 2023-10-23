package main

import (
	"path/filepath"

	"github.com/coghost/xparse"
	"github.com/coghost/xutil"
	"github.com/gookit/goutil/fsutil"
)

func main() {
	home := xparse.GetProjectHome("xparse")
	raw := fsutil.MustReadFile(filepath.Join(home, "/examples/blind/asset/meta01.html"))
	ps, err := NewBasicParser(341, raw, filepath.Join(home, "/examples/blind/asset/341.yaml"))
	xparse.UpdateRefiners(ps)

	xutil.PanicIfErr(err)
	ps.DoParse()
	ps.PrettifyData()
}
