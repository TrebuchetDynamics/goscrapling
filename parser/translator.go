package parser

import "github.com/TrebuchetDynamics/goscrapling/core/translator"

func CSSToXPath(selector string) (string, error) {
	return translator.CSSToXPath(selector)
}
