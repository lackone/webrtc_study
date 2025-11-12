package msg

type Heartbeat struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
