package xparse

import (
	"fmt"
	"os"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/iancoleman/strcase"
	"github.com/thoas/go-funk"
)

const (
	hint = `

func (p *%[3]s) %[1]s(raw ...interface{}) interface{} {
	// TODO: raw[0] is the interface of string value parsed
	// TODO: raw[1] is map/*config.Config
	// TODO: raw[2] is *goquery.Selection/gjson.Result
	txt := p.GetStrBySplitAtIndex(raw[0], "", -1)
	return txt
}
`

	hintFn = `
%[4]s
WARN: WHY GOT THIS PROMPT?
Maybe you've missed one of following methods:

- RECOMMENDED: you can call xparse.UpdateRefiners(p) before DoParse
  + this only need once
- or you can manually assign it to p.Refiners by:
  + p.Refiners["%[1]s"] = p.%[1]s
  + every new refiner is required
%[4]s
`
)

var hintSep = strings.Repeat("-", 32) //nolint

type RefOpts struct {
	methods  []string
	hintType int
}

type RefOptFunc func(o *RefOpts)

func bindRefOpts(opt *RefOpts, opts ...RefOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithMethods(marr []string) RefOptFunc {
	return func(o *RefOpts) {
		o.methods = append(o.methods, marr...)
	}
}

func WithHintType(i int) RefOptFunc {
	return func(o *RefOpts) {
		o.hintType = i
	}
}

func prompt(iface interface{}, snakeMtdName, mtdName string, opts ...RefOptFunc) {
	opt := RefOpts{}
	bindRefOpts(&opt, opts...)

	prmType := fmt.Sprintf("%T", iface)
	arr := strings.Split(prmType, ".")
	prmType = arr[len(arr)-1]

	xpretty.RedPrintf(`Cannot find Refiner: (%s or %s)`, snakeMtdName, mtdName)
	xpretty.RedPrintf(`Please add following method:`)
	xpretty.GreenPrintf(hint, mtdName, snakeMtdName, prmType, hintSep)
	xpretty.YellowPrintf(hintFn, mtdName, snakeMtdName, prmType, hintSep)

	os.Exit(0)
}

// UpdateRefiners binds all refiners to parser
func UpdateRefiners(parser interface{}, opts ...RefOptFunc) {
	opt := RefOpts{hintType: 1}
	bindRefOpts(&opt, opts...)

	Invoke(parser, "Scan")

	attrs, _ := GetField(parser, "AttrToBeRefined").Interface().([]string)
	attrs = append(attrs, opt.methods...)

	bindRefiners(parser, attrs, opts...)
}

func bindRefiners(parser interface{}, attrs []string, opts ...RefOptFunc) {
	refiners, _ := GetField(parser, "Refiners").Interface().(map[string]func(raw ...interface{}) interface{})

	//nolint:revive,stylecheck
	for _, mtd_name := range attrs {
		mtdName := strcase.ToCamel(mtd_name)
		method := GetMethod(parser, mtdName)

		if funk.IsEmpty(method) {
			prompt(parser, mtd_name, mtdName, opts...)
		}

		refiners[mtdName], _ = method.Interface().(func(raw ...interface{}) interface{})
	}
}
