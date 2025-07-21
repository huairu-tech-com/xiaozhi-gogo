package src

import (
	"encoding/json"

	"github.com/hertz-contrib/websocket"
	"github.com/rs/zerolog/log"
)

type Emotion string

const (
	EmotionNeutral     Emotion = "neutral"
	EmotionHappy       Emotion = "happy"
	EmotionLaughing    Emotion = "laughing"
	EmotionFunny       Emotion = "funny"
	EmotionSad         Emotion = "sad"
	EmotionAngry       Emotion = "angry"
	EmotionCrying      Emotion = "crying"
	EmotionLoving      Emotion = "loving"
	EmotionEmbarrassed Emotion = "embarrassed"
	EmotionSurprised   Emotion = "surprised"
	EmotionShocked     Emotion = "shocked"
	EmotionThinking    Emotion = "thinking"
	EmotionWinking     Emotion = "winking"
	EmotionCool        Emotion = "cool"
	EmotionRelaxed     Emotion = "relaxed"
	EmomtionDelicious  Emotion = "delicious"
	EmotionKissy       Emotion = "kissy"
	EmotionConfident   Emotion = "confident"
	EmotionSleepy      Emotion = "sleepy"
	EmotionSilly       Emotion = "silly"
	EmotionConfused    Emotion = "confused"
)

var emotionEmoji = map[Emotion]string{
	EmotionNeutral:     "",
	EmotionHappy:       "😊",
	EmotionLaughing:    "😂",
	EmotionFunny:       "🤡",
	EmotionSad:         "😢",
	EmotionAngry:       "😠",
	EmotionCrying:      "😭",
	EmotionLoving:      "🥰",
	EmotionEmbarrassed: "😳",
	EmotionShocked:     "😱",
	EmotionSurprised:   "😮",
	EmotionThinking:    "🤔",
	EmotionWinking:     "😉",
	EmotionCool:        "😎",
	EmotionRelaxed:     "😌",
	EmomtionDelicious:  "😋",
	EmotionKissy:       "😘",
	EmotionConfident:   "😏",
	EmotionSleepy:      "😴",
	EmotionSilly:       "🤪",
	EmotionConfused:    "😕",
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
		"text":       text,
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
		"text":       text,
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
