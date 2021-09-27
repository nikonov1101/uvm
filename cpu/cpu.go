package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/defines"
)

type flags struct {
	zero  bool
	carry bool
	halt  bool
}

func (f *flags) String() string {
	return fmt.Sprintf("Z: %v | C: %v | H: %v", f.zero, f.carry, f.halt)
}

type CPU struct {
	ROM [defines.ROMSize]uint8
	RAM [defines.RAMSize]uint8

	registers [defines.RegisterCount]uint8
	stack     *stack
	flags     *flags

	// program counter
	pc uint16

	// syntax and opcodes
	syn *asm.Syntax
}

func NewCPU(syn *asm.Syntax) *CPU {
	return &CPU{
		ROM:       [defines.ROMSize]uint8{},
		RAM:       [defines.RAMSize]uint8{},
		registers: [defines.RegisterCount]uint8{},
		stack:     newStack(defines.StackDepth),
		flags:     &flags{},
		pc:        0,
		syn:       syn,
	}
}

func (cpu *CPU) Run() {
	for {
		// load next value from mem,
		// must be an instruction
		v := cpu.ROM[cpu.pc]

		// decode instruction
		// note: panics on invalid input
		next := cpu.decodeInstruction(v)

		// load operands
		for i := 0; i < next.operandCount; i++ {
			// calculate next mem address
			cpu.pc++
			// fetch memory
			// todo: do it via something like MAR/MDR, as real hardware do
			//  or just emulate it, at least it will looks hacky.
			given := cpu.ROM[cpu.pc]
			expected := next.operands[i]

			// sanity check
			opName := checkOperand(given, expected.opType)
			// XXX debug
			fmt.Printf("  operand %s loaded\n", opName)

			// store within instruction
			next.operands[i].value = given
		}

		// XXX print instruction with operators loaded
		fmt.Printf("at PC = %d (0x%02x) -> RUN %s\n", cpu.pc, cpu.pc, next)

		next.execute(cpu)

		// dump CPU state after the each instruction
		fmt.Println("======== CPU state ========")
		fmt.Printf("PC = %d\n", cpu.pc)
		fmt.Printf("flags:\n  %v\n", cpu.flags)
		fmt.Printf("registers:\n  ")
		for i := 0; i < defines.RegisterCount; i++ {
			fmt.Printf("r%d = %02x", i, cpu.registers[i])
			if i+1 != defines.RegisterCount {
				fmt.Printf(" | ")
			}
		}
		fmt.Printf("\n===========================\n\n")

		if cpu.flags.halt {
			return
		}
	}
}

// decodeInstruction checks that given opcode exists,
// if so, annotates it with desired operand types
// and the instruction name (just for the debug purposes).
func (cpu *CPU) decodeInstruction(opcode uint8) instruction {
	var operandsForOpCode []asm.OperandType
	var ok bool
	var mnemonic string

	// we'd like to have a mnemonic for given opcode,
	// so walk through the whole syntax definition.
	for name, opcodes := range *cpu.syn {
		// does this mnemonic implements given opcode?
		operandsForOpCode, ok = opcodes[opcode]
		if !ok {
			continue
		}
		mnemonic = name
		break
	}

	if !ok {
		panic(fmt.Sprintf("invalid instruiction %2x", opcode))
	}

	var instructionOperands []operand
	for _, op := range operandsForOpCode {
		// note: just a dirty crutch to add two address bytes for instruction.
		// Need to find a smarter way to handle such situation.
		if op == asm.OperandAddr {
			instructionOperands = append(instructionOperands, operand{opType: asm.OperandAddr}, operand{opType: asm.OperandAddr})
		} else {
			instructionOperands = append(instructionOperands, operand{opType: op})
		}
	}

	return instruction{
		name:         mnemonic,
		opCode:       opcode,
		operandCount: len(instructionOperands),
		operands:     instructionOperands,
	}
}
