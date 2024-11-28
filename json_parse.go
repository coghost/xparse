package xparse

import (
	"fmt"
	"os"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

type JSONParser struct {
	*Parser
}

func NewJSONParser(rawData []byte, ymlMap ...[]byte) *JSONParser {
	parser := &JSONParser{
		NewParser(rawData, ymlMap...),
	}

	parser.Spawn(rawData, ymlMap...)

	return parser
}

func (p *JSONParser) Spawn(raw []byte, ymlCfg ...[]byte) {
	p.LoadConfig(ymlCfg...)
	p.LoadRootSelection(raw)
}

func (p *JSONParser) LoadRootSelection(raw []byte) {
	p.RawData = string(raw)
	p.Root = gjson.Parse(string(raw))
}

func (p *JSONParser) DoParse() {
	p.runCheck()

	for key, cfg := range p.config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.rankOffset = 0
			result, _ := p.Root.(gjson.Result)
			p.parseDom(key, cfgType, result, p.ParsedData, _layerForRank)
		default:
			fmt.Fprint(os.Stderr, xpretty.Redf(_nonMapHint, key, cfg))
			continue
		}
	}

	p.PostDoParse()
	p.RefineJobsWithPreset()
}

func (p *JSONParser) parseDom(key string, cfg interface{}, result gjson.Result, data map[string]interface{}, layer int) {
	p.appendNestedKeys(key)
	defer p.popNestedKeys()

	b := p.isRequiredKey(key)
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
		p.handleMap(key, v, result, data, layer)
	default:
		panic(xpretty.Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]interface{})", key, cfg))
	}
}

func (p *JSONParser) getSelectionAttr(key string, cfg map[string]interface{}, result gjson.Result) interface{} {
	var raw interface{}
	raw = result.String()
	raw = p.stripChars(key, raw, cfg)
	raw = p.refineAttr(key, raw, cfg, result)
	raw = p.advancedPostRefineAttr(raw, cfg)

	return p.convertToType(raw, cfg)
}

func (p *JSONParser) getSelectionSliceAttr(key string, cfg map[string]interface{}, resultArr []gjson.Result) interface{} {
	var raw []string
	for _, v := range resultArr {
		raw = append(raw, v.String())
	}

	joiner := p.getJoinerOrDefault(cfg, AttrJoinerSep)
	v := p.refineAttr(key, strings.Join(raw, joiner), cfg, resultArr)
	v = p.advancedPostRefineAttr(v, cfg)

	return p.convertToType(v, cfg)
}

func (p *JSONParser) getSelectionMapAttr(key string, cfg map[string]interface{}, results map[string]gjson.Result) interface{} {
	dat := make(map[string]string)
	for k, v := range results {
		dat[k] = v.String()
	}

	str, _ := Stringify(dat)
	v := p.refineAttr(key, str, cfg, results)
	v = p.advancedPostRefineAttr(v, cfg)

	return p.convertToType(v, cfg)
}

func (p *JSONParser) handleStr(key string, sel string, result gjson.Result, data map[string]interface{}) {
	data[key] = result.Get(sel).String()
}

func (p *JSONParser) handleMap(
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
		p.parseDomNodes(cfg, dom, subData)

		if layer == _layerForRank {
			p.FocusedStub = dom
		}

	case []gjson.Result:
		var allSubData []map[string]interface{}

		for _, result := range dom {
			if layer == _layerForRank {
				p.FocusedStub = result
			}

			// only calculate rank at first layer
			if layer == _layerForRank {
				p.setRank(cfg)
			}

			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parseDomNodes(cfg, result, subData)
		}

		data[key] = allSubData
	}
}

func (p *JSONParser) getResultAtIndex(_ interface{}, result gjson.Result, index int) (rs gjson.Result) {
	arr := result.Array()
	if len(arr) > index {
		return arr[index]
	}

	return
}

