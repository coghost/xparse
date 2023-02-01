package xparse

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coghost/xpretty"
	"github.com/gookit/goutil/fsutil"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type JsonParserSuite struct {
	suite.Suite
	parser *JsonParser

	rawJson []byte
	rawYaml []byte

	examples_home string
}

func TestJsonParser(t *testing.T) {
	suite.Run(t, new(JsonParserSuite))
}

func (s *JsonParserSuite) SetupSuite() {
	xpretty.Initialize(xpretty.WithColor(true), xpretty.WithDummyLog(true))
	home := GetProjectHome("xparse")
	s.examples_home = home
	s.rawJson = fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed.json"))
	s.rawYaml = fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_json.yaml"))
	s.parser = NewJsonParser(s.rawJson, s.rawYaml)
}

func getBytes(path string) []byte {
	home := GetProjectHome("xparse")
	return fsutil.MustReadFile(filepath.Join(home, fmt.Sprintf("/examples/%s", path)))
}

func (s *JsonParserSuite) TearDownSuite() {
}

func (s *JsonParserSuite) Test_00_gjson_adotb() {
	res := gjson.Parse(string(s.rawJson))
	job := res.Array()[0]
	want := "Front-End Engineer – 2023 (US)"
	got := job.Get("jobs.0.title").String()
	s.Equal(want, got)
}

func (s *JsonParserSuite) Test_01_init() {
	p := s.parser
	p.ToggleDevMode(true)
	p.DoParse()
}

func (s *JsonParserSuite) Test_02_array_as_root() {
	home := GetProjectHome("xparse")
	rawJson := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_array_as_root.json"))
	rawYaml := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_array_as_root.yaml"))

	p := NewJsonParser(rawJson, rawYaml)
	p.DoParse()

	p1 := NewJsonParser(s.rawJson, s.rawYaml)
	p1.DoParse()
	s.Equal(p.ParsedData, p1.ParsedData)
}

