# xparse

parse raw html or raw json file to structured data with yaml config file

## demo

```yaml
__raw:
  site_url: https://xkcd.com/
  test_keys:
    - bottom.comic_links.*
    - middle.ctitle
    - middle.transcript

bottom:
  _locator: div#bottom
  comic_links:
    _locator: div#comicLinks>a
    _index: ~
    text:
    href:
      _attr: href
      _attr_refine: enrich_url

middle:
  _locator: div#middleContainer
  ctitle: div#ctitle
  transcript: div#transcript
```

## constants

all reserved keys when we used to write yaml config file to map the HTML/JSON

```go
package xparse

const (
	// Index is used to get all or one elem from results
	//  - nil value/not existed: get all
	//  - array value: [0, 1] get elems[0] and elems[1]
	//  - single value: 0 get elems[0]
	// index has 4 types:
	//  1. without index
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	Index = "_index"

	// Locator is the path/selector we used to find elem we want
	Locator = "_locator"

	// ExtractPrevElem
	// in most cases, we can use locator to get the elem we want,
	// but in some rare cases, there is no proper locator to use, so we have to use this to get prev elem
	ExtractPrevElem = "_extract_prev"

	// Attr
	// by default we use the text of elem, but we can always specify the attr we want
	// this is useful when parsing info from HTML
	Attr = "_attr"

	// AttrRefine
	//  - bool(true): will automatically generate a method name
	//  - string(_name): will add prefix refine so "_xxx" will be renamed to "_refine_name"
	//  - string(refine_xxx/_refine_xxx): will be it
	//  - string(not started with _): will be it
	AttrRefine = "_attr_refine"

	// AttrJoiner attr joiner
	AttrJoiner = "_joiner"

	// PostJoin is called when all attrs (as array) are parsed,
	// it transforms the attrs array to string by joining the joiner
	PostJoin = "_post_join"

	// Strip is a simple refiner
	//  - if `_strip: true` or _strip not existed, will do `strings.TrimSpace`
	//  - if `_strip: str` goes with a str, will do `strings.ReplaceAll(raw, str, "")`
	//
	//  WARN: this is called by default, you should use `_strip: false` to disable it
	Strip = "_strip"

	// Type is a simple type converter, returns the type specified,
	//  without `_type: b/i/f`, returns as string
	//
	//  - b:bool
	//  - i:int
	//  - f:float
	Type = "_type"
)

const (
	// JsonArrayRootLocator is a hard-coded symbol,
	//
	// since JSON is built on two structures:
	//   - A collection of name/value pairs.
	//   - An ordered list of values.
	//
	// and there is no root locator for ordered list,
	// so we use this symbol when json file is with ordered list of values like: `[{...}, {...}]`
	JsonArrayRootLocator = "*/*"
)

const (
	// _prefixRefine defines the word we use as the prefix of method of attr refiner
	_prefixRefine = "_refine"

	// AttrJoinerSep is a separator used to join an array to string
	AttrJoinerSep = "|||"
)

const (
	// AttrJoinElemsText is a hard-coded value used after `reserved key: _attr`
	//  used only when parsing HTML
	//  it joins all elems inner text to string
	//  WARN: this is rarely used, you can always find another to do same thing
	AttrJoinElemsText = "__join_text"

	// RefineWithKeyName is a hard-coded symbol, on behalf of the key name as a refiner method
	//  we can use `_attr_refine: __key` instead of `_attr_refine: _refine_a_changeable_name`
	/**
	root:
		# ...
		a_changeable_name:
			_locator: div.xxx
			_attr: title
			_attr_refine: __key
	**/
	RefineWithKeyName = "__key"

	// PrefixLocatorStub
	// is used for multiple locators those not in same stub,
	// so we recalculated from its base locator(as is the map root)
	/**
		jobs:
			_locator: jobs
			_index:
			# ...
			taxo:
				_locator: taxonomyAttributes
				_index: 0
				attr:
				_locator:
					- attributes
					- ___.salarySnippet
	**/
	//  - so `___.salarySnippet` will be calculated as `jobs[jobs.index].salarySnippet`
	//  - and `attributes` still is calculated from `taxo`
	PrefixLocatorStub = "___"
)

const (
	AttrTypeB = "b"
	AttrTypeF = "f"
	AttrTypeI = "i"
)

```

## Reserved keys/TODO

```go
package xparse

// file: const.go

// TODO:
const (
	ABBR_ATTR            = "_a"
	ABBR_ATTR_REFINE     = "_ar"
	ABBR_INDEX           = "_i"
	ABBR_STRIP           = "_s"
	ABBR_LOCATOR         = "_l"
	ABBR_LOCATOR_EXTRACT = "_le"
	ABBR_PREFIX_EXTRACT  = "_e"
	ABBR_PREFIX_REFINE   = "_r"
	ABBR_TYPE            = "_t"
)

// TODO:
var reservedWords = map[string]string{
	"attr":          "_attr",
	"attr_refine":   "_attr_refine",
	"prefix_refine": "_refine",
	"index":         "_index",
	"strip":         "_strip",
	"locator":       "_locator",
	"type":          "_type",

	"joiner":          "_joiner",
	"locator_extract": "_locator_extract",
	"prefix_extract":  "_extract",
}
```
