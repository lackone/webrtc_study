package msg

type Msg struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
