package xparse_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/coghost/xparse"
	"github.com/coghost/xpretty"

	"github.com/gookit/config/v2"
	"github.com/gookit/goutil/fsutil"
	"github.com/stretchr/testify/suite"
)

type ParserSuite struct {
	suite.Suite
	parser *xparse.Parser

	rawHtml []byte
	rawYaml []byte
}

func TestParser(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

func refine_image_1_src(raw ...interface{}) interface{} {
	cfg := raw[1].(*config.Config)
	domain := cfg.String("__raw.site_url")
	uri := xparse.EnrichUrl(raw[0], domain)
	return uri
}

func (s *ParserSuite) _refine_alt_alt(raw ...interface{}) interface{} {
	cfg := raw[0]
	return cfg
}

func (s *ParserSuite) SetupSuite() {
	home := xparse.GetProjectHome("xparse")
	s.rawHtml = fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd_353.html"))
	s.rawYaml = fsutil.MustReadFile(filepath.Join(home, "/examples/xkcd/xkcd.yaml"))
	// s.rawHtml = fsutil.MustReadFile("./examples/xkcd/xkcd_353.html")
	// s.rawYaml = fsutil.MustReadFile("./examples/xkcd/xkcd.yaml")
	s.parser = xparse.NewParser(s.rawHtml, s.rawYaml)
	s.parser.Refiners["refine_image_1_src"] = refine_image_1_src
	s.parser.Refiners["_refine_alt_alt"] = s._refine_alt_alt
}

func (s *ParserSuite) TearDownSuite() {
}

func (s *ParserSuite) Test01_00PanicsWithUnsupportedType() {
	str := `__raw:
page:
  title: head>title
  footnote:
  - div.ok
  - div.fail
non_test_keys:
  - div.fail
`
	yml := []byte(str)
	ps := xparse.NewParser(s.rawHtml, yml)
	s.Panics(func() {
		ps.DoParse()
	})
}

func (s *ParserSuite) Test01_01PanicsWithIndexE1() {
	str := `__raw:
middle_container:
  _locator: div#middleContainer
  ctitle: div#ctitle
  comic_nav:
    _locator: ul.comicNav
    _index:
    - b
    - a
`
	yml := []byte(str)
	ps := xparse.NewParser(s.rawHtml, yml)
	s.Panics(func() {
		ps.DoParse()
	})
}

func (s *ParserSuite) Test01_02PanicsWithIndexE2() {
	str := `__raw:
middle_container:
  _locator: div#middleContainer
  ctitle: div#ctitle
  comic_nav:
    _locator: ul.comicNav
    _index:
        fail: href
        panic: enrich_url
`
	yml := []byte(str)
	ps := xparse.NewParser(s.rawHtml, yml)
	s.Panics(func() {
		ps.DoParse()
	})
}

func (s *ParserSuite) Test01_03PanicsWithRefineMethod() {
	str := `__raw:
middle_container:
  _locator: div#middleContainer
  ctitle:
    _locator: div#ctitle
    _attr_refine: enrich_url
  ctitle1:
    _locator: div#ctitle
    _attr_refine: true
  comic_nav:
    _locator: ul.comicNav
    _index: 0
    nav:
      _index: ~
      _locator: li>a
      text:
      href:
        _attr: href
        _attr_refine:
          - enrich_url
      rel:
        _attr: rel
      accesskey:
        _attr: accesskey
  comic:
    _locator: div#comic>img
    _attr:
      - src
      - title
      - alt
  transcript: div#transcript
`
	yml := []byte(str)
	ps := xparse.NewParser(s.rawHtml, yml)

	_refine_ctitle1 := func(raw ...interface{}) interface{} {
		return ""
	}
	ps.Refiners["_refine_ctitle1"] = _refine_ctitle1
	s.Panics(func() {
		ps.DoParse()
	})
}

func (s *ParserSuite) Test02_00DataStr() {
	str := `__raw:
  test_keys:
    - non_map

non_map: div.abc
non_test_keys: div.non`
	yml := []byte(str)
	ps := xparse.NewParser(s.rawHtml, yml)
	ps.DoParse()
	raw, e := ps.DataAsJson()
	s.Nil(e)
	s.NotNil(raw)
	s.Equal(raw, "{}")
}

func (s *ParserSuite) Test03_00InRealWorld() {
	xpretty.ToggleColor(true)
	s.parser.DoParse()

	dat, err := json.Marshal(s.parser.ParsedData)
	s.Nil(err)

	exp := `
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
	s.JSONEq(exp, string(dat))
}
