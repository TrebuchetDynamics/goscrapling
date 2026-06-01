package progress

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/surfaces"

func RenderBuilderLoopHandoff(p *Progress) string {
	return surfaces.RenderBuilderLoopHandoff(p)
}

func RenderAgentQueue(p *Progress) string {
	return surfaces.RenderAgentQueue(p)
}

func RenderNextSlices(p *Progress) string {
	return surfaces.RenderNextSlices(p)
}

func RenderBlockedSlices(p *Progress) string {
	return surfaces.RenderBlockedSlices(p)
}

func RenderUmbrellaCleanup(p *Progress) string {
	return surfaces.RenderUmbrellaCleanup(p)
}

func ReplaceMarker(input, kind, body string) (string, error) {
	return surfaces.ReplaceMarker(input, kind, body)
}
