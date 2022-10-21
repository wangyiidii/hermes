package message

type Type = int32

const (
	OnlineType     Type = iota // 上线信息
	ClientInfoType             // 客户端信息
	ClipboardType              // 粘贴板信息
)
