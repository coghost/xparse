package xparse

type VerifyOp int

const (
	VerifyPrintNone VerifyOp = iota
	VerifyPrintAll
	VerifyPrintMissed
)

type VerifyOpts struct {
	level VerifyOp

	stubKey string

	color bool
}

type VerifyOptFunc func(o *VerifyOpts)

func bindVerifyOpts(opt *VerifyOpts, opts ...VerifyOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithOuputLevel(i VerifyOp) VerifyOptFunc {
	return func(o *VerifyOpts) {
		o.level = i
	}
}

func WithStubKey(s string) VerifyOptFunc {
	return func(o *VerifyOpts) {
		o.stubKey = s
	}
}

func WithColor(b bool) VerifyOptFunc {
	return func(o *VerifyOpts) {
		o.color = b
	}
}
