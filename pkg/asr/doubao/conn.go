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
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	// https://www.volcengine.com/docs/6561/1354869?lang=zh#demo
	DoubaoStreamAsrEndpoint = "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel"
)

const (
	// 小时版
	DoubaoModelDuration = "volc.bigasr.sauc.duration"
	// 并发版
	DoubaoModelConcurrent = "volc.bigasr.sauc.concurrent"
)

var DefaultDialer = func(ctx context.Context, cfg *AsrDoubaoConfig) (*AsrDoubaoConn, error) {
	headers := http.Header{}
	headers.Set("Host", cfg.Host)
	headers.Set("X-Api-App-Key", cfg.ApiKey)
	headers.Set("X-Api-Access-Key", cfg.AccessKey)
	headers.Set("X-Api-Resource-Id", DoubaoModelDuration)

	connectId := uuid.New().String()
	headers.Set("X-Api-Connect-Id", connectId)

	fmt.Printf("Connecting to Doubao ASR with connectId: %s\n", connectId)
	fmt.Printf("Using API Key: %+v\n", headers)
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: false},
	}

	conn, resp, err := dialer.DialContext(ctx, DoubaoStreamAsrEndpoint, headers)
	doubaoConn := &AsrDoubaoConn{
		conn:      conn,
		connectId: connectId,
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Response Stats: %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "Response Headers: %v\n", resp.Header)
		bytes, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Response Body: %s\n", string(bytes))
		return nil, err
	}

	doubaoConn.respCh = make(chan *asr.AsrResponse, 10) // buffered channel for responses
	doubaoConn.ctx, doubaoConn.cancel = context.WithCancel(ctx)
	doubaoConn.ttLogid = resp.Header.Get("X-Tt-Logid")

	fmt.Printf("Connected to Doubao ASR with connectId: %s, ttLogid: %s\n", doubaoConn.connectId, doubaoConn.ttLogid)

	if err := doubaoConn.sendParameter(); err != nil {
		return nil, err
	}

	go doubaoConn.enterLoop()

	return doubaoConn, err
}

// 协议版本， 当前仅有一个
const ProtocolVersion = 0b0001 & 0xFF

// 包类型
const (
	MessageTypeFullClientRequest  = 0b0001 & 0xFF
	MessageTypeAudioOnlyRequest   = 0b0010 & 0xFF
	MessageTypeFullServerResponse = 0b1001 & 0xFF
	MessageTypeServerError        = 0b1111 & 0xFF
)

// Serialization Method
const (
	SerializationMethodJson = 0b0000 & 0xFF
	SerializationMethodNone = 0b0001 & 0xFF
)

// Message Compression
const (
	CompressionNone = 0b0000 & 0xFF
	CompressionGzip = 0b0001 & 0xFF
)

// Message audio request last flag
const (
	NoSequence     = 0b0000 & 0xFF // 全部音频数据
	PosSequence    = 0b0001 & 0xFF // 最后一个音频数据包
	NegSequence    = 0b0010 & 0xFF //
	NegWithSequnce = 0b0011 & 0xFF //
)

// Server response message type
type ServerMessageType string

const (
	// full server response message
	SeverMessageTypeFullServerResponse ServerMessageType = "full_server_response" // 全部音频数据包
	// indicates an error message from the server
	ServerMessageTypeServerError ServerMessageType = "server_error" // 错误信息
)

const (
	ServerErrorSuccess        = 20000000 // 成功
	ServerErrorInvalidRequest = 40000001 // 请求参数无效

	ServerErrorEmptyAudio   = 45000002 // 空音频
	ServerErrorTimeout      = 45000081 // 音频过短
	ServerErrorAudioFormat  = 45000151 // 音频格式不正确
	SererErrorInternalError = 55000031 // 服务内部处理错误
)

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
	conn      *websocket.Conn
	connectId string // 客户端的 connect ID
	ttLogid   string // 服务端的 trace ID

	respCh chan *asr.AsrResponse // 响应通道

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 上下文取消函数
}

func (conn *AsrDoubaoConn) Err() error {
	return conn.ctx.Err()
}

func (conn *AsrDoubaoConn) Pressure() int32 {
	return 0
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
	return conn.close()
}

func (conn *AsrDoubaoConn) close() error {
	if conn.conn != nil {
		return conn.conn.Close()
	}

	if conn.ctx.Err() != nil {
		conn.cancel()
	}

	if conn.respCh != nil {
		close(conn.respCh)
		conn.respCh = nil
	}

	return nil
}

