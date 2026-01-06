// Package bitpacking provides functions for encoding and decoding packed integer values.
package bitpacking

import "unsafe"

func EncodedSize(n int, bpk int) int {
	switch bpk {
	case 4:
		return (n + 1) / 2
	case 8:
		return n
	case 12:
		return (n + 1) / 2 * 3
	case 16:
		return n * 2
	default:
		panic("bpk must be 4, 8, 12, or 16")
	}
}

// Encode8 packs uint8 values into bytes using the specified bits-per-key (bpk).
// When bpk < 8, the lower bpk bits of each value are used.
//
// For bpk=4, two values are packed per byte (lower 4 bits each).
// For bpk=8, values are copied directly.
//
// Panics if bpk is not 4 or 8.
func Encode8(input []uint8, bpk int, output []byte) {
	switch bpk {
	case 4:
		if len(input) > 1 {
			pairs := unsafe.Slice((*[2]uint8)(unsafe.Pointer(unsafe.SliceData(input))), len(input)/2)
			_ = output[len(pairs)-1]
			for i, p := range pairs {
				output[i] = (p[0] & 0x0F) | (p[1]&0x0F)<<4 //gcassert:bce
			}
		}
		if len(input)%2 == 1 {
			output[len(input)/2] = input[len(input)-1] & 0x0F
		}
	case 8:
		copy(output[:len(input)], input)
	default:
		panic("bpk must be 4 or 8")
	}
}

// Encode16 packs uint16 values into bytes using the specified bits-per-key
// (bpk). When bpk < 16, the lower bpk bits of each value are used.
//
// For bpk=12, two values (a, b) are packed into 3 bytes:
//   - byte0 = low 8 bits of a
//   - byte1 = high 4 bits of a (low nibble) | low 4 bits of b (high nibble)
//   - byte2 = high 8 bits of b
//
// For bpk=16, values are stored as 2 bytes each in little-endian order.
//
// Panics if bpk is not 12 or 16.
func Encode16(input []uint16, bpk int, output []byte) {
	switch bpk {
	case 12:
		// Verify the buffer is large enough.
		_ = output[(len(input)+1)/2*3-1]
		outTrios := unsafe.Slice((*[3]byte)(unsafe.Pointer(unsafe.SliceData(output))), (len(input)+1)/2)
		if len(input) > 1 {
			// Cast to slices of arrays to elide bound checks.
			inPairs := unsafe.Slice((*[2]uint16)(unsafe.Pointer(unsafe.SliceData(input))), len(input)/2)
			_ = outTrios[len(inPairs)-1]
			for i, p := range inPairs {
				a, b := p[0], p[1]
				//gcassert:bce
				outTrios[i] = [3]byte{
					byte(a & 0xFF),
					byte((a>>8)&0x0F) | byte(b&0x0F)<<4,
					byte(b >> 4),
				}
			}
		}
		if len(input)%2 == 1 {
			a := input[len(input)-1]
			outTrios[len(outTrios)-1] = [3]byte{
				byte(a & 0xFF),
				byte((a >> 8) & 0x0F),
				0,
			}
		}
	case 16:
		// Verify the buffer is large enough.
		_ = output[len(input)*2-1]
		// Cast to slices of arrays to elide bound checks.
		outPairs := unsafe.Slice((*[2]byte)(unsafe.Pointer(unsafe.SliceData(output))), len(input))
		_ = outPairs[len(input)-1]
		for i, v := range input {
			//gcassert:bce
			outPairs[i] = [2]byte{
				byte(v & 0xFF),
				byte(v >> 8),
			}
		}
	default:
		panic("bpk must be 12 or 16")
	}
}

// Decode returns the i-th value from packed data, assuming it was encoded with the given bpk.
// Supports bpk = 4, 8, 12, 16.
// Panics if bpk is not one of these values.
func Decode(data []byte, i uint, bpk int) uint16 {
	switch bpk {
	case 8:
		return uint16(data[i])
	case 16:
		_ = data[i*2+1]
		return uint16(unsafeGet(data, i*2)) | uint16(unsafeGet(data, i*2+1))<<8
	case 4:
		shift := (i & 1) * 4
		return uint16((data[i>>1] >> shift) & 0x0F)
	case 12:
		base := (i >> 1) * 3
		_ = data[base+2]
		w := uint32(unsafeGet(data, base)) | uint32(unsafeGet(data, base+1))<<8 | uint32(unsafeGet(data, base+2))<<16
		shift := (i & 1) * 12
		return uint16((w >> shift) & 0xFFF)
	default:
		panic("bpk must be 4, 8, 12, or 16")
	}
}

// Decode3 returns the i1-th, i2-th, and i3-th value from packed data, assuming
// it was encoded with the given bpk.
//
// Supports bpk = 4, 8, 12, 16. Panics if bpk is not one of these values.
func Decode3(data []byte, i1, i2, i3 uint, bpk int) (uint16, uint16, uint16) {
	maxIdx := max(max(i1, i2), i3)
	switch bpk {
	case 8:
		_ = data[maxIdx]
		return uint16(unsafeGet(data, i1)), uint16(unsafeGet(data, i2)), uint16(unsafeGet(data, i3))
	case 16:
		_ = data[maxIdx*2+1]
		return uint16(unsafeGet(data, i1*2)) | uint16(unsafeGet(data, i1*2+1))<<8,
			uint16(unsafeGet(data, i2*2)) | uint16(unsafeGet(data, i2*2+1))<<8,
			uint16(unsafeGet(data, i3*2)) | uint16(unsafeGet(data, i3*2+1))<<8
	case 4:
		_ = data[maxIdx>>1]
		shift1 := (i1 & 1) * 4
		shift2 := (i2 & 1) * 4
		shift3 := (i3 & 1) * 4
		return uint16((unsafeGet(data, i1>>1) >> shift1) & 0x0F),
			uint16((unsafeGet(data, i2>>1) >> shift2) & 0x0F),
			uint16((unsafeGet(data, i3>>1) >> shift3) & 0x0F)
	case 12:
		_ = data[(maxIdx>>1)*3+2]
		base1 := (i1 >> 1) * 3
		base2 := (i2 >> 1) * 3
		base3 := (i3 >> 1) * 3
		w1 := uint32(unsafeGet(data, base1)) | uint32(unsafeGet(data, base1+1))<<8 | uint32(unsafeGet(data, base1+2))<<16
		w2 := uint32(unsafeGet(data, base2)) | uint32(unsafeGet(data, base2+1))<<8 | uint32(unsafeGet(data, base2+2))<<16
		w3 := uint32(unsafeGet(data, base3)) | uint32(unsafeGet(data, base3+1))<<8 | uint32(unsafeGet(data, base3+2))<<16
		shift1 := (i1 & 1) * 12
		shift2 := (i2 & 1) * 12
		shift3 := (i3 & 1) * 12
		return uint16((w1 >> shift1) & 0xFFF),
			uint16((w2 >> shift2) & 0xFFF),
			uint16((w3 >> shift3) & 0xFFF)
	default:
		panic("bpk must be 4, 8, 12, or 16")
	}
}

func unsafeGet(data []byte, index uint) byte {
	return *(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(data)), index))
}
