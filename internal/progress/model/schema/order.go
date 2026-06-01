package schema

import "sort"

func SortedPhaseKeys(p *Progress) []string {
	keys := make([]string, 0, len(p.Phases))
	for key := range p.Phases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func SortedSubphaseKeys(phase Phase) []string {
	keys := make([]string, 0, len(phase.Subphases))
	for key := range phase.Subphases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
