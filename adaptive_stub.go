package goscrapling

import "errors"

var (
	errAdaptiveNotImplemented = errors.New("goscrapling: adaptive behavior is not implemented")
	ErrMissingStore           = errors.New("goscrapling: missing adaptive store")
	ErrNilElement             = errors.New("goscrapling: nil element")
	ErrEmptyIdentifier        = errors.New("goscrapling: empty identifier")
	ErrInvalidSelector        = errors.New("goscrapling: invalid selector")
	ErrInvalidPercentage      = errors.New("goscrapling: invalid percentage")
)

type Match struct {
	Element *Element
	Score   float64
}
