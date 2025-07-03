package doubao

import (
	"context"
	"os"
	"testing"

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
