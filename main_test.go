package main

import (
	"math/rand"
	"testing"

	"github.com/faiface/beep"
	"gotest.tools/v3/assert"
)

const Size int = 131072
const ChunkSize int = 64

var Ref = make([][2]float64, ChunkSize)

func init() {
	for i := range Ref {
		Ref[i][0] = float64(i)
		Ref[i][1] = float64(i)
	}
}

func TestChunk(t *testing.T) {
	size := 64
	avail := Size
	mock := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		assert.Equal(t, size, len(samples))
		if avail == 0 {
			return 0, false
		}
		var i int
		for i = 0; i < size; i++ {
			if avail == 0 {
				return i, true
			}
			samples[i] = Ref[i]
			avail--
		}
		return i, true
	})

	c := Chunk(mock, size)
	samples := make([][2]float64, 512)
	for s := 0; s < Size; {
		read := rand.Intn(512)
		n, ok := c.Stream(samples[0:read])
		assert.Assert(t, ok)
		if Size-s < read {
			assert.Equal(t, Size-s, n)
		} else {
			assert.Equal(t, read, n)
		}
		for i := 0; i < n; i++ {
			assert.Equal(t, Ref[(s+i)%size], samples[i])
		}
		s += n
	}
	n, ok := c.Stream(samples)
	assert.Assert(t, !ok)
	assert.Equal(t, 0, n)
}
