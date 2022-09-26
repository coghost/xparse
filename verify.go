package xparse

import (
	"fmt"
	"strings"

	"github.com/coghost/xpretty"
	"github.com/tidwall/gjson"
)

// Verify if rawJson contains keys or not, and print(or not) missed keys according to level
//
// level: 0(no print)/1(print all)/2(print only missed keys)
//
//	@return failed_keys
func Verify(rawJson string, keys []string, level int) (failed map[string][]string) {
	failed = make(map[string][]string)
	result := gjson.Get(rawJson, "jobs")
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
			if level == 1 || !ok {
				arr = append(arr, _fn("\t| %s: %s\n", key, val))
			}
			rank = _rank
		}
		v1 := fn("%3s.", rank)
		if len(arr) != 0 {
			arr = Insert(arr, 0, v1)
		}
		if level != 0 {
			fmt.Print(strings.Join(arr, ""))
		}
		return true
	})
	return failed
}