// this is the first handshake message send and response
func (conn *AsrDoubaoConn) sendParameter() error {
	var err error
	var mt int
	var bytes []byte
	var n int
	var fullClientPacket []byte

	w, err := conn.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		goto end
	}

	fullClientPacket, err = conn.buildFullClientPacket()
	if err != nil {
		goto end
	}

	n, err = w.Write(fullClientPacket)
	if err != nil {
		goto end
	}

	if len(fullClientPacket) != n {
		err = errors.New("failed to write full client packet")
		goto end
	}

	w.Close()

	mt, bytes, err = conn.conn.ReadMessage()
	if err != nil {
		err = errors.Wrap(err, "failed to read message after sending full client packet")
		goto end
	}

	if mt != websocket.BinaryMessage {
		err = errors.New("expected binary message, got text message")
		goto end
	}

	if len(bytes) < 4 {
		err = errors.New("message too short, expected at least 4 bytes")
		goto end
	}

	if checkMessageType(bytes[1]) == ServerMessageTypeServerError {
		err = errors.Errorf("expected server error message, got %d",
			conn.parseServerResponseError(bytes))
		goto end
	}

	if checkMessageType(bytes[1]) == SeverMessageTypeFullServerResponse {
		_, err := conn.parseFullServerResponse(bytes)
		if err != nil {
			goto end
		}
	}

end:
	if err != nil {
		conn.close()
	}

	return err
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
	// payload.Request.VadSegmentDuration = 3000
	// payload.Request.EndWindowSize = 800
	// payload.Request.FroceToSpeechTime = 10000
	// payload.Request.Corpus.BoostingTableId = ""
	// payload.Request.Corpus.BosstingTableName = ""
	// payload.Request.Corpus.CorrectTableID = ""

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	packetBuffer := bytes.Buffer{}
	binary.Write(&packetBuffer, binary.BigEndian, header)
	binary.Write(&packetBuffer, binary.BigEndian, uint32(len(payloadBytes)))
	binary.Write(&packetBuffer, binary.BigEndian, payloadBytes)

	return packetBuffer.Bytes(), nil
}

func (conn *AsrDoubaoConn) buildAudioOnlyRequestPacket(audioData []byte, seq int32, isLastSnippet bool) ([]byte, error) {
	packetIndicator := NoSequence // 默认是最后一个音频数据包
	if isLastSnippet {
		packetIndicator = NegSequence // 最后一个音频数据包
		seq = -seq
	}
	header := conn.buildHeader(MessageTypeAudioOnlyRequest,
		int8(packetIndicator),
		SerializationMethodNone,
		CompressionGzip)

	var buffer bytes.Buffer
	zipwriter := gzip.NewWriter(&buffer)
	if _, err := zipwriter.Write(audioData); err != nil {
		return nil, errors.Wrap(err, "failed to create zip writer")
	}
	zipwriter.Close()
	zippedData := buffer.Bytes()
	packetBuffer := bytes.Buffer{}
	binary.Write(&packetBuffer, binary.BigEndian, header)
	// binary.Write(&packetBuffer, binary.BigEndian, seq)
	binary.Write(&packetBuffer, binary.BigEndian, uint32(len(zippedData)))
	binary.Write(&packetBuffer, binary.BigEndian, zippedData)
	return packetBuffer.Bytes(), nil
}

func (conn *AsrDoubaoConn) enterLoop() error {
	go func() {
		if err := conn.readText(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading text: %v\n", err)
		}
	}()

	<-conn.ctx.Done()
	return conn.ctx.Err()
}

func (conn *AsrDoubaoConn) readText() error {
	defer func() {
		fmt.Println("11111111111111111111111111111111111111111111")
		fmt.Println("11111111111111111111111111111111111111111111")
		fmt.Println("11111111111111111111111111111111111111111111")
		fmt.Println("11111111111111111111111111111111111111111111")
		fmt.Println("11111111111111111111111111111111111111111111")
		conn.cancel()
		conn.conn.Close()
	}()

	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("======== Dida dida ==== \n")
		case <-conn.ctx.Done():
			return conn.ctx.Err()
		default:
			conn.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			mt, bytes, err := conn.conn.ReadMessage()
			fmt.Printf("DOUBAO => ASR: %d, length: %d\n", mt, len(bytes))

			if err != nil {
				return errors.Wrap(err, "failed to read message in readText")
			}

			if mt != websocket.BinaryMessage {
				return errors.New("expected binary message, got text message")
			}

			if len(bytes) < 4 {
				return errors.New("message too short, expected at least 4 bytes")
			}

			switch checkMessageType(bytes[1]) {
			case SeverMessageTypeFullServerResponse:
				payload, err := conn.parseFullServerResponse(bytes)
				if err != nil {
					return errors.Wrap(err, "failed to parse full server response")
				}
				fmt.Printf("DOUBAO => Server : %+v\n", payload)
				// conn.respCh <- &asr.AsrResponse{
				// 	Text: payload.Result.Text,
				// }

			case ServerMessageTypeServerError:
				err := conn.parseServerResponseError(bytes)
				if err != nil {
					return errors.Wrap(err, "failed to parse server response error")
				}

			default:
				return errors.Errorf("unknown message type: %d", bytes[0])
			}
		}
	}
}

