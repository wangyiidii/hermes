package message

type Type = int32

const (
	ClientInfoType Type = iota // 客户端信息
	ClipboardType              // 粘贴板信息
)
