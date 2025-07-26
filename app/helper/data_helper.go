package helper

import (
	"encoding/json"
	"strings"
	"time"
)

func FilterJSON(model interface{}, fields []string) map[string]interface{} {
	data, _ := json.Marshal(model)
	var all map[string]interface{}
	_ = json.Unmarshal(data, &all)

	if len(fields) == 0 {
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

func IsEmptyValue(v interface{}) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		return val == ""
	case int:
		return val == 0
	case int64:
		return val == 0
	case float64:
		return val == 0
	case *time.Time:
		return val == nil
	case *JSONTime:
		return val == nil
	default:
		return false
	}
}

func FilterOutDefaulted(
	cols []string,
	vals []interface{},
	defaultCols []string,
) (filteredCols []string, filteredVals []interface{}) {
	defaultSet := make(map[string]struct{}, len(defaultCols))
	for _, c := range defaultCols {
		defaultSet[c] = struct{}{}
	}

	for i, col := range cols {
		val := vals[i]
		if _, isDefaultable := defaultSet[col]; isDefaultable && IsEmptyValue(val) {
			continue
		}
		filteredCols = append(filteredCols, col)
		filteredVals = append(filteredVals, val)
	}
	return filteredCols, filteredVals
}

func BuildRowTokens(
	allCols []string,
	vals []interface{},
	defaultCols []string,
) (rowSQL string, argsOut []interface{}) {
	defaultSet := make(map[string]struct{}, len(defaultCols))
	for _, dc := range defaultCols {
		defaultSet[dc] = struct{}{}
	}

	tokens := make([]string, len(allCols))
	for i, col := range allCols {
		v := vals[i]
		if _, isDefault := defaultSet[col]; isDefault && IsEmptyValue(v) {
			tokens[i] = "DEFAULT"
		} else {
			tokens[i] = "?"
			argsOut = append(argsOut, v)
		}
	}
	rowSQL = "(" + strings.Join(tokens, ", ") + ")"
	return rowSQL, argsOut
}
