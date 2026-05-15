package fetchers

import (
	"github.com/TrebuchetDynamics/goscrapling/core/storage"
	"github.com/TrebuchetDynamics/goscrapling/engines/toolbelt"
)

type Store = storage.Store
type Response = toolbelt.Response
type ResponseOptions = toolbelt.ResponseOptions
type RequestMetadata = toolbelt.RequestMetadata

var NewResponse = toolbelt.NewResponse
