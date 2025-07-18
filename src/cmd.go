package src

import (
	"encoding/json"

	"github.com/hertz-contrib/websocket"
	"github.com/rs/zerolog/log"
)

const (
	EmotionNeutral     = "neutral"
	EmotionHappy       = "happy"
	EmotionLaughing    = "laughing"
	EmotionFunny       = "funny"
	EmotionSad         = "sad"
	EmotionAngry       = "angry"
	EmotionCrying      = "crying"
	EmotionLoving      = "loving"
	EmotionEmbarrassed = "embarrassed"
	EmotionSurprised   = "surprised"
	EmotionShocked     = "shocked"
	EmotionThinking    = "thinking"
	EmotionWinking     = "winking"
	EmotionCool        = "cool"
	EmotionRelaxed     = "relaxed"
	EmomtionDelicious  = "delicious"
	EmotionKissy       = "kissy"
	EmotionConfident   = "confident"
	EmotionSleepy      = "sleepy"
	EmotionSilly       = "silly"
	EmotionConfused    = "confused"
)

var emotionEmoji = map[string]string{
	"neutral":     "😐",
	"happy":       "😊",
	"laughing":    "😂",
	"funny":       "🤡",
	"sad":         "😢",
	"angry":       "😠",
	"crying":      "😭",
	"loving":      "🥰",
	"embarrassed": "😳",
	"surprised":   "😮",
	"shocked":     "😱",
	"thinking":    "🤔",
	"winking":     "😉",
	"cool":        "😎",
	"relaxed":     "😌",
	"delicious":   "😋",
	"kissy":       "😘",
	"confident":   "😏",
	"sleepy":      "😴",
	"silly":       "🤪",
	"confused":    "😕",
}

const (
	CmdTypeTTS    string = "tts"
	CmdTypeSTT    string = "stt"
	CmdTypeLLM    string = "llm"
	CmdTypeSystem string = "system"
	CmdTypeAlert  string = "alert"
)

func (s *Session) cmdTTSStart() error {
	jsonData := map[string]string{
		"type":       CmdTypeTTS,
		"state":      "start",
		"session_id": s.sessionId,
	}
	log.Debug().Msgf("cmdTTSStart: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdTTSStop(raw []byte) error {
	jsonData := map[string]string{
		"type":       CmdTypeTTS,
		"state":      "stop",
		"session_id": s.sessionId,
	}
	log.Debug().Msgf("cmdTTSStop: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdTTSSentenceStart(text string) error {
	jsonData := map[string]string{
		"type":       CmdTypeTTS,
		"state":      "sentence_start",
		"session_id": s.sessionId,
		text:         text,
	}
	log.Debug().Msgf("cmdTTSSentenceStart: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdSTT(text string) error {
	jsonData := map[string]string{
		"type":       CmdTypeSTT,
		"session_id": s.sessionId,
		text:         text,
	}
	log.Debug().Msgf("cmdSTT: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdLLM(emotion string) error {
	jsonData := map[string]string{
		"type":       CmdTypeLLM,
		"session_id": s.sessionId,
		"emotion":    emotion,
	}
	log.Debug().Msgf("cmdLLM: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdSystem() error {
	jsonData := map[string]string{
		"type":       CmdTypeSystem,
		"command":    "reboot",
		"session_id": s.sessionId,
	}
	log.Debug().Msgf("cmdSystem: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdAlert(status, message, emotion string) error {
	jsonData := map[string]interface{}{
		"type":       CmdTypeAlert,
		"status":     status,
		"message":    message,
		"session_id": s.sessionId,
		"emotion":    emotion,
	}
	log.Debug().Msgf("cmdTTSAlert: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}

func (s *Session) cmdEmotion(emotion string) error {
	return s.cmdLLM(emotion)
}
