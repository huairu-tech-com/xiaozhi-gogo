package tts

import "context"

type TTSResponse struct {
	IsStart bool   `json:"is_start"` // Indicates if this is the start of a TTS response
	IsEnd   bool   `json:"is_end"`   // Indicates if this is the end of a TTS response
	Text    string `json:"text"`     // The text to be spoken
	Audio   []byte `json:"audio"`    // The audio data in bytes
	Err     error  `json:"-"`        // Error if any occurred during processing
}

type TTS interface {
	GenerateAudio(ctx context.Context, text string, speed float32) ([]byte, error)
}
