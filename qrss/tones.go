package qrss

import (
	"math"
)

const BaseHz = 500 // 500
const StepHz = 5 // 100 // 5

const SampleHz = 44100 // 8000  // samples per second
const GradualSamples = SampleHz / 5 // A fifth of a second

const WholeNote = 6 * SampleHz // seconds
const HalfNote = 3 * SampleHz // seconds

func Play(levels []byte) []byte {
	var z []byte
	var prev byte = 255
	for _, b := range levels {
		if b == prev {
			prev = 255
			continue
		}
		if b == 0 {
			z = append(z, Gap()...)
			prev = 255
		} else {
			z = append(z, Boop(b)...)
			prev = b
		}
	}
	return z
}

func Boop(level byte) []byte {
	var z []byte
	hz := BaseHz + float64(level) * StepHz
	for t := 0; t < WholeNote; t++ {
		var gain float64
		switch {
		case t < GradualSamples: {
			x := (float64(t) / float64(GradualSamples)) * math.Pi
			y := math.Cos(x)
			gain = 0.5 - y / 2.0
			}
		case WholeNote - t < GradualSamples: {
			x := (float64(WholeNote - t) / float64(GradualSamples)) * math.Pi
			y := math.Cos(x)
			gain = 0.5 - y / 2.0
			}
		default: {
			gain = 1.0
			}
		}

		theta := float64(t) * hz * (2.0 * math.Pi) / SampleHz
		volts := gain * math.Sin(theta)
		short := int16(10000 * volts)
		z = append(z, byte(short >> 8), byte(short))
	}
	return z
}

func Gap() []byte {
	var z []byte
	for t := 0; t < HalfNote; t++ {
		z = append(z, 0, 0)
	}
	return z
}
