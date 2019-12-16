package cpu

import (
	"fmt"
)

const (
	_ = iota
	operandValue
	operandRegister
	operandAddress

	asmNOP  = 0x00
	asmJUMP = 0x01
	asmPUSH = 0x02
	asmPOP  = 0x03
	asmHALT = 0x09

	asmADDRegReg = 0x10
	asmADDRegVal = 0x11

	asmMOVRegReg  = 0x20
	asmMOVRegVal  = 0x21
	asmMOVRegAddr = 0x22
)

type operand struct {
	opType int
	value  uint8
}

// checkOperand checks that given value can be
// an operand of the expected type.
// returns human readable name for debug
func checkOperand(v uint8, typ int) string {
	switch typ {
	case operandValue:
		// value can be any value, nothing to do here
		return fmt.Sprintf("#%d", v)
	case operandRegister:
		if v >= RegisterCount {
			panic("invalid register operand")
		}
		return fmt.Sprintf("r%d", v)
	case operandAddress:
		if int(v) >= ROMSize {
			panic("mem address operand is out of memory")
		}
		return fmt.Sprintf("$%d", v)
	default:
		panic("operand type must be defined")
	}
}

type instruction struct {
	// just a name, like MOV, XOR, JUMP
	name string
	// what to do
	opCode uint8
	// how many operands we need to fetch from memory
	operandCount uint16
	// on which data we need to preform operation
	operands []operand
}

// execute the instruction
// can touch:
//   * registers
//   * flag register
//   * program counter
// note that in must increase PC by one
//   if it's regular instruction (not JUMP)
func (in *instruction) execute(cpu *CPU) {
	switch in.opCode {
	case asmNOP:
		// just do nothing

	case asmJUMP:
		// go to address, DO NOT increment PC by one
		cpu.pc = uint16(in.operands[0].value)
		return

	case asmADDRegReg:
		// todo: handle carry, overflow and zero
		// store result in the first operand register
		r0 := in.operands[0].value
		r1 := in.operands[1].value
		cpu.registers[r0] += cpu.registers[r1]

	case asmMOVRegVal:
		cpu.registers[in.operands[0].value] = in.operands[1].value

	case asmHALT:
		cpu.flags.halt = true
		return

	case asmPUSH:
		regVal := cpu.registers[in.operands[0].value]
		cpu.stack.push(regVal)
	case asmPOP:
		cpu.registers[in.operands[0].value] = cpu.stack.pop()

	default:
		panic(fmt.Sprintf("dunno how to execute instruction %2x", in.opCode))
	}

	// go to next instruction
	cpu.pc++
}

// decodeInstruction checks that given opcode exists,
// if so, annotates it with desired operand types
// and the instruction name (just for the debug purposes).
func decodeInstruction(v uint8) instruction {
	var m = map[uint8]instruction{
		// do nothing
		asmNOP: {name: "NOP", operandCount: 0},
		// load new address to PC
		asmJUMP: {name: "JUMP", operandCount: 1, operands: []operand{{opType: operandAddress}}},
		// PUSH r1
		asmPUSH: {name: "PUSH", operandCount: 1, operands: []operand{{opType: operandRegister}}},
		// POP r1
		asmPOP: {name: "POP", operandCount: 1, operands: []operand{{opType: operandRegister}}},
		// stop all the things
		asmHALT: {name: "HALT", operandCount: 0},

		// ADD r1 + r2
		asmADDRegReg: {name: "ADD", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandRegister}}},
		// ADD reg + #val
		asmADDRegVal: {name: "ADD", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandValue}}},
		// MOV r1 <- r2
		asmMOVRegReg: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandRegister}}},
		// MOV r1 <- #val
		asmMOVRegVal: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandValue}}},
		// MOV r1 <- $addr
		asmMOVRegAddr: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandAddress}}},
	}

	ins, ok := m[v]
	if !ok {
		panic(fmt.Sprintf("invalid instruiction %2x", v))
	}

	ins.opCode = v
	return ins
}
