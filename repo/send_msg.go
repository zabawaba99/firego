package repo

type sendMsg struct {
	Action action      `json:"a"`
	ReqNum float64     `json:"r"`
	Data   interface{} `json:"b"`
}
