package main

import (
	"encoding/binary"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 48000

type sound int

const (
	soundO sound = iota
	soundX
	soundBoardWon
	soundBoardDrawn
	soundWin
	soundLoss
	soundDraw
	soundCount
)

// sounds plays short synthesised effects.
type sounds struct {
	ctx  *audio.Context
	pcm  [soundCount][]byte
	mute bool
}

func newSounds() *sounds {
	s := &sounds{ctx: audio.NewContext(sampleRate)}
	s.pcm[soundO] = notes(0.18, 0.09, 587)
	s.pcm[soundX] = notes(0.18, 0.09, 494)
	s.pcm[soundBoardWon] = notes(0.16, 0.09, 659, 880)
	s.pcm[soundBoardDrawn] = notes(0.12, 0.14, 330)
	s.pcm[soundWin] = notes(0.16, 0.11, 523, 659, 784, 1047)
	s.pcm[soundLoss] = notes(0.14, 0.17, 392, 330, 262)
	s.pcm[soundDraw] = notes(0.12, 0.13, 440, 440)
	return s
}

func (s *sounds) play(snd sound) {
	if !s.mute {
		s.ctx.NewPlayerF32FromBytes(s.pcm[snd]).Play()
	}
}

// notes synthesises a sequence of soft bell-like notes, each lasting duration seconds,
// as 32-bit float stereo samples.
func notes(volume, duration float64, frequencies ...float64) []byte {
	perNote := int(duration * sampleRate)
	tail := int(0.25 * sampleRate) // the last note rings a little longer
	total := perNote*len(frequencies) + tail
	pcm := make([]byte, 0, total*8)
	for i := range total {
		v := 0.0
		for n, f := range frequencies {
			t := float64(i-n*perNote) / sampleRate
			if t < 0 {
				continue
			}
			attack := math.Min(t/0.004, 1)
			decay := math.Exp(-t * 18)
			v += attack * decay * (math.Sin(2*math.Pi*f*t) + 0.25*math.Sin(4*math.Pi*f*t))
		}
		sample := math.Float32bits(float32(volume * v))
		pcm = binary.LittleEndian.AppendUint32(pcm, sample) // left
		pcm = binary.LittleEndian.AppendUint32(pcm, sample) // right
	}
	return pcm
}
