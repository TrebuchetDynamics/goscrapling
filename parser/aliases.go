package parser

import (
	"github.com/TrebuchetDynamics/goscrapling/core/customtypes"
	"github.com/TrebuchetDynamics/goscrapling/core/storage"
)

type Store = storage.Store
type Key = storage.Key
type Fingerprint = storage.Fingerprint
type ScoreComponent = storage.ScoreComponent

type TextHandler = customtypes.TextHandler
type TextHandlers = customtypes.TextHandlers
type AttributesHandler = customtypes.AttributesHandler

var newAttributesHandler = customtypes.NewAttributesHandler
