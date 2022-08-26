package xparse

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/config/v2"
	"github.com/iancoleman/strcase"
	"github.com/shomali11/util/xconversions"
	"github.com/thoas/go-funk"
)

type Parser struct {
	Config *config.Config
	Root   *goquery.Selection

	// map to config
	ParsedData map[string]interface{}

	// test mode
	// isTestMode bool
	testKeys []string

	// selectedKeys []string
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

func (p *Parser) LoadRootSelection(raw []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	PanicIfErr(err)
	p.Root = doc.Selection
}

func (p *Parser) LoadConfig(ymlCfg []byte) {
	p.Config = Yaml2Config(ymlCfg)
	p.testKeys = p.Config.Strings("__raw.test_keys")
}

func (p *Parser) DataStr() (string, error) {
	return xconversions.Stringify(p.ParsedData)
}

func (p *Parser) DoParse() {
	for key, cfg := range p.Config.Data() {
		if !p.filterKey(key) {
			continue
		}

		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseDom(key, cfgType, p.Root, p.ParsedData)
		default:
			fmt.Println(Redf("[NON-MAP] {%v:%v}, please move into a map instead", key, cfg))
			continue
		}
	}
}

func (p *Parser) filterKey(key string) (b bool) {
	if strings.HasPrefix(key, "__") {
		return
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
		panic(Redf("unknown type of (%v:%v), only support (1:string or 2:map[string]interface{})", key, cfg))
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
		for _, gs := range dom {
			subData := make(map[string]interface{})
			allSubData = append(allSubData, subData)

			p.parse_dom_nodes(cfg, gs, subData)
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
				panic(Redf("all indexes should be int, but (%s is %T: %v)\n", key, val, val))
			}
		}
		return d
	default:
		panic(Redf("index should be int or []interface{}, but (%s is %T: %v)\n", key, val, val))
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
		panic(Redf("unknown type of dom %s:%v %v", key, cfg, dom))
	}
}

func (p *Parser) getSelectionAttr(key string, cfg map[string]interface{}, selection *goquery.Selection) interface{} {
	raw := p.getRawAttr(cfg, selection)
	raw = p.refineAttr(key, raw, cfg)
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
		panic(Redf("attr should be (string or []interface{}), but (%s is %T: %v)\n", attr, attrType, attrType))
	}
}

func (p *Parser) TrimSpace(txt string, cfg map[string]interface{}) string {
	if cfg[STRIPPED] == nil {
		return txt
	}
	return strings.TrimSpace(txt)
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

func (p *Parser) refineAttr(key string, raw interface{}, cfg map[string]interface{}) interface{} {
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
				fmt.Println(Redf(`Cannot find Refiner: (%s or %s)`, mtd_name, MtdName))
				fmt.Println(Greenf(`Please assign it to parser.Refiners by either one:
  - parser.Refiners["%s"] = your_func
  - parser.Refiners["%s"] = your_func`, mtd_name, MtdName))
				os.Exit(0)
			}
		}
		return injectFn(raw, p.Config)
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
		panic(Redf("refine method should be (bool or str), but (%s is %T: %v)\n", key, mtd, mtd))
	}

	return mtdName
}

func (p *Parser) EnrichUrl(raw interface{}) interface{} {
	domain := p.Config.String("__raw.site_url")
	uri := EnrichUrl(raw, domain)
	return uri
}
