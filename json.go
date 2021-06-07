package cosmovisor

import (
	"encoding/json"
	"strings"
)

// isJSONLog returns whether the string contains json, and if so, returns that json.
func isJSONLog(s string) (string, bool) {
	beg := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if beg == end || beg > end {
		return "", false
	}

	return s[beg:end+1], true
}

type jsonLogMessage struct {
	Err     string `json:"err"`
	Message string `json:"message"`
}

func parseJSONLog(s string) (jsonLogMessage, error) {
	m := jsonLogMessage{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return jsonLogMessage{}, err
	}

	return m, nil
}
