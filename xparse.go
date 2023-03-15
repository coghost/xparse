package xparse

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/coghost/xdtm"
	"github.com/coghost/xpretty"
	"github.com/ghodss/yaml"
	"github.com/gookit/config/v2"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

const (
	_nonMapHint = "[NON-MAP] {%v:%v}, please move into a map instead"
)

const (
	_layerForRank = iota + 1
	_layerForOthers
)

type Parser struct {
	sourceData []byte
	sourceYaml [][]byte

	config *config.Config

	Root interface{}

	// this is a map's stub, check PrefixLocatorStub for more info
	FocusedStub interface{}

	RawData string

	// devMode
	devMode bool

	rank       int
	rankOffset int
	// use the real order of page or not (which is same as _index:)
	rankAsIndex bool

	// PID parser uniqid
	PID string

	// map to config
	ParsedData map[string]interface{}

	// testKeys, only keys in testKeys will be parsed, and .rank is parsed by default
	testKeys []string

	// verify keys, keys will be verified
	verifyKeys []string

	nestedKeys []string

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

func NewParser(raw []byte, ymlMap ...[]byte) *Parser {
	return &Parser{
		sourceData: raw,
		sourceYaml: ymlMap,
		config:     &config.Config{},
		ParsedData: make(map[string]interface{}),
		Refiners:   make(map[string]func(args ...interface{}) interface{}),

		rankAsIndex: false,
	}
}

func (p *Parser) ToggleDevMode(b bool) {
	p.devMode = b
}

func (p *Parser) Debug(key interface{}, raw ...interface{}) {
	if p.devMode {
		xpretty.GreenPrintf(fmt.Sprintf("[%d] %v: (%v)", p.rank, key, raw[0]))
	}
}

func (p *Parser) LoadConfig(ymlCfg ...[]byte) {
	p.config = Yaml2Config(ymlCfg...)
	p.testKeys = p.config.Strings("__raw.test_keys")
	p.verifyKeys = p.config.Strings("__raw.verify_keys")
}

func (p *Parser) GetVerifyKeys() (arr []string) {
	return p.verifyKeys
}

func (p *Parser) BindPresetData(dat map[string]interface{}) {
	for k, v := range dat {
		_, b := p.ParsedData[k]
		if b {
			continue
		}

		if funk.IsEmpty(v) {
			continue
		}

		p.ParsedData[k] = v
	}
}

// GetRawInfo
//
// get raw info's value in config file
//   - if args is empty, will return __raw's value
//   - else return the first value in args
func (p *Parser) GetRawInfo(args ...string) map[string]interface{} {
	key := FirstOrDefaultArgs("__raw", args...)
	raw := p.config.Data()[key]
	return raw.(map[string]interface{})
}

func (p *Parser) GetParsedData() map[string]interface{} {
	return p.ParsedData
}

func (p *Parser) PrettifyData(args ...interface{}) {
	xpretty.PrettyMap(p.ParsedData)
}

func (p *Parser) PrettifyJsonData(args ...interface{}) {
	xpretty.PrettyJson(p.MustDataAsJson(args...))
}

// DataAsJson returns a string of args[0] or p.ParsedData and error
func (p *Parser) DataAsJson(args ...interface{}) (string, error) {
	if len(args) != 0 {
		return Stringify(args[0])
	} else {
		return Stringify(p.ParsedData)
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
	for key, cfg := range p.config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]interface{}:
			p.parseAttrs("", key, cfgType)
		default:
			fmt.Println(xpretty.Redf(_nonMapHint, key, cfg))
			continue
		}
	}
}

