package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/defines"
	"github.com/sshaman1101/uvm/math"
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
	case asm.OperandValue:
		// value can be any value, nothing to do here
		return fmt.Sprintf("#%02X", v)
	case asm.OperandReg:
		if v >= defines.RegisterCount {
			panic("invalid register operand")
		}
		return fmt.Sprintf("r%02d", v)
	case asm.OperandAddr:
		return fmt.Sprintf("$%02X", v)
	default:
		panic("operand type must be defined")
	}
}

type instruction struct {
	// just a name, like MOV, XOR, JUMP
	name string
	// what to do
	opCode uint8
	// on which data we need to preform operation,
	// note: when the only instruction fetched, we know the
	// operand **types** from the instruction code itself.
	// the actual value of operands (register names, address values)
	// will be fetched from a memory during the instruction fetch cycle.
	operands []operand
}

func (in instruction) OperandSize() int {
	size := 0
	for _, op := range in.operands {
		size += op.opType.Size()
	}
	return size
}

func (in instruction) String() string {
	ops := ""
	for _, o := range in.operands {
		ops += fmt.Sprintf("%s %0X, ", o.opType, o.value)
	}
	return in.name + " " + ops
}

// asAddress returns 16 bit address from given operand indexes
func (in instruction) asAddress(seg uint8, loOp, hiOp int) uint32 {
	lo := in.operands[loOp].value
	hi := in.operands[hiOp].value
	return uint32(seg)<<16 | uint32(hi)<<8 | uint32(lo)
}

// execute the instruction
// can touch:
//   - registers
//   - flag register
//   - program counter (when jumps)
func (cpu *CPU) execute(in instruction) {
	switch in.opCode {
	case asm.OpNOP:
		// just do nothing

	case asm.OpJUMP:
		cpu.pc = in.asAddress(cpu.segmentSelectorReg, 0, 1)

	case asm.OpJUMPIF_EQ:
		if cpu.flags.zero {
			cpu.pc = in.asAddress(cpu.segmentSelectorReg, 0, 1)
		}

	case asm.OpJUMPIF_NE:
		if !cpu.flags.zero {
			cpu.pc = in.asAddress(cpu.segmentSelectorReg, 0, 1)
		}

	case asm.OpCP:
		r0 := cpu.generalPurposeReg[in.operands[0].value]
		r1 := cpu.generalPurposeReg[in.operands[1].value]
		cpu.flags.zero = r0 == r1
		cpu.flags.carry = false

	case asm.OpCPI:
		r0 := cpu.generalPurposeReg[in.operands[0].value]
		v1 := in.operands[1].value
		cpu.flags.zero = r0 == v1
		cpu.flags.carry = false

	case asm.OpADDRegReg:
		r0 := in.operands[0].value
		r1 := in.operands[1].value

		result, carry := math.Add8(cpu.generalPurposeReg[r0], cpu.generalPurposeReg[r1])
		cpu.generalPurposeReg[r0] = result

		cpu.flags.zero = result == 0
		cpu.flags.carry = carry

	case asm.OpADDRegVal:
		reg := in.operands[0].value
		value := in.operands[1].value

		result, carry := math.Add8(cpu.generalPurposeReg[reg], value)
		cpu.generalPurposeReg[reg] = result

		cpu.flags.zero = result == 0
		cpu.flags.carry = carry

	case asm.OpMOVRegVal:
		reg := in.operands[0].value
		val := in.operands[1].value
		cpu.generalPurposeReg[reg] = val

	case asm.OpMOVRegReg:
		dstReg := in.operands[0].value
		srcReg := in.operands[1].value
		cpu.generalPurposeReg[dstReg] = cpu.generalPurposeReg[srcReg]

	case asm.OpLOAD: // LOAD reg <- $mem
		addr := in.asAddress(cpu.segmentSelectorReg, 1, 2)
		val := cpu.mem[addr]

		reg := in.operands[0].value
		cpu.generalPurposeReg[reg] = val

	case asm.OpSTORE: // STORE $mem <- reg
		reg := in.operands[2].value
		val := cpu.generalPurposeReg[reg]

		addr := in.asAddress(cpu.segmentSelectorReg, 0, 1)
		cpu.mem[addr] = val

	case asm.OpHALT:
		cpu.flags.halt = true

	case asm.OpPUSH:
		reg := in.operands[0].value
		val := cpu.generalPurposeReg[reg]
		sp := 0x00FFFFFF & cpu.sp
		cpu.mem[sp] = val
		// TODO(nikonov): any kind of stack guards, maybe?
		cpu.sp--

	case asm.OpPOP:
		sp := 0x00FFFFFF & cpu.sp
		val := cpu.mem[sp]
		// TODO(nikonov): any kind of stack guards, maybe?
		cpu.sp++

		reg := in.operands[0].value
		cpu.generalPurposeReg[reg] = val

	case asm.OpCLEAR:
		reg := in.operands[0].value
		cpu.generalPurposeReg[reg] = 0
		cpu.flags.zero = true
		cpu.flags.carry = false

	case asm.OpINC:
		reg := in.operands[0].value
		val := cpu.generalPurposeReg[reg]

		result, carry := math.Add8(val, 1)
		cpu.generalPurposeReg[reg] = result
		cpu.flags.zero = result == 0
		cpu.flags.carry = carry

	default:
		panic(fmt.Sprintf("dunno how to execute instruction %2x (%s)", in.opCode, in.name))
	}
}
