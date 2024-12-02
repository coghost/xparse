package xparse

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

// Mock
//
//   - raw[0]: the parsed text
//   - raw[1]: *config.Config
//   - raw[2]: *goquery.Selection / gjson.Result
func (p *Parser) Mock(raw ...interface{}) interface{} {
	return raw[0]
}

func (p *Parser) RefineURL(raw ...interface{}) interface{} {
	return p.EnrichUrl(raw...)
}

func (p *Parser) RefineUrl(raw ...interface{}) interface{} { //nolint
	return p.EnrichUrl(raw...)
}

func (p *Parser) EnrichURL(raw ...interface{}) interface{} {
	return p.EnrichUrl(raw...)
}

func (p *Parser) EnrichUrl(raw ...interface{}) interface{} { //nolint
	domain := p.config.String("__raw.site_url")
	uri := EnrichURL(domain, raw[0])

	return uri
}

func (p *Parser) ToFloat(raw ...interface{}) interface{} {
	return ToFixed(cast.ToFloat64(raw[0]), _precision)
}

func (p *Parser) BindRank(raw ...interface{}) interface{} {
	return p.rank
}

func (p *Parser) RefineRank(raw ...interface{}) interface{} {
	return p.rank
}

func (p *Parser) RefineUpper(raw ...interface{}) interface{} {
	return strings.ToUpper(cast.ToString(raw[0]))
}

func (p *Parser) RefineLower(raw ...interface{}) interface{} {
	return strings.ToLower(cast.ToString(raw[0]))
}

func (p *Parser) EnsureNotSlice(v interface{}, methodName string) {
	switch v.(type) {
	case []interface{}, []string, []int:
		xpretty.RedPrintln(fmt.Sprintf("\n⚠️  WARNING: %s received a slice, expected a single value.\n"+
			"💡 Hint: Did you pass 'raw' instead of 'raw[0]'?\n", methodName))
		os.Exit(0)
	}
}

func (p *Parser) ToString(v interface{}) string {
	p.EnsureNotSlice(v, "ToString")
	return cast.ToString(v)
}

func (p *Parser) ToInt64(v interface{}) int64 {
	p.EnsureNotSlice(v, "ToInt64")
	return cast.ToInt64(v)
}

func (p *Parser) SafeGetConfigKey(raw interface{}, key string) string {
	return SafeGetFromMap[string](raw, key)
}

// StringToBinary converts various boolean string representations to binary integers (1/0)
// Returns 1 for the following true values (case-insensitive):
//   - "true", "t"
//   - "1"
//   - "yes", "y"
//   - "on"
//
// Returns 0 for any other values, including empty strings and nil
func (p *Parser) StringToBinary(raw interface{}) int {
	return StringToBinary(p.ToString(raw))
}

// TrimByFields removes all "\r\n\t" and keep one space at most
//
//   - 1. strings.TrimSpace
//   - 2. strings.Join(strings.Fields(s), " ")
func (p *Parser) TrimByFields(raw ...interface{}) interface{} {
	rawStr, _ := raw[0].(string)
	s := strings.TrimSpace(rawStr)

	return strings.Join(strings.Fields(s), " ")
}

// Trim alias of TrimByFields
func (p *Parser) Trim(raw ...interface{}) interface{} {
	return p.TrimByFields(raw...)
}

// SplitAtIndex splits a string by separator and returns the element at specified index
//
// Parameters:
//   - raw: input value to be converted to string
//   - sep: separator to split the string
//   - index: desired index (negative index counts from end)
//
// Returns:
//   - The element at index after splitting
//   - If sep is empty or not found, returns original string
//   - If index out of bounds, returns first/last element
func (p *Parser) SplitAtIndex(raw interface{}, sep string, index int) string {
	return NewSplitter(raw, sep, index).String()
}

// Deprecated: Use SplitAtIndex instead. This function will be removed in a future version.
//
// GetStrBySplitAtIndex splits raw to slice and returns element at index
//
//   - if sep not in raw, returns raw
//   - if index < 0, reset index to len() + index
//   - if index > total length, returns the last one
//   - else returns element at index
func (p *Parser) GetStrBySplitAtIndex(raw interface{}, sep string, index int) string {
	return NewSplitter(raw, sep, index).String()
}

func (p *Parser) GetStrBySplit(raw interface{}, sep string, offset int, withSep bool) string {
	s := cast.ToString(raw)

	val, b := GetStrBySplit(s, sep, offset)
	if !b {
		return s
	}

	if withSep {
		return sep + val
	}

	return val
}

func (p *Parser) RefineDotNumber(raw ...interface{}) interface{} {
	v, err := CharToNum(raw[0].(string))
	if err != nil {
		return raw
	}

	return v
}

func (p *Parser) RefineCommaNumber(raw ...interface{}) interface{} {
	v, err := CharToNum(raw[0].(string), Chars(","))
	if err != nil {
		return raw
	}

	return v
}

// RefineAttrByIndex
// usage of attributes:
//
//   - _joiner: ","
//   - _attr_refine: _attr_by_index
//   - _attr_index: 0
func (p *Parser) RefineAttrByIndex(raw ...interface{}) interface{} {
	cfg, _ := raw[1].(map[string]interface{})
	idx := 0
	idxTxt, b := cfg[AttrIndex]

	if b {
		idx = cast.ToInt(idxTxt)
	}

	sep := AttrJoinerSep
	if txt, b := cfg[AttrJoiner]; b {
		sep, _ = txt.(string)
	}

	txt := p.SplitAtIndex(raw[0], sep, idx)
	txt = strings.TrimSpace(txt)

	return txt
}

func (p *Parser) RefineEncodedJSON(raw ...interface{}) interface{} {
	return p.RefineEncodedJson(raw...)
}

func (p *Parser) RefineEncodedJson(raw ...interface{}) interface{} { //nolint
	txt := p.SplitAtIndex(raw[0], "", -1)

	content, err := base64.StdEncoding.DecodeString(txt)
	if err != nil {
		log.Warn().Err(err).Msg("cannot decode string")
		return nil
	}

	var data map[string]interface{}

	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Warn().Err(err).Msg("cannot unmarshal str to json")
		return nil
	}

	return data
}

func (p *Parser) RefineJobsWithPreset() {
	p.refineJobs()
	p.refineJob()
}

func (p *Parser) refineJobs() {
	jobs, ok := p.GetParsedData("jobs").([]map[string]interface{})
	if !ok {
		return
	}

	for _, job := range jobs {
		p.AppendPresetData(job)
	}
}

func (p *Parser) refineJob() {
	job, ok := p.GetParsedData("job").(map[string]interface{})
	if !ok {
		return
	}

	p.AppendPresetData(job)
}
