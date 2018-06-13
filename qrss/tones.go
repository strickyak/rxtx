package qrss

import (
	"crypto/rand"
	"io"
	"log"
	"math"
	"math/big"
)

type Volt float64

type ToneGen struct {
	SampleRate float64 // Samples per second (Hz)
	ToneLen    float64 // Length in seconds
	RampLen    float64 // ramp-up, ramp-down in seconds
	BaseHz     float64
	StepHz     float64
}

// How many ticks (at the SampleRate) are in a whole tone (time length ToneLen).
func (tg ToneGen) WholeTicks() float64 {
	return tg.SampleRate * tg.ToneLen
}

// How many ticks (at the SampleRate) are in the raised-cosine-RampUp/RampDown time (time length RampLen).
func (tg ToneGen) RampTicks() float64 {
	return tg.SampleRate * tg.RampLen
}

// Turn a sequence of tones into voltage samples.  Special case Tone 0 creates a gap (silence) of whole tone length.
func (tg ToneGen) PlayTones(tones []Tone, vv chan Volt) {
	for _, b := range tones {
		tg.Boop(b, b, Both, vv)
	}
}

func (tg ToneGen) PlayTonePairs(tonePairs []TonePair, vv chan Volt) {
	for _, p := range tonePairs {
		tg.Boop(p.A, p.B, p.Fade, vv)
	}
}

// Boop writes voltages in range [-1.0, +1.0] to the channel vv, for tones sliding from tone1 to tone2, which might be the same tone.
// Notice Boop(0, _, _, _) produces silence.
func (tg ToneGen) Boop(tone1, tone2 Tone, fe FadeEnd, vv chan Volt) {
	hz1 := tg.BaseHz + float64(tone1)*tg.StepHz
	hz2 := tg.BaseHz + float64(tone2)*tg.StepHz

	wholeTicks := int(tg.WholeTicks())
	for t := 0; t < wholeTicks; t++ {
		if tone1 == 0 {
			vv <- Volt(0.0)
			continue
		}

		// Portion ranges 0.0 to almost 1.0.
		portion := float64(t) / float64(wholeTicks)
		// Interpolate part of the way between hz1 and hz2.
		hz := hz1 + portion*(hz2-hz1)
		log.Printf("%06d: %8.0f hz (%5.1f, %5.1f)", t, hz, tone1, tone2)

		// Apply a raised-cosine envelope to the first and last RampTicks ticks.
		var envelopeGain float64
		switch {
		case fe != Right && t < int(tg.RampTicks()): // First RampTicks, gain goes from 0.0 to 1.0
			{
				x := (float64(t) / tg.RampTicks()) * math.Pi
				y := math.Cos(x)
				envelopeGain = 0.5 - y/2.0
			}
		case fe != Left && int(tg.WholeTicks())-t < int(tg.RampTicks()): // Last RampTicks, gain goes from 1.0 to 0.0.
			{
				x := ((tg.WholeTicks() - float64(t)) / tg.RampTicks()) * math.Pi
				y := math.Cos(x)
				envelopeGain = 0.5 - y/2.0
			}
		default: // Middle of the Boop has full envelopeGain 1.0.
			{
				envelopeGain = 1.0
			}
		}

		// The angle theta depends on the ticks and the frequency hz.
		theta := float64(t) * hz * (2.0 * math.Pi) / tg.SampleRate
		// Take the sin of the angle, and multiply by the envelopeGain.
		v := envelopeGain * math.Sin(theta)
		vv <- Volt(v)
	}
}

const MaxShort = 0x7FFF

// EmitVolts consumes the volts from the channel vv, which use range [-1.0, +1.0].
// It multiplies by an overall gain, converts to signed int16, and writes to the writer in big-endian format.
// When the input volts channel has no more, we write true to the done channel,
// so the main program can exit.
func EmitVolts(vv chan Volt, gain float64, w io.Writer, done chan bool) {
	for {
		volt, ok := <-vv
		if !ok {
			break
		}
		y := gain * float64(volt)
		// Clip at +/- 1 unit.
		if y > 1.0 {
			y = 1.0
		}
		if y < -1.0 {
			y = -1.0
		}
		yShort := int(MaxShort * y)

		buf := []byte{
			byte(255 & (yShort >> 8)),
			byte(255 & yShort),
		}
		w.Write(buf)
	}
	done <- true
}

func Random(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(r.Int64())
}
