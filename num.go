package xparse

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(MaxUint >> 1)
	MinInt  = -MaxInt - 1
)

const (
	_roundSignBase = 0.5
	_precision     = 2
)

var ErrNoNumbers = errors.New("no number found")

type NumOpts struct {
	chars string
	dft   any
}

type NumOptFunc func(o *NumOpts)

// Chars the chars will be kept in CharToNum
// usually used as "decimal" like "."
func Chars(s string) NumOptFunc {
	return func(o *NumOpts) {
		o.chars = s
	}
}

func Dft(i any) NumOptFunc {
	return func(o *NumOpts) {
		o.dft = i
	}
}

func bindOpts(opt *NumOpts, opts ...NumOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

// CharToNum extract `number+Chars` from source str
//
//	the extracted value could be float value, so convert to float first, then return int by default
func CharToNum(rawStr string, opts ...NumOptFunc) (v any, e error) {
	opt := NumOpts{chars: ".", dft: 1}
	bindOpts(&opt, opts...)

	a := "[0-9" + opt.chars + "]+"
	re := regexp.MustCompile(a)
	c := re.FindAllString(rawStr, -1)
	joinStr := strings.Join(c, "")

	if strings.Contains(opt.chars, ",") {
		joinStr = strings.ReplaceAll(joinStr, ",", ".")
	}

	if joinStr == "" {
		return joinStr, ErrNoNumbers
	}

	switch opt.dft.(type) {
	case int:
		v, e := cast.ToFloat64E(joinStr)
		if e != nil {
			return nil, e
		}

		return cast.ToIntE(v)
	case int64:
		// v could be float value
		v, e := cast.ToFloat64E(joinStr)
		if e != nil {
			return nil, e
		}

		return cast.ToInt64E(v)
	case float32:
		return cast.ToFloat32E(joinStr)
	case float64:
		return cast.ToFloat64E(joinStr)
	default:
		return cast.ToStringE(joinStr)
	}
}

func MustCharToNum(s string, opts ...NumOptFunc) (v any) {
	v, e := CharToNum(s, opts...)
	if e != nil {
		PanicIfErr(e)
	}

	return v
}

func NumF64KMFromStr(str string, opts ...NumOptFunc) (i float64, b bool) {
	unit := 1.0

	if strings.Contains(strings.ToUpper(str), "K") {
		unit = 1000.0
	}

	if strings.Contains(strings.ToUpper(str), "M") {
		unit = 1000000.0
	}

	opt := NumOpts{chars: ".", dft: 1}
	bindOpts(&opt, opts...)

	if !strings.Contains(opt.chars, ".") {
		opt.chars += "."
	}

	v := MustCharToNum(str, Chars(opt.chars), Dft(opt.dft))
	if v == nil {
		return
	}

	return cast.ToFloat64(v) * unit, true
}

func MustF64KMFromStr(str string, opts ...NumOptFunc) float64 {
	if v, b := NumF64KMFromStr(str, opts...); !b {
		panic(fmt.Sprintf("no number found in %s", str))
	} else {
		return v
	}
}

func Round(num float64) int {
	return int(num + math.Copysign(_roundSignBase, num))
}

func ToFixed(num float64, precision ...int) float64 {
	p := FirstOrDefaultArgs(_precision, precision...)
	output := math.Pow(10, float64(p)) //nolint:mnd

	return float64(Round(num*output)) / output
}

func ToFixedStr(num float64, precision ...int) string {
	p := FirstOrDefaultArgs(_precision, precision...)
	output := ToFixed(num, p)

	return cast.ToString(output)
}
