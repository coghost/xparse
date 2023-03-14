package xparse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
	"golang.org/x/net/html"
)

type HtmlParser struct {
	*Parser
}

func NewHtmlParser(rawHtml []byte, ymlMap ...[]byte) *HtmlParser {
	p := &HtmlParser{
		NewParser(rawHtml, ymlMap...),
	}
	p.Spawn(rawHtml, ymlMap...)
	return p
}

func (p *HtmlParser) Spawn(raw []byte, ymlCfg ...[]byte) {
	p.LoadConfig(ymlCfg...)
	p.LoadRootSelection(raw)
}

func (p *HtmlParser) LoadRootSelection(raw []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	PanicIfErr(err)
	p.Root = doc.Selection
}

func (p *HtmlParser) DoParse() {
	p.runCheck()
	for key, cfg := range p.config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.rankOffset = 0
			p.parseDom(key, cfgType, p.Root.(*goquery.Selection), p.ParsedData, _layerForRank)
		default:
			fmt.Println(xpretty.Redf(_nonMapHint, key, cfg))
			continue
		}
	}
	p.PostDoParse()
}

// parseDom
// only support two data type
// 1. str
// 2. map[string]interface{}
func (p *HtmlParser) parseDom(key string, cfg interface{}, selection *goquery.Selection, data map[string]interface{}, layer int) {
	p.checkNestedKeys(key)
	defer p.popNestedKeys()

	b := p.isRequiredKey(key)
	// xpretty.DummyLog(key, p.testKeys, b, p.forceParsedKey, p.nestedKeys)
	if !b {
		return
	}

	if funk.IsEmpty(cfg) {
		data[key] = p.getSelectionAttr(key, map[string]interface{}{key: ""}, selection)
		return
	}

	switch v := cfg.(type) {
	case string:
		// the recursive end condition
		p.handleStr(key, v, selection, data)
	case map[string]interface{}:
		p.handleMap(key, v, selection, data, layer)
	default:
		panic(xpretty.Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]interface{})", key, cfg))
	}
}

func (p *HtmlParser) handleStr(key string, sel string, selection *goquery.Selection, data map[string]interface{}) {
	data[key] = selection.Find(sel).First().Text()
}

// handleMap
//  1. find all matched elems
//     1.1. found only 1 node
//     1.2. found more than 1 node
func (p *HtmlParser) handleMap(
	key string,
	cfg map[string]interface{},
	selection *goquery.Selection,
	data map[string]interface{},
	layer int,
) {
	if p.isLeaf(cfg) {
		p.getNodesAttrs(key, cfg, selection, data)
		return
	}

	elems, _ := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Selection:
		subData := make(map[string]interface{})
		data[key] = subData
		p.parseDomNodes(cfg, dom, subData)

	case []*goquery.Selection:
		var allSubData []map[string]interface{}
		for _, gs := range dom {
			if layer == _layerForRank {
				p.FocusedStub = gs
			}

			// only calculate rank at first layer
			if layer == _layerForRank {
				p.setRank(cfg)
			}

			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parseDomNodes(cfg, gs, subData)
		}
		data[key] = allSubData
	}
}

func (p *HtmlParser) parseDomNodes(
	cfg map[string]interface{},
	selection *goquery.Selection,
	data map[string]interface{},
) {
	for k, sc := range cfg {
		if strings.HasPrefix(k, "_") {
			continue
		}
		p.parseDom(k, sc, selection, data, _layerForOthers)
	}
}

func (p *HtmlParser) getAllElems(key string, cfg map[string]interface{}, selection *goquery.Selection) (iface interface{}, isComplexSel bool) {
	sel := cfg[Locator]
	if sel == nil {
		return selection, isComplexSel
	}

	isComplexSel = true

	switch sel := sel.(type) {
	case string:
		if !strings.Contains(sel, ",") {
			iface, isComplexSel = p.getOneSelector(key, sel, cfg, selection)
		} else {
			iface = p.getElemsOneByOne(key, strings.Split(sel, ","), cfg, selection)
		}
	case []interface{}:
		var ss []string
		for _, v := range sel {
			ss = append(ss, v.(string))
		}
		iface = p.getElemsOneByOne(key, ss, cfg, selection)
	case map[string]interface{}:
		dat := make(map[string]*goquery.Selection)
		backup := selection

		for k, v := range sel {
			v, backup = p.handleStub(v, backup)
			res, _ := p.getOneSelector(key, v, cfg, backup)
			dat[k] = res.(*goquery.Selection)
		}
		iface = dat
	default:
		panic(fmt.Sprintf("unsupported key (%T: %s)", sel, sel))
	}

	return
}

