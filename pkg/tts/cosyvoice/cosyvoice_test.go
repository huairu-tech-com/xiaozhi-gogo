package cosyvoice

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildSiliconClient(apiKey, baseURL, voice string) *Tts {
	t := NewTts(apiKey, baseURL, voice)
	return t
}

func TestCreateClient(t *testing.T) {
	ins := buildSiliconClient(
		os.Getenv("SILICONFLOW_API_KEY"),
		os.Getenv("SILICONFLOW_BASE_URL"),
		os.Getenv("SILICONFLOW_VOICE"))

	assert.NotNil(t, ins, "Silicon TTS client should not be nil")
}

func TestGenerateAudio(t *testing.T) {
	ins := buildSiliconClient(
		os.Getenv("SILICONFLOW_API_KEY"),
		os.Getenv("SILICONFLOW_BASE_URL"),
		os.Getenv("SILICONFLOW_VOICE"))

	assert.NotNil(t, ins, "Silicon TTS client should not be nil")

	buffer, err := ins.GenerateAudio(context.Background(), "你好， 这是来自小智发出的声音", 1.2)
	assert.NoError(t, err, "Silicon TTS should not return an error")
	assert.NotEmpty(t, buffer, "Silicon TTS should return a non-empty audio response")

	tmpFile, err := os.CreateTemp("", "example.*.wav")
	assert.NoError(t, err, "Failed to create temporary file for audio output")
	assert.NotNil(t, tmpFile, "Temporary file for audio output should not be nil")

	ioutil.WriteFile(tmpFile.Name(), buffer, 0644)
	t.Logf("Audio written to temporary file: %s", tmpFile.Name())
	defer os.Remove(tmpFile.Name()) // Clean up

}
