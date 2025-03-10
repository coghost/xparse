package xparse

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/coghost/xdtm"
	"github.com/coghost/xparse/plugin/js"
	"github.com/coghost/xparse/plugin/py3"
	"github.com/coghost/xpretty"
	"github.com/ghodss/yaml"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
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

const (
	_rangeIndexLen = 2
)

const (
	skippedKeySymbol = "__"
)

type Parser struct {
	sourceData []byte
	sourceYaml [][]byte

	config *config.Config

	Root any

	// this is a map's stub, check const.PrefixLocatorStub for more info
	FocusedStub any

	RawData string

	// devMode
	devMode bool

	rank       int
	rankOffset int
	// use the real order of page or not (which is same as _index:)
	rankAsIndex bool

	// PID parser uniqid
	PID string

	presetData map[string]any

	// map to config
	ParsedData map[string]any

	// testKeys, only keys in testKeys will be parsed, and .rank is parsed by default
	testKeys []string
	// nestedKeysForCheckingTestKeys
	//  > example(job.yaml):
	//  -----
	//   job:
	//     datePosted:
	//     address:
	// 	     region:
	//  -----
	//
	//  so the nestedKeys can be
	//  - []
	//  - ["job"]->["job", "datePosted"]->["job"]
	//  - ["job"]->["job", "address"]->["job", "address", "region"]->["job", "address"]->["job"]
	//  - []
	nestedKeysForCheckingTestKeys []string

	// verify keys, keys will be verified
	verifyKeys []string

	// Refiners is a map of:
	//
	//  > string: func
	//  - string is name we defined
	//  - func has three params:
	//    + first params is string, which is the raw str get from html (usually by get_text/get_attr)
	//    + second params is the *config.Config (which is rarely used)
	//    + third params is *goquery.Selection
	Refiners map[string]func(raw ...any) any

	AttrToBeRefined []string
}

func NewParser(raw []byte, ymlMap ...[]byte) *Parser {
	return &Parser{
		sourceData: raw,
		sourceYaml: ymlMap,
		config:     &config.Config{},
		ParsedData: make(map[string]any),
		Refiners:   make(map[string]func(args ...any) any),

		rankAsIndex: false,
	}
}

func (p *Parser) ToggleDevMode(b bool) {
	p.devMode = b
}

func (p *Parser) Debug(key any, raw ...any) {
	if p.devMode {
		xpretty.GreenPrintf(fmt.Sprintf("[%d] %v: (%v)", p.rank, key, raw[0]))
	}
}

func (p *Parser) LoadConfig(ymlCfg ...[]byte) {
	p.config = Yaml2Config(ymlCfg...)
	p.testKeys = p.config.Strings("__raw.test_keys")
	p.verifyKeys = p.config.Strings("__raw.verify_keys")
}

func (p *Parser) VerifyKeys() (arr []string) {
	return p.verifyKeys
}

func (p *Parser) BindPresetData(dat map[string]any) {
	if dat == nil {
		return
	}

	p.presetData = dat
}

func (p *Parser) GetPresetData() map[string]any {
	return p.presetData
}

func (p *Parser) ExtraInfo() map[string]any {
	return nil
}

func (p *Parser) AppendPresetData(data map[string]any) {
	pd := p.GetPresetData()
	for k, v := range pd {
		_, b := data[k]
		if !b {
			data[k] = v
		}
	}

	// try add parser unique id to data
	_, found := data["site"]
	if !found && p.PID != "" {
		data["site"] = p.PID
	}

	// try add p.PID to external_id
	v, found := data["external_id"]
	if !found {
		return
	}

	s, _ := v.(string)
	pre := p.PID + "_"

	if found && p.PID != "" && s != "" && !strings.HasPrefix(s, pre) {
		data["external_id"] = p.PID + "_" + s
	}
}

func (p *Parser) MustMandatoryFields(got, wanted []string) {
	if len(got) == 0 || len(wanted) == 0 {
		return
	}

	a, _ := funk.DifferenceString(got, wanted)
	if len(a) != 0 {
		log.Fatal().Msg(xpretty.Yellowf("unwanted keys %q found, please check if typo or missing", a))
	}
}

