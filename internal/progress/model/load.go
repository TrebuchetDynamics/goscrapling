package model

import (
	"encoding/json"
	"os"
)

func Load(path string) (*Progress, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var progress Progress
	if err := json.Unmarshal(body, &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}
