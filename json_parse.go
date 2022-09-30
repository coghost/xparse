package xparse

import (
	"fmt"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/gookit/config/v2"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

const (
	layerWithRank = 1
	layerOthers   = 2
)

type JsonParser struct {
	Parser
}

func NewJsonParser(rawHtml, ymlMap []byte) *JsonParser {
	p := &JsonParser{
		Parser{
			Config:     &config.Config{},
			ParsedData: make(map[string]interface{}),
			Refiners:   make(map[string]func(args ...interface{}) interface{}),
		},
	}

	p.Spawn(rawHtml, ymlMap)
	return p
}

func (p *JsonParser) Spawn(raw, ymlCfg []byte) {
	p.LoadConfig(ymlCfg)
	p.LoadRootSelection(raw)
}

func (p *JsonParser) LoadRootSelection(raw []byte) {
	p.RawData = string(raw)
	p.JRoot = gjson.Parse(string(raw))
}

func (p *JsonParser) DoParse() {
	for key, cfg := range p.Config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseDom(key, cfgType, p.JRoot, p.ParsedData, layerWithRank)
		default:
			fmt.Println(xpretty.Redf("[NON-MAP] {%v:%v}, please move into a map instead", key, cfg))
			continue
		}
	}
}

func (p *JsonParser) parseDom(key string, cfg interface{}, result gjson.Result, data map[string]interface{}, layer int) {
	p.checkNestedKeys(key)
	defer p.popNestedKeys()

	b := p.requiredKey(key)
	// xpretty.DummyLog(key, p.testKeys, b, p.forceParsedKey, p.nestedKeys)
	if !b {
		return
	}

	if funk.IsEmpty(cfg) {
		data[key] = p.getSelectionAttr(key, map[string]interface{}{key: ""}, result)
		return
	}

	switch v := cfg.(type) {
	case string:
		// the recursive end condition
		p.handleStr(key, v, result, data)
	case map[string]interface{}:
		p.handle_map(key, v, result, data, layer)
	default:
		panic(xpretty.Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]interface{})", key, cfg))
	}
}

func (p *JsonParser) getSelectionAttr(key string, cfg map[string]interface{}, result gjson.Result) interface{} {
	var raw interface{}
	raw = result.String()
	raw = p.refineAttr(key, raw, cfg, result)

	t, o := cfg[TYPE]
	if o {
		switch t {
		case "b":
			return cast.ToBool(raw)
		case "i":
			return cast.ToInt(raw)
		case "f":
			return cast.ToFloat64(raw)
		}
	}

	return raw
}

func (p *JsonParser) handleStr(key string, sel string, result gjson.Result, data map[string]interface{}) {
	data[key] = result.Get(sel).String()
}

func (p *JsonParser) handle_map(
	key string,
	cfg map[string]interface{},
	result gjson.Result,
	data map[string]interface{},
	layer int,
) {
	if p.isLeaf(cfg) {
		p.getNodesAttrs(key, cfg, result, data)
		return
	}

	elems := p.getAllElems(key, cfg, result)
	switch dom := elems.(type) {
	case gjson.Result:
		subData := make(map[string]interface{})
		data[key] = subData
		p.parse_dom_nodes(cfg, dom, subData)

	case []gjson.Result:
		var allSubData []map[string]interface{}
		for _, gs := range dom {
			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parse_dom_nodes(cfg, gs, subData)
			// only calculate rank at first layer
			if layer == 1 {
				p.rank++
			}
		}
		data[key] = allSubData
	}
}

func (p *JsonParser) getIndex(sel interface{}, result gjson.Result, index int) (rs gjson.Result) {
	arr := result.Array()
	if len(arr) > index {
		return arr[index]
	}

	return
}

func (p *JsonParser) getAllElems(key string, cfg map[string]interface{}, result gjson.Result) (iface interface{}) {
	sel := cfg[LOCATOR]
	if sel == nil {
		return result
	}

	if sel == _ORDERED_LIST_CONST {
		result = gjson.Parse(p.RawData)
	} else {
		result = result.Get(sel.(string))
	}

	index, exist := cfg[INDEX]
	if index == nil {
		if !exist {
			return p.getIndex(sel, result, 0)
			// arr := result.Get(sel.(string)).Array()
			// if len(arr) > 0 {
			// 	return arr[0]
			// }
		}

		return result.Array()
	}

	switch val := index.(type) {
	case int:
		return p.getIndex(sel, result, val)
		// arr := result.Get(sel.(string)).Array()
		// if len(arr) > val {
		// 	return arr[val]
		// }
		// return
	case []interface{}:
		var d []gjson.Result
		for _, v := range val {
			switch v := v.(type) {
			case int:
				r := p.getIndex(sel, result, v)
				d = append(d, r)
				// d = append(d, result.Get(fmt.Sprintf("%v.%v", sel, v)))
			default:
				panic(xpretty.Redf("all indexes should be int, but (%s is %T: %v)\n", key, val, val))
			}
		}
		return d
	default:
		panic(xpretty.Redf("index should be int or []interface{}, but (%s is %T: %v)\n", key, val, val))
	}
}

func (p *JsonParser) parse_dom_nodes(
	cfg map[string]interface{},
	result gjson.Result,
	data map[string]interface{},
) {
	for k, sc := range cfg {
		if strings.HasPrefix(k, "_") {
			continue
		}
		p.parseDom(k, sc, result, data, layerOthers)
	}
}

func (p *JsonParser) getNodesAttrs(
	key string,
	cfg map[string]interface{},
	selection gjson.Result,
	data map[string]interface{},
) {
	elems := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case gjson.Result:
		data[key] = p.getSelectionAttr(key, cfg, dom)

	case []gjson.Result:
		var subData []interface{}
		for _, dm := range dom {
			d := p.getSelectionAttr(key, cfg, dm)
			subData = append(subData, d)
		}
		data[key] = subData
	default:
		panic(xpretty.Redf("unknown type of dom %s:%v %v", key, cfg, dom))
	}
}
