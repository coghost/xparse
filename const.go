package xparse

// JSON is built on two structures:
//   - A collection of name/value pairs.
//   - An ordered list of values.
//
// so we define ORDERED_LIST_SYMBOL to get all values as list/array
// `ORDERED_LIST_CONST` is used for json file like: `[{...}, {...}]` which has no name-value pair
const _ORDERED_LIST_CONST = "*/*"

// not used
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

// not used
var ReservedWords = map[string]string{
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

const (
	INDEX   = "_index"
	LOCATOR = "_locator"

	ATTR = "_attr"
	// _attr_refine:
	//  - bool(true): will automatically generate a method name
	//  - string(_name): will add prefix refine so "_xxx" will be renamed to "_refine_name"
	//  - string(refine_xxx/_refine_xxx): will be it
	//  - string(not started with _): will be it
	ATTR_REFINE   = "_attr_refine"
	PREFIX_REFINE = "_refine"

	// a simple refiner, will do strings.Strip
	STRIP = "_strip"

	// for now support
	//  - b:bool
	//  - i:int
	//  - f:float
	TYPE = "_type"

	// base locator is used for multiple locators not in same stub
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
	//  - ___.salarySnippet will be parsed to jobs[index].salarySnippet
	//  - attributes at the same time will be calculated from taxo
	PREFIX_LOCATOR_STUB = "___"
)

const (
	ATTR_TYPE_B = "b"
	ATTR_TYPE_F = "f"
	ATTR_TYPE_I = "i"
)
