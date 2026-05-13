package goscrapling

import "errors"

var errAdaptiveNotImplemented = errors.New("goscrapling: adaptive behavior is not implemented")

type Match struct {
	Element *Element
	Score   float64
}
