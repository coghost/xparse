package xparse

import (
	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
)

// mustCfgLocator must get the locator config
func mustCfgLocator(cfg map[string]interface{}) interface{} {
	return must(cfgLocator, cfg)
}

func cfgLocator(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, Locator, LocatorAbbr)
}

// mustCfgRaw retrieves the raw configuration value from the config map.
// It returns the value associated with the Raw key, ignoring any errors.
// If the key doesn't exist, it returns nil.
func mustCfgRaw(cfg map[string]interface{}) interface{} {
	key, _ := getConfig(cfg, Raw)
	return key
}

// mustCfgIndex must get _index's config
func mustCfgIndex(cfg map[string]interface{}) interface{} {
	return must(cfgIndex, cfg)
}

func cfgIndex(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, Index, IndexAbbr)
}

// mustCfgAttrRefine must get _attr_refine's config
func mustCfgAttrRefine(cfg map[string]interface{}) interface{} {
	return must(cfgAttrRefine, cfg)
}

func cfgAttrRefine(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, AttrRefine, AttrRefineAbbr)
}

// mustCfgIndex must get _index's config
func mustCfgType(cfg map[string]interface{}) interface{} { //nolint
	return must(cfgType, cfg)
}

func cfgType(cfg map[string]interface{}) (interface{}, bool) {
	return getConfig(cfg, Type, TypeAbbr)
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
