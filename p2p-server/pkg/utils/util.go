package utils

import "encoding/json"

func Marshal(d any) string {
	if data, err := json.Marshal(d); err != nil {
		return ""
	} else {
		return string(data)
	}
}

func Unmarshal(str string) map[string]any {
	var data map[string]any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return nil
	}
	return data
}