func (p *JSONParser) getAllElems(key string, cfg map[string]interface{}, result gjson.Result) (iface interface{}, isComplexSel bool) {
	sel := mustCfgLocator(cfg)
	if sel == nil {
		return result, false
	}

	switch sel := sel.(type) {
	case string:
		if sel == JSONArrayRootLocator {
			result = gjson.Parse(p.RawData)
		} else {
			result = result.Get(sel)
		}

		iface = p.getOneSelector(key, sel, cfg, result)
	case []interface{}:
		arr := []gjson.Result{}
		backup := result

		for _, v := range sel {
			v, backup = p.handleStub(v, backup)
			v1, _ := v.(string)
			result = backup.Get(v1)

			res := p.getOneSelector(key, v, cfg, result)
			gRes, _ := res.(gjson.Result)
			arr = append(arr, gRes)
		}

		iface = arr
		isComplexSel = true
	case map[string]interface{}:
		dat := make(map[string]gjson.Result)
		backup := result

		for selK, v := range sel {
			v, backup = p.handleStub(v, backup)
			v1, _ := v.(string)
			result = backup.Get(v1)

			res := p.getOneSelector(key, v, cfg, result)
			gRes, _ := res.(gjson.Result)

			dat[selK] = gRes
		}

		iface = dat
		isComplexSel = true
	default:
		panic(fmt.Sprintf("unsupported key (%T: %s)", sel, sel))
	}

	return iface, isComplexSel
}

func (p *JSONParser) handleStub(raw interface{}, result gjson.Result) (interface{}, gjson.Result) {
	rawStr, _ := raw.(string)
	arr := strings.Split(rawStr, ".")

	if arr[0] == PrefixLocatorStub {
		raw = strings.Join(arr[1:], ".")
		result, _ = p.FocusedStub.(gjson.Result)
	}

	return raw, result
}

func (p *JSONParser) getOneSelector(key string, sel interface{}, cfg map[string]interface{}, result gjson.Result) (iface interface{}) {
	index, existed := cfgIndex(cfg)
	if index == nil {
		if !existed {
			return p.getResultAtIndex(sel, result, 0)
		}

		return result.Array()
	}

	switch val := index.(type) {
	case int, int64, uint64:
		return p.getResultAtIndex(sel, result, cast.ToInt(val))
	case string:
		total := len(result.Array())

		indexes := ParseNumberRanges(val)
		if len(indexes) != 0 {
			var d []gjson.Result

			for _, idx := range indexes {
				i := idx
				if idx < 0 {
					i = idx + total
				}

				d = append(d, result.Array()[i])
			}

			return d
		}

		arr := strings.Split(val, ",")
		if len(arr) != _rangeIndexLen {
			panic(xpretty.Redf("range index format must be (a-b), but (%s is %T: %v)", key, val, val))
		}

		start, end := 0, total

		if v := arr[0]; v != "" {
			start = refineIndex(key, v, total)
		}

		if v := arr[1]; v != "" {
			end = refineIndex(key, v, total)
		}

		var d []gjson.Result
		for i := start; i < end; i++ {
			d = append(d, result.Array()[i])
		}

		return d
	case []interface{}:
		var resArr []gjson.Result

		for _, v := range val {
			switch v := v.(type) {
			case int, uint64, int64:
				r := p.getResultAtIndex(sel, result, cast.ToInt(v))
				resArr = append(resArr, r)
			default:
				panic(xpretty.Redf("all indexes should be int, but (%s is %T: %v)", key, val, val))
			}
		}

		return resArr
	default:
		panic(xpretty.Redf("index should be int or []interface{}, but (%s is %T: %v)", key, val, val))
	}
}

func (p *JSONParser) parseDomNodes(
	cfg map[string]interface{},
	result gjson.Result,
	data map[string]interface{},
) {
	for k, sc := range cfg {
		if strings.HasPrefix(k, "_") {
			continue
		}

		p.parseDom(k, sc, result, data, _layerForOthers)
	}
}

func (p *JSONParser) getNodesAttrs(
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
