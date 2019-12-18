package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
)

const (
	// todo: move it to the `asm` package
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
	opType asm.OperandType
	value  uint8
}

// checkOperand checks that given value can be
// an operand of the expected type.
// returns human readable name for debug
func checkOperand(v uint8, typ asm.OperandType) string {
	switch typ {
	case asm.OperandVal:
		// value can be any value, nothing to do here
		return fmt.Sprintf("#%d", v)
	case asm.OperandReg:
		if v >= RegisterCount {
			panic("invalid register operand")
		}
		return fmt.Sprintf("r%d", v)
	case asm.OperandAddr:
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
	operandCount int
	// on which data we need to preform operation
	operands []operand
}

// calculateAddress calculates 16-bit address using
// LO and HI 8-bit address parts
func (in *instruction) calculateAddress(lo, hi uint8) uint16 {
	// shift HI by 8 bits, then OR with LO to fill the rest
	return uint16(hi)<<8 | uint16(lo)
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
		cpu.pc = in.calculateAddress(in.operands[0].value, in.operands[1].value)
		return

	case asmADDRegReg:
		// todo: handle carry, overflow and zero
		// store result in the first operand register
		r0 := in.operands[0].value
		r1 := in.operands[1].value
		cpu.registers[r0] += cpu.registers[r1]

	case asmMOVRegVal:
		cpu.registers[in.operands[0].value] = in.operands[1].value
	case asmMOVRegAddr:
		addr := in.calculateAddress(in.operands[1].value, in.operands[2].value)
		// load ROM value at given ADDR into the operand's REG
		cpu.registers[in.operands[0].value] = cpu.ROM[addr]

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
func (cpu *CPU) decodeInstruction(v uint8) instruction {
	operandsForOpCode, ok := (*cpu.opCodes)[v]
	if !ok {
		panic(fmt.Sprintf("invalid instruiction %2x", v))
	}

	var instructionOperands []operand
	for _, op := range operandsForOpCode {
		// note: just a dirty crutch to add two address bytes for instruction.
		// Need find a smarter way to handle this situation.
		if op == asm.OperandAddr {
			instructionOperands = append(instructionOperands, operand{opType: asm.OperandAddr}, operand{opType: asm.OperandAddr})
		} else {
			instructionOperands = append(instructionOperands, operand{opType: op})
		}
	}

	return instruction{
		// todo: WTF with name?
		name:         fmt.Sprintf("%02x", v),
		opCode:       v,
		operandCount: len(instructionOperands),
		operands:     instructionOperands,
	}
}
