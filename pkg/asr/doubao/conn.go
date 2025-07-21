package doubao

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Header struct {
	ProtocolVersion     byte // 协议版本
	HeaderSize          byte // 头部大小
	MessageType         byte // 消息类型
	Flags               byte // 标志位
	SerializationMethod byte // 序列化方法
	Compression         byte // 压缩方法
	Reserved            byte // 保留字节，当前设置为 0
}

func (h Header) ToBytes() []byte {
	bytes := make([]byte, 4)
	bytes[0] = (h.ProtocolVersion << 4) | h.HeaderSize
	bytes[1] = (h.MessageType << 4) | h.Flags
	bytes[2] = (h.SerializationMethod << 4) | h.Compression
	bytes[3] = 0 // reserved byte, currently set to 0
	return bytes
}

type FullClientRequestPayload struct {
	User struct {
		Uid        string `json:"uid"`         // 用户 ID
		Did        string `json:"did"`         // 设备 ID
		Platform   string `json:"platform"`    // 平台
		SdkVersion string `json:"sdk_version"` // SDK 版本
		AppVersion string `json:"app_version"` // 应用版本
	} `json:"user"`
	Audio struct {
		Format   string `json:"format"`   // 音频格式 pcm(pcm_s16le) / wav(pcm_s16le) / ogg
		Codec    string `json:"codec"`    // 音频编解码器 raw / opus，默认为 raw(pcm) 。
		Rate     int    `json:"rate"`     // 音频采样率 16000
		Bits     int    `json:"bits"`     // 音频位数 16
		Channel  int    `json:"channel"`  // 音频通道数 1
		Language string `json:"language"` // 语言
	} `json:"audio"`

	Request struct {
		ModelName          string `json:"model_name"`           // 模型名称
		EnableItn          bool   `json:"enable_itn"`           // 是否开启 ITN
		EnablePunc         bool   `json:"enable_punc"`          // 是否开启标点符号
		EnableDdc          bool   `json:"enable_ddc"`           // 是否开启 DDC
		ShowUtterances     bool   `json:"show_utterances"`      // 是否返回分段结果
		ResultType         string `json:"result_type"`          // 结果类型 full / segment，默认为 full
		VadSegmentDuration int    `json:"vad_segment_duration"` // VAD 分段时长，单位为毫秒，默认 1000ms
		EndWindowSize      int    `json:"end_window_size"`      // 结束窗口大小，单位为毫秒，默认 1000ms
		FroceToSpeechTime  int    `json:"froce_to_speech_time"` // 强制语音时间，单位为毫秒，默认 0ms

		Corpus struct {
			BoostingTableId   string `json:"boosting_table_id"`   // 语料库 ID
			BosstingTableName string `json:"boosting_table_name"` // 语料库名称
			CorrectTableID    string `json:"correct_table_id"`    // 纠错表 ID
			Context           string `json:"context"`             // 上下文信息
		} `con:"corpus"`
	}
}

type FullServerResponsePacketPayload struct {
	Result struct {
		Text       string `json:"text"` // 识别结果文本
		Utterances []struct {
			Text      string `json:"text"`       // 分段结果文本
			StartTime int32  `json:"start_time"` // 分段开始时间，单位为毫秒
			EndTime   int32  `json:"end_time"`   // 分段结束时间，单位为毫秒
			Definite  bool   `json:"definite"`   // 是否为最终结果
			Words     []struct {
				BlankDuration int32  `json:"blank_duration"` // 单词间的空白时长，单位为毫秒
				EndTime       int32  `json:"end_time"`       // 单词结束时间，单位为毫秒
				StartTime     int32  `json:"start_time"`     // 单词开始时间，单位为毫秒
				Text          string `json:"text"`           // 单词文本
			} `json:"words"` // 分段结果中的单词信息
		}
	} `json:"result"`
}

// https://www.volcengine.com/docs/6561/1354869
type AsrDoubaoConn struct {
	ctx       context.Context // 上下文
	conn      *websocket.Conn
	connectId string // 客户端的 connect ID

	ttLogid string                  // 服务端的 trace ID
	respCh  chan<- *asr.AsrResponse // 响应通道
}

