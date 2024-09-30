package filter

import (
	"encoding/binary"
	"github.com/naveen246/slatedb-go/slatedb/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetSpecifiedBitOnly(t *testing.T) {
	cases := []struct {
		buf      []byte
		expected []byte
		bit      uint64
	}{
		{[]byte{0xF0, 0xAB, 0x9C}, []byte{0xF8, 0xAB, 0x9C}, 3},
		{[]byte{0xF0, 0xAB, 0x9C}, []byte{0xF0, 0xAF, 0x9C}, 10},
	}

	for _, c := range cases {
		setBit(c.bit, c.buf)
		assert.Equal(t, c.expected, c.buf)
	}

	nBytes := 4
	for byt := 0; byt < nBytes; byt++ {
		for i := 0; i < 8; i++ {
			buf := make([]byte, nBytes)
			bit := byt*8 + i
			setBit(uint64(bit), buf)

			for unset := 0; unset < nBytes; unset++ {
				if unset != byt {
					assert.Equal(t, byte(0), buf[unset])
				} else {
					assert.Equal(t, byte(1<<i), buf[byt])
				}
			}
		}
	}
}

func TestSetBitsDoesntUnsetBits(t *testing.T) {
	buf := []byte{0xFF, 0xFF, 0xFF}

	for i := 0; i < 24; i++ {
		setBit(uint64(i), buf)
		for j := 0; j < len(buf); j++ {
			assert.Equal(t, byte(0xFF), buf[j])
		}
	}
}

func TestCheckBits(t *testing.T) {
	numBytes := 4
	for i := 0; i < numBytes; i++ {
		for b := 0; b < 8; b++ {
			bit := i*8 + b
			buf := make([]byte, numBytes)
			buf[i] = 1 << b
			for checked := 0; checked < numBytes*8; checked++ {
				assert.Equal(t, bit == checked, checkBit(uint64(checked), buf))
			}
		}
	}
}

func TestComputeProbes(t *testing.T) {
	hash := uint64(0xDF77EF56DEADBEEF)
	probes := probesForKey(hash, 7, 1000000)
	expected := []uint32{
		928559, // h1
		107781, // h1 + h2
		287004, // h1 + h2 + h2 + 1
		466229, // h1 + h2 + h2 + 1 + h2 + 1 + 2
		645457, // h1 + h2 + h2 + 1 + h2 + 1 + 2 + h2 + 1 + 2 + 3
		824689, 3926,
	}
	assert.Equal(t, expected, probes)
}

func TestFilterEffective(t *testing.T) {
	keysToTest := uint32(100000)
	keySize := common.SizeOfUint32InBytes
	builder := NewBloomFilterBuilder(10)

	var i uint32
	for i = 0; i < keysToTest; i++ {
		bytes := make([]byte, keySize)
		binary.BigEndian.PutUint32(bytes, i)
		builder.AddKey(bytes)
	}
	filter := builder.Build()

	// check all entries in filter
	for i = 0; i < keysToTest; i++ {
		bytes := make([]byte, keySize)
		binary.BigEndian.PutUint32(bytes, i)
		assert.True(t, filter.HasKey(bytes))
	}

	// check false positives
	fp := uint32(0)
	for i := keysToTest; i < keysToTest*2; i++ {
		bytes := make([]byte, keySize)
		binary.BigEndian.PutUint32(bytes, i)
		if filter.HasKey(bytes) {
			fp += 1
		}
	}

	// observed fp is 0.00744
	assert.True(t, float32(fp)/float32(keysToTest) < 0.01)
}
