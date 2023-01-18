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
	_DEFAULT_STUB = "jobs"
)

// Verify if rawJson contains keys or not, and print(or not) missed keys according to level
//
//	printLevel:
//	 - 0(print none)
//	 - 1(print all)
//	 - 2(print only missed keys)
//
//	  @return failed_keys
func Verify_v1(rawJson string, keys []string, opts ...VerifyOptFunc) (failed map[string][]string) {
	opt := VerifyOpts{stubKey: "jobs", level: VerifyPrintAll}
	bindVerifyOpts(&opt, opts...)

	failed = make(map[string][]string)
	result := gjson.Get(rawJson, opt.stubKey)

	result.ForEach(func(_, value gjson.Result) bool {
		fn := xpretty.Greenf
		rank := ""
		arr := []string{}
		for _, key := range keys {
			ok := true
			_rank, val := value.Get("rank").Raw, value.Get(key).String()
			_fn := xpretty.Greenf
			if val == "" {
				_fn = xpretty.Redfu
				fn = _fn
				failed[key] = append(failed[key], _rank)
				ok = false
			}
			if opt.level == VerifyPrintAll || !ok {
				arr = append(arr, _fn("\t| %s: %s\n", key, val))
			}
			rank = _rank
		}
		v1 := fn("%3s.", rank)
		if len(arr) != 0 {
			arr = Insert(arr, 0, v1)
		}
		if opt.level != VerifyPrintNone {
			fmt.Print(strings.Join(arr, ""))
		}
		return true
	})

	if funk.IsEmpty(result) {
		xpretty.RedPrintf("Verify failed: rawJson has no key (%s), please check the rootKey passed in\n", opt.stubKey)
	}
	return failed
}

func Verify(rawJson string, keys []string, opts ...VerifyOptFunc) (failed map[string][]string, allResp map[string]map[int][]string) {
	sym := "┃"
	opt := VerifyOpts{level: VerifyPrintAll, stubKey: _DEFAULT_STUB, color: true}
	bindVerifyOpts(&opt, opts...)
	xpretty.ToggleColor(opt.color)

	failed = make(map[string][]string)
	root := gjson.Parse(rawJson)

	// get the stubKeys: keys directly in the root node
	stubKeys := make(map[string][]int)
	// dict of stub key and keys in the stub key
	grpKeys := make(map[string][]string)

	allResults := make(map[string]map[int][]string)
	allResp = make(map[string]map[int][]string)

	// first go through the keys to be verified to get all stubkeys with its items' ranks
	for _, key := range keys {
		k := strings.Split(key, ".")[0]

		// if there's no dot found, we'll add the stubkey to keys
		if !strings.Contains(key, ".") {
			k = opt.stubKey
			key = opt.stubKey + "." + key
		}

		if !strings.Contains(key, "#") {
			key = opt.stubKey + ".#." + strings.Split(key, ".")[1]
		}

		k = fmt.Sprintf("%s.#.rank", k)

		// same stub key will be handled only once
		if _, b := stubKeys[k]; !b {
			for _, v := range root.Get(k).Array() {
				stubKeys[k] = append(stubKeys[k], cast.ToInt(v.Int()))
			}
			allResults[k] = make(map[int][]string)
		}

		getKeyValAsArray(root, key, stubKeys[k], allResults[k])
		grpKeys[k] = append(grpKeys[k], key)
	}

	for stk, ranks := range stubKeys {
		wanted := strings.Split(stk, "#")[0]
		wantStub := wanted[:len(wanted)-1]

		arr := []string{}

		if len(stubKeys) > 1 {
			v := xpretty.Cyanf("%[2]s*\n%[1]s", strings.Repeat("-", 32), wanted)
			arr = append(arr, v)
		}

		results := allResults[stk]
		gks := grpKeys[stk]

		// assign to response
		allResp[wantStub] = results

		for _, rank := range ranks {
			res := results[rank]

			for i, v := range res {
				grpKey := gks[i]
				if !strings.HasPrefix(grpKey, wanted) {
					continue
				}

				abbrKey := strings.ReplaceAll(grpKey, fmt.Sprintf("%s#.", wanted), "")

				colorize := xpretty.Greenf
				if v == "" {
					colorize = xpretty.Redfu
					failed[wantStub] = append(failed[wantStub], fmt.Sprintf("%d:%s", rank, abbrKey))
				}

				output := colorize("%3d.\t%s %s: %s", rank, sym, abbrKey, v)
				if i != 0 {
					output = fmt.Sprintf("%3s \t", "") + colorize("%s %s: %s", sym, abbrKey, v)
				}
				arr = append(arr, output)
			}
		}
		fmt.Println(strings.Join(arr, "\n"))
	}

	return failed, allResp
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
