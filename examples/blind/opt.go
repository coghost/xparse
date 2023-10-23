package main

import "github.com/coghost/xparse"

type NewParserFunc func(site int, raw []byte, ymlFile string, opts ...OptFunc) (p xparse.IParser, err error)

type Opts struct {
	parentSites []string
}

type OptFunc func(o *Opts)

func BindOpts(opt *Opts, opts ...OptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithParentSites(arr []string) OptFunc {
	return func(o *Opts) {
		o.parentSites = arr
	}
}