func (p *HtmlParser) handleStub(raw interface{}, result *goquery.Selection) (interface{}, *goquery.Selection) {
	ar1 := strings.Split(raw.(string), ".")
	if ar1[0] == PrefixLocatorStub {
		raw = strings.Join(ar1[1:], ".")
		result = p.FocusedStub.(*goquery.Selection)
	}
	return raw, result
}

func (p *HtmlParser) getElemsOneByOne(key string, selArr []string, cfg map[string]interface{}, selection *goquery.Selection) (iface []*goquery.Selection) {
	// selArr := strings.Split(sel, ",")
	var resArr []*goquery.Selection
	backup := selection

	for _, v := range selArr {
		v1, backup := p.handleStub(v, backup)
		v = v1.(string)
		elem, _ := p.getOneSelector(key, v, cfg, backup)
		switch val := elem.(type) {
		case *goquery.Selection:
			resArr = append(resArr, val)
		case []*goquery.Selection:
			resArr = append(resArr, val...)
		}
	}
	return resArr
}

func (p *HtmlParser) getOneSelector(key string, sel interface{}, cfg map[string]interface{}, selection *goquery.Selection) (iface interface{}, isComplexSel bool) {
	elems := selection.Find(sel.(string))
	index := cfg[Index]
	isComplexSel = strings.Contains(sel.(string), ",")

	iface = p.handleNullIndexOnly(key, isComplexSel, cfg, elems)
	if iface != nil {
		return
	}

	switch val := index.(type) {
	case int:
		iface = elems.Eq(cast.ToInt(val))
	case uint64:
		iface = elems.Eq(cast.ToInt(val))
	case string:
		arr := strings.Split(val, ",")
		if len(arr) != 2 {
			panic(xpretty.Redf("range index format must be (a-b), but (%s is %T: %v)\n", key, val, val))
		}
		start, end := 0, len(elems.Nodes)
		if v := arr[0]; v != "" {
			start = getIndex(key, v, len(elems.Nodes))
		}
		if v := arr[1]; v != "" {
			end = getIndex(key, v, len(elems.Nodes))
		}
		var d []*goquery.Selection
		for i := start; i < end; i++ {
			d = append(d, elems.Eq(i))
		}
		iface = d
	case []interface{}:
		var d []*goquery.Selection
		for _, v := range val {
			switch v := v.(type) {
			case int:
				d = append(d, elems.Eq(v))
			case uint64:
				d = append(d, elems.Eq(cast.ToInt(v)))
			default:
				panic(xpretty.Redf("all indexes should be int, but (%s is %T: %v)\n", key, val, val))
			}
		}
		iface = d
	default:
		panic(xpretty.Redf("index should be int or []interface{}, but (%s is %T: %v)\n", key, val, val))
	}

	return
}

func getIndex(key string, intStr string, total int) int {
	end, err := cast.ToIntE(intStr)
	if err != nil {
		panic(xpretty.Redf("range index must be number, but (%s is %T: %v)", key, intStr, intStr))
	}
	if end < 0 {
		end += total
	}

	return end
}

func (p *HtmlParser) handleNullIndexOnly(key string, isComplexSel bool, cfg map[string]interface{}, elems *goquery.Selection) interface{} {
	// index has 4 types:
	//  1. without index equal with 3(index:0)
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	// if index existed, just return nil
	index, existed := cfg[Index]
	if index != nil {
		return nil
	}

	// a index which not existed, is a shortcut for index:0, so just return the first selection
	if !existed {
		if isComplexSel {
			return p.getAllSelections(elems)
		}
		if s, ok := cfg[ExtractPrevElem]; !ok {
			return elems.First()
		} else {
			return p.extractNode(key, s, elems)
		}
	}

	// if index is yaml's null: '~' or null
	return p.getAllSelections(elems)
}

func (p *HtmlParser) extractNode(key string, sel interface{}, elems *goquery.Selection) interface{} {
	switch s := sel.(type) {
	case bool:
		return elems.Prev()
	case string:
		if s == "__prev" {
			return elems.Prev()
		} else {
			return elems.PrevFiltered(s)
		}
	default:
		panic(xpretty.Redf("action _extract_prev only support bool and string, but (%s's %v is %T: %v)", key, ExtractPrevElem, sel, sel))
	}
}

