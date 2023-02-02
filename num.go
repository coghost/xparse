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

var ErrorNoNumbers = errors.New("no number found")

type NumOpts struct {
	chars string
	dft   interface{}
}

type NumOptFunc func(o *NumOpts)

func Chars(s string) NumOptFunc {
	return func(o *NumOpts) {
		o.chars = s
	}
}

func Dft(i interface{}) NumOptFunc {
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
func CharToNum(s string, opts ...NumOptFunc) (v interface{}, e error) {
	opt := NumOpts{chars: ".", dft: 1}
	bindOpts(&opt, opts...)

	a := "[0-9" + opt.chars + "]+"
	re := regexp.MustCompile(a)
	c := re.FindAllString(s, -1)
	r := strings.Join(c, "")

	if r == "" {
		return r, ErrorNoNumbers
	}

	switch opt.dft.(type) {
	case int:
		v, e := cast.ToFloat64E(r)
		if e != nil {
			return nil, e
		}
		return cast.ToIntE(v)
	case int64:
		// v could be float value
		v, e := cast.ToFloat64E(r)
		if e != nil {
			return nil, e
		}
		return cast.ToInt64E(v)
	case float32:
		return cast.ToFloat32E(r)
	case float64:
		return cast.ToFloat64E(r)
	default:
		return cast.ToStringE(r)
	}
}

func MustCharToNum(s string, opts ...NumOptFunc) (v interface{}) {
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
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision ...int) float64 {
	p := FirstOrDefaultArgs(2, precision...)
	output := math.Pow(10, float64(p))
	return float64(Round(num*output)) / output
}

func ToFixedStr(num float64, precision ...int) string {
	p := FirstOrDefaultArgs(2, precision...)
	output := ToFixed(num, p)
	return cast.ToString(output)
}
