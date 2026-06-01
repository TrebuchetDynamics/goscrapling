package surfaces

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress"
)

// Write regenerates all marker-delimited builder-loop progress surfaces.
func Write(stdout io.Writer, root string, p *progress.Progress) error {
	for _, target := range []struct {
		path string
		kind string
		body string
	}{
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md", kind: "builder-loop-handoff", body: progress.RenderBuilderLoopHandoff(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/assignable/agent-queue.md", kind: "agent-queue", body: progress.RenderAgentQueue(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/assignable/next-slices.md", kind: "next-slices", body: progress.RenderNextSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked/blocked-slices.md", kind: "blocked-slices", body: progress.RenderBlockedSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md", kind: "agent-queue", body: progress.RenderAgentQueue(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md", kind: "next-slices", body: progress.RenderNextSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked-slices.md", kind: "blocked-slices", body: progress.RenderBlockedSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/cleanup/umbrella-cleanup.md", kind: "umbrella-cleanup", body: progress.RenderUmbrellaCleanup(p)},
	} {
		if err := rewriteMarker(filepath.Join(root, target.path), target.kind, target.body); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "progress: regenerated %s\n", target.path)
	}
	return nil
}

func rewriteMarker(path, kind, body string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	output, err := progress.ReplaceMarker(string(input), kind, body)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
