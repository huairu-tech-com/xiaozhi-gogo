package cosyvoice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
)

const (
	TTSModel     = "FunAudioLLM/CosyVoice2-0.5B"
	DefaultVoice = "benjamin"
)

var (
	VoiceList = []string{
		"alex",
		"anna",
		"bella",
		"benjamin",
		"charles",
		"clarie",
		"diana",
		"david",
	}
)

type Tts struct {
	apiKey  string
	baseURL string
	voice   string

	client *openai.Client
}

func NewTts(apiKey, baseURL, voice string) *Tts {
	t := &Tts{
		apiKey:  apiKey,
		baseURL: "https://api.siliconflow.cn/v1/audio/speech",
		voice:   fmt.Sprintf("%s:%s", TTSModel, DefaultVoice),
	}

	if lo.Contains(VoiceList, voice) {
		t.voice = fmt.Sprintf("%s:%s", TTSModel, voice)
	}

	config := openai.DefaultConfig(t.apiKey)
	config.BaseURL = t.baseURL
	t.client = openai.NewClientWithConfig(config)

	return t
}

func (t *Tts) GenerateAudio(ctx context.Context, text string, speed float32) ([]byte, error) {
	if t.client == nil {
		return nil, errors.New("TTS client is not initialized")
	}

	data := map[string]interface{}{
		"model":           TTSModel,
		"input":           text,
		"voice":           t.voice,
		"response_format": "pcm",
		"sample_rate":     44100,
		"stream":          true,
		"gain":            0.0,
		"speed":           speed,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal TTS request data")
	}

	req, _ := http.NewRequest(
		http.MethodPost,
		t.baseURL,
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create TTS request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS request failed with status code: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read TTS response body")
	}

	return responseBody, nil
}
