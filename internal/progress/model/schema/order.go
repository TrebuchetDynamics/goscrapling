package schema

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema/ordering"

func SortedPhaseKeys(p *Progress) []string {
	return ordering.Keys(p.Phases)
}

func SortedSubphaseKeys(phase Phase) []string {
	return ordering.Keys(phase.Subphases)
}
