package xparse

import (
	"fmt"
	"testing"

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
	Structify(s.rawJson, &dat)
	// keys := GetKeys(dat["friends"], "friends")

	all := []string{}
	GetMapKeys(&all, dat)
	fmt.Println(all)
	// PrintAllKeys(dat)
}

func (s *UtilSuite) Test02_CutStr() {
	raw := "a,b,c,d,e"
	v, b := GetStrBySplit(raw, ",", 6)
	s.Equal("e", v)
	s.Equal(true, b)

	v, b = GetStrBySplit(raw, ",", -1)
	s.Equal("e", v)
	s.Equal(true, b)
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
			"site":     895,
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
