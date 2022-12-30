package xparse

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/coghost/xpretty"
	"github.com/ghodss/yaml"
	"github.com/gookit/config/v2"
	"github.com/iancoleman/strcase"
	"github.com/k0kubun/pp/v3"
	"github.com/shomali11/util/xconversions"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

const (
	nonMapHint = "[NON-MAP] {%v:%v}, please move into a map instead"
)

const (
	layerForRank = iota + 1
	layerForOthers
)

type Parser struct {
	Config *config.Config
	Root   *goquery.Selection
	JRoot  gjson.Result

	// this is a map's stub, check PREFIX_LOCATOR_STUB for more info
	FocusedStub interface{}

	RawData string

	// devMode
	devMode bool
	rank    int

	// map to config
	ParsedData map[string]interface{}

	// testKeys, only keys in testKeys will be parsed
	testKeys []string
	//
	forceParsedKey bool
	nestedKeys     []string

	// selectedKeys []string

	// Refiners is a map of
	//  > string: func
	//  - string is name we defined
	//  - func has three params:
	//    + first params is string, which is the raw str get from html (usually by get_text/get_attr)
	//    + second params is the *config.Config (which is rarely used)
	//    + third params is *goquery.Selection
	Refiners map[string]func(raw ...interface{}) interface{}

	AttrToBeRefined []string
}

func NewHtmlParser(rawHtml, ymlMap []byte) *Parser {
	p := &Parser{
		Config:     &config.Config{},
		ParsedData: make(map[string]interface{}),
		Refiners:   make(map[string]func(args ...interface{}) interface{}),
	}
	p.Spawn(rawHtml, ymlMap)

	return p
}

func (p *Parser) Spawn(raw, ymlCfg []byte) {
	p.LoadConfig(ymlCfg)
	p.LoadRootSelection(raw)
}

func (p *Parser) ToggleDevMode(b bool) {
	p.devMode = b
}

func (p *Parser) Debug(key interface{}, raw ...interface{}) {
	if p.devMode {
		pp.Println(fmt.Sprintf("[%d] %v: (%v)", p.rank, key, raw[0]))
	}
}

func (p *Parser) LoadRootSelection(raw []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	PanicIfErr(err)
	p.Root = doc.Selection
}

func (p *Parser) LoadConfig(ymlCfg []byte) {
	p.Config = Yaml2Config(ymlCfg)
	p.testKeys = p.Config.Strings("__raw.test_keys")
}

func (p *Parser) GetRawInfo() map[string]interface{} {
	raw := p.Config.Data()["__raw"]
	return raw.(map[string]interface{})
}

func (p *Parser) GetParsedData() map[string]interface{} {
	return p.ParsedData
}

func (p *Parser) PrettifyJsonData(args ...interface{}) {
	xpretty.PrettyJson(p.MustDataAsJson(args...))
}

func (p *Parser) DataAsJson(args ...interface{}) (string, error) {
	if len(args) != 0 {
		return xconversions.Stringify(args[0])
	} else {
		return xconversions.Stringify(p.ParsedData)
	}
}

func (p *Parser) MustDataAsJson(args ...interface{}) string {
	raw, err := p.DataAsJson(args...)
	PanicIfErr(err)
	return raw
}

func (p *Parser) DataAsYaml(args ...interface{}) (string, error) {
	raw, err := p.DataAsJson(args...)
	if err != nil {
		return raw, err
	}
	v, e := yaml.JSONToYAML([]byte(raw))
	return string(v), e
}

func (p *Parser) MustDataAsYaml(args ...interface{}) string {
	raw, err := p.DataAsYaml(args...)
	PanicIfErr(err)
	return raw
}

func (p *Parser) Scan() {
	for key, cfg := range p.Config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseAttrs("", key, cfgType)
		default:
			fmt.Println(xpretty.Redf(nonMapHint, key, cfg))
			continue
		}
	}
}

