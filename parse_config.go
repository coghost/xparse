package xparse

import (
	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
)

// mustCfgLocator must get the locator config
func mustCfgLocator(cfg map[string]any) any {
	return must(cfgLocator, cfg)
}

func cfgLocator(cfg map[string]any) (any, bool) {
	return getConfig(cfg, Locator, LocatorAbbr)
}

// mustCfgRaw retrieves the raw configuration value from the config map.
// It returns the value associated with the Raw key, ignoring any errors.
// If the key doesn't exist, it returns nil.
func mustCfgRaw(cfg map[string]any) any {
	key, _ := getConfig(cfg, Raw)
	return key
}

// mustCfgIndex must get _index's config
func mustCfgIndex(cfg map[string]any) any {
	return must(cfgIndex, cfg)
}

func cfgIndex(cfg map[string]any) (any, bool) {
	return getConfig(cfg, Index, IndexAbbr)
}

// mustCfgAttrRefine must get _attr_refine's config
func mustCfgAttrRefine(cfg map[string]any) any {
	return must(cfgAttrRefine, cfg)
}

func cfgAttrRefine(cfg map[string]any) (any, bool) {
	return getConfig(cfg, AttrRefine, AttrRefineAbbr)
}

// mustCfgIndex must get _index's config
func mustCfgType(cfg map[string]any) any { //nolint
	return must(cfgType, cfg)
}

func cfgType(cfg map[string]any) (any, bool) {
	return getConfig(cfg, Type, TypeAbbr)
}

func getConfig(cfg map[string]any, keys ...string) (any, bool) {
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

func must(fn func(map[string]any) (any, bool), cfg map[string]any) any {
	v, _ := fn(cfg)
	return v
}
