package jsonfile

import (
	"encoding/json"
	"os"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/schema"
)

func Load(path string) (*schema.AppMap, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var appMap schema.AppMap
	if err := json.Unmarshal(body, &appMap); err != nil {
		return nil, err
	}
	return &appMap, nil
}
