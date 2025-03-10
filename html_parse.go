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

type HTMLParser struct {
	*Parser
}

func NewHTMLParser(rawHTML []byte, ymlMap ...[]byte) *HTMLParser {
	p := &HTMLParser{
		Parser: NewParser(rawHTML, ymlMap...),
	}
	p.Spawn(rawHTML, ymlMap...)

	return p
}

func (p *HTMLParser) Spawn(raw []byte, ymlCfg ...[]byte) {
	p.LoadConfig(ymlCfg...)
	p.LoadRootSelection(raw)
}

func (p *HTMLParser) LoadRootSelection(raw []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	PanicIfErr(err)

	p.Root = doc.Selection
}

func (p *HTMLParser) DoParse() {
	p.runCheck()

	for key, cfg := range p.config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]any:
			p.rankOffset = 0
			selection, _ := p.Root.(*goquery.Selection)
			p.parseDom(key, cfgType, selection, p.ParsedData, _layerForRank)
		default:
			xpretty.RedPrintf(_nonMapHint, key, cfg)
			continue
		}
	}

	p.PostDoParse()
	p.RefineJobsWithPreset()
}

// parseDom
// only support two data type
// 1. str
// 2. map[string]any
func (p *HTMLParser) parseDom(key string, cfg any, selection *goquery.Selection, data map[string]any, layer int) {
	p.appendNestedKeys(key)
	defer p.popNestedKeys()

	b := p.isRequiredKey(key)
	// xpretty.DummyLog(key, p.testKeys, b, p.forceParsedKey, p.nestedKeys)
	if !b {
		return
	}

	if funk.IsEmpty(cfg) {
		data[key] = p.getSelectionAttr(key, map[string]any{key: ""}, selection)
		return
	}

	switch v := cfg.(type) {
	case string:
		// the recursive end condition
		p.handleStr(key, v, selection, data)
	case map[string]any:
		p.handleMap(key, v, selection, data, layer)
	default:
		panic(xpretty.Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]any)", key, cfg))
	}
}

func (p *HTMLParser) handleStr(key string, sel string, selection *goquery.Selection, data map[string]any) {
	data[key] = selection.Find(sel).First().Text()
}

// handleMap
//  1. find all matched elems
//     1.1. found only 1 node
//     1.2. found more than 1 node
func (p *HTMLParser) handleMap(
	key string,
	cfg map[string]any,
	selection *goquery.Selection,
	data map[string]any,
	layer int,
) {
	if p.isLeaf(cfg) {
		p.getNodesAttrs(key, cfg, selection, data)
		return
	}

	elems, _ := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Selection:
		subData := make(map[string]any)
		data[key] = subData
		p.parseDomNodes(cfg, dom, subData)
	case []*goquery.Selection:
		var allSubData []map[string]any

		for _, selection := range dom {
			if layer == _layerForRank {
				p.FocusedStub = selection
			}

			// only calculate rank at first layer
			if layer == _layerForRank {
				p.setRank(cfg)
			}

			subData := make(map[string]any)
			allSubData = append(allSubData, subData)

			p.parseDomNodes(cfg, selection, subData)
		}

		data[key] = allSubData
	}
}

func (p *HTMLParser) parseDomNodes(
	cfg map[string]any,
	selection *goquery.Selection,
	data map[string]any,
) {
	for k, sc := range cfg {
		if strings.HasPrefix(k, "_") {
			continue
		}

		p.parseDom(k, sc, selection, data, _layerForOthers)
	}
}

func (p *HTMLParser) getAllElems(key string,
	cfg map[string]any,
	selection *goquery.Selection,
) (iface any, isComplexSel bool) {
	selCfg := mustCfgLocator(cfg)
	if selCfg == nil {
		return selection, false
	}

	isComplexSel = true

	switch selCfg := selCfg.(type) {
	case string:
		if !strings.Contains(selCfg, ",") {
			// cfg likes: (_locator: div.title)
			// so selCfg is div.title
			iface, isComplexSel = p.getOneSelector(key, selCfg, cfg, selection)
		} else {
			// cfg likes: `_locator: div.title,h2.title,h3.title`
			// so selCfg is `div.title,h2.title,h3.title`
			iface = p.getElemsOneByOne(key, strings.Split(selCfg, ","), cfg, selection)
		}
	case []any:
		var selArr []string
		// cfg likes
		// _locator:
		//   - div.title
		//   - h2.title
		// so selCfg is slice of ["div.title", "h2.title"]
		for _, v := range selCfg {
			s, ok := v.(string)
			if !ok {
				panic(fmt.Sprintf("selector of %s require string, but got %v", key, v))
			}

			selArr = append(selArr, s)
		}

		iface = p.getElemsOneByOne(key, selArr, cfg, selection)
	case map[string]any:
		// cfg likes
		// _locator:
		//   divTitle: div.title
		//   h2Title: h2.title
		// so selCfg is map of {"divTitle":"div.title", "h2Title":"span.ratingNumber"}
		dat := make(map[string]*goquery.Selection)
		backup := selection

		for dataKey, subCfg := range selCfg {
			subCfg, backup = p.handleStub(subCfg, backup)

			res, _ := p.getOneSelector(key, subCfg, cfg, backup)
			dat[dataKey], _ = res.(*goquery.Selection)
		}

		iface = dat
	default:
		panic(fmt.Sprintf("unsupported key (%T: %s)", selCfg, selCfg))
	}

	return iface, isComplexSel
}

