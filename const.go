package xparse

const (
	// Index is used to get all or one elem from results, and support two format `_index or _i`
	//  - nil value/not existed: get all
	//  - array value: [0, 1] get elems[0] and elems[1]
	//  - single value: 0 get elems[0]
	// index has 4 types:
	//  1. without index
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	Index = "_index"

	// Locator is the path/selector we used to find elem we want, and support two format `_locator or _l`
	//
	// Supported types:
	//  > string:
	//   _locator: string
	//
	//  > list:
	//   _locator:
	//     - div.001
	//     - div.002
	//     - div.003
	//
	//  > map:
	//   _locator:
	//     key1: div.001
	//     key2: div.002
	//     key3: div.003
	Locator = "_locator"

	// ExtractPrevElem
	// in most cases, we can use locator to get the elem we want,
	// but in some rare cases, there is no proper locator to use, so we have to use this to get prev elem
	ExtractPrevElem = "_extract_prev"

	ExtractParent = "_extract_parent"

	// Attr
	// by default we use the text of elem, but we can always specify the attr we want
	// this is useful when parsing info from HTML
	//
	//  - if _attr is '__html' will return the raw HTML
	Attr = "_attr"

	// AttrRefine, and support two format `_attr_refine or _ar`
	//  - bool(true): will automatically generate a method name
	//  - string(_name): will add prefix refine so "_xxx" will be renamed to "_refine_name"
	//  - string(refine_xxx/_refine_xxx): will be it
	//  - string(not started with _): will be it
	AttrRefine = "_attr_refine"

	// AttrJoiner attr joiner
	AttrJoiner = "_joiner"

	// AttrIndex
	//   - _joiner: ","
	//   - _attr_refine: _attr_by_index
	//   - _attr_index: 0
	AttrIndex = "_attr_index"

	AttrRegex = "_attr_regex"

	// AttrPython run python script directly(python environment is required), and the print will be used as the attr value.
	// i.e.:
	//
	//   import sys
	//   raw = sys.argv[1] # raw is globally registered, so we can it directly.
	//   # previous two line is automatically added to following to code.
	//   arr = raw.split("_")
	//   print(arr[1]) # this is required, we need the output value as refined attr value.
	//
	// > please check `examples/html_yaml/0900.yaml` for demo.
	AttrPython = "_attr_python"

	// AttrJS like python, but with js.
	// i.e.:
	//   arr = raw.split("_") // by default, raw is registered
	//   refined = arr[1] // refined is required, it the value we get from js.
	//
	//  - please check `examples/html_yaml/0901.yaml` for demo.
	//  - JavaScript library: underscore(https://underscorejs.org/) is supported by default"
	AttrJS = "_attr_js"

	// PostJoin is called when all attrs (as array) are parsed,
	// it transforms the attrs array to string by joining the joiner
	PostJoin = "_post_join"

	// Strip is a simple refiner
	//  - if `_strip: true` or _strip not existed, will do `strings.TrimSpace`
	//  - if `_strip: str` goes with a str, will do `strings.ReplaceAll(raw, str, "")`
	//  - if `_strip: ["(", ")"]`, will replace one by one
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
	LocatorAbbr    = "_l"
	IndexAbbr      = "_i"
	AttrRefineAbbr = "_ar"
	TypeAbbr       = "_t"
)

const (
	// JSONArrayRootLocator is a hard-coded symbol,
	//
	// since JSON is built on two structures:
	//   - A collection of name/value pairs.
	//   - An ordered list of values.
	//
	// and there is no root locator for ordered list,
	// so we use this symbol when json file is with ordered list of values like: `[{...}, {...}]`
	JSONArrayRootLocator = "*/*"
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

	// time

	// AttrTypeT quick mode
	AttrTypeT = "t"
	// AttrTypeT1 search mode, a bit slower
	AttrTypeT1 = "t1"
)