func (s *JsonParserSuite) Test_03_simple() {
	rawYaml := `
jobs:
  _locator: jobs
  _index:
  title: title
  rank:
    _attr_refine: bind_rank
`

	p := NewJsonParser(s.rawJson, []byte(rawYaml))
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":  0,
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank":  1,
				"title": "Remote Python Prep Instructor (Part-Time)",
			},
			{
				"rank":  2,
				"title": "Machine Learning Apprentice",
			},
			{
				"rank":  3,
				"title": "Actuarial Data Entry Temp",
			},
			{
				"rank":  4,
				"title": "Python Engineer",
			},
			{
				"rank":  5,
				"title": "Data Analyst (Remote)",
			},
			{
				"rank":  6,
				"title": "Python Developer",
			},
			{
				"rank":  7,
				"title": "Linguistic QA",
			},
			{
				"rank":  8,
				"title": "Associate Data Analyst (Python)",
			},
			{
				"rank":  9,
				"title": "Backend Python Developer [REMOTE]",
			},
			{
				"rank":  10,
				"title": "Applied Biomechanics Researcher",
			},
			{
				"rank":  11,
				"title": "Python Developer",
			},
			{
				"rank":  12,
				"title": "Software Engineer - Undergrad New College Grad - Multiple Locations - 2023",
			},
			{
				"rank":  13,
				"title": "Python Developer",
			},
			{
				"rank":  14,
				"title": "Data Mentor (Part-time)",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *JsonParserSuite) Test_04_index() {
	rawYaml := `
jobs:
  _locator: jobs
  _index:
    - 0
    - 11
  title: title
  rank:
    _attr_refine: bind_rank
`

	p := NewJsonParser(s.rawJson, []byte(rawYaml))
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":  0,
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank":  1,
				"title": "Python Developer",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *JsonParserSuite) Test_05_type() {
	rawYaml := getBytes("json_yaml/05.yaml")
	p := NewJsonParser(s.rawJson, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank": 0,
				"reviews": map[string]interface{}{
					"count":   88390,
					"count_f": 88390.000000,
					"link":    "/cmp/Amazon.com/reviews",
					"rated":   false,
				},
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank": 1,
				"reviews": map[string]interface{}{
					"count":   0,
					"count_f": 0.000000,
					"link":    "",
					"rated":   false,
				},
				"title": "Python Developer",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

func refineMax(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	if v == "" {
		return v
	}
	return "max: $" + v
}

func refineSalaryMin(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	if v == "" {
		return v
	}
	return "min: $" + v
}

func (s *JsonParserSuite) Test_0601_attrRefineManually() {
	rawYaml := getBytes("json_yaml/0601.yaml")
	p := NewJsonParser(s.rawJson, rawYaml)
	p.Refiners["RefineMax"] = refineMax
	p.Refiners["RefineSalaryMin"] = refineSalaryMin
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank": 0,
				"salary": map[string]interface{}{
					"max": "",
					"min": "",
				},
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank": 1,
				"salary": map[string]interface{}{
					"max": "max: $94426.57",
					"min": "min: $74573.43",
				},
				"title": "Python Developer",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

type Parser2 struct {
	*JsonParser
}

func newParser2(rawData, ymlMap []byte) *Parser2 {
	return &Parser2{
		NewJsonParser(rawData, ymlMap),
	}
}

func (p *Parser2) RefineMax(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func (p *Parser2) RefineSalaryMin(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func (s *JsonParserSuite) Test_0602_attrRefineAutoFind() {
	rawYaml := getBytes("json_yaml/0602.yaml")
	p := newParser2(s.rawJson, rawYaml)
	UpdateRefiners(p)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank": 0,
				"salary": map[string]interface{}{
					"max": "",
					"min": "",
				},
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank": 1,
				"salary": map[string]interface{}{
					"max": "94426.57",
					"min": "74573.43",
				},
				"title": "Python Developer",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

func (s *JsonParserSuite) Test_0701_locator_gjson_multipaths() {
	rawYaml := getBytes("json_yaml/0701.yaml")
	p := NewJsonParser(s.rawJson, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":    0,
				"salary":  "{\"extractedSalary\":{ \"max\": 120000, \"min\": 120000, \"type\": \"yearly\" }}",
				"salary2": "{\"extractedSalary\":{ \"max\": 120000, \"min\": 120000, \"type\": \"yearly\" },\"salarySnippet\":{ \"salaryTextFormatted\": false, \"source\": \"EXTRACTION\", \"text\": \"$120,000 a year\" }}",
				"title":   "Front-End Engineer – 2023 (US)",
			},
			{
				"rank":    1,
				"salary":  "{\"estimatedSalary\":{\n        \"formattedRange\": \"$74.6K - $94.4K a year\",\n        \"max\": 94426.57,\n        \"min\": 74573.43,\n        \"type\": \"YEARLY\"\n      }}",
				"salary2": "{\"estimatedSalary\":{\n        \"formattedRange\": \"$74.6K - $94.4K a year\",\n        \"max\": 94426.57,\n        \"min\": 74573.43,\n        \"type\": \"YEARLY\"\n      },\"salarySnippet\":{ \"salaryTextFormatted\": false }}",
				"title":   "Python Developer",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *JsonParserSuite) Test_0702_locator_list() {
	rawYaml := getBytes("json_yaml/0702.yaml")
	p := NewJsonParser(s.rawJson, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":    0,
				"salary":  "|||{ \"max\": 120000, \"min\": 120000, \"type\": \"yearly\" }",
				"salary2": "|||{ \"max\": 120000, \"min\": 120000, \"type\": \"yearly\" }|||{ \"salaryTextFormatted\": false, \"source\": \"EXTRACTION\", \"text\": \"$120,000 a year\" }",
				"title":   "Front-End Engineer – 2023 (US)",
			},
			{
				"rank":    1,
				"salary":  "{\n        \"formattedRange\": \"$74.6K - $94.4K a year\",\n        \"max\": 94426.57,\n        \"min\": 74573.43,\n        \"type\": \"YEARLY\"\n      }|||",
				"salary2": "{\n        \"formattedRange\": \"$74.6K - $94.4K a year\",\n        \"max\": 94426.57,\n        \"min\": 74573.43,\n        \"type\": \"YEARLY\"\n      }||||||{ \"salaryTextFormatted\": false }",
				"title":   "Python Developer",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

func (s *JsonParserSuite) Test_0703_locator_map() {
	rawYaml := getBytes("json_yaml/0703.yaml")

	p := NewJsonParser(s.rawJson, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":    0,
				"salary":  "{\"esti\":\"\",\"extract\":\"{ \\\"max\\\": 120000, \\\"min\\\": 120000, \\\"type\\\": \\\"yearly\\\" }\"}",
				"salary2": "{\"esti\":\"\",\"extract\":\"{ \\\"max\\\": 120000, \\\"min\\\": 120000, \\\"type\\\": \\\"yearly\\\" }\",\"snipt\":\"{ \\\"salaryTextFormatted\\\": false, \\\"source\\\": \\\"EXTRACTION\\\", \\\"text\\\": \\\"$120,000 a year\\\" }\"}",
				"salary3": "{\"kept\":\"iam not changed\"}",
				"title":   "Front-End Engineer – 2023 (US)",
			},
			{
				"rank":    1,
				"salary":  "{\"esti\":\"{\\n        \\\"formattedRange\\\": \\\"$74.6K - $94.4K a year\\\",\\n        \\\"max\\\": 94426.57,\\n        \\\"min\\\": 74573.43,\\n        \\\"type\\\": \\\"YEARLY\\\"\\n      }\",\"extract\":\"\"}",
				"salary2": "{\"esti\":\"{\\n        \\\"formattedRange\\\": \\\"$74.6K - $94.4K a year\\\",\\n        \\\"max\\\": 94426.57,\\n        \\\"min\\\": 74573.43,\\n        \\\"type\\\": \\\"YEARLY\\\"\\n      }\",\"extract\":\"\",\"snipt\":\"{ \\\"salaryTextFormatted\\\": false }\"}",
				"salary3": "{\"kept\":\"iam not changed\"}",
				"title":   "Python Developer",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

func RefineAttr(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func (s *JsonParserSuite) Test_0704_locator_list2() {
	rawYaml := getBytes("json_yaml/0704.yaml")

	p := NewJsonParser(s.rawJson, rawYaml)
	p.Refiners["RefineAttr"] = RefineAttr
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank": 0,
				"taxo": map[string]interface{}{
					"attr": "Full-time|||{ \"salaryTextFormatted\": false, \"source\": \"EXTRACTION\", \"text\": \"$120,000 a year\" }",
				},
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank": 1,
				"taxo": map[string]interface{}{
					"attr": "Full-time|||{ \"salaryTextFormatted\": false }",
				},
				"title": "Python Developer",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

type Parser3 struct {
	*JsonParser
}

func newParser3(rawData, ymlMap []byte) *Parser3 {
	return &Parser3{
		NewJsonParser(rawData, ymlMap),
	}
}

func (p *Parser3) RefineTaxoAttrArr(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	arr := strings.Split(v, AttrJoinerSep)
	return arr
}

func (p *Parser3) RefineTaxoAttrMap(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	d := make(map[string]interface{})
	json.Unmarshal([]byte(v), &d)
	return d
}

func (s *JsonParserSuite) Test_0801_refineComplexSel() {
	rawYaml := getBytes("json_yaml/0801.yaml")
	p := newParser3(s.rawJson, rawYaml)
	UpdateRefiners(p)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank": 0,
				"taxo": map[string]interface{}{
					"attr_arr": []string{
						"Full-time",
						"$120,000 a year",
					},
				},
				"tier": map[string]interface{}{
					"attr_map": map[string]interface{}{
						"label": "/career/front-end-developer/salaries/Seattle--WA",
						"snip":  "$120,000 a year",
					},
				},
				"title": "Front-End Engineer – 2023 (US)",
			},
			{
				"rank": 1,
				"taxo": map[string]interface{}{
					"attr_arr": []string{
						"Part-time",
						"$40 an hour",
					},
				},
				"tier": map[string]interface{}{
					"attr_map": map[string]interface{}{
						"label": "/career/instructor/salaries",
						"snip":  "$40 an hour",
					},
				},
				"title": "Remote Python Prep Instructor (Part-Time)",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}
