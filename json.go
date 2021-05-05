package cosmovisor

import "encoding/json"

func isJSONLog(s string) bool {
	return s[0] == '{' && s[len(s)-1] == '}'
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
