package cosmovisor

import "encoding/json"

func isJsonLog(s string) bool {
	return s[0] == '{' && s[len(s)-1] == '}'
}

type jsonLogMessage struct {
	Err     string `json:"err"`
	Message string `json:"message"`
}

func parseJsonLog(s string) (jsonLogMessage, error) {
	m := jsonLogMessage{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return jsonLogMessage{}, nil
	}

	return m, nil
}
