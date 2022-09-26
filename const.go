package xparse

var ReservedWords = map[string]string{
	"attr":            "_attr",
	"attr_refine":     "_attr_refine",
	"children":        "_children",
	"index":           "_index",
	"joiner":          "_joiner",
	"striped":         "_striped",
	"locator":         "_locator",
	"locator_extract": "_locator_extract",
	"prefix_extract":  "_extract",
	"prefix_refine":   "_refine",
}

const (
	ATTR            = "_attr"
	ATTR_REFINE     = "_attr_refine"
	CHILDREN        = "_children"
	INDEX           = "_index"
	JOINER          = "_joiner"
	STRIP           = "_strip"
	LOCATOR         = "_locator"
	LOCATOR_EXTRACT = "_locator_extract"
	PREFIX_EXTRACT  = "_extract"
	PREFIX_REFINE   = "_refine"
)
