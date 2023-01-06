package xparse

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/ghodss/yaml"
	"github.com/gookit/config/v2"
	"github.com/iancoleman/strcase"
	"github.com/shomali11/util/xconversions"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
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
	Root   interface{}

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

	// verify keys, keys will be verified
	verifyKeys []string

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

func (p *Parser) Spawn(raw, ymlCfg []byte) {
	p.LoadConfig(ymlCfg)
	p.LoadRootSelection(raw)
}

func (p *Parser) ToggleDevMode(b bool) {
	p.devMode = b
}

func (p *Parser) Debug(key interface{}, raw ...interface{}) {
	if p.devMode {
		xpretty.GreenPrintf(fmt.Sprintf("[%d] %v: (%v)", p.rank, key, raw[0]))
	}
}

func (p *Parser) LoadConfig(ymlCfg []byte) {
	p.Config = Yaml2Config(ymlCfg)
	p.testKeys = p.Config.Strings("__raw.test_keys")
	p.verifyKeys = p.Config.Strings("__raw.verify_keys")
}

func (p *Parser) GetVerifyKeys() (arr []string) {
	return p.verifyKeys
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

// func (p *Parser) DoParse() {
// }

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

func (p *Parser) isLeaf(cfg map[string]interface{}) bool {
	for k := range cfg {
		// if key starts with _, means has child node
		if !strings.HasPrefix(k, "_") {
			return false
		}
	}
	return true
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
