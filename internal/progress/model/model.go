package model

import (
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/jsonfile"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/selection"
)

type Status = schema.Status

const (
	StatusComplete   = schema.StatusComplete
	StatusInProgress = schema.StatusInProgress
	StatusPlanned    = schema.StatusPlanned
)

type Meta = schema.Meta
type UpstreamMeta = schema.UpstreamMeta
type BuilderLoopMeta = schema.BuilderLoopMeta
type Progress = schema.Progress
type Phase = schema.Phase
type Subphase = schema.Subphase
type Item = schema.Item
type Row = schema.Row
type Counts = schema.Counts
type Stats = schema.Stats

func Load(path string) (*Progress, error) {
	return jsonfile.Load(path)
}

func AgentQueueRows(p *Progress) []Row {
	return selection.AgentQueueRows(p)
}

func BlockedRows(p *Progress) []Row {
	return selection.BlockedRows(p)
}

func UmbrellaRows(p *Progress) []Row {
	return selection.UmbrellaRows(p)
}

func Rows(p *Progress) []Row {
	return selection.Rows(p)
}

func SortedPhaseKeys(p *Progress) []string {
	return schema.SortedPhaseKeys(p)
}

func SortedSubphaseKeys(phase Phase) []string {
	return schema.SortedSubphaseKeys(phase)
}