func (p *Parser) parseAttrs(parentKey, key string, config interface{}) {
	switch cfg := config.(type) {
	case map[string]interface{}:
		if p.isLeaf(cfg) {
			if _, b := cfg["_attr_refine"]; !b {
				return
			}
			attr := cfg[ATTR]
			refine, b := cfg[ATTR_REFINE]
			if !b {
				return
			}

			name := p.getRefineMethodName(key, refine, attr)
			name = strcase.ToCamel(name)
			p.AttrToBeRefined = append(p.AttrToBeRefined, name)
			p.AttrToBeRefined = funk.UniqString(p.AttrToBeRefined)

			return
		}

		for k, c := range cfg {
			p.parseAttrs(key, k, c)
		}
	default:
		return
	}
}

func (p *Parser) runCheck() {
}

func (p *Parser) DoParse() {
	p.runCheck()
	for key, cfg := range p.Config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseDom(key, cfgType, p.Root, p.ParsedData, layerForRank)
		default:
			fmt.Println(xpretty.Redf(nonMapHint, key, cfg))
			continue
		}
	}
}

func (p *Parser) popNestedKeys() {
	if len(p.nestedKeys) == 0 {
		return
	}
	p.nestedKeys = p.nestedKeys[:len(p.nestedKeys)-1]
}

func (p *Parser) checkNestedKeys(key string) bool {
	if !strings.Contains(key, "__") {
		p.nestedKeys = append(p.nestedKeys, key)
	}
	for _, tk := range p.testKeys {
		for _, nk := range p.nestedKeys {
			_tk := strings.ReplaceAll(tk, ".*", "")
			b := strings.Contains(tk, ".*") && strings.Contains(nk, _tk)
			if b {
				p.forceParsedKey = b
				// xpretty.DummyErrorLog(key, p.forceParsedKey)
				return true
			}
		}
	}
	p.forceParsedKey = false
	return false
}

func (p *Parser) requiredKey(key string) (b bool) {
	if strings.HasPrefix(key, "__") {
		return
	}

	if !p.devMode {
		return true
	}

	if p.forceParsedKey {
		return true
	}

	if funk.NotEmpty(p.testKeys) && !funk.Contains(p.testKeys, key) {
		return
	}

	return true
}

// parseDom
// only support two data type
// 1. str
// 2. map[string]interface{}
func (p *Parser) parseDom(key string, cfg interface{}, selection *goquery.Selection, data map[string]interface{}, layer int) {
	p.checkNestedKeys(key)
	defer p.popNestedKeys()

	b := p.requiredKey(key)
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
		p.handle_map(key, v, selection, data, layer)
	default:
		panic(xpretty.Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]interface{})", key, cfg))
	}
}

func (p *Parser) handleStr(key string, sel string, selection *goquery.Selection, data map[string]interface{}) {
	data[key] = selection.Find(sel).First().Text()
}

// handle_map
//  1. find all matched elems
//     1.1. found only 1 node
//     1.2. found more than 1 nodes
func (p *Parser) handle_map(
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
		p.parse_dom_nodes(cfg, dom, subData)

	case []*goquery.Selection:
		var allSubData []map[string]interface{}
		for _, gs := range dom {
			if layer == layerForRank {
				p.FocusedStub = gs
			}

			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parse_dom_nodes(cfg, gs, subData)
			// only calculate rank at first layer
			if layer == layerForRank {
				p.rank++
			}
		}
		data[key] = allSubData
	}
}

func (p *Parser) isLeaf(cfg map[string]interface{}) bool {
	for k := range cfg {
		// if key starts with _, means has child node
		if !strings.HasPrefix(k, "_") {
			return false
		}
	}
	return true
}

func (p *Parser) parse_dom_nodes(
	cfg map[string]interface{},
	selection *goquery.Selection,
	data map[string]interface{},
) {
	for k, sc := range cfg {
		if strings.HasPrefix(k, "_") {
			continue
		}
		p.parseDom(k, sc, selection, data, layerForOthers)
	}
}

