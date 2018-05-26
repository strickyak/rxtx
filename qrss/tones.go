package qrss

import (
	"crypto/rand"
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

func (tg ToneGen) Play(tones []Tone) []Volt {
	var z []Volt
	for _, b := range tones {
		z = append(z, tg.Boop(b, b)...)
	}
	return z
}

// Notice Boop(0) produces silence.
func (tg ToneGen) Boop(tone1, tone2 Tone) []Volt {
	var z []Volt
	hz1 := tg.BaseHz + float64(tone1)*tg.StepHz
	hz2 := tg.BaseHz + float64(tone2)*tg.StepHz
	wholeTicks := int(tg.WholeTicks())
	for t := 0; t < wholeTicks; t++ {

		portion := float64(t) / float64(wholeTicks)
		hz := hz1 + float64(hz2-hz1)*portion

		if tone1 == 0 && !*WITH_TAILS {
			z = append(z, Volt(0.0))
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
		z = append(z, Volt(v))
	}
	return z
}

const MaxShort = 0x7FFF

func VoltsToS16be(vv []Volt, gain float64) []byte {
	var z []byte
	for i := 0; i < len(vv); i++ {
		v := gain * float64(vv[i])
		// Clip at +/- 1 unit.
		if v > 1.0 {
			v = 1.0
		}
		if v < -1.0 {
			v = -1.0
		}
		short := int(MaxShort * v)

		z = append(z, byte(255&(short>>8)), byte(255&short))

	}
	return z
}

func Random(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(r.Int64())
}
