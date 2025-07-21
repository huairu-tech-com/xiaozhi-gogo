package asr

import "time"

type AsrResponse struct {
	IsFinish bool
	Success  bool  // Indicates if the ASR request was successful
	Err      error // Error message if the request failed
	Text     string
}

type AsrService interface {
	SendAudio(pcm []byte, isLastFrame bool, timeout time.Duration) error
	SetResponseCh(chan<- *AsrResponse)
	Close() error
}
