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
	hint0 = ``
	hint1 = `

func (p *%[3]s) %[1]s(raw ...interface{}) interface{} {
	// TODO: raw[0] is the interface of string value parsed
	// TODO: raw[1] is map/*config.Config
	// TODO: raw[2] is *goquery.Selection/gjson.Result
	txt := p.GetStrBySplitAtIndex(raw[0], "", -1)
	return txt
}
`
	hint2 = `

func (p *%[3]s) %[1]s(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	// TODO: raw[1] is map/*config.Config
	// TODO: raw[2] is *goquery.Selection/gjson.Result
	return v
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

func prompt(any interface{}, mtd_name, MtdName string, opts ...RefOptFunc) {
	opt := RefOpts{}
	bindRefOpts(&opt, opts...)

	tp := fmt.Sprintf("%T", any)
	arr := strings.Split(tp, ".")
	tp = arr[len(arr)-1]

	hint := hint1
	switch opt.hintType {
	case 2:
		hint = hint2
	default:
		hint = hint1
	}

	fmt.Println(xpretty.Redf(`Cannot find Refiner: (%s or %s)`, mtd_name, MtdName))
	fmt.Println(xpretty.Redf(`Please add following method:`))
	fmt.Println(xpretty.Greenf(hint, MtdName, mtd_name, tp, strings.Repeat("-", 32)))
	fmt.Println(xpretty.Yellowf(hintFn, MtdName, mtd_name, tp, strings.Repeat("-", 32)))

	os.Exit(0)
}

// UpdateRefiners binds all refiners to parser
func UpdateRefiners(p interface{}, opts ...RefOptFunc) {
	opt := RefOpts{hintType: 1}
	bindRefOpts(&opt, opts...)

	Invoke(p, "Scan")

	attrs := GetField(p, "AttrToBeRefined").Interface().([]string)
	attrs = append(attrs, opt.methods...)

	bindRefiners(p, attrs, opts...)
}

func bindRefiners(p interface{}, attrs []string, opts ...RefOptFunc) {
	refiners := GetField(p, "Refiners").Interface().(map[string]func(raw ...interface{}) interface{})

	for _, mtd_name := range attrs {
		MtdName := strcase.ToCamel(mtd_name)
		method := GetMethod(p, MtdName)
		if funk.IsEmpty(method) {
			prompt(p, mtd_name, MtdName, opts...)
		}
		refiners[MtdName] = method.Interface().(func(raw ...interface{}) interface{})
	}
}
