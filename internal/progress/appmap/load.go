package appmap

import (
	"encoding/json"
	"os"
)

func LoadAppMap(path string) (*AppMap, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var appMap AppMap
	if err := json.Unmarshal(body, &appMap); err != nil {
		return nil, err
	}
	return &appMap, nil
}
