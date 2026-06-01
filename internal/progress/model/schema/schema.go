package schema

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema/document"

type Status = document.Status

const (
	StatusComplete   = document.StatusComplete
	StatusInProgress = document.StatusInProgress
	StatusPlanned    = document.StatusPlanned
)

type Meta = document.Meta
type UpstreamMeta = document.UpstreamMeta
type BuilderLoopMeta = document.BuilderLoopMeta
type Progress = document.Progress
type Phase = document.Phase
type Subphase = document.Subphase
type Item = document.Item
type Row = document.Row
type Counts = document.Counts
type Stats = document.Stats
