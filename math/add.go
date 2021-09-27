package math

// Add8 sums 2 8-bit values, returns the result and carry flag).
// Basically a shameless copy of bits.Add32().
func Add8(v1, v2 uint8) (uint8, bool) {
	sum16 := uint16(v1) + uint16(v2)
	sum8 := uint8(sum16)
	carryOut := uint8(sum16 >> 8)
	return sum8, carryOut > 0
}
