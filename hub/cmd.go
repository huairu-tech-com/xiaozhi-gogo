package hub

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

func (s *Session) cmdTTSStart() error {
	jsonData := map[string]string{
		"type":  "tts",
		"state": "start",
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
		"type":  "tts",
		"state": "stop",
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
		"type":  "tts",
		"state": "sentence_start",
		text:    text,
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
		"type": "tts",
		text:   text,
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
		"type":    "tts",
		"emotion": emotion,
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
		"type":    "system",
		"command": "reboot",
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
		"type":    "alert",
		"status":  status,
		"message": message,
		"emotion": emotion,
	}
	log.Debug().Msgf("cmdTTSVolume: %+v", jsonData)

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	defer w.Close()
	return json.NewEncoder(w).Encode(jsonData)
}
