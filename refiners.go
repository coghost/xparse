package xparse

import (
	"fmt"
	"os"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

func prompt(any interface{}, mtd_name, MtdName string) {
	tp := fmt.Sprintf("%T", any)
	arr := strings.Split(tp, ".")
	tp = arr[len(arr)-1]

	fmt.Println(xpretty.Redf(`Cannot find Refiner: (%s or %s)`, mtd_name, MtdName))
	fmt.Println(xpretty.Redf(`Please add following method:`))
	fmt.Println(xpretty.Greenf(`

func (p *%[3]s) %[1]s(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	// TODO: raw[1] is *config.Config
	// TODO: raw[2] is *goquery.Selection/gjson.Result
	return v
}
`, MtdName, mtd_name, tp, strings.Repeat("-", 32)))

	fmt.Println(xpretty.Yellowf(`
%[4]s
WARN: WHY GOT THIS PROMPT?
Maybe you've missed one of following methods:

- RECOMMENDED: you can call xparse.UpdateRefiners(p) before DoParse
  + this only need once
- or you can manually assign it to p.Refiners by:
  + p.Refiners["%[1]s"] = p.%[1]s
  + every new refiner is required
%[4]s
`, MtdName, mtd_name, tp, strings.Repeat("-", 32)))
	os.Exit(0)
}

func UpdateRefiners(p interface{}, methodNames ...string) {
	Invoke(p, "Scan")

	attrs := GetField(p, "AttrToBeRefined").Interface().([]string)
	attrs = append(attrs, methodNames...)

	bindRefiners(p, attrs...)
}

func bindRefiners(p interface{}, attrs ...string) {
	refiners := GetField(p, "Refiners").Interface().(map[string]func(raw ...interface{}) interface{})

	for _, mtd_name := range attrs {
		MtdName := strcase.ToCamel(mtd_name)
		method := GetMethod(p, MtdName)
		if funk.IsEmpty(method) {
			prompt(p, mtd_name, MtdName)
		}
		refiners[MtdName] = method.Interface().(func(raw ...interface{}) interface{})
	}
}

// Mock
//
//   - raw[0]: the parsed text
//   - raw[1]: *config.Config
//   - raw[2]: *goquery.Selection / gjson.Result
func (p *Parser) Mock(raw ...interface{}) interface{} {
	return raw[0]
}

func (p *Parser) RefineUrl(raw ...interface{}) interface{} {
	return p.EnrichUrl(raw...)
}

func (p *Parser) EnrichUrl(raw ...interface{}) interface{} {
	domain := p.Config.String("__raw.site_url")
	uri := EnrichUrl(raw[0], domain)
	return uri
}

func (p *Parser) ToFloat(raw ...interface{}) interface{} {
	return ToFixed(cast.ToFloat64(raw), 2)
}

func (p *Parser) BindRank(raw ...interface{}) interface{} {
	return p.rank
}

// TrimByFields removes all "\r\n\t" and keep one space at most
//
//   - 1. strings.TrimSpace
//   - 2. strings.Join(strings.Fields(s), " ")
func (p *Parser) TrimByFields(raw ...interface{}) interface{} {
	s := strings.TrimSpace(raw[0].(string))
	return strings.Join(strings.Fields(s), " ")
}

func (p *Parser) Trim(raw ...interface{}) interface{} {
	return p.TrimByFields(raw...)
}
