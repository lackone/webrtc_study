package msg

type Close struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}
