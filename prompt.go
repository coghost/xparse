package xparse

import (
	"fmt"
	"os"
	"strings"

	"github.com/coghost/xpretty"
)

// PromptStyle defines how missing refiner messages are handled
type PromptStyle int

const (
	// PromptError returns error without exiting
	PromptError PromptStyle = iota
	// PromptPrint only prints warning
	PromptPrint
)

// RefinerHint provides template for missing refiner method
type RefinerHint struct {
	MethodTemplate string
	WarningMessage string
	Separator      string
}

// DefaultHint provides default templates for missing refiners
var DefaultHint = RefinerHint{
	MethodTemplate: `
func (p *%[3]s) %[1]s(raw ...interface{}) interface{} {
    // TODO: raw[0] is the interface of string value parsed
    // TODO: raw[1] is map/*config.Config
    // TODO: raw[2] is *goquery.Selection/gjson.Result%[4]s
    txt := p.SplitAtIndex(raw[0], "", -1)
    return txt
}
`,
	WarningMessage: `
%[4]s
WARN: WHY GOT THIS PROMPT?
Maybe you've missed one of following methods:

- RECOMMENDED: you can call xparse.UpdateRefiners(p) before DoParse
  + this only need once
- or you can manually assign it to p.Refiners by:
  + p.Refiners["%[1]s"] = p.%[1]s
  + every new refiner is required
%[4]s
`,
	Separator: strings.Repeat("-", 32),
}

// PromptConfig controls prompt behavior
type PromptConfig struct {
	Style PromptStyle
	Hint  RefinerHint

	ExtraHints string
}

func NewPromptConfig(hints ...string) *PromptConfig {
	return &PromptConfig{
		Style:      PromptPrint,
		Hint:       DefaultHint,
		ExtraHints: FirstOrDefaultArgs("", hints...),
	}
}

// prompts handles missing refiner notifications
func prompts(iface interface{}, snakeMtdName, mtdName string, cfg *PromptConfig) error {
	if cfg == nil {
		cfg = NewPromptConfig()
	}

	prmType := fmt.Sprintf("%T", iface)
	arr := strings.Split(prmType, ".")
	prmType = arr[len(arr)-1]

	msg := fmt.Sprintf("Missing Refiner: (%s or %s)\n", snakeMtdName, mtdName)

	switch cfg.Style {
	case PromptError:
		return fmt.Errorf(msg)
	default:
		xpretty.RedPrintf(msg)
		xpretty.GreenPrintf(cfg.Hint.MethodTemplate, mtdName, snakeMtdName, prmType, cfg.ExtraHints)
		xpretty.YellowPrintf(cfg.Hint.WarningMessage, mtdName, snakeMtdName, prmType, cfg.Hint.Separator)
		os.Exit(0)

		return nil
	}
}
