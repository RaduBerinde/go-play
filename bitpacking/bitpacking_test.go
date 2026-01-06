package bitpacking

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode8_BPK4(t *testing.T) {
	input := []uint8{0x0A, 0x0B, 0x0C, 0x0D}
	output := make([]byte, 2)
	Encode8(input, 4, output)
	// Expected: [0xBA, 0xDC] (0x0A in low nibble, 0x0B in high nibble, etc.)
	require.Equal(t, []byte{0xBA, 0xDC}, output)

	Encode8(input[:3], 4, output)
	require.Equal(t, []byte{0xBA, 0x0C}, output)
	Encode8(input[:2], 4, output)
	require.Equal(t, []byte{0xBA}, output[:1])
	Encode8(input[:1], 4, output)
	require.Equal(t, []byte{0x0A}, output[:1])
}

func TestEncode8_BPK8(t *testing.T) {
	input := []uint8{0x12, 0x34, 0x56}
	output := make([]byte, 3)
	Encode8(input, 8, output)
	require.Equal(t, input, output)
}

func TestEncode16_BPK12(t *testing.T) {
	// Two 12-bit values: 0x123 and 0x456
	input := []uint16{0x123, 0x456}
	output := make([]byte, 3)
	Encode16(input, 12, output)
	require.Equal(t, []byte{0x23, 0x61, 0x45}, output)

	input = []uint16{0x123, 0x456, 0x789}
	output = make([]byte, 6)
	Encode16(input, 12, output)
	require.Equal(t, []byte{0x23, 0x61, 0x45, 0x89, 0x07, 0x00}, output)
}

func TestEncode16_BPK16(t *testing.T) {
	input := []uint16{0x1234, 0x5678}
	output := make([]byte, 4)
	Encode16(input, 16, output)

	// Little-endian: 0x1234 -> [0x34, 0x12], 0x5678 -> [0x78, 0x56]
	expected := []byte{0x34, 0x12, 0x78, 0x56}
	for i, v := range expected {
		if output[i] != v {
			t.Errorf("index %d: expected 0x%02X, got 0x%02X", i, v, output[i])
		}
	}
}

func TestDecode_BPK4(t *testing.T) {
	data := []byte{0xBA, 0xDC}

	tests := []struct {
		i        uint
		expected uint16
	}{
		{0, 0x0A},
		{1, 0x0B},
		{2, 0x0C},
		{3, 0x0D},
	}

	for _, tc := range tests {
		got := Decode(data, tc.i, 4)
		if got != tc.expected {
			t.Errorf("Decode(data, %d, 4): expected 0x%X, got 0x%X", tc.i, tc.expected, got)
		}
	}
}

func TestDecode_BPK8(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56}

	for i, v := range data {
		got := Decode(data, uint(i), 8)
		if got != uint16(v) {
			t.Errorf("Decode(data, %d, 8): expected 0x%X, got 0x%X", i, v, got)
		}
	}
}

func TestDecode_BPK12(t *testing.T) {
	// Encoded as: [low8_a=0x23, high4_a|low4_b<<4=0x61, high8_b=0x45]
	data := []byte{0x23, 0x61, 0x45}

	got0 := Decode(data, 0, 12)
	if got0 != 0x123 {
		t.Errorf("Decode(data, 0, 12): expected 0x123, got 0x%X", got0)
	}

	got1 := Decode(data, 1, 12)
	if got1 != 0x456 {
		t.Errorf("Decode(data, 1, 12): expected 0x456, got 0x%X", got1)
	}
}

func TestDecode_BPK16(t *testing.T) {
	data := []byte{0x34, 0x12, 0x78, 0x56}

	got0 := Decode(data, 0, 16)
	if got0 != 0x1234 {
		t.Errorf("Decode(data, 0, 16): expected 0x1234, got 0x%X", got0)
	}

	got1 := Decode(data, 1, 16)
	if got1 != 0x5678 {
		t.Errorf("Decode(data, 1, 16): expected 0x5678, got 0x%X", got1)
	}
}

func TestRoundTrip(t *testing.T) {
	for range 100 {
		n := 1 + rand.IntN(100)
		input := make([]uint16, n)
		for i := range input {
			input[i] = uint16(rand.Uint32())
		}
		input8 := make([]uint8, n)
		for i := range input {
			input8[i] = uint8(input[i])
		}

		for _, bpk := range []int{4, 8, 12, 16} {
			encoded := make([]byte, EncodedSize(n, bpk))
			if bpk <= 8 {
				Encode8(input8, bpk, encoded)
			} else {
				Encode16(input, bpk, encoded)
			}
			for i := range input {
				require.Equal(t, Decode(encoded, uint(i), bpk), input[i]&(1<<bpk-1))
			}
			for range 10 {
				a, b, c := rand.IntN(n), rand.IntN(n), rand.IntN(n)
				va, vb, vc := Decode3(encoded, uint(a), uint(b), uint(c), bpk)
				require.Equal(t, va, input[a]&(1<<bpk-1))
				require.Equal(t, vb, input[b]&(1<<bpk-1))
				require.Equal(t, vc, input[c]&(1<<bpk-1))
			}
		}
	}
}
