package asr

import "time"

type AsrResponse struct {
	Text string
}

type AsrService interface {
	SendAudio(pcm []byte, seq int, isLastFrame bool, timeout time.Duration) error
	ResponseCh() chan *AsrResponse
	Close() error
}
