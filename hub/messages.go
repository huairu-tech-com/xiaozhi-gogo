package hub

import (
	"github.com/bytedance/sonic"
)

type MessageType string

const (
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
	AudioModeAuto     AudioMode = "auto"
	AudioModeManual   AudioMode = "manual"
	AudioModeRealtime AudioMode = "realtime"
)

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

// Hello
type Hello struct {
	MetaMessage

	Version     int32  `json:"version"`   // 版本号
	Transport   string `json:"transport"` // 传输方式
	AudioParams struct {
		Format        string `json:"format"`         // 音频格式
		SampleRate    int32  `json:"sample_rate"`    // 采样率
		Channels      int32  `json:"channels"`       // 声道数
		FrameDuration int32  `json:"frame_duration"` // 帧时长，单位为毫秒
	} `json:"audio_params"` // 音频参数
}

// 开始监听
type ListenStart struct {
	MetaMessage
	Text string `json:"text"` // 唤醒词文本
}

// 停止监听
type ListenStop struct {
}

// 唤醒词检测
type ListenDetect struct {
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
