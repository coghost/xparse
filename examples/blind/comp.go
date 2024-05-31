package main

import (
	"fmt"
	"strings"

	"github.com/coghost/xparse"
	"github.com/thoas/go-funk"
	"github.com/ungerik/go-dry"
)

type BasicParser struct {
	*xparse.HTMLParser
}

func NewBasicParser(site int, raw []byte, ymlFile string, opts ...OptFunc) (cp *BasicParser, err error) {
	ymlArr := getYamlConfigs(ymlFile, opts...)
	p := &BasicParser{
		xparse.NewHTMLParser(raw, ymlArr...),
	}
	p.PID = fmt.Sprintf("%d", site)

	return p, nil
}

func getYamlConfigs(ymlFile string, opts ...OptFunc) [][]byte {
	opt := Opts{}
	BindOpts(&opt, opts...)

	var ymlArr [][]byte

	opt.parentSites = append(opt.parentSites, ymlFile)
	for _, i := range opt.parentSites {
		yml := loadYamlConfig(i)
		ymlArr = append(ymlArr, yml)
	}

	return ymlArr
}

func loadYamlConfig(ymlFile string) []byte {
	v, err := dry.FileGetBytes(ymlFile)
	if err != nil {
		panic(err)
	}

	if funk.IsEmpty(v) {
		panic(fmt.Sprintf("empty file of %s", ymlFile))
	}

	return v
}

func (p *BasicParser) RefineReviews(raw ...interface{}) interface{} {
	txt := p.GetStrBySplitAtIndex(raw[0], "\n", 0)
	txt = p.GetStrBySplitAtIndex(txt, "(", 1)
	txt = strings.ReplaceAll(txt, ",", "")

	return xparse.MustCharToNum(txt)
}
