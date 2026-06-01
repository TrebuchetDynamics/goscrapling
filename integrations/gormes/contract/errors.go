package contract

import "errors"

// ErrUnknownTool is returned when a Gormes extraction adapter receives an unsupported tool name.
var ErrUnknownTool = errors.New("unknown gormes browser extraction tool")
