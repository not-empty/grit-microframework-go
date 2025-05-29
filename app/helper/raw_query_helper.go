package helper

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var (
	rawMu      sync.RWMutex
	rawQueries = make(map[string]map[string]string)
)

func RegisterRawQueries(table string, queries map[string]string) {
	rawMu.Lock()
	defer rawMu.Unlock()
	rawQueries[table] = queries
}

func GetRawQuery(table, name string) (string, bool) {
	rawMu.RLock()
	defer rawMu.RUnlock()
	qm, found := rawQueries[table]
	if !found {
		return "", false
	}
	sql, ok := qm[name]
	return sql, ok
}

var rawDenySubstr = []string{
	";",
	"--",
	"/*",
	"*/",
}

var rawDenyWords = []string{
	"drop",
	"alter",
	"truncate",
	"delete",
	"update",
	"insert",
	"create",
	"merge",
	"replace",
	"grant",
	"revoke",
	"commit",
	"rollback",
	"savepoint",
	"lock",
	"unlock",
	"exec",
	"call",
	"use",
	"set",
	"limit",
	"offset",
	"join",
}

var rawAllowList = []string{
	"select",
	"with",
}

func CheckRawQueryAllowed(query string) (bool, error) {
	lower := strings.ToLower(query)

	for _, bad := range rawDenySubstr {
		if strings.Contains(lower, bad) {
			return false, fmt.Errorf("forbidden substring in query: %s", bad)
		}
	}

	nonWordRE := regexp.MustCompile(`[^a-z0-9_]+`)

	tokens := nonWordRE.Split(lower, -1)
	for _, tok := range tokens {
		for _, bad := range rawDenyWords {
			if tok == bad {
				return false, fmt.Errorf("forbidden keyword in query: %s", bad)
			}
		}
	}

	trimmed := strings.TrimSpace(lower)
	for _, prefix := range rawAllowList {
		if strings.HasPrefix(trimmed, prefix) {
			return true, nil
		}
	}

	return false, fmt.Errorf("only %v queries are allowed", rawAllowList)
}

func ExtractRawParams(query string) []string {
	re := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	matches := re.FindAllStringSubmatch(query, -1)
	seen := map[string]bool{}
	var params []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			params = append(params, name)
		}
	}
	return params
}

func ValidateRawParams(query string, params map[string]any) error {
	required := ExtractRawParams(query)
	for _, name := range required {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing parameter: %s", name)
		}
	}
	reqSet := map[string]bool{}
	for _, name := range required {
		reqSet[name] = true
	}
	for k := range params {
		if !reqSet[k] {
			return fmt.Errorf("unexpected parameter: %s", k)
		}
	}
	return nil
}

func PrepareRawQuery(query string, params map[string]any) (string, []interface{}) {
	re := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	matches := re.FindAllStringSubmatch(query, -1)

	args := make([]interface{}, 0, len(matches))
	for _, match := range matches {
		key := match[1]
		query = strings.Replace(query, ":"+key, "?", 1)
		args = append(args, params[key])
	}

	query = fmt.Sprintf("%s LIMIT %d", query, 25)
	return query, args
}
