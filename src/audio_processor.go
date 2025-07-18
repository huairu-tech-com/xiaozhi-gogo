package src

import (
	"fmt"
	"sync"
	"time"

	"github.com/baabaaox/go-webrtcvad"
	"github.com/pkg/errors"
	opus "github.com/qrtc/opus-go"
)

var (
	SampleRate = 16000
	BitRate    = 16
	VadLen     = 320
)

const MaxFrameLen = 100
const FrameSize = 320

var (
	ErrNoAudioFrame     = errors.New("no audio frame available")
	ErrAudioLenNotAlign = errors.New("audio length is not aligned with frame size")
)

type AudioProcessor struct {
	lock                 sync.Mutex
	frames               [][]byte
	tempFrame            []byte
	previousFrameActivte bool
	seq                  int
	longPausing          bool
	lastVoiceDetectedAt  time.Time

	vadPositiveOnly bool
	vadInst         webrtcvad.VadInst
	opusDecoder     *opus.OpusDecoder
}

func NewAudioProcessor(vadPositiveOnly bool) (*AudioProcessor, error) {
	ab := &AudioProcessor{
		frames:               make([][]byte, 0),
		tempFrame:            make([]byte, 0),
		previousFrameActivte: false,
		lock:                 sync.Mutex{},
		vadPositiveOnly:      vadPositiveOnly,
		lastVoiceDetectedAt:  time.Now().Add(time.Hour * 1000),
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

func (ab *AudioProcessor) PushOpus(bytes []byte) error {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	outbuf := make([]byte, 4096)
	n, err := ab.opusDecoder.Decode(bytes, outbuf)
	if err != nil {
		return err
	}

	if n != 1920 {
		panic(fmt.Sprintf("opus decode error, expected 1920 bytes, got %d bytes", n))
	}

	vadPositiveCount := 0
	for i := 0; i < n; i += VadLen {
		chunkActive, err := webrtcvad.Process(ab.vadInst, SampleRate, outbuf[i:i+VadLen], VadLen)
		if err != nil {
			return err
		}

		if chunkActive {
			vadPositiveCount += 1
		}
	}

	if vadPositiveCount > 3 {
		if !ab.previousFrameActivte {
			ab.seq += 1
			ab.frames = append(ab.frames, ab.tempFrame)
		}

		ab.seq += 1
		ab.longPausing = false
		ab.frames = append(ab.frames, outbuf[:n])
		ab.lastVoiceDetectedAt = time.Now()
	}

	ab.previousFrameActivte = vadPositiveCount > 3
	ab.tempFrame = make([]byte, 0, n)
	copy(ab.tempFrame[:], outbuf[:n])

	// detect long pause in conversation state
	if !ab.longPausing && ab.detectLongPause() {
		ab.seq = 0
		ab.longPausing = true
		ab.frames = append(ab.frames, ab.tempFrame) // mark the end of a conversation
	}

	return nil
}

// marks the end of a conversation
func (ab *AudioProcessor) detectLongPause() bool {
	return time.Now().After(ab.lastVoiceDetectedAt.Add(2 * time.Second))
}

func (ab *AudioProcessor) PopPCMWithVoice() (bytes []byte, seq int, isLast bool, err error) {
	if len(ab.frames) == 0 {
		return nil, ab.seq, ab.longPausing, ErrNoAudioFrame
	}

	ab.lock.Lock()
	defer ab.lock.Unlock()

	firstFrame := ab.frames[0]
	ab.frames = ab.frames[1:]

	return firstFrame, ab.seq, ab.longPausing, nil
}

func (ab *AudioProcessor) IsEmpty() bool {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	return len(ab.frames) == 0
}

func (ab *AudioProcessor) Size() int {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	return len(ab.frames)
}

func (ab *AudioProcessor) Clear() {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	ab.frames = make([][]byte, 0)
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
