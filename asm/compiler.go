package asm

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sshaman1101/uvm/defines"
)

type (
	OperandType string
	nodeType    uint8
)

const (
	OperandReg   OperandType = "reg"
	OperandValue OperandType = "val"
	OperandAddr  OperandType = "addr"

	_ nodeType = iota
	nodeInstruction
	nodeText
	nodeByte
)

func newNodeType(s string) nodeType {
	switch s {
	case ".text":
		return nodeText
	case ".byte":
		return nodeByte
	default:
		return nodeInstruction
	}
}

func (n nodeType) String() string {
	switch n {
	case nodeInstruction:
		return "instruction"
	case nodeText:
		return ".text"
	case nodeByte:
		return ".byte"
	default:
		return "unknown"
	}
}

// parseTextOperand parses and returns its byte-code representation.
// if here is one-byte operands - only `lo` will be filled.
// for two-bytes operands (such as mem address) both of `hi` and `lo`
// will hold values;
// note: return order is lo, hi
func parseTextOperand(op string, typ OperandType) (uint8, uint8, error) {
	switch typ {
	case OperandReg:
		v, err := asRegister(op)
		return v, 0, err

	case OperandValue:
		v, err := asValue(op)
		return v, 0, err

	case OperandAddr:
		v, err := asAddress(op)
		if err != nil {
			return 0, 0, err
		}

		// present as two 8bit values
		lo := uint8(v)
		hi := uint8(v >> 8)
		return lo, hi, nil

	default:
		return 0, 0, fmt.Errorf("unknown type %s", typ)
	}
}

// assemble turns instruction and operands text into the machine codes
func assemble(ins string, ops []string) []uint8 {
	p, ok := Syntax[ins]
	if !ok {
		panic(fmt.Sprintf("unknown instruction %s", ins))
	}

	// go thorough all available combinations of operands for given instruction,
	// try to find matching one
	for opCode, expected := range p {
		if len(expected) != len(ops) {
			continue
		}

		// calculate matches, all operands must be matched with
		// the expected schema,
		// clear and repeat otherwise
		matched := 0
		// store matched operands bytecode representation
		var operandStack []uint8

		hi, lo := uint8(0), uint8(0)
		var err error

		for i := 0; i < len(expected); i++ {
			// which type of operand we expect for such command?
			expectedType := expected[i]
			lo, hi, err = parseTextOperand(ops[i], expectedType)
			// operand at this particular position matches with required,
			// 1. increment matches count,
			// 2. according to expected operator type determine
			//    how long operator is, in case of address store both values in LittleEndian
			//    otherwise store only `lo` one.
			if err == nil {
				matched++
				if expectedType == OperandReg || expectedType == OperandValue {
					operandStack = append(operandStack, lo)
				} else {
					operandStack = append(operandStack, lo, hi)
				}
			}
		}

		// all operators matched, now we have enough information
		// to generate machine code. Return instruction's opcode
		// followed by all of the arguments
		if matched == len(ops) {
			return append([]uint8{opCode}, operandStack...)
		}
	}

	return nil
}

// Compile compiles program loaded from reader (usually strings.Reader or os.File)
func Compile(textReader io.Reader) [1 << 16]uint8 {

	bin := [defines.ROMSize]uint8{}
	offset := uint16(0x00)
	lineNum := 0

	sk := bufio.NewScanner(textReader)
	for sk.Scan() {
		lineNum++

		line := strings.TrimSpace(sk.Text())
		if len(line) == 0 {
			continue
		}

		// remove comments
		parts := strings.Split(line, ";")
		line = strings.TrimSpace(parts[0])
		if len(line) == 0 {
			continue
		}

		// split into instruction + operands
		parts = strings.Split(line, " ")
		ins := parts[0]
		ops := parts[1:]

		// clean-up operands
		for i := 0; i < len(ops); i++ {
			ops[i] = strings.Replace(strings.TrimSpace(ops[i]), ",", "", -1)
		}

		// decide on what we're looking right now - operator or macro?
		// todo: labels
		insType := newNodeType(ins)

		fmt.Printf("ASM:%02d:\t`%s`\toperands: \t%s\n", lineNum, ins, ops)

		switch insType {
		case nodeInstruction:
			code := assemble(ins, ops)
			if code == nil {
				panic(fmt.Sprintf("failed to build instruction %s %v at %d: empty code returned", ins, ops, lineNum))
			}

			for i := range code {
				bin[offset] = code[i]
				offset++
			}

		case nodeText: // parse address behind .text macro, move mem pointer to that location
			if len(ops) != 1 {
				// we're expecting only memory address
				panic(fmt.Sprintf("invalid .text syntax at line %d", lineNum))
			}

			// parse given address
			addr, err := asAddress(ops[0])
			if err != nil {
				panic(fmt.Sprintf("failed to parse addr `%s` at line %d: %v", ops[0], lineNum, err))
			}

			// put the following program right after the given address
			offset = addr
		case nodeByte: // parse addr and value at .byte macro, but the byte at the given address
			if len(ops) != 2 {
				// we're expecting memory address and one byte value
				panic(fmt.Sprintf("invalid .byte syntax at line %d", lineNum))
			}

			// parse "operands"
			addr, err := asAddress(ops[0])
			if err != nil {
				panic(fmt.Sprintf("failed to parse addr `%s` at line %d: %v", ops[0], lineNum, err))
			}

			v, err := asValue(ops[1])
			if err != nil {
				panic(fmt.Sprintf("failed to parse value `%s` at line %d: %v", ops[1], lineNum, err))
			}

			bin[addr] = v
		}
	}

	return bin
}

func asAddress(s string) (uint16, error) {
	if s[0] != '$' {
		return 0, fmt.Errorf("syntax error: mem address (must starts with \"$\")")
	}

	strVal := s[1:]
	addr, err := strconv.ParseUint(strVal, 16, 16)
	if err != nil {
		return 0, err
	}

	return uint16(addr), nil
}

func asValue(s string) (uint8, error) {
	if s[0] != '#' {
		return 0, fmt.Errorf("syntax error: invalid value (must starts with \"#\")")
	}

	strVal := s[1:]
	v, err := strconv.ParseUint(strVal, 16, 8)
	if err != nil {
		return 0, err
	}

	return uint8(v), nil
}

func asRegister(s string) (uint8, error) {
	if s[0] != 'r' {
		return 0, fmt.Errorf("syntax error: invalid register name (must start with \"r\")")
	}

	if len(s) != 2 || s[1] < '0' || s[1] >= strconv.Itoa(defines.RegisterCount)[0] {
		return 0, fmt.Errorf("register name must be r0..r7")
	}

	return s[1] - '0', nil
}
