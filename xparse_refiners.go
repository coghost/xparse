package xparse

import (
	"strings"

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

func (p *Parser) RefineUrl(raw ...interface{}) interface{} {
	return p.EnrichUrl(raw...)
}

func (p *Parser) EnrichUrl(raw ...interface{}) interface{} {
	domain := p.config.String("__raw.site_url")
	uri := EnrichUrl(domain, raw[0])
	return uri
}

func (p *Parser) ToFloat(raw ...interface{}) interface{} {
	return ToFixed(cast.ToFloat64(raw[0]), 2)
}

func (p *Parser) BindRank(raw ...interface{}) interface{} {
	return p.rank
}

// TrimByFields removes all "\r\n\t" and keep one space at most
//
//   - 1. strings.TrimSpace
//   - 2. strings.Join(strings.Fields(s), " ")
func (p *Parser) TrimByFields(raw ...interface{}) interface{} {
	s := strings.TrimSpace(raw[0].(string))
	return strings.Join(strings.Fields(s), " ")
}

// Trim alias of TrimByFields
func (p *Parser) Trim(raw ...interface{}) interface{} {
	return p.TrimByFields(raw...)
}

// GetStrBySplitAtIndex
// split raw to slice and then return element at index
//
//   - if sep not in raw, returns raw
//   - if index < 0, reset index to len() + index
//   - if index > total length, returns the last one
//   - else returns element at index
func (p *Parser) GetStrBySplitAtIndex(raw interface{}, sep string, index int) string {
	str := cast.ToString(raw)
	if sep == "" || !strings.Contains(str, sep) {
		return str
	}

	arr := strings.Split(str, sep)
	if index > len(arr)-1 {
		index = len(arr) - 1
	} else if index < 0 {
		index = len(arr) + index
	}

	return arr[index]
}

func (p *Parser) GetStrBySplit(raw interface{}, sep string, offset int, withSep bool) string {
	s := cast.ToString(raw)
	v, b := GetStrBySplit(s, sep, offset)
	if !b {
		return s
	}
	if withSep {
		return sep + v
	}
	return v
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
	cfg := raw[1].(map[string]interface{})

	idx := 0
	idxTxt, b := cfg[AttrIndex]
	if b {
		idx = cast.ToInt(idxTxt)
	}

	sep := AttrJoinerSep
	if txt, b := cfg[AttrJoiner]; b {
		sep = txt.(string)
	}

	txt := p.GetStrBySplitAtIndex(raw[0], sep, idx)
	txt = strings.TrimSpace(txt)
	return txt
}
