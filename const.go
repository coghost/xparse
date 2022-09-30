package xparse

// JSON is built on two structures:
//   - A collection of name/value pairs.
//   - An ordered list of values.
//
// so we define ORDERED_LIST_SYMBOL to get all values as list/array
// `ORDERED_LIST_CONST` is used for json file like: `[{...}, {...}]` which has no name-value pair
const _ORDERED_LIST_CONST = "*/*"

var ReservedWords = map[string]string{
	"attr":            "_attr",
	"attr_refine":     "_attr_refine",
	"children":        "_children",
	"index":           "_index",
	"joiner":          "_joiner",
	"strip":           "_strip",
	"locator":         "_locator",
	"locator_extract": "_locator_extract",
	"prefix_extract":  "_extract",
	"prefix_refine":   "_refine",
	"type":            "_type",
}

const (
	ATTR            = "_attr"
	ATTR_REFINE     = "_attr_refine"
	INDEX           = "_index"
	STRIP           = "_strip"
	LOCATOR         = "_locator"
	LOCATOR_EXTRACT = "_locator_extract"
	PREFIX_EXTRACT  = "_extract"
	PREFIX_REFINE   = "_refine"
	TYPE            = "_type"
)

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
