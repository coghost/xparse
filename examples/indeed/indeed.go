package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"xparse"

	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/goutil/fsutil"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cast"
)

type IndeedParser struct {
	*xparse.Parser
}

func GetSalary(str string, currency string, args ...string) string {
	if currency != "" {
		arr := strings.Split(str, currency)
		str = arr[len(arr)-1]
	}

	sep := xparse.FirstOrDefaultArgs(" ", args...)
	str = strings.Split(str, sep)[0]

	return str
}

func (p *IndeedParser) refineJobId(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return strings.Join(strings.Split(v, "_")[1:], "_")
}

func (p *IndeedParser) refineSalary(raw ...interface{}) (salary interface{}) {
	rs := cast.ToString(raw[0])
	if rs == "" {
		return
	}
	arr := strings.Split(rs, " - ")
	if len(arr) == 1 {
		return raw[0]
	}
	_min := GetSalary(arr[0], "$")
	_max := GetSalary(arr[1], "$")
	return map[string]interface{}{
		"min":  xparse.MustF64KMFromStr(_min),
		"max":  xparse.MustF64KMFromStr(_max),
		"_min": _min,
		"_max": _max,
	}
}

func (p *IndeedParser) refineLocation(raw ...interface{}) interface{} {
	q := raw[2].(*goquery.Selection)
	txt := ""
	n := q.Find("a.more_loc")
	if n != nil {
		txt = n.Text()
	}

	v := cast.ToString(raw[0])
	v = strings.ReplaceAll(v, txt, "")
	return v
}

func (p *IndeedParser) RefineCompanyId(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func NewIndeedParser(html, yml []byte) *IndeedParser {
	return &IndeedParser{
		xparse.NewParser(html, yml),
	}
}

func main() {
	home := xparse.GetProjectHome("xparse")
	rawHtml := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed.html"))
	rawYaml := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed.yaml"))

	ps := NewIndeedParser(rawHtml, rawYaml)
	ps.ToggleDevMode(true)
	ps.Refiners["refine_job_id"] = ps.refineJobId
	ps.Refiners["refine_salary"] = ps.refineSalary
	ps.Refiners["_refine_location"] = ps.refineLocation
	ps.Refiners["RefineCompanyId"] = ps.RefineCompanyId
	ps.DoParse()

	rawJson := ps.MustDataAsJson()

	// verify all existed keys
	all := []string{}
	xparse.GetMapKeys(&all, ps.ParsedData["jobs"])
	xparse.Verify(rawJson, all, 0)

	// verify specified keys
	comp := []string{
		"company.name",
		"company.id",
		"company.location",
		"company.rating",
	}
	dat := xparse.Verify(rawJson, comp, 2)
	pp.Println(dat)

	ry := ps.MustDataAsYaml()
	fmt.Println(ry)
}
