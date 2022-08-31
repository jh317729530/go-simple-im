package model

type Packet struct {
	Length    uint32      // 包长度，4位
	Version   uint32      // 包版本
	Operation uint32      // 操作
	Body      *PacketBody // 包内容
}

// 业务内容，待定义
type PacketBody struct {
}

const Version1 uint32 = 1

const (
	HeartBeat      uint32 = 10 // 心跳包
	HeartBeatReply uint32 = 11 // 心跳包回复
)

const (
	PacketLengthSize = 4
	VersionSize      = 4
	OperationSize    = 4

	VersionOffset   = 0
	OperationOffset = VersionOffset + VersionSize
	BodyOffset      = OperationOffset + OperationSize
)
