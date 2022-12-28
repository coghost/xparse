package xparse

import (
	"fmt"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/gookit/config/v2"
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

func NewJsonParser(rawData, ymlMap []byte) *JsonParser {
	p := &JsonParser{
		Parser{
			Config:     &config.Config{},
			ParsedData: make(map[string]interface{}),
			Refiners:   make(map[string]func(args ...interface{}) interface{}),
		},
	}

	p.Spawn(rawData, ymlMap)
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
	p.runCheck()
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
	return p.convertToType(raw, cfg)
}

func (p *JsonParser) getSelectionSliceAttr(key string, cfg map[string]interface{}, resultArr []gjson.Result) interface{} {
	var raw []string
	for _, v := range resultArr {
		raw = append(raw, v.String())
	}
	v := p.refineAttr(key, strings.Join(raw, ATTR_SEP), cfg, resultArr)
	return p.convertToType(v, cfg)
}

func (p *JsonParser) getSelectionMapAttr(key string, cfg map[string]interface{}, results map[string]gjson.Result) interface{} {
	dat := make(map[string]string)
	for k, v := range results {
		dat[k] = v.String()
	}
	str, _ := Stringify(dat)
	v := p.refineAttr(key, str, cfg, results)
	return p.convertToType(v, cfg)
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

	elems, _ := p.getAllElems(key, cfg, result)
	switch dom := elems.(type) {
	case gjson.Result:
		subData := make(map[string]interface{})
		data[key] = subData
		p.parse_dom_nodes(cfg, dom, subData)
		if layer == layerWithRank {
			p.FocusedStub = dom
		}

	case []gjson.Result:
		var allSubData []map[string]interface{}
		for _, gs := range dom {
			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			if layer == layerWithRank {
				p.FocusedStub = gs
			}

			p.parse_dom_nodes(cfg, gs, subData)
			// only calculate rank at first layer
			if layer == layerWithRank {
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

func (p *JsonParser) getAllElems(key string, cfg map[string]interface{}, result gjson.Result) (iface interface{}, isComplexSel bool) {
	sel := cfg[LOCATOR]
	if sel == nil {
		return result, false
	}

	switch sel := sel.(type) {
	case string:
		if sel == _ORDERED_LIST_CONST {
			result = gjson.Parse(p.RawData)
		} else {
			result = result.Get(sel)
		}
		iface = p.getOneSelector(key, sel, cfg, result)
	case []interface{}:
		var arr []gjson.Result
		backup := result

		for _, v := range sel {
			ar1 := strings.Split(v.(string), ".")

			if ar1[0] == PREFIX_LOCATOR_STUB {
				v = strings.Join(ar1[1:], ".")
				backup = p.FocusedStub.(gjson.Result)
			}

			result = backup.Get(v.(string))
			res := p.getOneSelector(key, v, cfg, result).(gjson.Result)
			arr = append(arr, res)
		}
		iface = arr
		isComplexSel = true
	case map[string]interface{}:
		dat := make(map[string]gjson.Result)
		backup := result
		for k, v := range sel {
			result = backup.Get(v.(string))
			res := p.getOneSelector(key, v, cfg, result).(gjson.Result)
			dat[k] = res
		}
		iface = dat
		isComplexSel = true
	default:
		panic(fmt.Sprintf("unsupported key (%T: %s)", sel, sel))
	}

	return iface, isComplexSel
}

func (p *JsonParser) getOneSelector(key string, sel interface{}, cfg map[string]interface{}, result gjson.Result) (iface interface{}) {
	index, exist := cfg[INDEX]
	if index == nil {
		if !exist {
			return p.getIndex(sel, result, 0)
		}
		return result.Array()
	}

	switch val := index.(type) {
	case int:
		return p.getIndex(sel, result, val)
	case []interface{}:
		var d []gjson.Result
		for _, v := range val {
			switch v := v.(type) {
			case int:
				r := p.getIndex(sel, result, v)
				d = append(d, r)
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
	elems, complexSel := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case gjson.Result:
		data[key] = p.getSelectionAttr(key, cfg, dom)

	case []gjson.Result:
		if !complexSel {
			var subData []interface{}
			for _, dm := range dom {
				d := p.getSelectionAttr(key, cfg, dm)
				subData = append(subData, d)
			}
			data[key] = subData
		} else {
			data[key] = p.getSelectionSliceAttr(key, cfg, dom)
		}
	case map[string]gjson.Result:
		if !complexSel {
			subData := make(map[string]interface{})
			for k, dm := range dom {
				d := p.getSelectionAttr(key, cfg, dm)
				subData[k] = d
			}
			data[key] = subData
		} else {
			data[key] = p.getSelectionMapAttr(key, cfg, dom)
		}
	default:
		panic(xpretty.Redf("unknown type of dom %s:%v %v", key, cfg, dom))
	}
}
