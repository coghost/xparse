package xparse

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/coghost/xpretty"
	"github.com/spf13/cast"

	"github.com/gookit/goutil/fsutil"
	"github.com/stretchr/testify/suite"
)

type xkcdParser struct {
	*HTMLParser
}

func newXkcdParser(rawHTML, ymlMap []byte) *xkcdParser {
	return &xkcdParser{
		NewHTMLParser(rawHTML, ymlMap),
	}
}

type HTMLParserSuite struct {
	suite.Suite
	parser *xkcdParser

	rawHTML []byte
	rawYaml []byte
}

func TestHtmlParser(t *testing.T) {
	suite.Run(t, new(HTMLParserSuite))
}

func (p *xkcdParser) RefineAltAlt(raw ...interface{}) interface{} {
	return raw[0]
}

func (s *HTMLParserSuite) SetupSuite() {
	xpretty.Initialize(xpretty.WithNoColor(true), xpretty.WithDummyLog(true))
	home := GetProjectHome("xparse")
	s.rawHTML = fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd_353.html"))
	s.rawYaml = fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd.yaml"))
	// s.rawHtml = fsutil.MustReadFile("./examples/xkcd/xkcd_353.html")
	// s.rawYaml = fsutil.MustReadFile("./examples/xkcd/xkcd.yaml")
	s.parser = newXkcdParser(s.rawHTML, s.rawYaml)
	// s.parser.Refiners["refine_image_1_src"] = refine_image_1_src
	// s.parser.Refiners["_refine_alt_alt"] = s._refine_alt_alt
}

func (s *HTMLParserSuite) TearDownSuite() {
}

func (s *HTMLParserSuite) Test_0100PanicsWithUnsupportedType() {
	yml := getBytes("html_yaml/0100.yaml")
	ps := NewHTMLParser(s.rawHTML, yml)
	s.PanicsWithValue(
		"unknown type of (footnote:[div.ok div.fail]), only support (1:string or 2:map[string]interface{})",
		func() {
			ps.DoParse()
		})
}

func (s *HTMLParserSuite) Test_0101PanicsWithIndexE1() {
	yml := getBytes("html_yaml/0101.yaml")
	ps := NewHTMLParser(s.rawHTML, yml)
	s.PanicsWithValue(
		"all indexes should be int, but (comic_nav is []interface {}: [b a])",
		func() {
			ps.DoParse()
		})
}

func (s *HTMLParserSuite) Test_0102PanicsWithIndexE2() {
	yml := getBytes("html_yaml/0102.yaml")
	ps := NewHTMLParser(s.rawHTML, yml)
	s.PanicsWithValue(
		"index should be int/int64/uint64 or []interface{}, but (comic_nav is map[string]interface {}: map[fail:href panic:enrich_url])",
		func() {
			ps.DoParse()
		})
}

func (s *HTMLParserSuite) Test_0103PanicsWithRefineMethod() {
	yml := getBytes("html_yaml/0103.yaml")
	ps := NewHTMLParser(s.rawHTML, yml)

	refineCtitle := func(raw ...interface{}) interface{} {
		return ""
	}
	ps.Refiners["_refine_ctitle1"] = refineCtitle
	s.Panics(
		func() {
			ps.DoParse()
		})
}

func (s *HTMLParserSuite) Test_0200DataStr() {
	yml := getBytes("html_yaml/0200.yaml")
	ps := NewHTMLParser(s.rawHTML, yml)
	ps.DoParse()
	raw, e := ps.DataAsJSON()
	s.Nil(e)
	s.NotNil(raw)
	s.Equal("{}", raw)
}

