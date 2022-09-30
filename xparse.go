package xparse

import (
	"bytes"
	"fmt"
	"os"
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

type Parser struct {
	Config *config.Config
	Root   *goquery.Selection
	JRoot  gjson.Result

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
}

func NewParser(rawHtml, ymlMap []byte) *Parser {
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

func (p *Parser) DoParse() {
	for key, cfg := range p.Config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseDom(key, cfgType, p.Root, p.ParsedData)
		default:
			fmt.Println(xpretty.Redf("[NON-MAP] {%v:%v}, please move into a map instead", key, cfg))
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
func (p *Parser) parseDom(key string, cfg interface{}, selection *goquery.Selection, data map[string]interface{}) {
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
		p.handle_map(key, v, selection, data)
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
) {
	if p.isLeaf(cfg) {
		p.getNodesAttrs(key, cfg, selection, data)
		return
	}

	elems := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Selection:
		subData := make(map[string]interface{})
		data[key] = subData
		p.parse_dom_nodes(cfg, dom, subData)

	case []*goquery.Selection:
		var allSubData []map[string]interface{}
		p.rank = 0
		for _, gs := range dom {
			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parse_dom_nodes(cfg, gs, subData)
			p.rank++
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
		p.parseDom(k, sc, selection, data)
	}
}

func (p *Parser) getAllElems(key string, cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	sel := cfg[LOCATOR]
	if sel == nil {
		return selection
	}

	elems := selection.Find(sel.(string))

	index, exist := cfg[INDEX]
	if index == nil {
		if !exist {
			return elems.First()
		}

		var d []*goquery.Selection
		for i := range elems.Nodes {
			d = append(d, elems.Eq(i))
		}
		return d
	}

	switch val := index.(type) {
	case int:
		return elems.Eq(val)
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
		return d
	default:
		panic(xpretty.Redf("index should be int or []interface{}, but (%s is %T: %v)\n", key, val, val))
	}
}

func (p *Parser) getNodesAttrs(
	key string,
	cfg map[string]interface{},
	selection *goquery.Selection,
	data map[string]interface{},
) {
	// fmt.Printf("Got %v, %T, %v\n", key, cfg, cfg)
	elems := p.getAllElems(key, cfg, selection)

	switch dom := elems.(type) {
	case *goquery.Document:
		panic("found Doc, Selection Required!")

	case *goquery.Selection:
		data[key] = p.getSelectionAttr(key, cfg, dom)

	case []*goquery.Selection:
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

func (p *Parser) getSelectionAttr(key string, cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	raw := p.getRawAttr(cfg, selection)
	raw = p.stripChars(key, raw, cfg)
	raw = p.refineAttr(key, raw, cfg, selection)
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
		return raw
	}

	switch v := st.(type) {
	case string:
		return strings.ReplaceAll(raw.(string), v, "")
	}
	return raw
}

// Invoke
//
//	return Invoke(*p, mtdName, p.Config)
func Invoke(any interface{}, name string, args ...interface{}) reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	v := reflect.ValueOf(any).MethodByName(name)
	return v
}

func (p *Parser) refineAttr(key string, raw interface{}, cfg map[string]interface{}, selection interface{}) interface{} {
	attr := cfg[ATTR]
	refine := cfg[ATTR_REFINE]
	if refine == nil {
		return raw
	}

	mtd_name := p.getRefineMethodName(key, refine, attr)
	// automatically convert snake_case(which is written in yaml) to CamelCase
	MtdName := strcase.ToCamel(mtd_name)
	method := reflect.ValueOf(p).MethodByName(MtdName)
	if funk.IsEmpty(method) {
		injectFn, b := p.Refiners[MtdName]
		if !b {
			injectFn, b = p.Refiners[mtd_name]
			if !b {
				fmt.Println(xpretty.Redf(`Cannot find Refiner: (%s or %s)`, mtd_name, MtdName))
				fmt.Println(xpretty.Greenf(`Please add following method:

func (p %[3]T) %[1]s(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	// TODO:
}

then assign it to parser.Refiners by either one:
  - parser.Refiners["%[2]s"] = %[1]s
  - parser.Refiners["%[1]s"] = %[1]s`, MtdName, mtd_name, p))
				os.Exit(0)
			}
		}
		return injectFn(raw, p.Config, selection)
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

	return mtdName
}

func (p *Parser) EnrichUrl(raw interface{}) interface{} {
	domain := p.Config.String("__raw.site_url")
	uri := EnrichUrl(raw, domain)
	return uri
}

func (p *Parser) ToFloat(raw interface{}) float64 {
	return ToFixed(cast.ToFloat64(raw), 2)
}

func (p *Parser) BindRank(raw interface{}) interface{} {
	return p.rank
}
