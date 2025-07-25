package src

import (
	"github.com/bytedance/sonic"
)

type MessageType string

const (
	MessageTypeNone             MessageType = "none"
	MessageTypeRawAudio         MessageType = "raw_audio"
	MessageTypeHello            MessageType = "hello"
	MessageTypeListenStart      MessageType = "listen_start"
	MessageTypeListenStop       MessageType = "listen_stop"
	MessageTypeListenDetect     MessageType = "listen_detect"
	MessageTypeTTSStart         MessageType = "tts_start"
	MessageTypeTTSStop          MessageType = "tts_stop"
	MessageTypeTTSSentenceStart MessageType = "tts_sentence_start"
	MessageTypeAbort            MessageType = "abort"
	MessageTypeIOTDescribe      MessageType = "iot_describe"
	MessageTypeIOTStates        MessageType = "iot_states"
	MessageTypeLlm              MessageType = "llm"
)

type AudioMode string

const (
	AudioModeNone     AudioMode = "none"
	AudioModeAuto     AudioMode = "auto"
	AudioModeManual   AudioMode = "manual"
	AudioModeRealtime AudioMode = "realtime"
)

type TransportType string

const (
	TransportTypeWebsocket TransportType = "websocket"
	TransportTypeMQTT      TransportType = "mqtt"
)

type BinaryProtocol2 struct {
	Version     uint16 `json:"version"`      // 协议版本
	Type        uint16 `json:"type"`         // 消息类型
	Reserved    uint32 `json:"reserved"`     // 保留字段
	Timestamp   uint32 `json:"timestamp"`    // 时间戳，单位为毫秒
	PayloadSize uint32 `json:"payload_size"` // 有效载荷大小
	Payload     []byte `json:"payload"`      // 有效载荷
}

type BinaryProtocol3 struct {
	Type        uint8  `json:"type"`         // 消息类型
	Reserved    uint8  `json:"reserved"`     // 保留字段
	PayloadSize uint16 `json:"payload_size"` // 有效载荷大小
	Payload     []byte `json:"payload"`      // 有效载荷
}

type MetaMessage struct {
	Type  string `json:"type"`
	State string `json:"state"`

	IotStates   string `json:"states,omitempty"`
	IotDescribe string `json:"descriptors,omitempty"`
}

func (m MetaMessage) MessageType() MessageType {
	if m.Type == "hello" {
		return MessageTypeHello
	}

	if m.Type == "listen" && m.State == "start" {
		return MessageTypeListenStart
	}

	if m.Type == "listen" && m.State == "stop" {
		return MessageTypeListenStop
	}

	if m.Type == "listen" && m.State == "detect" {
		return MessageTypeListenDetect
	}

	if m.Type == "tts" && m.State == "start" {
		return MessageTypeTTSStart
	}

	if m.Type == "tts" && m.State == "stop" {
		return MessageTypeTTSStop
	}

	if m.Type == "tts" && m.State == "sentence_start" {
		return MessageTypeTTSSentenceStart
	}

	if m.Type == "iot" && len(m.IotDescribe) != 0 {
		return MessageTypeIOTDescribe
	}

	if m.Type == "iot" && len(m.IotStates) != 0 {
		return MessageTypeIOTStates
	}

	if m.Type == "abort" {
		return MessageTypeAbort
	}

	panic("invalid message type: " + m.Type)
}
func MessageFromBytes[T any](raw []byte) (*T, error) {
	var msg T
	if err := sonic.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type HelloAudioParams struct {
	SampleRate    int32  `json:"sample_rate"`    // 采样率
	Format        string `json:"format"`         // 音频格式
	Channels      int32  `json:"channels"`       // 声道数
	FrameDuration int32  `json:"frame_duration"` // 帧时长，单位为毫秒
}

// Hello
type Hello struct {
	MetaMessage

	Version   int32         `json:"version"`   // 版本号
	Transport TransportType `json:"transport"` // 传输方式
	Features  struct {
		MCP bool `json:"mcp"`
	} `json:"features,omitempty"` // 特性
	AudioParams HelloAudioParams `json:"audio_params"` // 音频参数
}

type HelloResponse struct {
	Type        MessageType      `json:"type"`         // 消息类型
	SessionId   string           `json:"session_id"`   // 会话ID
	Transport   TransportType    `json:"transport"`    // 传输方式
	AudioParams HelloAudioParams `json:"audio_params"` // 音频参数
}

// 开始监听
type ListenStart struct {
	MetaMessage
	SessionId string    `json:"session_id"` // 会话ID
	Mode      AudioMode `json:"mode"`       // 音频模式
}

// 停止监听
type ListenStop struct {
}

// 唤醒词检测
type ListenDetect struct {
	MetaMessage

	SessionId string `json:"session_id"` // 会话ID
	Text      string `json:"text"`       // 检测到的唤醒词文本
}

// TTS开始
type TTSStart struct {
	MessageType

	Text string `json:"text"` // TTS文本
}

// TTS结束
type TTSEnd struct {
	MessageType

	Text string `json:"text"` // TTS文本
}

// TTS新句子开始
type TTSentenceStart struct {
	MessageType

	Text string `json:"text"` // 句子文本
}

type Abort struct {
	MessageType

	SessionId string `json:"session_id"` // 会话ID
	Reason    string `json:"reason"`     // 终止原因
}

type Llm struct {
	MetaMessage
	Emotion string `json:"emotion"` // 情感
}

type IotDescribe struct {
	MetaMessage
	SessionId string `json:"session_id"` // 会话ID
}

type IotStates struct {
	MetaMessage
	SEssionId string `json:"session_id"` // 会话ID
}
type DownHello struct {
}