func (p *Parser) getAllElems(key string, cfg map[string]interface{}, selection *goquery.Selection) (iface interface{}, isComplexSel bool) {
	sel := cfg[LOCATOR]
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

func (p *Parser) handleStub(raw interface{}, result *goquery.Selection) (interface{}, *goquery.Selection) {
	ar1 := strings.Split(raw.(string), ".")
	if ar1[0] == PREFIX_LOCATOR_STUB {
		raw = strings.Join(ar1[1:], ".")
		result = p.FocusedStub.(*goquery.Selection)
	}
	return raw, result
}

func (p *Parser) getElemsOneByOne(key string, selArr []string, cfg map[string]interface{}, selection *goquery.Selection) (iface []*goquery.Selection) {
	// selArr := strings.Split(sel, ",")
	var resArr []*goquery.Selection
	backup := selection

	for _, v := range selArr {
		v1, backup := p.handleStub(v, backup)
		v = v1.(string)
		elem, _ := p.getOneSelector(key, v, cfg, backup)
		resArr = append(resArr, elem.(*goquery.Selection))
	}
	return resArr
}

func (p *Parser) getOneSelector(key string, sel interface{}, cfg map[string]interface{}, selection *goquery.Selection) (iface interface{}, isComplexSel bool) {
	elems := selection.Find(sel.(string))
	index, existed := cfg[INDEX]
	isComplexSel = strings.Contains(sel.(string), ",")

	iface = p.handleNullIndex(sel, index, existed, elems)
	if iface != nil {
		return
	}

	switch val := index.(type) {
	case int:
		iface = elems.Eq(val)
	case []interface{}:
		var d []*goquery.Selection
		for _, v := range val {
			switch v := v.(type) {
			case int:
				d = append(d, elems.Eq(v))
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

func (p *Parser) handleNullIndex(sel, index interface{}, existed bool, elems *goquery.Selection) interface{} {
	// index has 4 types:
	//  1. without index
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	isComplexSel := strings.Contains(sel.(string), ",")

	// if index existed, just return nil
	if index != nil {
		return nil
	}

	// index not existed, just return the first selection
	if !existed {
		if isComplexSel {
			return p.getAllSelections(elems)
		}
		return elems.First()
	}

	// if index is yaml's null: '~' or null
	return p.getAllSelections(elems)
}

func (p *Parser) getAllSelections(elems *goquery.Selection) []*goquery.Selection {
	var d []*goquery.Selection
	for i := range elems.Nodes {
		d = append(d, elems.Eq(i))
	}
	return d
}

func (p *Parser) getNodesAttrs(
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
			data[key] = subData
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

func (p *Parser) getSelectionSliceAttr(key string, cfg map[string]interface{}, resultArr []*goquery.Selection) interface{} {
	var resArr []string
	for _, v := range resultArr {
		raw := p.getRawAttr(cfg, v)
		resArr = append(resArr, raw.(string))
	}
	v := p.refineAttr(key, strings.Join(resArr, ATTR_SEP), cfg, resultArr)
	return p.convertToType(v, cfg)
}

func (p *Parser) getSelectionMapAttr(key string, cfg map[string]interface{}, results map[string]*goquery.Selection) interface{} {
	dat := make(map[string]string)

	for k, v := range results {
		raw := p.getRawAttr(cfg, v)
		dat[k] = raw.(string)
	}
	str, _ := Stringify(dat)
	v := p.refineAttr(key, str, cfg, results)
	return p.convertToType(v, cfg)
}

func (p *Parser) getSelectionAttr(key string, cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	raw := p.getRawAttr(cfg, selection)
	raw = p.stripChars(key, raw, cfg)
	raw = p.refineAttr(key, raw, cfg, selection)

	return p.convertToType(raw, cfg)
}

func (p *Parser) convertToType(raw interface{}, cfg map[string]interface{}) interface{} {
	t, o := cfg[TYPE]
	if o {
		switch t {
		case ATTR_TYPE_B:
			return cast.ToBool(raw)
		case ATTR_TYPE_I:
			return cast.ToInt(math.Round(cast.ToFloat64(raw)))
		case ATTR_TYPE_F:
			return cast.ToFloat64(raw)
		}
	}

	return raw
}

func (p *Parser) getRawAttr(cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	attr := cfg[ATTR]

	// fmt.Printf("Got %T, %v\n", attr, attr)
	if attr == nil {
		v := selection.Text()
		return p.TrimSpace(v, cfg)
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

func (p *Parser) TrimSpace(txt string, cfg map[string]interface{}) string {
	if cfg[STRIP] == nil {
		return txt
	}
	return strings.TrimSpace(txt)
}

func (p *Parser) stripChars(key string, raw interface{}, cfg map[string]interface{}) interface{} {
	st := cfg[STRIP]
	if st == true {
		return strings.TrimSpace(raw.(string))
	}

	switch v := st.(type) {
	case string:
		return strings.ReplaceAll(raw.(string), v, "")
	}
	return raw
}

func (p *Parser) isMethodExisted(mtd_name string) (rv reflect.Value, b bool) {
	// automatically convert snake_case(which is written in yaml) to CamelCase or camelCase
	// first check camelCase (private method preffered)
	// if not found then check CamelCase
	mtdName := strcase.ToLowerCamel(mtd_name)
	MtdName := strcase.ToCamel(mtd_name)

	method := reflect.ValueOf(p).MethodByName(mtdName)
	if funk.IsEmpty(method) {
		method = reflect.ValueOf(p).MethodByName(MtdName)
		if funk.IsEmpty(method) {
			return
		}
	}
	return method, true
}

func (p *Parser) getRefinerFn(mtd_name string) (func(raw ...interface{}) interface{}, bool) {
	mtdName := strcase.ToLowerCamel(mtd_name)
	MtdName := strcase.ToCamel(mtd_name)

	injectFn, b := p.Refiners[mtdName]
	if !b {
		injectFn, b = p.Refiners[MtdName]
		if !b {
			prompt(p, mtdName, MtdName)
		}
	}

	return injectFn, b
}

func (p *Parser) refineAttr(key string, raw interface{}, cfg map[string]interface{}, selection interface{}) interface{} {
	attr := cfg[ATTR]
	refine := cfg[ATTR_REFINE]
	if refine == nil {
		return raw
	}
	mtd_name := p.getRefineMethodName(key, refine, attr)
	method, ok := p.isMethodExisted(mtd_name)

	if !ok {
		injectFn, b := p.getRefinerFn(mtd_name)
		if b {
			return injectFn(raw, p.Config, selection)
		}
	}

	param := []reflect.Value{reflect.ValueOf(raw)}
	res := method.Call(param)

	return res[0].Interface()
}

func (p *Parser) getRefineMethodName(key string, refine, attr interface{}) string {
	var mtdName string
	switch mtd := refine.(type) {
	case bool:
		switch attr.(type) {
		case string:
			mtdName = fmt.Sprintf("%v_%v_%v", PREFIX_REFINE, key, attr)
		default:
			mtdName = fmt.Sprintf("%v_%v", PREFIX_REFINE, key)
		}
	case string:
		mtdName = mtd
	default:
		panic(xpretty.Redf("refine method should be (bool or str), but (%s is %T: %v)\n", key, mtd, mtd))
	}
	// auto add refine to method startswith "_" like "_abc"
	// so "_abc" will be converted to "refine_abc"
	if strings.HasPrefix(mtdName, "_") && !strings.HasPrefix(mtdName, PREFIX_REFINE) {
		mtdName = PREFIX_REFINE + mtdName
	}
	return mtdName
}
