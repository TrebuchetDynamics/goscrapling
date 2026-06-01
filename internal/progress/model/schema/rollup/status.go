package rollup

// Status derives an aggregate status from child statuses.
func Status[S comparable](children []S, complete, inProgress, planned S) S {
	if len(children) == 0 {
		return planned
	}
	allComplete := true
	anyStarted := false
	for _, status := range children {
		if status != complete {
			allComplete = false
		}
		if status == complete || status == inProgress {
			anyStarted = true
		}
	}
	switch {
	case allComplete:
		return complete
	case anyStarted:
		return inProgress
	default:
		return planned
	}
}