var DefaultDialer = func(ctx context.Context, cfg *AsrDoubaoConfig) (*AsrDoubaoConn, error) {
	headers := http.Header{}
	headers.Set("Host", cfg.Host)
	headers.Set("X-Api-App-Key", cfg.ApiKey)
	headers.Set("X-Api-Access-Key", cfg.AccessKey)
	headers.Set("X-Api-Resource-Id", DoubaoModelDuration)

	connectId := uuid.New().String()
	headers.Set("X-Api-Connect-Id", connectId)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: false},
	}

	log.Info().Msgf("Dialing Doubao ASR service at %s with connectId: %s", DoubaoStreamAsrEndpoint, connectId)
	conn, resp, err := dialer.DialContext(ctx, DoubaoStreamAsrEndpoint, headers)
	doubaoConn := &AsrDoubaoConn{
		ctx:       ctx,
		conn:      conn,
		connectId: connectId,
	}
	if err != nil {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Response Stats: %d\n", resp.StatusCode))
		sb.WriteString(fmt.Sprintf("Response Headers: %v\n", resp.Header))
		bytes, _ := io.ReadAll(resp.Body)

		sb.WriteString(fmt.Sprintf("Response Body: %s\n", string(bytes)))
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
		return nil, errors.Wrapf(err, "failed to dial websocket: %s", sb.String())
	}

	doubaoConn.ttLogid = resp.Header.Get("X-Tt-Logid")

	if err := doubaoConn.sendParameters(); err != nil {
		doubaoConn.Close()
		return nil, err
	}

	go func() {
		if err := doubaoConn.readLoop(); err != nil {
			log.Error().Err(err).Msg("AsrDoubaoConn read loop error")
			doubaoConn.Close()
		}
	}()

	return doubaoConn, err
}

func (conn *AsrDoubaoConn) SetResponseCh(ch chan<- *asr.AsrResponse) {
	conn.respCh = ch
}

func (conn *AsrDoubaoConn) String() string {
	var sb strings.Builder
	sb.WriteString("AsrDoubaoConn{")
	sb.WriteString("connectId: ")
	sb.WriteString(conn.connectId)
	sb.WriteString(", ttLogid: ")
	sb.WriteString(conn.ttLogid)
	sb.WriteString(", host: ")
	sb.WriteString("}")
	return sb.String()
}

func (conn *AsrDoubaoConn) Close() error {
	log.Info().Msgf("Closing AsrDoubaoConn: %s", conn.String())

	if conn.conn != nil {
		conn.conn.Close()
	}

	return nil
}

