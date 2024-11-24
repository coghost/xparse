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

// Core extraction configuration keys
const (
	// Index specifies which elements to extract from results
	// Formats: "_index" or "_i"
	// Values:
	//   - nil/not existed: get all elements
	//   - array: [0,1] gets elements[0] and elements[1]
	//   - single: 0 gets elements[0]
	// Index types:
	//  1. without index
	//  2. index: ~ (index is null)
	//  3. index: 0
	//  4. index: [0, 1, ...]
	//  5. index: 0,4 => 0,1,2,3
	Index = "_index"

	// Locator specifies the path/selector to find desired elements
	// Formats: "_locator" or "_l"
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

	// Element navigation keys
	// ExtractPrevElem is used when no proper locator exists
	// in most cases, we can use locator to get the elem we want,
	// but in some rare cases, there is no proper locator to use, so we have to use this to get prev elem
	ExtractPrevElem = "_extract_prev"
	ExtractParent   = "_extract_parent"
)

// Attribute related configuration keys
const (
	// Attr specifies which attribute to extract
	// Default is element text
	// Special value "__html" returns raw HTML
	Attr = "_attr"

	// AttrRefine specifies how to refine the extracted attribute
	// Formats: "_attr_refine" or "_ar"
	// Values:
	//   - bool(true): auto-generate method name
	//   - string(_name): adds prefix "refine" so "_xxx" becomes "_refine_name"
	//   - string(refine_xxx/_refine_xxx): used as-is
	//   - string(not started with _): used as-is
	AttrRefine = "_attr_refine"

	// AttrJoiner specifies the joiner for attributes
	AttrJoiner = "_joiner"

	// AttrIndex configuration:
	//   - _joiner: ","
	//   - _attr_refine: _attr_by_index
	//   - _attr_index: 0
	AttrIndex = "_attr_index"

	AttrRegex = "_attr_regex"

	// AttrPython runs Python script directly (requires Python environment)
	// Example:
	//   import sys
	//   raw = sys.argv[1] # raw is globally registered
	//   arr = raw.split("_")
	//   print(arr[1]) # required: output value as refined attr value
	AttrPython = "_attr_python"

	// AttrJS runs JavaScript code
	// Example:
	//   arr = raw.split("_") // raw is registered by default
	//   refined = arr[1] // refined is required value
	// Note: Underscore.js (https://underscorejs.org/) is supported by default
	AttrJS = "_attr_js"
)

// Post-processing configuration keys
const (
	// PostJoin joins parsed attributes array into string using joiner
	PostJoin = "_post_join"

	// Strip controls string trimming
	// Values:
	//   - if `_strip: true` or not existed: does strings.TrimSpace
	//   - if `_strip: str`: does strings.ReplaceAll(raw, str, "")
	//   - if `_strip: ["(", ")"]`: replaces one by one
	// Note: Called by default, use `_strip: false` to disable
	Strip = "_strip"

	// Type converts output to specified type
	// Without `_type: b/i/f`, returns as string
	// Values:
	//   - b: bool
	//   - i: int
	//   - f: float
	Type = "_type"
)

// Abbreviated keys
const (
	LocatorAbbr    = "_l"
	IndexAbbr      = "_i"
	AttrRefineAbbr = "_ar"
	TypeAbbr       = "_t"
)

// Special locators and internal constants
const (
	// JSONArrayRootLocator is used for JSON arrays without root object
	// Used when JSON file has ordered list of values like: `[{...}, {...}]`
	JSONArrayRootLocator = "*/*"

	// PrefixLocatorStub for multiple locators not in same stub
	// Recalculates from base locator (map root)
	// Example:
	//   jobs:
	//     _locator: jobs
	//     _index:
	//     taxo:
	//       _locator: taxonomyAttributes
	//       _index: 0
	//       attr:
	//       _locator:
	//         - attributes
	//         - ___.salarySnippet
	PrefixLocatorStub = "___"

	// _prefixRefine defines the word we use as the prefix of method of attr refiner
	_prefixRefine = "_refine"
	// AttrJoinerSep is a separator used to join an array to string
	AttrJoinerSep = "|||"
)

// Special attribute values
const (
	// AttrJoinElemsText joins all elements inner text to string
	// Used only when parsing HTML
	// Warning: Rarely used, consider alternatives
	AttrJoinElemsText = "__join_text"

	// AttrRawHTML returns the raw html of locator
	AttrRawHTML = "__html"

	// RefineWithKeyName uses key name as refiner method
	// Example:
	//   root:
	//     a_changeable_name:
	//       _locator: div.xxx
	//       _attr: title
	//       _attr_refine: __key
	RefineWithKeyName = "__key"
)

// Type constants
const (
	AttrTypeB = "b" // Boolean
	AttrTypeF = "f" // Float
	AttrTypeI = "i" // Integer

	// Time types
	AttrTypeT  = "t"  // Quick mode
	AttrTypeT1 = "t1" // Search mode
)
```