func (conn *AsrDoubaoConn) ResponseCh() chan *asr.AsrResponse {
	return conn.respCh
}

func (conn *AsrDoubaoConn) SendAudio(pcm []byte, seq int32, isLast bool) error {
	if len(pcm) < 100 {
		return nil
	}

	if conn.ctx.Err() != nil {
		return conn.ctx.Err()
	}

	fmt.Printf("XIAOZHI => server => ASR : %d\n", len(pcm))

	var err error
	var data []byte
	err = conn.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		goto end
	}

	data, err = conn.buildAudioOnlyRequestPacket(pcm, seq, isLast)
	if err != nil {
		goto end
	}

	if err := conn.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		goto end
	}

end:
	if err != nil {
		conn.cancel()
		conn.conn.Close()
	}

	return err
}

func checkMessageType(mt byte) ServerMessageType {
	if mt>>4&0x0F == MessageTypeFullServerResponse {
		return SeverMessageTypeFullServerResponse
	}

	if mt>>4&0x0F == MessageTypeServerError {
		return ServerMessageTypeServerError
	}

	return "unknown"
}

func (conn *AsrDoubaoConn) parseFullServerResponse(raw []byte) (*FullServerResponsePacketPayload, error) {
	// 4 header bytes + 4 payload size bytes + 4 sequence bytes
	if len(raw) < 12 {
		return nil, errors.New("raw data too short to parse full server response")
	}

	r := bytes.NewReader(raw)
	b0, _ := r.ReadByte()
	protocolVersion := (b0 >> 4) & 0x0F
	headerSize := b0 & 0x0F

	b1, _ := r.ReadByte()
	messageType := b1 >> 4 & 0x0F
	flags := b1 & 0x0F

	b2, _ := r.ReadByte()
	serializeMethod := b2 >> 4 & 0x0F
	compression := b2 & 0x0F

	b3, _ := r.ReadByte()
	reserved := b3

	var sequence uint32
	if err := binary.Read(r, binary.BigEndian, &sequence); err != nil {
		return nil, err
	}

	var payloadSize uint32
	if err := binary.Read(r, binary.BigEndian, &payloadSize); err != nil {
		return nil, err
	}

	log.Debug().Msgf("ProtocolVersion: %d, HeaderSize: %d, MessageType: %d, Flags: %d, SerializeMethod: %d, Compression: %d, Reserved: %d, PayloadSize: %d",
		protocolVersion, headerSize, messageType, flags, serializeMethod, compression, reserved, payloadSize)

	if len(raw) != int(12+payloadSize) {
		return nil, errors.Errorf("raw data length %d does not match expected length %d", len(raw), 8+payloadSize)
	}

	var payload FullServerResponsePacketPayload
	if json.NewDecoder(r).Decode(&payload) != nil {
		return nil, errors.New("failed to decode full server response payload")
	}

	return &payload, nil
}

func (conn *AsrDoubaoConn) parseServerResponseError(raw []byte) error {
	// 4 header bytes + 4 payload size bytes + 4 sequence bytes
	if len(raw) < 12 {
		return errors.New("raw data too short to parse full server response")
	}

	r := bytes.NewReader(raw)
	b0, _ := r.ReadByte()
	protocolVersion := (b0 >> 4) & 0x0F
	headerSize := b0 & 0x0F

	b1, _ := r.ReadByte()
	messageType := b1 >> 4 & 0x0F
	flags := b1 & 0x0F

	b2, _ := r.ReadByte()
	serializeMethod := b2 >> 4 & 0x0F
	compression := b2 & 0x0F

	b3, _ := r.ReadByte()
	reserved := b3

	var errMessageCode uint32
	var errMessageSize uint32

	if err := binary.Read(r, binary.BigEndian, &errMessageCode); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &errMessageSize); err != nil {
		return err
	}

	log.Debug().Msgf("ProtocolVersion: %d, HeaderSize: %d, MessageType: %d, Flags: %d, SerializeMethod: %d, Compression: %d, Reserved: %d, PayloadSize: %d",
		protocolVersion, headerSize, messageType, flags, serializeMethod, compression, reserved, errMessageSize)

	if len(raw) < 12+int(errMessageSize) {
		return errors.Errorf("raw data length %d does not match expected length %d", len(raw), 12+errMessageSize)
	}

	message := string(raw[12 : 12+errMessageSize])
	return errors.Errorf("server error: code %d, message: %s", errMessageCode, message)
}

func (conn *AsrDoubaoConn) buildHeader(mt, flags, serializeMethod, compression int8) []byte {
	headerSize := 0x01 // 4
	header := make([]byte, 4)
	header[0] = byte((ProtocolVersion << 4) | headerSize)
	header[1] = byte((mt << 4) | (flags & 0x0F))
	header[2] = byte((serializeMethod << 4) | compression)
	header[3] = 0 // reserved byte, currently set to 0
	return header
}