func (s *HTMLParserSuite) Test_0300InRealWorld() {
	UpdateRefiners(s.parser)
	s.parser.DoParse()
	dat := s.parser.MustDataAsJSON()

	want := `
{
	"bottom": {
		"comic": {
			"href": "https://xkcd.com/1732/"
		},
		"comic_links": [{
			"href": "http://threewordphrase.com/",
			"text": "Three Word Phrase"
		}, {
			"href": "https://www.smbc-comics.com/",
			"text": "SMBC"
		}, {
			"href": "https://www.qwantz.com",
			"text": "Dinosaur Comics"
		}, {
			"href": "https://oglaf.com/",
			"text": "Oglaf"
		}, {
			"href": "https://www.asofterworld.com",
			"text": "A Softer World"
		}, {
			"href": "https://buttersafe.com/",
			"text": "Buttersafe"
		}, {
			"href": "https://pbfcomics.com/",
			"text": "Perry Bible Fellowship"
		}, {
			"href": "https://questionablecontent.net/",
			"text": "Questionable Content"
		}, {
			"href": "http://www.buttercupfestival.com/",
			"text": "Buttercup Festival"
		}, {
			"href": "https://www.mspaintadventures.com/",
			"text": "Homestuck"
		}, {
			"href": "https://www.jspowerhour.com/",
			"text": "Junior Scientist Power Hour"
		}, {
			"href": "https://medium.com/civic-tech-thoughts-from-joshdata/so-you-want-to-reform-democracy-7f3b1ef10597",
			"text": "Tips on technology and government"
		}, {
			"href": "https://www.nytimes.com/interactive/2017/climate/what-is-climate-change.html",
			"text": "Climate FAQ"
		}, {
			"href": "https://twitter.com/KHayhoe",
			"text": "Katharine Hayhoe"
		}],
		"comic_map": [{
			"alt": "Grownups",
			"coords": "0,0,100,100",
			"href": "https://xkcd.com/150/"
		}, {
			"alt": "Circuit Diagram",
			"coords": "104,0,204,100",
			"href": "https://xkcd.com/730/"
		}, {
			"alt": "Angular Momentum",
			"coords": "208,0,308,100",
			"href": "https://xkcd.com/162/"
		}, {
			"alt": "Self-Description",
			"coords": "312,0,412,100",
			"href": "https://xkcd.com/688/"
		}, {
			"alt": "Alternative Energy Revolution",
			"coords": "416,0,520,100",
			"href": "https://xkcd.com/556/"
		}],
		"feed": [{
			"href": "https://xkcd.com/rss.xml",
			"text": "RSS Feed"
		}]
	},
	"middle_container": {
		"comic": {
			"alt": "Python",
			"src": "//imgs.xkcd.com/comics/python.png",
			"title": "I wrote 20 short programs in Python yesterday.  It was wonderful.  Perl, I'm leaving you."
		},
		"comic_nav": {
			"nav": [{
				"accesskey": "",
				"href": "https://xkcd.com/1/",
				"rel": "",
				"text": "|\u003c"
			}, {
				"accesskey": "p",
				"href": "https://xkcd.com/352/",
				"rel": "prev",
				"text": "\u003c Prev"
			}, {
				"accesskey": "",
				"href": "https://c.xkcd.com/random/comic/",
				"rel": "",
				"text": "Random"
			}, {
				"accesskey": "n",
				"href": "https://xkcd.com/354/",
				"rel": "next",
				"text": "Next \u003e"
			}, {
				"accesskey": "",
				"href": "https://xkcd.com/",
				"rel": "",
				"text": "\u003e|"
			}]
		},
		"ctitle": "Python",
		"transcript": "[[ Guy 1 is talking to Guy 2, who is floating in the sky ]]\nGuy 1: You're flying! How?\nGuy 2: Python!\nGuy 2: I learned it last night! Everything is so simple!\nGuy 2: Hello world is just 'print \"Hello, World!\" '\nGuy 1: I dunno... Dynamic typing? Whitespace?\nGuy 2: Come join us! Programming is fun again! It's a whole new world up here!\nGuy 1: But how are you flying?\nGuy 2: I just typed 'import antigravity'\nGuy 1: That's it?\nGuy 2: ...I also sampled everything in the medicine cabinet for comparison.\nGuy 2: But i think this is the python.\n{{ I wrote 20 short programs in Python yesterday.  It was wonderful.  Perl, I'm leaving you. }}"
	},
	"page": {
		"by_multiple_locators": ["Archive", "What If?", "Blag", "How To", "Store", "About", "Feed", "Email", "TW", "FB", "IG", "", "", "|\u003c", "\u003c Prev", "Random", "Next \u003e", "\u003e|", "|\u003c", "\u003c Prev", "Random", "Next \u003e", "\u003e|", "https://xkcd.com/353/", "https://imgs.xkcd.com/comics/python.png", "", "RSS Feed", "Atom Feed", "Email", "Three Word Phrase", "SMBC", "Dinosaur Comics", "Oglaf", "A Softer World", "Buttersafe", "Perry Bible Fellowship", "Questionable Content", "Buttercup Festival", "Homestuck", "Junior Scientist Power Hour", "Tips on technology and government", "Climate FAQ", "Katharine Hayhoe", "Creative Commons Attribution-NonCommercial 2.5 License", "More details"],
		"footnote": "xkcd.com is best viewed with Netscape Navigator 4.0 or below on a Pentium 3±1 emulated in Javascript on an Apple IIGSat a screen resolution of 1024x1. Please enable your ad blockers, disable high-heat drying, and remove your devicefrom Airplane Mode and set it to Boat Mode. For security reasons, please leave caps lock on while browsing.",
		"license": "\n\nThis work is licensed under a\nCreative Commons Attribution-NonCommercial 2.5 License.\n\nThis means you're free to copy and share these comics (but not to sell them). More details.\n",
		"license1": "This work is licensed under a\nCreative Commons Attribution-NonCommercial 2.5 License.\n\nThis means you're free to copy and share these comics (but not to sell them). More details.",
		"title": "xkcd: Python"
	},
	"top_container": {
		"first_link": "Archive",
		"top_left": [{
			"href": "https://xkcd.com/archive",
			"text": "Archive"
		}, {
			"href": "https://what-if.xkcd.com",
			"text": "What If?"
		}, {
			"href": "https://blag.xkcd.com",
			"text": "Blag"
		}, {
			"href": "https://xkcd.com/how-to/",
			"text": "How To"
		}, {
			"href": "https://store.xkcd.com/",
			"text": "Store"
		}, {
			"href": "https://xkcd.com/about",
			"text": "About"
		}, {
			"href": "https://xkcd.com/atom.xml",
			"text": "Feed"
		}, {
			"href": "https://xkcd.com/newsletter/",
			"text": "Email"
		}, {
			"href": "https://twitter.com/xkcd/",
			"text": "TW"
		}, {
			"href": "https://www.facebook.com/TheXKCD/",
			"text": "FB"
		}, {
			"href": "https://www.instagram.com/xkcd/",
			"text": "IG"
		}, {
			"href": "https://xkcd.com/",
			"text": ""
		}, {
			"href": "https://blacklivesmatter.com/",
			"text": ""
		}],
		"top_right": {
			"masthead": {
				"image": {
					"alt": "xkcd.com logo",
					"src": "https://xkcd.com/s/0b7742.png"
				},
				"slogan": "A webcomic of romance, sarcasm, math, and language."
			},
			"news": {
				"links": {
					"href": "https://blacklivesmatter.com/",
					"text": ""
				}
			}
		}
	}
}
	`
	s.JSONEq(want, dat)
}

