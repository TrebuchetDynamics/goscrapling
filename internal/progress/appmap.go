package progress

import appmapmodel "github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap"

type AppMap = appmapmodel.AppMap
type AppMapMeta = appmapmodel.AppMapMeta
type AppMapEntry = appmapmodel.AppMapEntry
type AppMapRef = appmapmodel.AppMapRef

func LoadAppMap(path string) (*AppMap, error) {
	return appmapmodel.LoadAppMap(path)
}

func ValidateAppMap(m *AppMap) error {
	return appmapmodel.ValidateAppMap(m)
}

func ValidateAppMapCoverage(repoRoot string, m *AppMap) error {
	return appmapmodel.ValidateAppMapCoverage(repoRoot, m)
}

func ValidateAppMapReferences(repoRoot string, m *AppMap, p *Progress) error {
	return appmapmodel.ValidateAppMapReferences(repoRoot, m, p)
}

func RenderAppMapMarkdown(m *AppMap) string {
	return appmapmodel.RenderAppMapMarkdown(m)
}
