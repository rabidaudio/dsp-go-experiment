package main

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

const N = 512

var Kernel = make([]float64, 64)

func init() {
	exponentialLowPassKernel(Kernel, 1)
}

func main() {
	f, err := os.Open("saw.wav")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(Chunk(streamer, 512), beep.Callback(func() {
		done <- true
	})))

	<-done
}

func Chunk(wrapped beep.Streamer, size int) beep.Streamer {
	buf := make([][2]float64, size, size)
	buffered := 0
	return beep.StreamerFunc(func(samples [][2]float64) (int, bool) {
		n := 0
		for n < len(samples) {
			if buffered > 0 {
				// read any buffered data into sample
				for i := size - buffered; i < size && n < len(samples); i++ {
					samples[n] = buf[i]
					n++
					buffered--
				}
			} else if len(samples)-n < size {
				// read into the buffer instead, so that we can send a partial amount outs
				nn, ok := wrapped.Stream(buf)
				if !ok {
					break
				}
				buffered += nn
				continue // on the next loop we'll copy the buffer into the samples
			} else {
				// chunk the samples to the desired size
				nn, ok := wrapped.Stream(samples[n : n+size])
				if !ok {
					break
				}
				n += nn
			}
		}
		return n, n > 0
	})
}

// func LowPass(stream beep.Streamer, kernel []float64) beep.Streamer {
// 	sbuffer := make([]float64, N, N)
// 	outbuffer := make([]float64, N+len(kernel), N+len(kernel))
// 	tail := make([]float64, len(kernel), len(kernel))
// 	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
// 		if len(samples) != N {
// 			panic(fmt.Sprintf("expected sample size %v but was %v", N, len(samples)))
// 		}
// 		if n, ok = stream.Stream(samples); !ok {
// 			return
// 		}
// 		for channel := 0; channel < 2; channel++ {
// 			for i := 0; i < n; i += N {
// 				sbuffer[i] = samples[i][channel]
// 			}
// 			convole(sbuffer, kernel, outbuffer)
// 			for i := 0; i < n; i += N {
// 				samples[i][channel] = outbuffer[i]
// 			}
// 			// add the tail of the last samples and save the new tail
// 			// for the next loop
// 			for t := range tail {
// 				samples[t][channel] += tail[t]
// 				tail[t] = outbuffer[N+t]
// 			}
// 		}
// 		return
// 	})
// }

func convole(sample, kernel, out []float64) {
	for i := range out {
		out[i] = 0
	}

	for i := range sample {
		for j := range kernel {
			out[i+j] += sample[i] * kernel[j]
		}
	}
}

func exponentialLowPassKernel(out []float64, decay float64) {
	var sum float64 = 0
	for i := range out {
		v := 1 * math.Exp(-1*decay*float64(i))
		out[i] = v
		sum += v
	}
	for i := range out {
		// normalize
		out[i] /= sum
	}
}