func refineAltAlt(raw ...interface{}) interface{} {
	return raw[0]
}

func (s *HTMLParserSuite) Test_0301DevMode() {
	rawYaml := getBytes("html_yaml/0301.yaml")
	p := NewHTMLParser(s.rawHTML, rawYaml)
	p.ToggleDevMode(true)
	p.Refiners["RefineAltAlt"] = refineAltAlt
	p.DoParse()

	want := map[string]interface{}{
		"bottom": map[string]interface{}{
			"comic_links": []map[string]interface{}{
				{
					"href": "http://threewordphrase.com/",
					"text": "Three Word Phrase",
				},
				{
					"href": "https://www.smbc-comics.com/",
					"text": "SMBC",
				},
				{
					"href": "https://www.qwantz.com",
					"text": "Dinosaur Comics",
				},
				{
					"href": "https://oglaf.com/",
					"text": "Oglaf",
				},
				{
					"href": "https://www.asofterworld.com",
					"text": "A Softer World",
				},
				{
					"href": "https://buttersafe.com/",
					"text": "Buttersafe",
				},
				{
					"href": "https://pbfcomics.com/",
					"text": "Perry Bible Fellowship",
				},
				{
					"href": "https://questionablecontent.net/",
					"text": "Questionable Content",
				},
				{
					"href": "http://www.buttercupfestival.com/",
					"text": "Buttercup Festival",
				},
				{
					"href": "https://www.mspaintadventures.com/",
					"text": "Homestuck",
				},
				{
					"href": "https://www.jspowerhour.com/",
					"text": "Junior Scientist Power Hour",
				},
				{
					"href": "https://medium.com/civic-tech-thoughts-from-joshdata/so-you-want-to-reform-democracy-7f3b1ef10597",
					"text": "Tips on technology and government",
				},
				{
					"href": "https://www.nytimes.com/interactive/2017/climate/what-is-climate-change.html",
					"text": "Climate FAQ",
				},
				{
					"href": "https://twitter.com/KHayhoe",
					"text": "Katharine Hayhoe",
				},
			},
		},
		"middle": map[string]interface{}{
			"ctitle":     "Python",
			"transcript": "[[ Guy 1 is talking to Guy 2, who is floating in the sky ]]\nGuy 1: You're flying! How?\nGuy 2: Python!\nGuy 2: I learned it last night! Everything is so simple!\nGuy 2: Hello world is just 'print \"Hello, World!\" '\nGuy 1: I dunno... Dynamic typing? Whitespace?\nGuy 2: Come join us! Programming is fun again! It's a whole new world up here!\nGuy 1: But how are you flying?\nGuy 2: I just typed 'import antigravity'\nGuy 1: That's it?\nGuy 2: ...I also sampled everything in the medicine cabinet for comparison.\nGuy 2: But i think this is the python.\n{{ I wrote 20 short programs in Python yesterday.  It was wonderful.  Perl, I'm leaving you. }}",
		},
	}
	s.Equal(want, p.ParsedData)
}

