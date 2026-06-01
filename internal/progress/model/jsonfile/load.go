package jsonfile

import (
	"encoding/json"
	"os"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema"
)

func Load(path string) (*schema.Progress, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var progress schema.Progress
	if err := json.Unmarshal(body, &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}