func (conn *AsrDoubaoConn) sendParameters() error {
	var err error
	var mt int
	var bytes []byte
	var n int
	var fullClientPacket []byte

	w, err := conn.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	fullClientPacket, err = conn.buildFullClientPacket()
	if err != nil {
		return err
	}

	n, err = w.Write(fullClientPacket)
	if err != nil {
		return err
	}

	if len(fullClientPacket) != n {
		return errors.New("failed to write full client packet")
	}
	w.Close()

	mt, bytes, err = conn.conn.ReadMessage()
	if err != nil {
		return err
	}

	if mt != websocket.BinaryMessage {
		return errors.New("expected binary message, got text message")
	}

	if len(bytes) < 4 {
		return errors.New("message too short, expected at least 4 bytes")
	}

	if parseMessageType(bytes[1]) == ServerMessageTypeServerError {
		_, err := conn.parseServerResponseError(bytes)
		if err != nil {
			return err
		}
	}

	if parseMessageType(bytes[1]) == SeverMessageTypeFullServerResponse {
		_, _, err := conn.parseFullServerResponse(bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn *AsrDoubaoConn) buildFullClientPacket() ([]byte, error) {
	header := conn.buildHeader(
		MessageTypeFullClientRequest,
		NoSequence,
		SerializationMethodJson,
		CompressionNone)

	var payload FullClientRequestPayload
	payload.User.Uid = "test_user"
	payload.User.Did = "test_device"
	payload.User.Platform = "Linux"
	payload.User.SdkVersion = "1.0.0"
	payload.User.AppVersion = "1.0.0"
	payload.Audio.Format = "pcm"
	payload.Audio.Codec = "raw"
	payload.Audio.Rate = 16000
	payload.Audio.Bits = 16
	payload.Audio.Channel = 1
	payload.Audio.Language = "zh-CN"
	payload.Request.ModelName = "bigmodel"
	payload.Request.EnableItn = true
	payload.Request.EnablePunc = true
	payload.Request.EnableDdc = false
	payload.Request.ResultType = "single"
	payload.Request.ShowUtterances = false

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	packetBuffer := bytes.Buffer{}
	binary.Write(&packetBuffer, binary.BigEndian, header.ToBytes())
	binary.Write(&packetBuffer, binary.BigEndian, uint32(len(payloadBytes)))
	binary.Write(&packetBuffer, binary.BigEndian, payloadBytes)

	return packetBuffer.Bytes(), nil
}

func (conn *AsrDoubaoConn) buildAudioOnlyRequestPacket(audioData []byte, isLastFrame bool) ([]byte, error) {
	packetIndicator := NoSequence
	if isLastFrame {
		packetIndicator = NegSequence
	}
	header := conn.buildHeader(MessageTypeAudioOnlyRequest,
		byte(packetIndicator),
		SerializationMethodNone,
		CompressionGzip)

	var buffer bytes.Buffer
	zipwriter := gzip.NewWriter(&buffer)
	if _, err := zipwriter.Write(audioData); err != nil {
		return nil, err
	}
	zipwriter.Close()

	zippedData := buffer.Bytes()
	packetBuffer := bytes.Buffer{}
	binary.Write(&packetBuffer, binary.BigEndian, header.ToBytes())
	binary.Write(&packetBuffer, binary.BigEndian, uint32(len(zippedData)))
	binary.Write(&packetBuffer, binary.BigEndian, zippedData)
	return packetBuffer.Bytes(), nil
}

func (conn *AsrDoubaoConn) readLoop() error {
	for {
		select {
		case <-conn.ctx.Done():
			return conn.ctx.Err()
		default:
			mt, bytes, err := conn.conn.ReadMessage()
			if err != nil {
				return err
			}

			if mt != websocket.BinaryMessage {
				return errors.New("expected binary message, got text message")
			}

			if len(bytes) < 4 {
				return errors.New("message too short, expected at least 4 bytes")
			}

			switch parseMessageType(bytes[1]) {
			case SeverMessageTypeFullServerResponse:
				header, payload, err := conn.parseFullServerResponse(bytes)
				if err != nil {
					return err
				}

				if conn.respCh != nil {
					conn.respCh <- &asr.AsrResponse{
						IsFinish: header.Flags&NegSequence != 0,
						Success:  true,
						Text:     payload.Result.Text,
						Err:      nil,
					}
				}

			case ServerMessageTypeServerError:
				if _, err := conn.parseServerResponseError(bytes); err != nil {
					if conn.respCh != nil {
						conn.respCh <- &asr.AsrResponse{
							IsFinish: true,
							Success:  false,
							Text:     "",
							Err:      err,
						}
					}

					return err
				}

			default:
				if conn.respCh != nil {
					conn.respCh <- &asr.AsrResponse{
						Success: false,
						Text:    "",
						Err:     errors.Errorf("unknown message type: %d", bytes[0]),
					}
				}

				return err
			}
		}
	}
}

func (conn *AsrDoubaoConn) SendAudio(pcm []byte,
	isLast bool,
	timeout time.Duration) error {
	var err error
	var data []byte
	err = conn.conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	data, err = conn.buildAudioOnlyRequestPacket(pcm, isLast)
	if err != nil {
		return err
	}

	if err := conn.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func parseMessageType(mt byte) ServerMessageType {
	if mt>>4&0x0F == MessageTypeFullServerResponse {
		return SeverMessageTypeFullServerResponse
	}

	if mt>>4&0x0F == MessageTypeServerError {
		return ServerMessageTypeServerError
	}

	return ServerMessageTypeUnknown
}

func (conn *AsrDoubaoConn) parseFullServerResponse(raw []byte) (*Header, *FullServerResponsePacketPayload, error) {
	// 4 header bytes + 4 payload size bytes + 4 sequence bytes
	if len(raw) < 12 {
		return nil, nil, errors.New("raw data too short to parse full server response")
	}

	var header Header
	r := bytes.NewReader(raw)
	b0, _ := r.ReadByte()
	protocolVersion := (b0 >> 4) & 0xFF
	headerSize := b0 & 0xFF
	header.ProtocolVersion = protocolVersion
	header.HeaderSize = headerSize

	b1, _ := r.ReadByte()
	messageType := b1 >> 4 & 0xFF
	flags := b1 & 0xFF
	header.MessageType = messageType
	header.Flags = flags

	b2, _ := r.ReadByte()
	serializeMethod := b2 >> 4 & 0xFF
	compression := b2 & 0xFF
	header.SerializationMethod = serializeMethod
	header.Compression = compression

	b3, _ := r.ReadByte()
	reserved := b3
	header.Reserved = reserved

	var sequence uint32
	if err := binary.Read(r, binary.BigEndian, &sequence); err != nil {
		return nil, nil, errors.Wrap(err, "failed to read sequence number")
	}

	var payloadSize uint32
	if err := binary.Read(r, binary.BigEndian, &payloadSize); err != nil {
		return nil, nil, errors.Wrap(err, "failed to read payload size")
	}

	if len(raw) != int(12+payloadSize) {
		return nil, nil, errors.Errorf("raw data length %d does not match expected length %d", len(raw), 8+payloadSize)
	}

	var payload FullServerResponsePacketPayload
	if json.NewDecoder(r).Decode(&payload) != nil {
		return nil, nil, errors.New("failed to decode full server response payload")
	}

	return &header, &payload, nil
}

func (conn *AsrDoubaoConn) parseServerResponseError(raw []byte) (*Header, error) {
	// 4 header bytes + 4 payload size bytes + 4 sequence bytes
	if len(raw) < 12 {
		return nil, errors.New("raw data too short to parse full server response")
	}

	header := &Header{}
	r := bytes.NewReader(raw)
	b0, _ := r.ReadByte()
	header.ProtocolVersion = (b0 >> 4) & 0x0F
	header.HeaderSize = b0 & 0x0F

	b1, _ := r.ReadByte()
	header.MessageType = b1 >> 4 & 0x0F
	header.Flags = b1 & 0x0F

	b2, _ := r.ReadByte()
	header.SerializationMethod = b2 >> 4 & 0x0F
	header.Compression = b2 & 0x0F

	b3, _ := r.ReadByte()
	header.Reserved = b3

	var errMessageCode uint32
	var errMessageSize uint32

	if err := binary.Read(r, binary.BigEndian, &errMessageCode); err != nil {
		return nil, errors.Wrap(err, "failed to read error message code")
	}

	if err := binary.Read(r, binary.BigEndian, &errMessageSize); err != nil {
		return nil, errors.Wrap(err, "failed to read error message size")
	}

	if len(raw) < 12+int(errMessageSize) {
		return nil, errors.Errorf("raw data length %d does not match expected length %d", len(raw), 12+errMessageSize)
	}

	message := string(raw[12 : 12+errMessageSize])
	return header, errors.Errorf("server error: code %d, message: %s", errMessageCode, message)
}

func (conn *AsrDoubaoConn) buildHeader(mt, flags, serializeMethod, compression byte) *Header {
	header := &Header{}
	header.ProtocolVersion = ProtocolVersion
	header.HeaderSize = byte(0x01)
	header.MessageType = mt
	header.Flags = flags
	header.SerializationMethod = serializeMethod
	header.Compression = compression
	header.Reserved = 0 // 保留字节，当前设置为 0

	return header
}