func getIndeedHTMLData(fname string) (b1, b2 []byte) {
	rawYaml := getBytes("html_yaml/" + fname)
	rawHTML := getBytes("indeed/indeed.html")

	return rawHTML, rawYaml
}

func (s *HTMLParserSuite) Test_0400_index() {
	rawHTML, rawYaml := getIndeedHTMLData("0400.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":  0,
				"title": "Python Software Engineer",
			},
			{
				"rank":  1,
				"title": "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *HTMLParserSuite) Test_0500_type() {
	rawHTML, rawYaml := getIndeedHTMLData("0500.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":       0,
				"rating":     2.000000,
				"rating_b":   false,
				"rating_i":   2,
				"rating_non": "",
				"title":      "Python Software Engineer",
			},
			{
				"rank":       1,
				"rating":     3.400000,
				"rating_b":   false,
				"rating_i":   3,
				"rating_non": "",
				"title":      "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}
	s.Equal(want, p.ParsedData)
}

type htmlParser1 struct {
	*HTMLParser
}

func newHTMLParser1(rawHTML, ymlMap []byte) *htmlParser1 {
	return &htmlParser1{
		NewHTMLParser(rawHTML, ymlMap),
	}
}

func (p *htmlParser1) RefineRating(raw ...interface{}) interface{} {
	v := cast.ToFloat64(raw[0])
	return v
}

func (p *htmlParser1) RefineLevel(raw ...interface{}) interface{} {
	switch v := cast.ToFloat64(raw[0]); {
	case v <= 0, v > 5:
		return ""
	case v <= 2:
		return "D"
	case v <= 3:
		return "C"
	case v <= 4:
		return "B"
	case v <= 5:
		return "A"
	default:
		return ""
	}
}

func (p *htmlParser1) GenLevel(raw ...interface{}) interface{} {
	v := p.RefineLevel(raw...)
	return fmt.Sprintf("G-%v", v)
}

func (s *HTMLParserSuite) Test_0601_attrRefineManually() {
	rawHTML, rawYaml := getIndeedHTMLData("0601.yaml")
	p := newHTMLParser1(rawHTML, rawYaml)
	p.Refiners["RefineRating"] = p.RefineRating
	p.Refiners["RefineLevel"] = p.RefineLevel
	p.Refiners["GenLevel"] = p.GenLevel
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":           0,
				"rating":         2.000000,
				"rating_level":   "D",
				"rating_level_2": "D",
				"rating_level_3": "G-D",
				"title":          "Python Software Engineer",
			},
			{
				"rank":           1,
				"rating":         3.400000,
				"rating_level":   "B",
				"rating_level_2": "B",
				"rating_level_3": "G-B",
				"title":          "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *HTMLParserSuite) Test_0602_attrRefineAuto() {
	rawHTML, rawYaml := getIndeedHTMLData("0601.yaml")
	p := newHTMLParser1(rawHTML, rawYaml)
	UpdateRefiners(p)
	p.DoParse()

	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"rank":           0,
				"rating":         2.000000,
				"rating_level":   "D",
				"rating_level_2": "D",
				"rating_level_3": "G-D",
				"title":          "Python Software Engineer",
			},
			{
				"rank":           1,
				"rating":         3.400000,
				"rating_level":   "B",
				"rating_level_2": "B",
				"rating_level_3": "G-B",
				"title":          "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

type htmlParser2 struct {
	*HTMLParser
}

func newHTMLParser2(rawHTML, ymlMap []byte) *htmlParser2 {
	return &htmlParser2{
		NewHTMLParser(rawHTML, ymlMap),
	}
}

func (p *htmlParser2) RefineCompInfo(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func (p *htmlParser2) RefineCompInfoArr(raw ...interface{}) interface{} {
	v := cast.ToString(raw[0])
	return v
}

func (s *HTMLParserSuite) Test_0701_complexSel() {
	rawHTML, rawYaml := getIndeedHTMLData("0701.yaml")
	p := newHTMLParser2(rawHTML, rawYaml)
	UpdateRefiners(p)
	p.DoParse()
	want := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"comp_info": []interface{}{
					"Estimated $102K - $129K a year",
					"2.0",
					"Zelis",
					"Python Software Engineer",
					"Remote",
				},
				"comp_info_arr": []interface{}{
					"Estimated $102K - $129K a year",
					"2.0",
					"Zelis",
					"Python Software Engineer",
					"Remote",
				},
				"comp_info_map": "{\"location\":\"Remote\",\"name\":\"Zelis\",\"rating\":\"2.0\",\"salary\":\"Estimated $102K - $129K a year\",\"title\":\"Python Software Engineer\"}",
				"rank":          0,
				"restub_arr": map[string]interface{}{
					"comp_info": []string{
						"Estimated $102K - $129K a year",
						"2.0",
						"Zelis",
						"Python Software Engineer",
						"Remote",
					},
				},
				"restub_map": map[string]interface{}{
					"comp_info": "{\"location\":\"Remote\",\"name\":\"Zelis\",\"rating\":\"2.0\",\"salary\":\"Estimated $102K - $129K a year\",\"title\":\"Python Software Engineer\"}",
				},
			},
			{
				"comp_info": []interface{}{
					"",
					"3.4",
					"CrowdStrike",
					"Data Scientist, Malware Detections Team (Remote)",
					"+1 locationRemote",
				},
				"comp_info_arr": []interface{}{
					"",
					"3.4",
					"CrowdStrike",
					"Data Scientist, Malware Detections Team (Remote)",
					"+1 locationRemote",
				},
				"comp_info_map": "{\"location\":\"+1 locationRemote\",\"name\":\"CrowdStrike\",\"rating\":\"3.4\",\"salary\":\"\",\"title\":\"Data Scientist, Malware Detections Team (Remote)\"}",
				"rank":          1,
				"restub_arr": map[string]interface{}{
					"comp_info": []string{
						"",
						"3.4",
						"CrowdStrike",
						"Data Scientist, Malware Detections Team (Remote)",
						"+1 locationRemote",
					},
				},
				"restub_map": map[string]interface{}{
					"comp_info": "{\"location\":\"+1 locationRemote\",\"name\":\"CrowdStrike\",\"rating\":\"3.4\",\"salary\":\"\",\"title\":\"Data Scientist, Malware Detections Team (Remote)\"}",
				},
			},
		},
	}

	s.Equal(want, p.ParsedData)
}

func (s *HTMLParserSuite) Test_0800() {
	rawHTML, rawYaml := getIndeedHTMLData("0800.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	failed, all := Verify(p.MustDataAsJSON(), p.VerifyKeys())

	s.Empty(failed)
	wantAll := map[string]map[int][]string{
		"jobs": {
			0: []string{
				"Python Software Engineer",
			},
			1: []string{
				"Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}
	s.Equal(wantAll, all)
}

func (s *HTMLParserSuite) Test_0801() {
	rawHTML, rawYaml := getIndeedHTMLData("0801.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	failed, all := Verify(p.MustDataAsJSON(), p.VerifyKeys())

	wantF := map[string][]string{
		"jobs": {
			"0:listing_date",
			"1:listing_date",
		},
	}

	wantAll := map[string]map[int][]string{
		"jobs": {
			0: []string{
				"Python Software Engineer",
				"",
			},
			1: []string{
				"Data Scientist, Malware Detections Team (Remote)",
				"",
			},
		},
		"pages": {
			0: []string{
				"job_a619997ec53df4dc",
			},
			1: []string{
				"job_8cd20f584d7164c7",
			},
		},
	}

	s.Equal(wantF, failed)
	s.Equal(wantAll, all)
}

func (s *HTMLParserSuite) Test_0802() {
	rawHTML, rawYaml := getIndeedHTMLData("0802.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	failed, all := Verify(p.MustDataAsJSON(), p.VerifyKeys())

	wantF := map[string][]string{
		"jobs": {
			"0:listing_date",
			"1:listing_date",
		},
	}
	wantAll := map[string]map[int][]string{
		"jobs": {
			0: []string{
				"Python Software Engineer",
				"job_a619997ec53df4dc",
				"",
			},
			1: []string{
				"Data Scientist, Malware Detections Team (Remote)",
				"job_8cd20f584d7164c7",
				"",
			},
		},
		"pages": {
			0: []string{
				"Python Software Engineer",
			},
			1: []string{
				"Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}
	s.Equal(wantF, failed)
	s.Equal(wantAll, all)
}

func (s *HTMLParserSuite) Test_0900() {
	rawHTML, rawYaml := getIndeedHTMLData("0900.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	wantData := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"company": map[string]interface{}{
					"name": "CrowdStrike",
				},
				"id":           "8cd20f584d7164c7",
				"listing_date": "",
				"rank":         0,
				"remote":       "Remote",
				"title":        "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}

	s.Equal(wantData, p.ParsedData)
}

func (s *HTMLParserSuite) Test_0901() {
	rawHTML, rawYaml := getIndeedHTMLData("0901.yaml")
	p := NewHTMLParser(rawHTML, rawYaml)
	p.DoParse()

	wantData := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"company": map[string]interface{}{
					"name": "CrowdStrike",
				},
				"id":           "8cd20f584d7164c7",
				"listing_date": "",
				"rank":         0,
				"remote":       "Remote",
				"title":        "Data Scientist, Malware Detections Team (Remote)",
			},
		},
	}

	s.Equal(wantData, p.ParsedData)
}
