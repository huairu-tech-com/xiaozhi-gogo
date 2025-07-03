package doubao

// https://www.volcengine.com/docs/6561/1354869?lang=zh#demo
const (
	DoubaoStreamAsrEndpoint = "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel"
)

const (
	DoubaoModelDuration   = "volc.bigasr.sauc.duration"   // 小时版
	DoubaoModelConcurrent = "volc.bigasr.sauc.concurrent" // 并发版
)

// 协议版本， 当前仅有一个
const ProtocolVersion = byte(0b0001 & 0xFF)

// 包类型
const (
	MessageTypeFullClientRequest  = byte(0b0001 & 0xFF)
	MessageTypeAudioOnlyRequest   = byte(0b0010 & 0xFF)
	MessageTypeFullServerResponse = byte(0b1001 & 0xFF)
	MessageTypeServerError        = byte(0b1111 & 0xFF)
)

// Serialization Method
const (
	SerializationMethodJson = byte(0b0000 & 0xFF)
	SerializationMethodNone = byte(0b0001 & 0xFF)
)

// Message Compression
const (
	CompressionNone = byte(0b0000 & 0xFF)
	CompressionGzip = byte(0b0001 & 0xFF)
)

// Message audio request last flag
const (
	NoSequence     = byte(0b0000 & 0xFF) // 全部音频数据
	PosSequence    = byte(0b0001 & 0xFF) // 最后一个音频数据包
	NegWithSequnce = byte(0b0011 & 0xFF)
	NegSequence    = byte(0b0010 & 0xFF)
)

// Server response message type
type ServerMessageType string

const (
	// full server response message
	SeverMessageTypeFullServerResponse ServerMessageType = "full_server_response" // 全部音频数据包
	// indicates an error message from the server
	ServerMessageTypeServerError ServerMessageType = "server_error" // 错误信息
)

const (
	ServerErrorSuccess        = 20000000 // 成功
	ServerErrorInvalidRequest = 40000001 // 请求参数无效
	ServerErrorEmptyAudio     = 45000002 // 空音频
	ServerErrorTimeout        = 45000081 // 音频过短
	ServerErrorAudioFormat    = 45000151 // 音频格式不正确
	SererErrorInternalError   = 55000031 // 服务内部处理错误
)
