package xparse

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

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

func Yaml2Config(raw ...[]byte) (cf *config.Config) {
	cf = config.New("")
	cf.AddDriver(yamlv3.Driver)
	err := cf.LoadSources(config.Yaml, raw[0], raw[1:]...)
	PanicIfErr(err)
	return cf
}

func EnrichUrl(domain string, raw interface{}) interface{} {
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
// please make sure you supplied a unique projectName
// and the full name of project directory
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

// GetStrBySplit split raw str with separator and join from offset
//
//	example:
//	 raw = "a,b,c,d,e"
//	 v, b := GetStrBySplit(raw, ",", 1)
//	 // v = "bcde", b = true
//
//	 v, b := GetStrBySplit(raw, "_", 1)
//	 // v = "a,b,c,d,e", b = false
//
// @return string
// @return bool
func GetStrBySplit(raw string, sep string, offset int) (string, bool) {
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

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	v := reflect.ValueOf(any).MethodByName(name)
	return v.Call(inputs)
}

func GetMethod(any interface{}, key string) reflect.Value {
	return reflect.ValueOf(any).MethodByName(key)
}

func GetField(any interface{}, key string) reflect.Value {
	return reflect.ValueOf(any).Elem().FieldByName(key)
}

// Stringify returns a string representation
func Stringify(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Structify returns the original representation
func Structify(data string, value interface{}) error {
	return json.Unmarshal([]byte(data), value)
}

// JoinUrlWithRef joins a URI reference from a base URL
func JoinUrlWithRef(baseUrl, refUrl string) (*url.URL, error) {
	u, err := url.Parse(refUrl)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(u), nil
}
