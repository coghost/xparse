package xparse

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
)

type Basic interface {
	bool | int | float32 | float64 | string
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Yaml2Config(raw []byte) (cf *config.Config) {
	cf = config.New("")
	cf.AddDriver(yamlv3.Driver)
	err := cf.LoadSources(config.Yaml, raw)
	PanicIfErr(err)
	return cf
}

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
	STRIPPED        = "_striped"
	LOCATOR         = "_locator"
	LOCATOR_EXTRACT = "_locator_extract"
	PREFIX_EXTRACT  = "_extract"
	PREFIX_REFINE   = "_refine"
)

var RedPrintf = color.New(color.FgRed, color.Bold).PrintfFunc()
var CyanPrintf = color.New(color.FgCyan, color.Bold).PrintfFunc()
var YellowPrintf = color.New(color.FgYellow, color.Bold).PrintfFunc()
var GreenPrintf = color.New(color.FgGreen, color.Bold).PrintfFunc()

var Red = color.New(color.FgRed, color.Bold).SprintFunc()
var Redf = color.New(color.FgRed, color.Bold).SprintfFunc()
var Green = color.New(color.FgGreen, color.Bold).SprintFunc()
var Greenf = color.New(color.FgGreen, color.Bold).SprintfFunc()
var White = color.New(color.FgHiWhite, color.Bold).SprintFunc()
var Whitef = color.New(color.FgHiWhite, color.Bold).SprintfFunc()
var Yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
var Yellowf = color.New(color.FgYellow, color.Bold).SprintfFunc()

// red foreground underline
var Redfu = color.New(color.FgRed, color.Bold, color.Underline).SprintfFunc()
var Redfc = color.New(color.FgRed, color.Bold, color.CrossedOut).SprintfFunc()

func SetColor(b bool) {
	color.NoColor = !b
}

func EnrichUrl(raw interface{}, domain string) interface{} {
	uri := raw.(string)
	pu, err := url.Parse(uri)
	PanicIfErr(err)

	if pu.Scheme != "" {
		return raw
	}

	if domain == "" {
		return raw
	}

	base, err := url.Parse(domain)
	PanicIfErr(err)

	uri = base.ResolveReference(pu).String()
	return uri
}

// GetProjectHome
//
// get the full path of current project, which is separated by projectName,
// please make sure you supplied an unique projectName
// and the fullname of project directory
func GetProjectHome(projectName string) string {
	pwd, _ := os.Getwd()
	arr := strings.Split(pwd, projectName)
	home := filepath.Join(arr[0], projectName)
	return home
}

// FirstOrDefaultArgs
//
// return the first args value, if args not empty
// else return default value
func FirstOrDefaultArgs[T Basic](dft T, args ...T) (val T) {
	val = dft
	if len(args) > 0 {
		val = args[0]
	}
	return val
}

// Insert
func Insert[T Basic](a []T, index int, value T) []T {
	// nil or empty slice or after last element
	if len(a) == index {
		return append(a, value)
	}
	// index < len(a)
	a = append(a[:index+1], a[index:]...)
	a[index] = value
	return a
}

// CutStrBySeparator: split raw str with separator and join from offset
//
//	example:
//	 raw = "a,b,c,d,e"
//	 v, b := CutStrBySeparator(raw, ",", 1)
//	 // v = "bcde", b = true
//
//	 v, b := CutStrBySeparator(raw, "_", 1)
//	 // v = "a,b,c,d,e", b = false
//
// @return string
// @return bool
func CutStrBySeparator(raw string, sep string, offset int) (string, bool) {
	if strings.Contains(raw, sep) {
		arr := strings.Split(raw, sep)
		i := offset
		if n := len(arr) - 1; n < offset {
			i = n
		}
		if offset < 0 {
			i = len(arr) + offset
		}
		return strings.Join(arr[i:], sep), true
	}
	return raw, false
}

func GetType(obj interface{}) string {
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

// GetMapKeys
func GetMapKeys(all *[]string, data interface{}, args ...string) {
	prefix := FirstOrDefaultArgs("", args...)

	var dat map[string]interface{}
	switch d := data.(type) {
	case []map[string]interface{}:
		dat = d[0]
	case map[string]interface{}:
		dat = d
	case []interface{}:
		switch d1 := d[0].(type) {
		case map[string]interface{}:
			dat = d1
		default:
			*all = append(*all, prefix)
			return
		}
	default:
		panic(fmt.Sprintf("not supported type found: (%T)", d))
	}

	for key, v := range dat {
		if prefix != "" {
			key = prefix + "." + key
		}
		switch t := v.(type) {
		case nil:
			// json.null
			*all = append(*all, key)
		case bool:
			// json.booleans
			*all = append(*all, key)
		case float64:
			// json.numbers
			*all = append(*all, key)
		case string:
			// json.strings
			*all = append(*all, key)
		case map[string]interface{}:
			// json.Object
			GetMapKeys(all, t, key)
		case []interface{}:
			// json.array
			// all = append(all, key)
			GetMapKeys(all, t, key)
		default:
			/** following are non json type **/
			*all = append(*all, key)
		}
	}
}