// GetRawInfo
//
// get raw info's value in config file
//   - if args is empty, will return __raw's value
//   - else return the first value in args
func (p *Parser) RawInfo(args ...string) map[string]any {
	key := FirstOrDefaultArgs("__raw", args...)
	raw := p.config.Data()[key]
	rawInfo, _ := raw.(map[string]any)

	return rawInfo
}

func (p *Parser) GetParsedData(args ...string) any {
	if len(args) == 0 {
		return p.ParsedData
	}

	return p.ParsedData[args[0]]
}

func (p *Parser) PrettifyData(args ...any) error {
	return xpretty.PrettyMap(p.ParsedData)
}

func (p *Parser) PrettifyJSONData(args ...any) error {
	return xpretty.PrettyJSON(p.MustDataAsJSON(args...))
}

// DataAsJson returns a string of args[0] or p.ParsedData and error
func (p *Parser) DataAsJSON(args ...any) (string, error) {
	if len(args) != 0 {
		key, _ := args[0].(string)

		v, ok := p.ParsedData[key]
		if !ok {
			return "", fmt.Errorf("cannot get data for key: %s", args[0]) //nolint
		}

		return Stringify(v)
	}

	return Stringify(p.ParsedData)
}

func (p *Parser) MustDataAsJSON(args ...any) string {
	raw, err := p.DataAsJSON(args...)
	PanicIfErr(err)

	return raw
}

func (p *Parser) DataAsStruct(structObj any, args ...any) error {
	raw, err := p.DataAsJSON(args...)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(raw), structObj)
}

func (p *Parser) DataAsYaml(args ...any) (string, error) {
	raw, err := p.DataAsJSON(args...)
	if err != nil {
		return raw, err
	}

	v, e := yaml.JSONToYAML([]byte(raw))

	return string(v), e
}

func (p *Parser) MustDataAsYaml(args ...any) string {
	raw, err := p.DataAsYaml(args...)
	PanicIfErr(err)

	return raw
}

func (p *Parser) Scan() {
	for key, cfg := range p.config.Data() {
		switch cfgType := cfg.(type) {
		case map[string]any:
			p.parseAttrs("", key, cfgType)
		default:
			xpretty.RedPrintf(_nonMapHint, key, cfg)
			continue
		}
	}
}

