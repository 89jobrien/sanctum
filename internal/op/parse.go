package op

import (
	"encoding/json"
	"fmt"
)

type opAccountJSON struct {
	URL       string `json:"url"`
	Shorthand string `json:"shorthand"`
}

func parseAccountJSON(data []byte) ([]Account, error) {
	var raw []opAccountJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("op: parse account list: %w", err)
	}
	accounts := make([]Account, 0, len(raw))
	for _, r := range raw {
		accounts = append(accounts, Account{URL: r.URL, Shorthand: r.Shorthand})
	}
	return accounts, nil
}
