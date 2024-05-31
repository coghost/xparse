package main

import (
	"path/filepath"

	"github.com/coghost/xparse"
	"github.com/gookit/goutil/fsutil"
)

func main() {
	home := xparse.GetProjectHome("xparse")
	raw := fsutil.MustReadFile(filepath.Join(home, "/examples/blind/asset/meta01.html"))
	ps, err := NewBasicParser(341, raw, filepath.Join(home, "/examples/blind/asset/341.yaml"))
	xparse.UpdateRefiners(ps)

	xparse.PanicIfErr(err)
	ps.DoParse()
	ps.PrettifyData()
}
