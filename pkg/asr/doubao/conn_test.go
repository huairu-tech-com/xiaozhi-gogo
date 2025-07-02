package doubao

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sampleRate = 44100
	frequency  = 440.0
)

func buildConnection() *AsrDoubaoConn {
	config := &AsrDoubaoConfig{
		Host:       "openspeech.bytedance.com",
		ResourceId: DoubaoModelDuration,
		Model:      "bigmodel",
	}

	config.ApiKey = os.Getenv("DOUBAO_API_KEY")
	config.AccessKey = os.Getenv("DOUBAO_ACCESS_KEY")

	conn, err := DefaultDialer(context.Background(), config)
	if err != nil {
		panic(err)
	}

	return conn
}

func TestDefaultDialer(t *testing.T) {
	conn := buildConnection()
	if conn == nil {
		t.Fatal("expected a valid connection, got nil")
	}

	defer conn.Close()

	assert.NotNil(t, conn, "expected a valid connection, got nil")
}

func TestSendAudio(t *testing.T) {
	conn := buildConnection()
	assert.NotNil(t, conn, "expected a valid connection, got nil")
	defer conn.Close()
	buf := make([]byte, 44100*2) // 1 second buffer (16-bit = 2 bytes)

	for k := 0; k < 5; k++ { // Send 5 chunks of audio
		for i := 0; i < len(buf)/2; i++ {
			t := float64(i) / float64(sampleRate)
			val := math.Sin(2.0 * math.Pi * frequency * t)
			// Convert to 16-bit PCM
			pcmVal := int16(val * 32767)
			// Little endian
			buf[i*2] = byte(pcmVal)
			buf[i*2+1] = byte(pcmVal >> 8)
		}

		if err := conn.SendAudio(buf); err != nil {
			assert.Fail(t, "failed to send audio: %v", err)
		}
		time.Sleep(1 * time.Second) // Simulate 1 second of audio

		t.Logf("Sent audio chunk of size: %d bytes", len(buf))
	}
}
