package xparse

import (
	"fmt"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

const (
	_defaultStubKey = "jobs"
)

func Verify(rawJSON string, keys []string, opts ...VerifyOptFunc) (failed map[string][]string, allResp map[string]map[int][]string) {
	sym := "┃"
	opt := VerifyOpts{level: VerifyPrintAll, stubKey: _defaultStubKey, color: true}
	bindVerifyOpts(&opt, opts...)
	xpretty.SetNoColor(!opt.color)

	failed = make(map[string][]string)
	root := gjson.Parse(rawJSON)

	// get the stubKeys: keys directly in the root node
	stubKeys := make(map[string][]int)
	// dict of stub key and keys in the stub key
	grpKeys := make(map[string][]string)

	allResults := make(map[string]map[int][]string)
	allResp = make(map[string]map[int][]string)

	// first go through the keys to be verified to get all stub keys with its items' ranks
	for _, key := range keys {
		baseKey := strings.Split(key, ".")[0]

		// if there's no dot found, we'll add the stub key to keys
		if !strings.Contains(key, ".") {
			baseKey = opt.stubKey
			key = opt.stubKey + "." + key
		}

		if !strings.Contains(key, "#") {
			key = opt.stubKey + ".#." + strings.Split(key, ".")[1]
		}

		baseKey = fmt.Sprintf("%s.#.rank", baseKey)

		// same stub key will be handled only once
		if _, b := stubKeys[baseKey]; !b {
			for _, v := range root.Get(baseKey).Array() {
				stubKeys[baseKey] = append(stubKeys[baseKey], cast.ToInt(v.Int()))
			}

			allResults[baseKey] = make(map[int][]string)
		}

		getKeyValAsArray(root, key, stubKeys[baseKey], allResults[baseKey])
		grpKeys[baseKey] = append(grpKeys[baseKey], key)
	}

	for stk, ranks := range stubKeys {
		wanted := strings.Split(stk, "#")[0]
		wantStub := wanted[:len(wanted)-1]

		var arr []string

		if len(stubKeys) > 1 {
			v := xpretty.Cyanf("%[2]s*\n%[1]s", strings.Repeat("-", 32), wanted) //nolint:mnd
			arr = append(arr, v)
		}

		results := allResults[stk]
		gks := grpKeys[stk]

		// assign to response
		allResp[wantStub] = results

		for _, rank := range ranks {
			res := results[rank]

			for i, resVal := range res {
				grpKey := gks[i]
				if !strings.HasPrefix(grpKey, wanted) {
					continue
				}

				subKey := strings.ReplaceAll(grpKey, fmt.Sprintf("%s#.", wanted), "")

				output := formatOutput(resVal, failed, wantStub, rank, subKey, opt, sym, i)
				if output != "" {
					arr = append(arr, output)
				}
			}
		}

		_, _ = xpretty.Println(strings.Join(arr, "\n"))
	}

	return failed, allResp
}

func formatOutput(resVal string, failed map[string][]string,
	wantStub string, rank int,
	abbrKey string, opt VerifyOpts,
	sym string, index int,
) string {
	colorize := xpretty.Greenf
	if resVal == "" {
		colorize = xpretty.Redfu

		failed[wantStub] = append(failed[wantStub], fmt.Sprintf("%d:%s", rank, abbrKey))
	}

	if opt.level == VerifyPrintNone {
		return ""
	}

	if opt.level == VerifyPrintMissed && resVal != "" {
		return ""
	}

	output := colorize("%3d.\t%s %s: %s", rank, sym, abbrKey, resVal)
	if index != 0 {
		output = fmt.Sprintf("%3s \t", "") + colorize("%s %s: %s", sym, abbrKey, resVal)
	}

	return output
}

func getKeyValAsArray(root gjson.Result, key string, indexes []int, results map[int][]string) {
	arr := root.Get(key).Array()
	if funk.IsEmpty(arr) {
		for i := range indexes {
			results[i] = append(results[i], "")
		}

		return
	}

	for i, v := range arr {
		idx := indexes[i]
		results[idx] = append(results[idx], v.String())
	}
}
