package schema

func (p *Progress) Stats() Stats {
	var stats Stats
	stats.Phases.Total = len(p.Phases)
	for _, phaseKey := range SortedPhaseKeys(p) {
		phase := p.Phases[phaseKey]
		phaseStatus := phase.derivedStatus()
		increment(&stats.Phases, phaseStatus)
		stats.Subphases.Total += len(phase.Subphases)
		for _, subphaseKey := range SortedSubphaseKeys(phase) {
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
	if len(phase.Subphases) == 0 {
		return StatusPlanned
	}
	allComplete := true
	anyStarted := false
	for _, subphase := range phase.Subphases {
		status := subphase.derivedStatus()
		if status != StatusComplete {
			allComplete = false
		}
		if status == StatusComplete || status == StatusInProgress {
			anyStarted = true
		}
	}
	switch {
	case allComplete:
		return StatusComplete
	case anyStarted:
		return StatusInProgress
	default:
		return StatusPlanned
	}
}

func (subphase Subphase) derivedStatus() Status {
	if len(subphase.Items) == 0 {
		return StatusPlanned
	}
	allComplete := true
	anyStarted := false
	for _, item := range subphase.Items {
		if item.Status != StatusComplete {
			allComplete = false
		}
		if item.Status == StatusComplete || item.Status == StatusInProgress {
			anyStarted = true
		}
	}
	switch {
	case allComplete:
		return StatusComplete
	case anyStarted:
		return StatusInProgress
	default:
		return StatusPlanned
	}
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
