package src

import (
	"testing"
)

func TestVadAudioFilter(t *testing.T) {
	filter := NewVADAudioFilter(3)

	stubs := []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{false, []byte{0x01, 0x02}, 0},
		{false, []byte{0x02, 0x02}, 0},
		{false, []byte{0x03, 0x02}, 0},
		{false, []byte{0x04, 0x02}, 0},
		{false, []byte{0x05, 0x02}, 0},
	}

	for _, stub := range stubs {
		if stub.audioLen != len(filter.Feed(stub.hasVoice, stub.audioFrame)) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(filter.Feed(stub.hasVoice, stub.audioFrame)))
		}
	}

	stubs = []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{true, []byte{0x01, 0x02}, 0},
		{true, []byte{0x02, 0x02}, 0},
		{true, []byte{0x03, 0x02}, 3},
		{true, []byte{0x04, 0x02}, 1},
		{true, []byte{0x05, 0x02}, 1},
	}

	for _, stub := range stubs {
		got := filter.Feed(stub.hasVoice, stub.audioFrame)
		for _, g := range got {
			t.Logf("got %X", g.frame[0])
		}

		if stub.audioLen != len(got) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(got))
		}
	}

	stubs = []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{false, []byte{0x01, 0x02}, 0},
		{true, []byte{0x02, 0x02}, 0},
		{true, []byte{0x03, 0x02}, 0},
		{true, []byte{0x04, 0x02}, 3},
		{true, []byte{0x05, 0x02}, 1},
	}

	for _, stub := range stubs {
		got := filter.Feed(stub.hasVoice, stub.audioFrame)
		for _, g := range got {
			t.Logf("got %X", g.frame[0])
		}

		if stub.audioLen != len(got) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(got))
		}
	}

	stubs = []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{false, []byte{0x01, 0x02}, 0},
		{true, []byte{0x02, 0x02}, 0},
		{false, []byte{0x03, 0x02}, 0},
		{true, []byte{0x04, 0x02}, 0},
		{true, []byte{0x05, 0x02}, 0},
	}

	for _, stub := range stubs {
		got := filter.Feed(stub.hasVoice, stub.audioFrame)
		for _, g := range got {
			t.Logf("got %X", g.frame[0])
		}

		if stub.audioLen != len(got) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(got))
		}
	}

	stubs = []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{false, []byte{0x01, 0x02}, 0},
		{true, []byte{0x02, 0x02}, 0},
		{true, []byte{0x03, 0x02}, 0},
		{false, []byte{0x04, 0x02}, 0},
		{true, []byte{0x05, 0x02}, 0},
	}

	for _, stub := range stubs {
		got := filter.Feed(stub.hasVoice, stub.audioFrame)
		for _, g := range got {
			t.Logf("got %X", g.frame[0])
		}

		if stub.audioLen != len(got) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(got))
		}
	}

	stubs = []struct {
		hasVoice   bool
		audioFrame []byte
		audioLen   int
	}{
		{false, []byte{0x01, 0x02}, 0},
		{true, []byte{0x02, 0x02}, 0},
		{true, []byte{0x03, 0x02}, 0},
		{true, []byte{0x04, 0x02}, 3},
		{false, []byte{0x05, 0x02}, 0},
	}

	for _, stub := range stubs {
		got := filter.Feed(stub.hasVoice, stub.audioFrame)
		for _, g := range got {
			t.Logf("got %X", g.frame[0])
		}

		if stub.audioLen != len(got) {
			t.Errorf("expected audio length %d, got %d", stub.audioLen, len(got))
		}
	}

}
