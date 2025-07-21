package src

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/baabaaox/go-webrtcvad"
	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr/doubao"
	"github.com/pkg/errors"
	opus "github.com/qrtc/opus-go"
)

var (
	SampleRate = 16000
	BitRate    = 16
	// 20ms of audio at 16kHz is 320 bytes
	VadBytesLen = 320
)

const MaxFrameLen = 100
const FrameSize = 320

var (
	ErrNoAudioFrame     = errors.New("no audio frame available")
	ErrAudioLenNotAlign = errors.New("audio length is not aligned with frame size")
)

type audio struct {
	hasVoice bool
	frame    []byte
	isLast   bool
}

type vadAudioFilter struct {
	threshold int // threshold for voice activity detection
	audios    []audio

	consecutiveVoiceCount int  // count of consecutive frames with voice activity
	preAudioHasVoice      bool // previous frame's voice activity status
	inVoice               bool // whether currently in a voice segment
}

func NewVADAudioFilter(threshold int) *vadAudioFilter {
	return &vadAudioFilter{
		threshold: threshold,
		audios:    make([]audio, 0),

		consecutiveVoiceCount: 0,
		preAudioHasVoice:      false,
	}
}

func (filter *vadAudioFilter) Feed(hasVoice bool, audioFrame []byte) []audio {
	a := audio{
		hasVoice: hasVoice,
		frame:    audioFrame,
		isLast:   filter.preAudioHasVoice && !hasVoice,
	}

	// this is the first frame
	if !filter.preAudioHasVoice && hasVoice {
		filter.audios = make([]audio, 0)
		filter.audios = append(filter.audios, a)
		filter.consecutiveVoiceCount = 1
	}

	if filter.preAudioHasVoice && hasVoice {
		filter.audios = append(filter.audios, a)
		filter.consecutiveVoiceCount += 1
	}

	if filter.preAudioHasVoice && !hasVoice {
		if filter.consecutiveVoiceCount < filter.threshold {
			filter.audios = make([]audio, 0)
		} else {
			filter.audios = append(filter.audios, a)
		}
	}

	filter.preAudioHasVoice = hasVoice
	if filter.consecutiveVoiceCount >= filter.threshold {
		audios := filter.audios
		filter.audios = make([]audio, 0)
		return audios
	}

	return nil
}

type AudioProcessor struct {
	ctx  context.Context
	lock sync.Mutex
	// a queue of audio frames

	asrResponseCh chan<- *asr.AsrResponse

	preFrameHasVoice bool
	prevFrame        []byte // previous audio frame for VAD processing

	vadInst     webrtcvad.VadInst
	opusDecoder *opus.OpusDecoder

	asrService asr.AsrService // ASR service for processing audio frames
	asrConfig  *config.AsrConfig

	audioFilter *vadAudioFilter // filter for audio frames
}

func NewAudioProcessor(ctx context.Context,
	asrConfg *config.AsrConfig,
	asrResponseCh chan<- *asr.AsrResponse) (*AudioProcessor, error) {
	ab := &AudioProcessor{
		ctx:              ctx,
		lock:             sync.Mutex{},
		preFrameHasVoice: false,

		asrConfig:     asrConfg,
		asrResponseCh: asrResponseCh,
		audioFilter:   NewVADAudioFilter(3),
	}

	var err error
	ab.opusDecoder, err = opus.CreateOpusDecoder(&opus.OpusDecoderConfig{
		SampleRate:  SampleRate,
		MaxChannels: 1,
	})
	if err != nil {
		return nil, err
	}

	ab.vadInst = webrtcvad.Create()
	webrtcvad.Init(ab.vadInst)

	err = webrtcvad.SetMode(ab.vadInst, 0)
	if err != nil {
		return nil, err
	}

	return ab, nil
}

func (ab *AudioProcessor) PushOpus(opusBytes []byte) error {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	pcmBytes := make([]byte, 4096)
	n, err := ab.opusDecoder.Decode(opusBytes, pcmBytes)
	if err != nil {
		return err
	}

	// 120ms of audio at 16kHz is 1920 bytes
	if n != 1920 {
		panic(fmt.Sprintf("opus decode error, expected 1920 bytes, got %d bytes", n))
	}

	vadPositiveCountInFrame := 0
	for i := 0; i < n; i += VadBytesLen {
		chunkActive, err := webrtcvad.Process(ab.vadInst, SampleRate, pcmBytes[i:i+VadBytesLen], VadBytesLen)
		if err != nil {
			return err
		}

		if chunkActive {
			vadPositiveCountInFrame += 1
		}
	}

	for _, v := range ab.audioFilter.Feed(vadPositiveCountInFrame > 3, pcmBytes[:n]) {
		if err := ab.sendAudioToAsrService(v.frame, v.isLast); err != nil {
			return errors.Wrap(err, "send audio to ASR service failed")
		}
	}

	return nil
}

func (ab *AudioProcessor) sendAudioToAsrService(audioFrame []byte, isLastFrame bool) error {
	if ab.asrService == nil {
		var err error
		doubaoConfig := doubao.DefaultConfig()
		doubaoConfig.ApiKey = ab.asrConfig.Doubao.ApiKey
		doubaoConfig.AccessKey = ab.asrConfig.Doubao.AccessKey

		ab.asrService, err = doubao.DefaultDialer(ab.ctx, doubaoConfig)
		if err != nil {
			return err
		}

		ab.asrService.SetResponseCh(ab.asrResponseCh)
	}

	if err := ab.asrService.SendAudio(audioFrame, isLastFrame, time.Second); err != nil {
		return err
	}

	if isLastFrame {
		time.Sleep(200 * time.Millisecond) // wait for ASR service to process the last frame
		ab.asrService.Close()
		ab.asrService = nil
	}

	return nil
}

func (ab *AudioProcessor) Close() {
	if ab.vadInst != nil {
		webrtcvad.Free(ab.vadInst)
	}

	if ab.opusDecoder != nil {
		ab.opusDecoder.Close()
		ab.opusDecoder = nil
	}
}