func (p *HtmlParser) getAllSelections(elems *goquery.Selection) []*goquery.Selection {
	var d []*goquery.Selection
	for i := range elems.Nodes {
		d = append(d, elems.Eq(i))
	}
	return d
}

func (p *HtmlParser) getNodesAttrs(
	key string,
	cfg map[string]interface{},
	selection *goquery.Selection,
	data map[string]interface{},
) {
	// fmt.Printf("Got %v, %T, %v\n", key, cfg, cfg)
	elems, complexSel := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Document:
		panic("found Doc, Selection Required!")

	case *goquery.Selection:
		data[key] = p.getSelectionAttr(key, cfg, dom)

	case []*goquery.Selection:
		if !complexSel {
			var subData []interface{}
			for _, dm := range dom {
				d := p.getSelectionAttr(key, cfg, dm)
				subData = append(subData, d)
			}
			d := p.postJoin(cfg, subData)
			data[key] = d
		} else {
			data[key] = p.getSelectionSliceAttr(key, cfg, dom)
		}
	case map[string]*goquery.Selection:
		if !complexSel {
			panic("not supported")
		}
		data[key] = p.getSelectionMapAttr(key, cfg, dom)
	default:
		panic(xpretty.Redf("unknown type of dom %s:%v %v", key, cfg, dom))
	}
}

func (p *HtmlParser) postJoin(cfg map[string]interface{}, data []interface{}) interface{} {
	pj, b := cfg[PostJoin]
	if !b {
		return data
	}

	joiner := p.getJoinerOr(cfg, "")
	switch v := pj.(type) {
	case string:
		joiner = v
	}

	var arr []string
	for _, v := range data {
		arr = append(arr, v.(string))
	}

	return strings.Join(arr, joiner)
}

func (p *HtmlParser) getSelectionSliceAttr(key string, cfg map[string]interface{}, resultArr []*goquery.Selection) interface{} {
	var resArr []string
	for _, v := range resultArr {
		raw := p.getRawAttr(cfg, v)
		resArr = append(resArr, raw.(string))
	}
	// joiner := p.getJoinerOr(cfg, AttrJoinerSep)
	// v := p.refineAttr(key, strings.Join(resArr, joiner), cfg, resultArr)
	v := p.refineAttr(key, resArr, cfg, resultArr)
	return p.convertToType(v, cfg)
}

func (p *HtmlParser) getSelectionMapAttr(key string, cfg map[string]interface{}, results map[string]*goquery.Selection) interface{} {
	dat := make(map[string]string)

	for k, v := range results {
		raw := p.getRawAttr(cfg, v)
		dat[k] = raw.(string)
	}
	str, _ := Stringify(dat)
	v := p.refineAttr(key, str, cfg, results)
	return p.convertToType(v, cfg)
}

func (p *HtmlParser) getSelectionAttr(key string, cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	raw := p.getRawAttr(cfg, selection)
	raw = p.stripChars(key, raw, cfg)
	raw = p.refineAttr(key, raw, cfg, selection)

	return p.convertToType(raw, cfg)
}

func (p *HtmlParser) getRawAttr(cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	attr := cfg[Attr]

	// fmt.Printf("Got %T, %v\n", attr, attr)
	if attr == nil {
		v := selection.Text()
		return p.TrimSpace(v, cfg)
	}

	if attr == AttrJoinElemsText {
		joiner := p.getJoinerOr(cfg, AttrJoinerSep)

		elems := selection.Contents()
		var arr []string
		for _, elem := range elems.Nodes {
			if elem.Type != html.TextNode {
				continue
			}
			v := elem.Data
			arr = append(arr, v)
		}

		return strings.Join(arr, joiner)
	}

	switch attrType := attr.(type) {
	case string:
		v := selection.AttrOr(attrType, "")
		return p.TrimSpace(v, cfg)
	case []interface{}:
		d := make(map[string]interface{})
		for _, at := range attrType {
			v := selection.AttrOr(at.(string), "")
			d[at.(string)] = p.TrimSpace(v, cfg)
		}
		return d
	default:
		panic(xpretty.Redf("attr should be (string or []interface{}), but (%s is %T: %v)\n", attr, attrType, attrType))
	}
}
