package helper

import (
	"encoding/json"
)

func FilterJSON(model interface{}, fields []string) map[string]interface{} {
	data, _ := json.Marshal(model)
	var all map[string]interface{}
	_ = json.Unmarshal(data, &all)

	if fields == nil || len(fields) == 0 {
		return all
	}

	filtered := make(map[string]interface{})
	for _, f := range fields {
		if val, ok := all[f]; ok {
			filtered[f] = val
		}
	}
	return filtered
}
