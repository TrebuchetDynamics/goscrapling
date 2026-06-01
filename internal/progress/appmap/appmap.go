package appmap

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/model"

type AppMap struct {
	Meta    AppMapMeta    `json:"meta"`
	Entries []AppMapEntry `json:"entries"`
}

type AppMapMeta struct {
	Version           string             `json:"version"`
	Upstream          model.UpstreamMeta `json:"upstream,omitempty"`
	GeneratedMarkdown string             `json:"generated_markdown,omitempty"`
	Py2ManyProbeDir   string             `json:"py2many_probe_dir,omitempty"`
}

type AppMapEntry struct {
	ID                     string      `json:"id"`
	Title                  string      `json:"title"`
	Upstream               []AppMapRef `json:"upstream"`
	FeatureAnchor          string      `json:"feature_anchor,omitempty"`
	BehaviorAtoms          []string    `json:"behavior_atoms,omitempty"`
	GoTarget               string      `json:"go_target,omitempty"`
	ProgressRows           []string    `json:"progress_rows,omitempty"`
	CoverageStatus         string      `json:"coverage_status"`
	TranslationSuitability string      `json:"translation_suitability"`
	StaticReferencePaths   []string    `json:"static_reference_paths,omitempty"`
	Notes                  []string    `json:"notes,omitempty"`
}

type AppMapRef struct {
	Ref     string   `json:"ref"`
	Kind    string   `json:"kind"`
	Symbols []string `json:"symbols,omitempty"`
}
