package asr

type AsrResponse struct {
	Text string
}

type AsrService interface {
	SendAudio(pcm []byte, seq int32, isLastFrame bool) error
	ResponseCh() chan *AsrResponse
	Close() error
}
