package src

import (
	"context"
	"fmt"

	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/tts"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/tts/cosyvoice"

	opus "github.com/qrtc/opus-go"
)

type TtsProcessor struct {
	ctx       context.Context         // context for managing cancellation and timeouts
	ttsConfig *config.CosyVoiceConfig // TTS configuration

	ttsSrv tts.TTS // TTS service interface
}

func NewTtsProcessor(
	ctx context.Context,
	ttsConfig *config.CosyVoiceConfig, // TTS configuration
) *TtsProcessor {
	t := &TtsProcessor{
		ctx:       ctx,
		ttsConfig: ttsConfig, // Extracting CosyVoice configuration
	}

	t.ttsSrv = cosyvoice.NewTts(ttsConfig.ApiKey,
		ttsConfig.BaseUrl,
		ttsConfig.Voice)

	return t
}

func (t *TtsProcessor) Push(text string) ([][]byte, error) {
	speed := 1
	pcm, err := t.ttsSrv.GenerateAudio(t.ctx, text, (float32)(speed))
	if err != nil {
		return nil, err
	}

	return t.pcmToOpusData([][]byte{pcm}, 16000, 1) // 假设采样率为16000Hz，单声道
}

func (t *TtsProcessor) pcmToOpusData(pcmSlices [][]byte, sampleRate int, channels int) ([][]byte, error) {
	if len(pcmSlices) == 0 {
		return nil, fmt.Errorf("PCM数据切片为空")
	}

	// 检查采样率是否支持
	supportedRates := map[int]bool{8000: true, 12000: true, 16000: true, 24000: true, 48000: true}
	if !supportedRates[sampleRate] {
		return nil, fmt.Errorf("采样率 %dHz 不被Opus支持，仅支持8000/12000/16000/24000/48000Hz", sampleRate)
	}

	// 创建Opus编码器
	encoder, err := opus.CreateOpusEncoder(&opus.OpusEncoderConfig{
		SampleRate:    sampleRate,
		MaxChannels:   channels,
		Application:   opus.AppVoIP,
		FrameDuration: opus.Framesize60Ms, // 使用60ms帧长
	})
	if err != nil {
		return nil, fmt.Errorf("创建Opus编码器失败: %v", err)
	}
	defer encoder.Close()

	// 所有编码后的Opus数据包
	var allOpusPackets [][]byte

	// 计算每帧样本数 (60ms帧)
	samplesPerFrame := (sampleRate * 60) / 1000 // 60ms帧
	// 每个样本的字节数 (16位 = 2字节)
	bytesPerSample := 2 * channels
	// 每帧字节数
	bytesPerFrame := samplesPerFrame * bytesPerSample

	for _, pcmSlice := range pcmSlices {
		if len(pcmSlice) == 0 {
			continue
		}

		// 确保PCM数据长度是偶数
		if len(pcmSlice)%2 != 0 {
			pcmSlice = pcmSlice[:len(pcmSlice)-1] // 截断最后一个字节
			if len(pcmSlice) == 0 {
				continue
			}
		}

		// 计算这个PCM片段可以分成多少帧
		numFrames := len(pcmSlice) / bytesPerFrame
		if len(pcmSlice)%bytesPerFrame != 0 {
			numFrames++ // 如果有剩余数据，额外增加一帧
		}

		// 逐帧处理PCM数据
		for frameIdx := 0; frameIdx < numFrames; frameIdx++ {
			frameStart := frameIdx * bytesPerFrame
			frameEnd := frameStart + bytesPerFrame

			// 确保不越界
			if frameEnd > len(pcmSlice) {
				frameEnd = len(pcmSlice)
			}

			// 当前帧的PCM数据
			framePcm := pcmSlice[frameStart:frameEnd]

			// 如果最后一帧数据不足，需要填充静音数据到完整帧大小
			if len(framePcm) < bytesPerFrame {
				paddedFrame := make([]byte, bytesPerFrame)
				copy(paddedFrame, framePcm)
				framePcm = paddedFrame
			}

			// 分配输出缓冲区 (Opus编码后的数据通常比PCM小)
			outBuf := make([]byte, len(framePcm))

			// 编码这一帧PCM数据到Opus
			n, err := encoder.Encode(framePcm, outBuf)
			if err != nil {
				continue // 跳过这一帧，继续处理下一帧
			}

			if n == 0 {
				continue // 跳过空帧
			}

			// 将编码后的Opus数据添加到结果集
			allOpusPackets = append(allOpusPackets, outBuf[:n])
		}
	}

	if len(allOpusPackets) == 0 {
		return nil, fmt.Errorf("所有PCM切片编码后为空")
	}

	return allOpusPackets, nil

}
