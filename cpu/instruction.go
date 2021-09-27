package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/defines"
	"github.com/sshaman1101/uvm/math"
)

const (
	// todo: move it to the `asm` package
	asmNOP   = 0x00
	asmJUMP  = 0x01
	asmPUSH  = 0x02
	asmPOP   = 0x03
	asmCLEAR = 0x04
	asmINC   = 0x05
	asmHALT  = 0x09

	asmADDRegReg = 0x10
	asmADDRegVal = 0x11

	asmMOVRegReg = 0x20
	asmMOVRegVal = 0x21
	asmLPM       = 0x22

	asmLOAD = 0x30

	asmSTORE = 0x40
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
		return fmt.Sprintf("#%02X", v)
	case asm.OperandReg:
		if v >= defines.RegisterCount {
			panic("invalid register operand")
		}
		return fmt.Sprintf("r%02d", v)
	case asm.OperandAddr:
		if int(v) >= defines.RAMSize {
			panic("mem address operand is out of memory")
		}
		return fmt.Sprintf("$%04X", v)
	default:
		panic("operand type must be defined")
	}
}

// calculateAddress calculates 16-bit address using
// LO and HI 8-bit address parts
func calculateAddress(lo, hi uint8) uint16 {
	// shift HI by 8 bits, then OR with LO to fill the rest
	return uint16(hi)<<8 | uint16(lo)
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

func (in instruction) String() string {
	ops := ""
	for _, o := range in.operands {
		ops += fmt.Sprintf("%s %0X, ", o.opType, o.value)
	}
	return in.name + " " + ops
}

// asAddress returns 16 bit address from given operand indexes
func (in *instruction) asAddress(loOp, hiOp int) uint16 {
	return calculateAddress(in.operands[loOp].value, in.operands[hiOp].value)
}

// execute the instruction
// can touch:
//   * registers
//   * flag register
//   * program counter
// note that in must increase PC by one
//   if it's regular instruction (not JUMP)
// TODO: seems like the dependency must be inverted:
//  the CPU executes the instruction, not vise versa.
func (in *instruction) execute(cpu *CPU) {
	switch in.opCode {
	case asmNOP:
		// just do nothing

	case asmJUMP:
		// go to address, DO NOT increment PC by one
		cpu.pc = in.asAddress(0, 1)
		return

	case asmADDRegReg:
		r0 := in.operands[0].value
		r1 := in.operands[1].value

		result, carry := math.Add8(cpu.registers[r0], cpu.registers[r1])
		cpu.registers[r0] = result

		cpu.flags.zero = result == 0
		cpu.flags.carry = carry
	case asmADDRegVal:
		reg := in.operands[0].value
		value := in.operands[1].value

		result, carry := math.Add8(cpu.registers[reg], value)
		cpu.registers[reg] = result

		cpu.flags.zero = result == 0
		cpu.flags.carry = carry

	case asmMOVRegVal:
		cpu.registers[in.operands[0].value] = in.operands[1].value
	case asmMOVRegReg:
		dstReg := in.operands[0].value
		srcReg := in.operands[1].value
		cpu.registers[dstReg] = cpu.registers[srcReg]

	case asmLPM: // load from program memory
		// calculate 16 bit address in ROM
		addr := in.asAddress(1, 2)
		// load value in the given register
		cpu.registers[in.operands[0].value] = cpu.ROM[addr]
		// todo: flags?

	case asmLOAD:
		addr := in.asAddress(1, 2)
		reg := in.operands[0].value
		val := cpu.RAM[addr]
		cpu.registers[reg] = val
		cpu.flags.zero = val == 0
	case asmSTORE: // addr, reg
		addr := in.asAddress(0, 1)
		reg := in.operands[2].value
		val := cpu.registers[reg]
		cpu.RAM[addr] = val

	case asmHALT:
		cpu.flags.halt = true
		return

	case asmPUSH:
		regVal := cpu.registers[in.operands[0].value]
		cpu.stack.push(regVal)
	case asmPOP:
		cpu.registers[in.operands[0].value] = cpu.stack.pop()
	case asmCLEAR:
		reg := in.operands[0].value
		cpu.registers[reg] = 0
		cpu.flags.zero = true
	case asmINC:
		reg := in.operands[0].value
		val := cpu.registers[reg]

		result, carry := math.Add8(val, 1)
		cpu.registers[reg] = result
		cpu.flags.zero = reg == 0
		cpu.flags.carry = carry

	default:
		panic(fmt.Sprintf("dunno how to execute instruction %2x (%s)", in.opCode, in.name))
	}

	// todo: math
	//  maybe it should be like:
	//  res = cpu.add(v1, v2)
	//  where res is a 8bit value, and flags are set by the method accordingly?

	// go to next instruction.
	// todo: something is wrong with this design.
	cpu.pc++
}
