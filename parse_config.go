package xparse

import (
	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
)

func mustCfgLocator(cfg map[string]interface{}) interface{} {
	return must(cfgLocator, cfg)
}

func cfgLocator(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, Locator, LocatorAbbr)
}

func mustCfgIndex(cfg map[string]interface{}) interface{} {
	return must(cfgIndex, cfg)
}

func cfgIndex(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, Index, IndexAbbr)
}

func mustCfgAttrRefine(cfg map[string]interface{}) interface{} {
	return must(cfgAttrRefine, cfg)
}

func cfgAttrRefine(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, AttrRefine, AttrRefineAbbr)
}

func getConfig(cfg map[string]interface{}, keys ...string) (interface{}, bool) {
	for _, k := range keys {
		sel, ok := cfg[k]
		if ok {
			return sel, true
		}
	}
	return nil, false
}

func refineIndex(key string, intStr string, total int) int {
	end, err := cast.ToIntE(intStr)
	if err != nil {
		panic(xpretty.Redf("range index must be number, but (%s is %T: %v)", key, intStr, intStr))
	}
	if end < 0 {
		end += total
	}

	return end
}

func must(fn func(map[string]interface{}) (interface{}, bool), cfg map[string]interface{}) interface{} {
	v, _ := fn(cfg)
	return v
}
