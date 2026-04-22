package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed news_keywords.json
var keywordHintsFS []byte

func LoadKeywordHints() (map[string][]string, error) {
	var hints map[string][]string
	if err := json.Unmarshal(keywordHintsFS, &hints); err != nil {
		return nil, fmt.Errorf("unmarshal embedded keyword hints: %w", err)
	}
	return hints, nil
}
