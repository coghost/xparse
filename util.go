package xparse

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
)

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
