package entity

type Notification struct {
	Receiver string `json:"receiver"`
	Message  string `json:"notification"`
	Subject  string `json:"subject"`
}
