package asm

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// mov
//  op: r, r
//  op: r, #
//  op: r, $

const (
	opTypeReg  = 1
	opTypeVal  = 2
	opTypeAddr = 3
)

// parseTextOperand parses and returns its byte-code representation.
// if here is one-byte operands - only `lo` will be filled.
// for two-bytes operands (such as mem address) both of `hi` and `lo`
// will hold values;
// note: return order is lo, hi
func parseTextOperand(op string, typ int) (uint8, uint8, error) {
	switch typ {
	case opTypeReg:
		if op[0] != 'r' {
			return 0, 0, fmt.Errorf("invalid register definition")
		}
		if op[1] < '0' && op[1] > '9' {
			return 0, 0, fmt.Errorf("invalid register name %s", op)
		}

		return op[1] - '0', 0, nil
	case opTypeVal:
		if op[0] != '#' {
			return 0, 0, fmt.Errorf("invalid value definition")
		}

		// remove the # char
		v := op[1:]
		// parse as uint8
		val, err := strconv.ParseUint(v, 16, 8)
		if err != nil {
			return 0, 0, err
		}

		return uint8(val), 0, err
	case opTypeAddr: // addr
		if op[0] != '$' {
			return 0, 0, fmt.Errorf("invalid addr definition")
		}

		// remove the $ char
		v := op[1:]
		// parse as uint16
		val, err := strconv.ParseUint(v, 16, 16)
		if err != nil {
			return 0, 0, err
		}

		// present as two 8bit values
		lo := uint8(val)
		hi := uint8(val >> 8)
		return lo, hi, nil
	default:
		return 0, 0, fmt.Errorf("unknown type %d", typ)
	}
}

// simple syntax tree, must be loaded from external source in the future
var syntax = map[string]map[uint8][]int{
	"NOP": {
		0x00: {},
	},
	"JUMP": {
		0x01: {opTypeAddr},
	},
	"PUSH": {
		0x02: {opTypeReg},
	},
	"POP": {
		0x03: {opTypeReg},
	},
	"HALT": {
		0x09: {},
	},
	"ADD": {
		0x10: {opTypeReg, opTypeReg},
		0x11: {opTypeReg, opTypeVal},
	},
	"MOV": {
		0x20: {opTypeReg, opTypeReg},
		0x21: {opTypeReg, opTypeVal},
		0x22: {opTypeReg, opTypeAddr},
	},
}

// assemble turns instruction and operands text into the machine codes
func assemble(ins string, ops []string) []uint8 {
	p, ok := syntax[ins]
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
				if expectedType == opTypeReg || expectedType == opTypeVal {
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

func Compile(prog string) [1 << 16]uint8 {
	rd := bufio.NewReader(strings.NewReader(prog))

	bin := [1 << 16]uint8{}
	offset := uint16(0x00)
	lineNum := 0

	for {
		lineNum++
		line, err := rd.ReadString('\n')
		if err != nil {
			fmt.Printf("err = %v\n", err)
			break
		}
		line = strings.TrimSpace(line)
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
		var insType string
		switch ins {
		case ".text":
			insType = "TEXT"
		case ".byte":
			insType = "BYTE"
		default:
			insType = "INSTR"
		}

		fmt.Printf("ASM: %s = `%s`, OPERANDS = `%s`\n", insType, ins, ops)

		if insType == "INSTR" {
			code := assemble(ins, ops)
			if code == nil {
				panic(fmt.Sprintf("failed to build instruction %s %v at %d: empty code returned", ins, ops, lineNum))
			}

			for i := range code {
				// fmt.Printf("ASM: %04x: %02x)\n", offset, code[i])
				bin[offset] = code[i]
				offset++
			}

		} else if insType == "TEXT" {
			if len(ops) != 1 {
				// we're expecting only memory address
				panic(fmt.Sprintf("invalid .text syntax at line %d", lineNum))
			}

			// parse given address
			addr := asAddress(ops[0], lineNum)

			// put the following program at text current .text addr
			offset = addr

		} else if insType == "BYTE" {
			if len(ops) != 2 {
				// we're expecting memory address and one byte value
				panic(fmt.Sprintf("invalid .byte syntax at line %d", lineNum))
			}

			// parse "operands"
			addr := asAddress(ops[0], lineNum)
			v := asValue(ops[1], lineNum)

			bin[addr] = v
		}

		fmt.Println()
	}

	return bin
}

func asAddress(s string, ln int) uint16 {
	if s[0] != '$' {
		panic(fmt.Sprintf("syntax error: invalid mem address (must starts with \"$\") at line %d", ln))
	}

	strVal := s[1:]
	addr, err := strconv.ParseUint(strVal, 16, 16)
	if err != nil {
		panic(fmt.Sprintf("failed to parse addr `%s` at line %d: %v", strVal, ln, err))
	}
	return uint16(addr)
}

func asValue(s string, ln int) uint8 {
	if s[0] != '#' {
		panic(fmt.Sprintf("syntax error: invalid value (must starts with \"#\") at line %d", ln))
	}

	strVal := s[1:]
	v, err := strconv.ParseUint(strVal, 16, 8)
	if err != nil {
		panic(fmt.Sprintf("failed to parse value `%s` at line %d: %v", strVal, ln, err))
	}
	return uint8(v)
}
