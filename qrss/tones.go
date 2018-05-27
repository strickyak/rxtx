package qrss

import (
	"crypto/rand"
	"io"
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

func (tg ToneGen) WholeTicks() float64 {
	return tg.SampleRate * tg.ToneLen
}

func (tg ToneGen) RampTicks() float64 {
	return tg.SampleRate * tg.RampLen
}

func (tg ToneGen) Play(tones []Tone, vv chan Volt) {
	for _, b := range tones {
		tg.Boop(b, b, vv)
	}
}

// Notice Boop(0) produces silence.
func (tg ToneGen) Boop(tone1, tone2 Tone, vv chan Volt) {
	hz1 := tg.BaseHz + float64(tone1)*tg.StepHz
	hz2 := tg.BaseHz + float64(tone2)*tg.StepHz
	wholeTicks := int(tg.WholeTicks())
	for t := 0; t < wholeTicks; t++ {

		portion := float64(t) / float64(wholeTicks)
		hz := hz1 + float64(hz2-hz1)*portion

		if tone1 == 0 && !*WITH_TAILS {
			vv <- Volt(0.0)
			continue
		}

		var gain float64
		switch {
		case t < int(tg.RampTicks()):
			{
				x := (float64(t) / tg.RampTicks()) * math.Pi
				y := math.Cos(x)
				gain = 0.5 - y/2.0
			}
		case int(tg.WholeTicks())-t < int(tg.RampTicks()):
			{
				x := ((tg.WholeTicks() - float64(t)) / tg.RampTicks()) * math.Pi
				y := math.Cos(x)
				gain = 0.5 - y/2.0
			}
		default:
			{
				gain = 1.0
			}
		}

		theta := float64(t) * hz * (2.0 * math.Pi) / tg.SampleRate
		v := gain * math.Sin(theta)
		vv <- Volt(v)
	}
}

const MaxShort = 0x7FFF

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
