package document

import (
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema/ordering"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema/rollup"
)

func (p *Progress) Stats() Stats {
	var stats Stats
	stats.Phases.Total = len(p.Phases)
	for _, phaseKey := range ordering.Keys(p.Phases) {
		phase := p.Phases[phaseKey]
		phaseStatus := phase.derivedStatus()
		increment(&stats.Phases, phaseStatus)
		stats.Subphases.Total += len(phase.Subphases)
		for _, subphaseKey := range ordering.Keys(phase.Subphases) {
			subphase := phase.Subphases[subphaseKey]
			increment(&stats.Subphases, subphase.derivedStatus())
			stats.Items.Total += len(subphase.Items)
			for _, item := range subphase.Items {
				increment(&stats.Items, item.Status)
			}
		}
	}
	return stats
}

func (phase Phase) derivedStatus() Status {
	statuses := make([]Status, 0, len(phase.Subphases))
	for _, subphase := range phase.Subphases {
		statuses = append(statuses, subphase.derivedStatus())
	}
	return rollup.Status(statuses, StatusComplete, StatusInProgress, StatusPlanned)
}

func (subphase Subphase) derivedStatus() Status {
	statuses := make([]Status, 0, len(subphase.Items))
	for _, item := range subphase.Items {
		statuses = append(statuses, item.Status)
	}
	return rollup.Status(statuses, StatusComplete, StatusInProgress, StatusPlanned)
}

func increment(counts *Counts, status Status) {
	switch status {
	case StatusComplete:
		counts.Complete++
	case StatusInProgress:
		counts.InProgress++
	default:
		counts.Planned++
	}
}
