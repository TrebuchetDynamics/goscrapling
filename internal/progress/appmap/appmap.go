package appmap

import (
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/jsonfile"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/markdown"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/schema"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/validation"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model"
)

type AppMap = schema.AppMap
type AppMapMeta = schema.AppMapMeta
type AppMapEntry = schema.AppMapEntry
type AppMapRef = schema.AppMapRef

func LoadAppMap(path string) (*AppMap, error) {
	return jsonfile.Load(path)
}

func ValidateAppMap(m *AppMap) error {
	return validation.Validate(m)
}

func ValidateAppMapCoverage(repoRoot string, m *AppMap) error {
	return validation.ValidateCoverage(repoRoot, m)
}

func ValidateAppMapReferences(repoRoot string, m *AppMap, p *model.Progress) error {
	return validation.ValidateReferences(repoRoot, m, p)
}

func RenderAppMapMarkdown(m *AppMap) string {
	return markdown.Render(m)
}