func (p *Parser) parseAttrs(parentKey, key string, config interface{}) {
	switch cfg := config.(type) {
	case map[string]interface{}:
		if p.isLeaf(cfg) {
			refine, b := cfg[AttrRefine]
			if !b {
				return
			}

			attr := cfg[Attr]
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

func (p *Parser) runCheck() {}

func (p *Parser) DoParse() {}

func (p *Parser) PostDoParse() {}

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
	return true
}

func (p *Parser) check(key string) bool {
	// len == 0, just return true
	if len(p.nestedKeys) < 1 {
		return true
	}

	nk := strings.Join(p.nestedKeys, ".")
	// len == 1, check if starts with nk or not
	if len(p.nestedKeys) == 1 {
		for _, tk := range p.testKeys {
			if strings.HasPrefix(tk, nk) {
				return true
			}
		}
		return false
	}

	// len p.nestedKeys > 1
	for _, tk := range p.testKeys {
		b := strings.Contains(tk, ".*")
		if b {
			_tk := strings.ReplaceAll(tk, ".*", "")
			if strings.HasPrefix(nk, _tk) {
				return true
			}
		} else {
			if strings.HasSuffix(nk, ".rank") {
				return true
			}
			if nk == tk {
				return true
			}
		}
	}
	return false
}

func (p *Parser) isRequiredKey(key string) (b bool) {
	if strings.HasPrefix(key, "__") {
		return
	}

	if !p.devMode {
		return true
	}

	if p.check(key) {
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

func (p *Parser) ToggleRankType(b bool) {
	p.rankAsIndex = b
}

func (p *Parser) setRank(cfg map[string]interface{}) {
	if cfg[Index] == nil {
		// when _index is nil, means use every item, so rank is same with offset
		p.rank = p.rankOffset
		p.rankOffset++
		return
	}

	switch idx := cfg[Index].(type) {
	case int:
		p.rank = idx
	case []interface{}:
		if p.rankAsIndex {
			p.rank = cast.ToInt(idx[p.rankOffset])
		} else {
			p.rank = p.rankOffset
		}
		p.rankOffset++
	default:
		panic(fmt.Sprintf("unsupported index for setRank %v", idx))
	}
}

func (p *Parser) convertToType(raw interface{}, cfg map[string]interface{}) interface{} {
	t, o := cfg[Type]
	if o {
		switch t {
		case AttrTypeB:
			return cast.ToBool(raw)
		case AttrTypeI:
			return cast.ToInt(math.Round(cast.ToFloat64(raw)))
		case AttrTypeF:
			return cast.ToFloat64(raw)
		case AttrTypeT:
			return p.formatDate(raw, false)
		case AttrTypeT1:
			return p.formatDate(raw, true)
		}
	}

	return raw
}

func (p *Parser) formatDate(raw interface{}, bySearch bool) interface{} {
	if v := xdtm.GetDateTimeStr(raw.(string), xdtm.WithBySearch(bySearch)); v != "" {
		return v
	}
	return raw
}

func (p *Parser) TrimSpace(txt string, cfg map[string]interface{}) string {
	st := cfg[Strip]
	if st == false {
		return txt
	}
	return strings.TrimSpace(txt)
}

func (p *Parser) stripChars(key string, raw interface{}, cfg map[string]interface{}) interface{} {
	switch v := raw.(type) {
	case string:
		return p.stripStrings(key, v, cfg)
	default:
		return v
	}
}

func (p *Parser) stripStrings(key string, raw interface{}, cfg map[string]interface{}) interface{} {
	st := cfg[Strip]
	if st == nil || st == true {
		return strings.TrimSpace(raw.(string))
	}

	switch v := st.(type) {
	case string:
		return strings.ReplaceAll(raw.(string), v, "")
	case []interface{}:
		val := raw.(string)
		for _, sub := range v {
			val = strings.ReplaceAll(val, sub.(string), "")
		}
		raw = val
	}
	return raw
}

func (p *Parser) isMethodExisted(mtd_name string) (rv reflect.Value, b bool) {
	// automatically convert snake_case(which is written in yaml) to CamelCase or camelCase
	// first check camelCase (private method preferred)
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
	attr := cfg[Attr]
	refine := cfg[AttrRefine]
	if refine == nil {
		return raw
	}
	mtd_name := p.getRefineMethodName(key, refine, attr)

	// refiners from parser-defined is prior than pre-defined
	injectFn, b := p.getRefinerFn(mtd_name)
	if b {
		// 1. with full config (*config.Config)
		// TODO: add a new key like `__return_config`
		//  - return injectFn(raw, p.config, selection)
		// 2. only current config (map)
		switch val := raw.(type) {
		case string:
			return injectFn(val, cfg, selection)
		case []string:
			var resp []interface{}
			for _, v := range val {
				resp = append(resp, injectFn(v, cfg, selection))
			}
			return resp
		default:
			panic(fmt.Sprintf("not supported type %s: %T, %v", key, val, val))
		}
	}

	// pre-defined methods
	method, ok := p.isMethodExisted(mtd_name)
	if ok {
		switch val := raw.(type) {
		case string:
			param := []reflect.Value{reflect.ValueOf(val), reflect.ValueOf(cfg), reflect.ValueOf(selection)}
			res := method.Call(param)
			return res[0].Interface()
		case []string:
			var resp []interface{}
			for _, v := range val {
				param := []reflect.Value{reflect.ValueOf(v), reflect.ValueOf(cfg), reflect.ValueOf(selection)}
				res := method.Call(param)
				resp = append(resp, res[0].Interface())
			}
			return resp
		default:
			panic(fmt.Sprintf("not supported type %s: %T, %v", key, val, val))
		}
	}

	return nil
}

func (p *Parser) getRefineMethodName(key string, refine, attr interface{}) string {
	var mtdName string
	switch mtd := refine.(type) {
	case bool:
		switch attr.(type) {
		case string:
			mtdName = fmt.Sprintf("%v_%v_%v", _prefixRefine, key, attr)
		default:
			mtdName = fmt.Sprintf("%v_%v", _prefixRefine, key)
		}
	case string:
		if mtd == RefineWithKeyName {
			mtdName = fmt.Sprintf("%v_%v", _prefixRefine, key)
		} else {
			mtdName = mtd
		}
	default:
		panic(xpretty.Redf("refine method should be (bool or str), but (%s is %T: %v)\n", key, mtd, mtd))
	}
	// auto add refine to method starts with "_" like "_abc"
	// so "_abc" will be converted to "refine_abc"
	if strings.HasPrefix(mtdName, "_") && !strings.HasPrefix(mtdName, _prefixRefine) {
		mtdName = _prefixRefine + mtdName
	}
	return mtdName
}

func (p *Parser) getJoinerOr(cfg map[string]interface{}, or string) string {
	joiner := or
	if j := cfg[AttrJoiner]; j != nil {
		joiner = j.(string)
	}
	return joiner
}
