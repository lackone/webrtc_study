package turn

import "encoding/json"

type TurnCreds struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Ttl      int      `json:"ttl"`
	Uris     []string `json:"uris"`
}

func (t *TurnCreds) Marshal() string {
	data, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(data)
}

func (t *TurnCreds) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), t)
}
