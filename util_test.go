package xparse_test

import (
	"fmt"
	"testing"

	"github.com/coghost/xparse"

	"github.com/shomali11/util/xconversions"
	"github.com/stretchr/testify/suite"
)

type UtilSuite struct {
	suite.Suite
	rawJson string
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupSuite() {
	s.rawJson = `
	{
		"name": {"first": "Tom", "last": "Anderson"},
		"age":37,
		"children": ["Sara","Alex","Jack"],
		"not_existed": null,
		"fav.movie": "Deer Hunter",
		"friends": [
		  {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
		  {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
		  {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
		]
	 }
`
}

func (s *UtilSuite) TearDownSuite() {
}

func (s *UtilSuite) Test01_GetKeys() {
	dat := make(map[string]interface{})
	xconversions.Structify(s.rawJson, &dat)
	// keys := xparse.GetKeys(dat["friends"], "friends")

	all := []string{}
	xparse.GetMapKeys(&all, dat)
	fmt.Println(all)
	// xparse.PrintAllKeys(dat)
}

func (s *UtilSuite) Test02_CutStr() {
	raw := "a,b,c,d,e"
	v, b := xparse.GetStrBySplit(raw, ",", 6)
	s.Equal("e", v)
	s.Equal(true, b)

	v, b = xparse.GetStrBySplit(raw, ",", -1)
	s.Equal("e", v)
	s.Equal(true, b)
}

func (s *UtilSuite) Test03() {
}