func (p *HTMLParser) handleStub(raw any, result *goquery.Selection) (any, *goquery.Selection) {
	key, ok := raw.(string)
	if !ok {
		panic(fmt.Sprintf("locator require string, but got (%T: %v)", raw, raw))
	}

	arr := strings.Split(key, ".")

	if arr[0] == PrefixLocatorStub {
		raw = strings.Join(arr[1:], ".")
		result, _ = p.FocusedStub.(*goquery.Selection)
	}

	return raw, result
}

func (p *HTMLParser) getElemsOneByOne(key string, selArr []string,
	cfg map[string]any, selection *goquery.Selection,
) []*goquery.Selection {
	var resArr []*goquery.Selection

	backup := selection

	for _, v := range selArr {
		v1, backup := p.handleStub(v, backup)
		v, _ = v1.(string)

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

func (p *HTMLParser) getOneSelector(key string, sel any,
	cfg map[string]any, selection *goquery.Selection,
) (any, bool) {
	selStr, _ := sel.(string)
	elems := selection.Find(selStr)
	index := mustCfgIndex(cfg)

	isComplexSel := strings.Contains(selStr, ",")

	iface := p.handleNullIndexOnly(key, isComplexSel, cfg, elems)
	if iface != nil {
		return iface, isComplexSel
	}

	switch val := index.(type) {
	case int, int64, uint64:
		iface = elems.Eq(cast.ToInt(val))
	case string:
		total := len(elems.Nodes)

		indexes := ParseNumberRanges(val)
		if len(indexes) != 0 {
			var selections []*goquery.Selection

			for _, idx := range indexes {
				i := idx
				if idx < 0 {
					i = idx + total
				}

				if i < total {
					selections = append(selections, elems.Eq(i))
				}
			}

			return selections, isComplexSel
		}

		arr := strings.Split(val, ",")
		if len(arr) != _rangeIndexLen {
			panic(xpretty.Redf("range index format must be (a-b), but (%s is %T: %v)", key, val, val))
		}

		start, end := 0, len(elems.Nodes)
		if v := arr[0]; v != "" {
			start = refineIndex(key, v, len(elems.Nodes))
		}

		if v := arr[1]; v != "" {
			end = refineIndex(key, v, len(elems.Nodes))
		}

		var d []*goquery.Selection
		for i := start; i < end; i++ {
			d = append(d, elems.Eq(i))
		}

		iface = d
	case []any:
		var selArr []*goquery.Selection

		for _, v := range val {
			switch v := v.(type) {
			case int, uint64, int64:
				selArr = append(selArr, elems.Eq(cast.ToInt(v)))
			default:
				panic(xpretty.Redf("all indexes should be int, but (%s is %T: %v)", key, val, val))
			}
		}

		iface = selArr
	default:
		panic(xpretty.Redf("index should be int/int64/uint64 or []any, but (%s is %T: %v)", key, val, val))
	}

	return iface, isComplexSel
}

func (p *HTMLParser) handleNullIndexOnly(key string, isComplexSel bool, cfg map[string]any, elems *goquery.Selection) any {
	// index has 4 types:
	//  1. without index equal with type:3(index:0)
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	// if index existed, just return nil
	index, existed := cfgIndex(cfg)
	if index != nil {
		return nil
	}

	// a index which not existed, is a shortcut for index:0, so just return the first selection
	if !existed {
		if isComplexSel {
			return p.getAllSelections(elems)
		}

		if s, ok := cfg[ExtractParent]; ok {
			return p.extractParent(key, s, elems)
		}

		if s, ok := cfg[ExtractPrevElem]; !ok {
			return elems.First()
		} else {
			return p.extractPrevNode(key, s, elems)
		}
	}

	// if index is yaml's null: '~' or null
	return p.getAllSelections(elems)
}

func (p *HTMLParser) extractParent(key string, sel any, elems *goquery.Selection) any {
	switch preSel := sel.(type) {
	case bool:
		return elems.Parent()
	case uint64, int, int64:
		n := cast.ToInt(preSel)
		for i := 0; i < n; i++ {
			elems = elems.Parent()
		}

		return elems
	}

	panic(xpretty.Redf("action _extract_parent only support bool and int, but (%s's %v is %T: %v)", key, ExtractParent, sel, sel))
}

func (p *HTMLParser) extractPrevNode(key string, sel any, elems *goquery.Selection) any {
	switch preSel := sel.(type) {
	case bool:
		return elems.Prev()
	case string:
		if preSel == "__prev" {
			return elems.Prev()
		} else {
			return elems.PrevFiltered(preSel)
		}
	default:
		panic(xpretty.Redf("action _extract_prev only support bool and string, but (%s's %v is %T: %v)", key, ExtractPrevElem, sel, sel))
	}
}

func (p *HTMLParser) getAllSelections(elems *goquery.Selection) []*goquery.Selection {
	var d []*goquery.Selection
	for i := range elems.Nodes {
		d = append(d, elems.Eq(i))
	}

	return d
}

func (p *HTMLParser) getNodesAttrs(
	key string,
	cfg map[string]any,
	selection *goquery.Selection,
	data map[string]any,
) {
	// first of all, check if _raw is set or not.
	if val := mustCfgRaw(cfg); val != nil && val != "" {
		data[key] = p.convertToType(val, cfg)
		return
	}

	elems, complexSel := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Document:
		panic("found Doc, Selection Required!")

	case *goquery.Selection:
		data[key] = p.getSelectionAttr(key, cfg, dom)

	case []*goquery.Selection:
		if !complexSel {
			var subData []any

			for _, dm := range dom {
				d := p.getSelectionAttr(key, cfg, dm)
				subData = append(subData, d)
			}

			d := p.postJoin(cfg, subData)
			data[key] = d
		} else {
			cpxIface := p.getSelectionSliceAttr(key, cfg, dom)

			switch ifc := cpxIface.(type) {
			case []string:
				var sd []any
				for _, k := range ifc {
					sd = append(sd, k)
				}

				data[key] = p.postJoin(cfg, sd)
			case []any:
				d := p.postJoin(cfg, ifc)
				data[key] = d
			default:
				data[key] = ifc
			}
		}
	case map[string]*goquery.Selection:
		if !complexSel {
			subData := make(map[string]any)

			for k, dm := range dom {
				d := p.getSelectionAttr(key, cfg, dm)
				subData[k] = d
			}

			data[key] = subData
		}

		data[key] = p.getSelectionMapAttr(key, cfg, dom)
	default:
		panic(xpretty.Redf("unknown type of dom %s:%v %v", key, cfg, dom))
	}
}

func (p *HTMLParser) postJoin(cfg map[string]any, data []any) any {
	postJoin, b := cfg[PostJoin]
	if !b {
		return data
	}

	joiner := p.getJoinerOrDefault(cfg, "")

	if v, ok := postJoin.(string); ok {
		joiner = v
	}

	var arr []string

	for _, v := range data {
		v1, _ := v.(string)
		arr = append(arr, v1)
	}

	return strings.Join(arr, joiner)
}

func (p *HTMLParser) getSelectionSliceAttr(key string, cfg map[string]any, resultArr []*goquery.Selection) any {
	var resArr []string

	for _, v := range resultArr {
		raw := p.getRawAttr(cfg, v)
		str, _ := raw.(string)
		resArr = append(resArr, str)
	}
	// joiner := p.getJoinerOr(cfg, AttrJoinerSep)
	// v := p.refineAttr(key, strings.Join(resArr, joiner), cfg, resultArr)
	v := p.refineAttr(key, resArr, cfg, resultArr)
	v = p.advancedPostRefineAttr(v, cfg)

	return p.convertToType(v, cfg)
}

func (p *HTMLParser) getSelectionMapAttr(key string, cfg map[string]any, results map[string]*goquery.Selection) any {
	dat := make(map[string]string)

	for k, v := range results {
		raw := p.getRawAttr(cfg, v)
		dat[k], _ = raw.(string)
	}

	str, _ := Stringify(dat)
	v := p.refineAttr(key, str, cfg, results)
	v = p.advancedPostRefineAttr(v, cfg)

	return p.convertToType(v, cfg)
}

func (p *HTMLParser) getSelectionAttr(key string, cfg map[string]any, selection *goquery.Selection) any {
	raw := p.getRawAttr(cfg, selection)
	raw = p.stripChars(key, raw, cfg)
	raw = p.refineAttr(key, raw, cfg, selection)
	raw = p.advancedPostRefineAttr(raw, cfg)

	return p.convertToType(raw, cfg)
}

func (p *HTMLParser) getRawAttr(cfg map[string]any, selection *goquery.Selection) any {
	attr := cfg[Attr]

	if attr == nil {
		v := selection.Text()
		return p.TrimSpace(v, cfg)
	}

	if attr == "__html" {
		v, err := selection.Html()
		if err != nil {
			panic(v)
		}

		return v
	}

	if attr == AttrJoinElemsText {
		var arr []string

		joiner := p.getJoinerOrDefault(cfg, AttrJoinerSep)
		elems := selection.Contents()

		for _, elem := range elems.Nodes {
			if elem.Type != html.TextNode {
				continue
			}

			arr = append(arr, elem.Data)
		}

		return strings.Join(arr, joiner)
	}

	switch attrType := attr.(type) {
	case string:
		v := selection.AttrOr(attrType, "")
		return p.TrimSpace(v, cfg)
	case []any:
		cplxAttr := make(map[string]any)

		for _, at := range attrType {
			atStr, _ := at.(string)
			v := selection.AttrOr(atStr, "")
			cplxAttr[atStr] = p.TrimSpace(v, cfg)
		}

		return cplxAttr
	default:
		panic(xpretty.Redf("attr should be (string or []any), but (%s is %T: %v)", attr, attrType, attrType))
	}
}