func (p *Parser) parseAttrs(_ string, key string, config any) {
	switch cfg := config.(type) {
	case map[string]any:
		if p.isLeaf(cfg) {
			refine, b := cfgAttrRefine(cfg)
			if !b {
				return
			}

			attr := cfg[Attr]
			name := p.convertAttrRefineToSnakeCaseName(key, refine, attr)
			name = GetCamelRefinerName(name)
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
	if len(p.nestedKeysForCheckingTestKeys) == 0 {
		return
	}

	p.nestedKeysForCheckingTestKeys = p.nestedKeysForCheckingTestKeys[:len(p.nestedKeysForCheckingTestKeys)-1]
}

func (p *Parser) appendNestedKeys(key string) {
	if strings.Contains(key, skippedKeySymbol) {
		return
	}

	p.nestedKeysForCheckingTestKeys = append(p.nestedKeysForCheckingTestKeys, key)
}

func (p *Parser) checkTestKeys(_ string) bool {
	// len == 0, just return true
	if len(p.nestedKeysForCheckingTestKeys) < 1 {
		return true
	}

	nestedKey := strings.Join(p.nestedKeysForCheckingTestKeys, ".")
	// len == 1, check if starts with nk or not
	if len(p.nestedKeysForCheckingTestKeys) == 1 {
		for _, tk := range p.testKeys {
			if strings.HasPrefix(tk, nestedKey) {
				return true
			}
		}

		return false
	}

	// len p.nestedKeys > 1
	for _, testKey := range p.testKeys {
		b := strings.Contains(testKey, ".*")
		if b {
			got := p.checkKey(testKey, nestedKey)
			if got {
				return true
			}
		} else {
			if strings.HasSuffix(nestedKey, ".rank") {
				return true
			}

			if nestedKey == testKey {
				return true
			}
		}
	}

	return false
}

func (*Parser) checkKey(testKey string, nestedKey string) bool {
	_tk := strings.ReplaceAll(testKey, ".*", "")
	return strings.HasPrefix(nestedKey, _tk)
}

func (p *Parser) isRequiredKey(key string) (b bool) {
	if strings.HasPrefix(key, "__") {
		return
	}

	if !p.devMode {
		return true
	}

	if p.checkTestKeys(key) {
		return true
	}

	if funk.NotEmpty(p.testKeys) && !funk.Contains(p.testKeys, key) {
		return
	}

	return true
}

func (p *Parser) isLeaf(cfg map[string]any) bool {
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

func (p *Parser) setRank(cfg map[string]any) {
	idxGot := mustCfgIndex(cfg)

	if idxGot == nil {
		// when _index is nil, means use every item, so rank is same with offset
		p.rank = p.rankOffset
		p.rankOffset++

		return
	}

	switch idx := idxGot.(type) {
	case int:
		p.rank = idx
	case []any:
		if p.rankAsIndex {
			p.rank = cast.ToInt(idx[p.rankOffset])
		} else {
			p.rank = p.rankOffset
		}

		p.rankOffset++
	case string:
		p.rank = p.rankOffset
		p.rankOffset++

		return
	default:
		panic(fmt.Sprintf("unsupported index for setRank %v", idx))
	}
}

func (p *Parser) convertToType(raw any, cfg map[string]any) any {
	t, o := cfgType(cfg)
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

func (p *Parser) formatDate(raw any, bySearch bool) any {
	rawStr, _ := raw.(string)
	if v := xdtm.GetDateTimeStr(rawStr, xdtm.WithBySearch(bySearch)); v != "" {
		return v
	}

	return raw
}

func (p *Parser) TrimSpace(txt string, cfg map[string]any) string {
	st := cfg[Strip]
	if st == false {
		return txt
	}

	return strings.TrimSpace(txt)
}

func (p *Parser) stripChars(key string, raw any, cfg map[string]any) any {
	switch v := raw.(type) {
	case string:
		return p.stripStrings(key, v, cfg)
	default:
		return v
	}
}

func (p *Parser) stripStrings(_ string, raw any, cfg map[string]any) any {
	rawStr, _ := raw.(string)

	st := cfg[Strip]
	if st == nil || st == true {
		return strings.TrimSpace(rawStr)
	}

	switch stripType := st.(type) {
	case string:
		return strings.ReplaceAll(rawStr, stripType, "")
	case []any:
		val := rawStr

		for _, sub := range stripType {
			subStr, _ := sub.(string)
			val = strings.ReplaceAll(val, subStr, "")
		}

		raw = val
	}

	return raw
}

func (p *Parser) isMethodExisted(snakeCaseName string) (rv reflect.Value, b bool) {
	// automatically convert snake_case(which is written in yaml) to CamelCase or camelCase
	// first check camelCase (private method preferred)
	// if not found then check CamelCase
	mtdName := GetLowerCamelRefinerName(snakeCaseName)
	MtdName := GetCamelRefinerName(snakeCaseName)

	method := reflect.ValueOf(p).MethodByName(mtdName)
	if funk.IsEmpty(method) {
		method = reflect.ValueOf(p).MethodByName(MtdName)
		if funk.IsEmpty(method) {
			return
		}
	}

	return method, true
}

func (p *Parser) getRefinerFn(snakeCaseName string) (func(raw ...any) any, bool) {
	mtdName := GetLowerCamelRefinerName(snakeCaseName)
	MtdName := GetCamelRefinerName(snakeCaseName)

	injectFn, found := p.Refiners[mtdName]
	if !found {
		injectFn, found = p.Refiners[MtdName]
		if !found {
			if fn, b := p.loadPreDefined(MtdName); b {
				return fn, b
			}

			typeName := getTypeNameFromInterface(p)
			_ = handleRefinerPrompt(typeName, MtdName, nil, true)

			os.Exit(0)
		}
	}

	return injectFn, found
}

func (p *Parser) loadPreDefined(mtdName string) (func(raw ...any) any, bool) {
	switch mtdName {
	case "BindRank":
		return p.BindRank, true
	case "RefineRank":
		return p.RefineRank, true
	case "EnrichURL":
		return p.EnrichURL, true
	case "EnrichUrl":
		return p.EnrichUrl, true
	case "RefineEncodedJson":
		return p.RefineEncodedJSON, true
	case "RefineEncodedJSON":
		return p.RefineEncodedJSON, true
	default:
		return nil, false
	}
}

func (p *Parser) refineByRe(raw any, cfg map[string]any) any {
	rgx, ok := cfg[AttrRegex]
	if !ok {
		return raw
	}

	regex, err := regexp.Compile(rgx.(string))
	if err != nil {
		log.Error().Err(err).Interface("regex", rgx).Msg("cannot compile regex")
	}

	rawStr, _ := raw.(string)

	return regex.FindString(rawStr)
}

func (p *Parser) refineByPython(raw any, cfg map[string]any) any {
	code, ok := cfg[AttrPython]
	if !ok {
		return raw
	}

	codeStr, _ := code.(string)
	rawStr, _ := raw.(string)

	resp, err := py3.Eval(codeStr, rawStr)
	if err != nil {
		log.Error().Err(err).Msg("cannot run python code")
	}

	return resp.RefinedString
}

func (p *Parser) refineByJS(raw any, cfg map[string]any) any {
	code, ok := cfg[AttrJS]
	if !ok {
		return raw
	}

	codeStr, _ := code.(string)
	rawStr, _ := raw.(string)

	resp, err := js.Eval(codeStr, rawStr)
	if err != nil {
		log.Error().Err(err).Msg("cannot run js code")
	}

	return resp.RefinedString
}

func (p *Parser) advancedPostRefineAttr(raw any, cfg map[string]any) any {
	raw = p.refineByRe(raw, cfg)
	raw = p.refineByPython(raw, cfg)
	raw = p.refineByJS(raw, cfg)

	return raw
}

func (p *Parser) refineAttr(key string, raw any, cfg map[string]any, selection any) any {
	attr := cfg[Attr]

	refine := mustCfgAttrRefine(cfg)
	if refine == nil {
		return raw
	}

	snakeCaseName := p.convertAttrRefineToSnakeCaseName(key, refine, attr)

	// refiners from parser-defined is prior than pre-defined
	injectFn, b := p.getRefinerFn(snakeCaseName)
	if b {
		// 1. with full config (*config.Config)
		// Not-Supported: add a new key like `__return_config`
		//  - return injectFn(raw, p.config, selection)
		// 2. only current config (map)
		switch val := raw.(type) {
		case string:
			return injectFn(val, cfg, selection)
		case []string:
			var resp []any
			for _, v := range val {
				resp = append(resp, injectFn(v, cfg, selection))
			}

			return resp
		default:
			panic(fmt.Sprintf("not supported type %s: %T, %v", key, val, val))
		}
	}

	// pre-defined methods
	method, ok := p.isMethodExisted(snakeCaseName)
	if ok {
		switch val := raw.(type) {
		case string:
			param := []reflect.Value{reflect.ValueOf(val), reflect.ValueOf(cfg), reflect.ValueOf(selection)}
			res := method.Call(param)

			return res[0].Interface()
		case []string:
			var resp []any

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

// convertAttrRefineToSnakeCaseName
//
//	@param key: the key of the stub
//	@param refiner: is the _attr_refine defined in yaml
//	@param attr: is the _attr key to be refined
func (p *Parser) convertAttrRefineToSnakeCaseName(key string, refiner, attr any) string {
	var snakeCaseName string

	switch mtd := refiner.(type) {
	case bool:
		switch attr.(type) {
		case string:
			snakeCaseName = fmt.Sprintf("%v_%v_%v", _prefixRefine, key, attr)
		default:
			snakeCaseName = fmt.Sprintf("%v_%v", _prefixRefine, key)
		}
	case string:
		if mtd == RefineWithKeyName {
			snakeCaseName = fmt.Sprintf("%v_%v", _prefixRefine, key)
		} else {
			snakeCaseName = mtd
		}
	default:
		panic(xpretty.Redf("refine method should be (bool or str), but (%s is %T: %v)", key, mtd, mtd))
	}
	// auto add refine to method starts with "_" like "_abc"
	// so "_abc" will be converted to "refine_abc"
	if strings.HasPrefix(snakeCaseName, "_") && !strings.HasPrefix(snakeCaseName, _prefixRefine) {
		snakeCaseName = _prefixRefine + snakeCaseName
	}

	return snakeCaseName
}

func (p *Parser) getJoinerOrDefault(cfg map[string]any, dft string) string {
	joiner := dft
	if j := cfg[AttrJoiner]; j != nil {
		joiner, _ = j.(string)
	}

	return joiner
}
