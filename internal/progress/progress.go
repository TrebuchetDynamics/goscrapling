package progress

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/model"

type Status = model.Status

const (
	StatusComplete   = model.StatusComplete
	StatusInProgress = model.StatusInProgress
	StatusPlanned    = model.StatusPlanned
)

type Meta = model.Meta
type UpstreamMeta = model.UpstreamMeta
type BuilderLoopMeta = model.BuilderLoopMeta
type Progress = model.Progress
type Phase = model.Phase
type Subphase = model.Subphase
type Item = model.Item
type Row = model.Row
type Counts = model.Counts
type Stats = model.Stats

func Load(path string) (*Progress, error) {
	return model.Load(path)
}

func AgentQueueRows(p *Progress) []Row {
	return model.AgentQueueRows(p)
}

func BlockedRows(p *Progress) []Row {
	return model.BlockedRows(p)
}

func UmbrellaRows(p *Progress) []Row {
	return model.UmbrellaRows(p)
}

func sortedPhaseKeys(p *Progress) []string {
	return model.SortedPhaseKeys(p)
}

func sortedSubphaseKeys(phase Phase) []string {
	return model.SortedSubphaseKeys(phase)
}
