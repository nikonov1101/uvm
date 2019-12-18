package asm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTextOperandRegisters(t *testing.T) {
	validInput := []struct {
		name     string
		expected uint8
	}{
		{"r0", 0},
		{"r1", 1},
		{"r2", 2},
		{"r3", 3},
		{"r4", 4},
		{"r5", 5},
		{"r6", 6},
		{"r7", 7},
	}

	invaildInput := []string{
		"rr",
		"r",
		"r-1",
		"r9",
		"r8",
		"rA",
		"rF",
		"$",
		"$1",
		"#",
		"#1",
	}

	for _, tt := range validInput {
		v, _, err := parseTextOperand(tt.name, OperandReg)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, v, tt.name)
	}

	for _, tt := range invaildInput {
		_, _, err := parseTextOperand(tt, OperandReg)
		require.Error(t, err, tt)
	}
}

func TestParseTextOperandValues(t *testing.T) {
	validInput := []struct {
		name     string
		expected uint8
	}{
		{"#0", 0},
		{"#00", 0},
		{"#1", 1},
		{"#01", 1},
		{"#29", 0x29},
		{"#0f", 0xf},
		{"#f", 0xf},
		{"#f1", 0xf1},
		{"#fa", 0xfa},
		{"#0fa", 0xfa},
		{"#ff", 0xff},
	}

	invaildInput := []string{
		"00",
		"0",
		"1",
		"123",
		"#",
		"$",
		"r",
		"#zx",
		"#123",  // overflow
		"#fafa", // overflow
		"#1101",
	}

	for _, tt := range validInput {
		v, _, err := parseTextOperand(tt.name, OperandVal)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, v, tt.name)
	}

	for _, tt := range invaildInput {
		_, _, err := parseTextOperand(tt, OperandVal)
		require.Error(t, err, tt)
	}
}

func TestParseTextOperandAddresses(t *testing.T) {
	validInput := []struct {
		name   string
		hi, lo uint8
	}{
		{"$0", 0, 0},
		{"$0000", 0, 0},
		{"$1", 0, 1},
		{"$01", 0, 1},
		{"$001", 0, 1},
		{"$0001", 0, 1},
		{"$dead", 0xde, 0xad},
		{"$1101", 0x11, 0x01},
		{"$1fa", 0x1, 0xfa},
		{"$fa1", 0xf, 0xa1},
	}

	invaildInput := []string{
		"$",
		"#",
		"r",
		"$q",
		"$11011",
		"$11q1",
	}

	for _, tt := range validInput {
		lo, hi, err := parseTextOperand(tt.name, OperandAddr)
		require.NoError(t, err)
		assert.Equal(t, tt.lo, lo, "lo "+tt.name)
		assert.Equal(t, tt.hi, hi, "hi ", tt.name)
	}

	for _, tt := range invaildInput {
		_, _, err := parseTextOperand(tt, OperandAddr)
		require.Error(t, err, tt)
	}
}
