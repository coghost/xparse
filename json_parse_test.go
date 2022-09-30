package xparse_test

import (
	"path/filepath"
	"testing"

	"github.com/coghost/xparse"
	"github.com/coghost/xpretty"
	"github.com/gookit/goutil/fsutil"
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type JsonParsesSuite struct {
	suite.Suite
	parser *xparse.JsonParser

	rawJson []byte
	rawYaml []byte
}

func TestJsonParser(t *testing.T) {
	suite.Run(t, new(JsonParsesSuite))
}

func (s *JsonParsesSuite) SetupSuite() {
	xpretty.Initialize(xpretty.WithColor(true), xpretty.WithDummyLog(true))
	home := xparse.GetProjectHome("xparse")
	s.rawJson = fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed.json"))
	s.rawYaml = fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_json.yaml"))
	s.parser = xparse.NewJsonParser(s.rawJson, s.rawYaml)
}

func (s *JsonParsesSuite) TearDownSuite() {
}

func (s *JsonParsesSuite) Test_00_simple() {
	res := gjson.Parse(string(s.rawJson))
	job := res.Array()[0]
	pp.Println(job.Get("jobs.0").String())
}

func (s *JsonParsesSuite) Test_01_init() {
	p := s.parser
	p.ToggleDevMode(true)
	p.DoParse()
	xpretty.PrettyJson(p.MustDataAsJson())
}

func (s *JsonParsesSuite) Test_02_array_as_root() {
	home := xparse.GetProjectHome("xparse")
	rawJson := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_array_as_root.json"))
	rawYaml := fsutil.MustReadFile(filepath.Join(home, "/examples/indeed/indeed_array_as_root.yaml"))

	p := xparse.NewJsonParser(rawJson, rawYaml)
	p.DoParse()
	xpretty.PrettyJson(p.MustDataAsJson())

	p1 := s.parser
	p1.DoParse()
	s.Equal(p.ParsedData, p1.ParsedData)
}
