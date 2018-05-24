package qrss

import (
	"math"
)

const BaseHz = 500 // 500
const StepHz = 5   // 100 // 5

type Volt float64

type ToneGen struct {
	SampleRate float64 // Samples per second (Hz)
	ToneLen    float64 // Length in seconds
	RampLen    float64 // ramp-up, ramp-down in seconds
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
		z = append(z, tg.Boop(b)...)
	}
	return z
}

// Notice Boop(0) produces silence.
func (tg ToneGen) Boop(tone Tone) []Volt {
	var z []Volt
	hz := BaseHz + float64(tone)*StepHz
	for t := 0; t < int(tg.WholeTicks()); t++ {
		if tone == 0 {
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
