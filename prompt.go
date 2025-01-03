package xparse

import (
	"errors"
	"fmt"
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

// DefaultSeparatorWidth defines the default width for separators
const DefaultSeparatorWidth = 32

// DefaultHint provides templates for refiner implementation and error messages
var DefaultHint = RefinerHint{
	MethodTemplate: `
func (p *%[2]s) %[1]s(raw ...interface{}) interface{} {
    // TODO: Implement refiner logic
    //
    // Parameters:
    // raw[0] interface{}        - Input string to be parsed (type: string)
    // raw[1] interface{}        - Configuration data (type: map/*config.Config)
    // raw[2] interface{}        - Parser context (type: *goquery.Selection/gjson.Result)
    // %[3]s
    // Default implementation (replace with actual logic)
    return p.SplitAtIndex(raw[0], "", -1)
}
`,
	WarningMessage: `
%[2]s
HINT: Missing Refiner Method

The method template above needs to be implemented as it's currently not registered.
REQUIRED: Implement the missing method

Note: If you've already implemented the method but still see this hint,
verify the registration:
1. [RECOMMENDED] Use xparse.UpdateRefiners(p)
   (Only needs to be called once before DoParse)

2. Or manually register:
   p.Refiners["%[1]s"] = p.%[1]s
%[2]s
`,
	Separator: strings.Repeat("-", DefaultSeparatorWidth),
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

var ErrPromptOnly = errors.New("prompt error")

// getTypeNameFromInterface extracts the type name from an interface
func getTypeNameFromInterface(iface interface{}) string {
	typeName := fmt.Sprintf("%T", iface)
	parts := strings.Split(typeName, ".")

	return parts[len(parts)-1]
}

// buildRefinerHintMessage formats the complete hint message
func buildRefinerHintMessage(typeName, mtdName string, cfg *PromptConfig, withWarningMsg bool) []string {
	if cfg == nil {
		cfg = NewPromptConfig()
	}

	messages := []string{
		fmt.Sprintf(cfg.Hint.MethodTemplate, mtdName, typeName, cfg.ExtraHints),
	}

	if withWarningMsg {
		messages = append(messages,
			fmt.Sprintf(cfg.Hint.WarningMessage, mtdName, cfg.Hint.Separator))
	}

	return messages
}

// handleRefinerPrompt handles missing refiner notifications
func handleRefinerPrompt(typeName string, method string, cfg *PromptConfig, withWarningMsg bool) error {
	if cfg == nil {
		cfg = NewPromptConfig()
	}

	if cfg.Style == PromptError {
		return fmt.Errorf("error happens: %w", ErrPromptOnly)
	}

	messages := buildRefinerHintMessage(typeName, method, cfg, withWarningMsg)
	xpretty.GreenPrintf(messages[0])

	if withWarningMsg {
		xpretty.YellowPrintf(messages[1])
	}

	return nil
}

func promptMissingRefiners(parser interface{}, missingMethods []string, opt RefOpts) {
	found := 0
	typeName := getTypeNameFromInterface(parser)

	if len(missingMethods) > 0 {
		base := "Missing following Refiners"
		xpretty.CyanPrintf("%[1]s\n%[2]s\n%[1]s\n", strings.Repeat("-", len(base)), base)
	}

	for _, mtdName := range missingMethods {
		found++
		_ = handleRefinerPrompt(typeName, mtdName, opt.promptCfg, found == len(missingMethods))
	}
}
