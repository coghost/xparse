package xparse

type IDev interface {
	ToggleDevMode(b bool)
	VerifyKeys() []string
}

type IConfig interface {
	RawInfo(args ...string) map[string]interface{}
}

type IData interface {
	BindPresetData(dat map[string]interface{})
	AppendPresetData(data map[string]interface{})

	LoadRootSelection([]byte)

	GetParsedData(keys ...string) interface{}
	DataAsJSON(args ...interface{}) (string, error)
	MustDataAsJSON(args ...interface{}) string
	PrettifyJSONData(args ...interface{}) error

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

func DoParse(parser IParser, opts ...ParseOptFunc) interface{} {
	opt := &ParseOpts{
		promptCfg: NewPromptConfig(),
	}
	bindParseOpts(opt, opts...)

	parser.BindPresetData(opt.preset)
	parser.ToggleDevMode(true)
	UpdateRefiners(parser, WithRefPromptConfig(opt.promptCfg))
	parser.DoParse()
	// parser.PostDoParse()

	return parser.GetParsedData()
}

type ParseOpts struct {
	dataAsSlice bool
	preset      map[string]interface{}
	rootKey     string
	promptCfg   *PromptConfig
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

// WithPresetData: used to bind page level data
func WithPresetData(preset map[string]interface{}) ParseOptFunc {
	return func(o *ParseOpts) {
		o.preset = preset
	}
}

func WithRootKey(s string) ParseOptFunc {
	return func(o *ParseOpts) {
		o.rootKey = s
	}
}

func WithPromptConfig(cfg *PromptConfig) ParseOptFunc {
	return func(o *ParseOpts) {
		o.promptCfg = cfg
	}
}
