package xparse

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/thoas/go-funk"
)

type RefOpts struct {
	methods  []string
	hintType int

	promptCfg *PromptConfig
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

func WithRefPromptConfig(cfg *PromptConfig) RefOptFunc {
	return func(o *RefOpts) {
		o.promptCfg = cfg
	}
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
	opt := RefOpts{hintType: 1}
	bindRefOpts(&opt, opts...)

	refiners, _ := GetField(parser, "Refiners").Interface().(map[string]func(raw ...interface{}) interface{})

	//nolint:revive,stylecheck
	for _, mtd_name := range attrs {
		mtdName := GetCamelRefinerName(mtd_name)
		method := GetMethod(parser, mtdName)

		if funk.IsEmpty(method) {
			_ = prompts(parser, mtd_name, mtdName, opt.promptCfg)
		}

		refiners[mtdName], _ = method.Interface().(func(raw ...interface{}) interface{})
	}
}

func GetCamelRefinerName(input string) string {
	return fixAcronyms(strcase.ToCamel(input))
}

func GetLowerCamelRefinerName(input string) string {
	return fixAcronyms(strcase.ToLowerCamel(input))
}

var commonAcronyms = map[string]string{
	"Id":   "ID",
	"Url":  "URL",
	"Uri":  "URI",
	"Json": "JSON",
	// Add more as needed
}

func fixAcronyms(s string) string {
	for k, v := range commonAcronyms {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}
