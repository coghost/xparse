package xparse

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilSuite struct {
	suite.Suite
	rawJSON string
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupSuite() {
	s.rawJSON = `
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
	err := Structify(s.rawJSON, &dat)
	s.Nil(err)
	// keys := GetKeys(dat["friends"], "friends")

	var all []string
	GetMapKeys(&all, dat)
	fmt.Println(all)
	// PrintAllKeys(dat)
}

func (s *UtilSuite) Test02_CutStr() {
	raw := "a,b,c,d,e"
	v, b := GetStrBySplit(raw, ",", 6)
	s.Equal("e", v)
	s.True(b)

	v, b = GetStrBySplit(raw, ",", -1)
	s.Equal("e", v)
	s.True(b)
}

func (s *UtilSuite) Test03_load() {
	r0 := getBytes("html_yaml/0000.yaml")
	r1 := getBytes("html_yaml/0001.yaml")
	cf := Yaml2Config(r1, r0)

	want := map[string]interface{}{
		"__raw": map[string]interface{}{
			"country": "CH",
			"language": []interface{}{
				"en",
			},
			"site":     uint64(895),
			"site_url": "https://www.jobisjob.ch/",
			"test_keys": []interface{}{
				"jobs.*",
			},
			"verify_keys": []interface{}{
				"salary_range",
				"listing_date",
				"external_id",
			},
		},
	}

	s.Equal(want, cf.Data())
}

func TestParseNumberRanges(t *testing.T) {
	assert := assert.New(t)

	// Empty input
	assert.Equal([]int{}, ParseNumberRanges(""))

	// Single numbers
	assert.Equal([]int{0}, ParseNumberRanges("0"))
	assert.Equal([]int{-1}, ParseNumberRanges("-1"))

	// Multiple numbers
	assert.Equal([]int{0, 1, 2, 3}, ParseNumberRanges("0,1,2,3"))
	assert.Equal([]int{-2, -1, 0, 1}, ParseNumberRanges("-2,-1,0,1"))

	// Inclusive ranges
	assert.Equal([]int{0, 1, 2, 3}, ParseNumberRanges("0-3"))
	assert.Equal([]int{-2, -1, 0, 1}, ParseNumberRanges("-2-1"))

	// Exclusive ranges
	assert.Equal([]int{0, 1, 2}, ParseNumberRanges("0~3"))
	assert.Equal([]int{-2, -1, 0}, ParseNumberRanges("-2~1"))

	// Mixed formats with spaces
	assert.Equal([]int{0, 3, 4, 5, 6, 7, 13, 14}, ParseNumberRanges("0, 3-7, 13~15"))
	assert.Equal([]int{-3, -2, -1, 0, 1, 2}, ParseNumberRanges("-3~0, 0-2"))
}
