package xparse

type IDev interface {
	ToggleDevMode(b bool)
	GetVerifyKeys() []string
}

type IConfig interface {
	GetRawInfo(args ...string) map[string]interface{}
}

type IData interface {
	BindPresetData(dat map[string]interface{})

	GetParsedData() interface{}
	// GetSliceData() []interface{}
	DataAsJson(args ...interface{}) (string, error)
	MustDataAsJson(args ...interface{}) string
	PrettifyJsonData(args ...interface{})

	DataAsYaml(args ...interface{}) (string, error)
	MustDataAsYaml(args ...interface{}) string

	MustMandatoryFields(got, want []string)

	PostDoParse()

	ExtraInfo() map[string]interface{}
}

type IParser interface {
	IConfig
	IDev
	IData

	DoParse()
}

func PreParse(p IParser, preset map[string]interface{}) {
}

func DoParse(p IParser, opts ...ParseOptFunc) interface{} {
	opt := &ParseOpts{}
	bindParseOpts(opt, opts...)

	p.BindPresetData(opt.preset)
	p.ToggleDevMode(true)
	UpdateRefiners(p)
	p.DoParse()
	p.PostDoParse()

	return p.GetParsedData()
}

func PostParse(p IParser) {
}

type ParseOpts struct {
	dataAsSlice bool

	preset map[string]interface{}
}

type ParseOptFunc func(o *ParseOpts)

func bindParseOpts(opt *ParseOpts, opts ...ParseOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithDataAsSlice(b bool) ParseOptFunc {
	return func(o *ParseOpts) {
		o.dataAsSlice = b
	}
}

func WithPresetData(preset map[string]interface{}) ParseOptFunc {
	return func(o *ParseOpts) {
		o.preset = preset
	}
}
