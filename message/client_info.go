package message

type ClientInfo struct {
	Name string `json:"name"`
	Os   string `json:"os"`
	Arch string `json:"arch"`
}
