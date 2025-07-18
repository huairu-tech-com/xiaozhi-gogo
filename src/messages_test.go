package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var helloRequest = []byte(`
{
	"type": "hello",
	"version": 1,
	"transport": "websocket",
	"audio_params": {
		"format": "opus",
		"sample_rate": 16000,
		"channels": 1,
		"frame_duration": 60
	}
}
`)

var IotDescribeRaw = []byte(`
{
    "session_id": "<会话ID>",
    "type": "iot",
    "descriptors": "<设备描述JSON>"
}
`)

func TestMetaMessageTypeRecognize(t *testing.T) {
	m, err := MessageFromBytes[MetaMessage](helloRequest)
	if err != nil {
		t.Fatalf("failed to parse message: %v", err)
	}

	assert.Equal(t, MessageTypeHello, m.MessageType(), "expected message type to be Hello")
}

func TestHelloMessageRecognize(t *testing.T) {
	m, err := MessageFromBytes[Hello](helloRequest)
	if err != nil {
		t.Fatalf("failed to parse UpHello message: %v", err)
	}

	assert.NotNil(t, m, "expected UpHello message to be parsed successfully")
	assert.Equal(t, int32(1), m.Version, "expected version to be 1")
	assert.Equal(t, TransportType("websocket"), m.Transport, "expected transport to be websocket")
	assert.Equal(t, "opus", m.AudioParams.Format, "expected audio format to be opus")
	assert.Equal(t, int32(16000), m.AudioParams.SampleRate, "expected sample rate to be 16000")
	assert.Equal(t, int32(1), m.AudioParams.Channels, "expected channels to be 1")
	assert.Equal(t, int32(60), m.AudioParams.FrameDuration, "expected frame duration to be 60")
	assert.Equal(t, string(MessageTypeHello), m.MetaMessage.Type, "expected frame duration to be 60")
}

func TestIotDescribeMessageRecognize(t *testing.T) {
	m, err := MessageFromBytes[IotDescribe](IotDescribeRaw)
	if err != nil {
		t.Fatalf("failed to parse UpIotDescribe message: %v", err)
	}

	assert.NotNil(t, m, "expected UpIotDescribe message to be parsed successfully")
	assert.Equal(t, MessageTypeIOTDescribe, m.MessageType(), "expected message type to be IOTDescribe")
	assert.NotEmpty(t, m.MetaMessage.IotDescribe, "expected descriptors to be non-empty")
}
